import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('errors');
const loginDuration = new Trend('login_duration');
const toursDuration = new Trend('tours_duration');
const profileDuration = new Trend('profile_duration');
const cartDuration = new Trend('cart_duration');
const createTourDuration = new Trend('create_tour_duration');

const BASE_URL = 'https://api-gateway-406932815122.europe-west1.run.app';

const USERS = [
  { username: 'admin', password: 'admin123' },
  { username: 'guide', password: 'guide123' },
];

export const options = {
  scenarios: {
    load_test: {
      executor: 'ramping-vus',
      startVUs: 1,
      stages: [
        { duration: '30s', target: 50 },
        { duration: '1m', target: 200 },
        { duration: '1m', target: 500 },
        { duration: '2m', target: 1000 },
        { duration: '1m', target: 1500 },
        { duration: '30s', target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<5000'],
    errors: ['rate<0.3'],
  },
};

function login(user) {
  const res = http.post(`${BASE_URL}/api/auth/login`, JSON.stringify(user), {
    headers: { 'Content-Type': 'application/json' },
  });
  loginDuration.add(res.timings.duration);

  if (res.status === 200) {
    const body = JSON.parse(res.body);
    return body.token;
  }
  return null;
}

export default function () {
  // Pick random user
  const user = USERS[Math.floor(Math.random() * USERS.length)];
  const token = login(user);

  if (!token) {
    errorRate.add(1);
    sleep(1);
    return;
  }
  errorRate.add(0);

  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };

  // GET published tours
  const toursRes = http.get(`${BASE_URL}/api/tours/published`, { headers });
  toursDuration.add(toursRes.timings.duration);
  check(toursRes, { 'tours 200': (r) => r.status === 200 });
  errorRate.add(toursRes.status !== 200 ? 1 : 0);
  sleep(0.5);

  // GET profile
  const profileRes = http.get(`${BASE_URL}/api/profile`, { headers });
  profileDuration.add(profileRes.timings.duration);
  check(profileRes, { 'profile 200': (r) => r.status === 200 });
  errorRate.add(profileRes.status !== 200 ? 1 : 0);
  sleep(0.5);

  // GET cart (admin/tourist only)
  if (user.username === 'admin') {
    const cartRes = http.get(`${BASE_URL}/api/cart`, { headers });
    cartDuration.add(cartRes.timings.duration);
    check(cartRes, { 'cart 200': (r) => r.status === 200 });
    errorRate.add(cartRes.status !== 200 ? 1 : 0);
    sleep(0.5);
  }

  // POST create tour (guide only)
  if (user.username === 'guide') {
    const tourPayload = JSON.stringify({
      name: `Load Test Tour ${Date.now()}`,
      description: 'Tour created during load testing',
      price: Math.floor(Math.random() * 100) + 10,
      difficulty: 'easy',
      tags: ['test'],
    });
    const createRes = http.post(`${BASE_URL}/api/tours`, tourPayload, { headers });
    createTourDuration.add(createRes.timings.duration);
    check(createRes, { 'create tour 200/201': (r) => r.status === 200 || r.status === 201 });
    errorRate.add(createRes.status !== 200 && createRes.status !== 201 ? 1 : 0);
    sleep(0.5);
  }

  // GET my tours (guide only)
  if (user.username === 'guide') {
    const myToursRes = http.get(`${BASE_URL}/api/tours/my`, { headers });
    check(myToursRes, { 'my tours 200': (r) => r.status === 200 });
    errorRate.add(myToursRes.status !== 200 ? 1 : 0);
    sleep(0.5);
  }

  sleep(1);
}

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