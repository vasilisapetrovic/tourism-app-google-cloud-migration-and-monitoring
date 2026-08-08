import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const loginDuration = new Trend('login_duration');
const getPublishedDuration = new Trend('get_published_duration');
const createTourDuration = new Trend('create_tour_duration');
const getProfileDuration = new Trend('get_profile_duration');
const getCartDuration = new Trend('get_cart_duration');

const BASE_URL = 'https://api-gateway-406932815122.europe-west1.run.app';

// Test scenarios - gradually increase load
export const options = {
  scenarios: {
    // Scenario 1: Ramp up gradually
    load_test: {
      executor: 'ramping-vus',
      startVUs: 1,
     stages: [
    { duration: '30s', target: 50 },     // Warmup
    { duration: '1m', target: 200 },     // Ramp up
    { duration: '1m', target: 500 },     // Heavy load
    { duration: '2m', target: 1000 },    // RED ZONE
    { duration: '1m', target: 1500 },    // EXTREME
    { duration: '30s', target: 0 },      // Ramp down
],
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<5000'],    // 95% requests under 5s
    errors: ['rate<0.3'],                  // Error rate under 30%
  },
};

// Login and get token
function login() {
  const payload = JSON.stringify({
    username: 'admin',
    password: 'admin123',
  });

  const params = {
    headers: { 'Content-Type': 'application/json' },
  };

  const res = http.post(`${BASE_URL}/api/auth/login`, payload, params);
  loginDuration.add(res.timings.duration);

  if (res.status === 200) {
    const body = JSON.parse(res.body);
    return body.token;
  }
  return null;
}

// Main test function - each virtual user runs this
export default function () {
  // 1. Login
  const token = login();
  if (!token) {
    errorRate.add(1);
    sleep(1);
    return;
  }
  errorRate.add(0);

  const authHeaders = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
  };

  // 2. Get published tours
  const publishedRes = http.get(`${BASE_URL}/api/tours/published`, authHeaders);
  getPublishedDuration.add(publishedRes.timings.duration);
  check(publishedRes, {
    'get published tours status 200': (r) => r.status === 200,
  });
  errorRate.add(publishedRes.status !== 200 ? 1 : 0);

  sleep(1);

  // 3. Get profile
  const profileRes = http.get(`${BASE_URL}/api/profile`, authHeaders);
  getProfileDuration.add(profileRes.timings.duration);
  check(profileRes, {
    'get profile status 200': (r) => r.status === 200,
  });
  errorRate.add(profileRes.status !== 200 ? 1 : 0);

  sleep(1);

  // 4. Get cart
  const cartRes = http.get(`${BASE_URL}/api/cart`, authHeaders);
  getCartDuration.add(cartRes.timings.duration);
  check(cartRes, {
    'get cart status 200': (r) => r.status === 200,
  });
  errorRate.add(cartRes.status !== 200 ? 1 : 0);

  sleep(1);

  // 5. Create tour (write operation - heavier load)
  const tourPayload = JSON.stringify({
    name: `Load Test Tour ${Date.now()}`,
    description: 'Tour created during load testing',
    price: Math.floor(Math.random() * 100) + 10,
    difficulty: 'easy',
    tags: ['test', 'loadtest'],
  });

  const createRes = http.post(`${BASE_URL}/api/tours`, tourPayload, authHeaders);
  createTourDuration.add(createRes.timings.duration);
  check(createRes, {
    'create tour status 200 or 201': (r) => r.status === 200 || r.status === 201,
  });
  errorRate.add(createRes.status !== 200 && createRes.status !== 201 ? 1 : 0);

  sleep(2);
}

// Summary after test
export function handleSummary(data) {
  const summary = {
    'Total Requests': data.metrics.http_reqs.values.count,
    'Failed Requests': data.metrics.http_req_failed ? data.metrics.http_req_failed.values.passes : 0,
    'Avg Response Time (ms)': Math.round(data.metrics.http_req_duration.values.avg),
    'P95 Response Time (ms)': Math.round(data.metrics.http_req_duration.values['p(95)']),
    'P99 Response Time (ms)': Math.round(data.metrics.http_req_duration.values['p(99)']),
    'Max Response Time (ms)': Math.round(data.metrics.http_req_duration.values.max),
    'Requests/sec': Math.round(data.metrics.http_reqs.values.rate * 100) / 100,
  };

  console.log('\n========== LOAD TEST SUMMARY ==========');
  for (const [key, value] of Object.entries(summary)) {
    console.log(`${key}: ${value}`);
  }
  console.log('========================================\n');

  return {
    'load-test-results.json': JSON.stringify(data, null, 2),
  };
}