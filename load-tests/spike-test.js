import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('errors');
const responseTrend = new Trend('response_time_trend');

const BASE_URL = 'https://api-gateway-406932815122.europe-west1.run.app';

const USERS = [
  { username: 'admin', password: 'admin123' },
  { username: 'guide', password: 'guide123' },
];

export const options = {
  scenarios: {
    spike: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '10s', target: 10 },
        { duration: '5s', target: 1000 },
        { duration: '1m', target: 1000 },
        { duration: '10s', target: 10 },
        { duration: '30s', target: 10 },
      ],
    },
  },
};

function login(user) {
  const res = http.post(`${BASE_URL}/api/auth/login`, JSON.stringify(user), {
    headers: { 'Content-Type': 'application/json' },
  });
  if (res.status === 200) {
    return JSON.parse(res.body).token;
  }
  return null;
}

export default function () {
  const user = USERS[Math.floor(Math.random() * USERS.length)];
  const token = login(user);

  if (!token) {
    errorRate.add(1);
    sleep(0.5);
    return;
  }

  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };

  const res = http.get(`${BASE_URL}/api/tours/published`, { headers });
  responseTrend.add(res.timings.duration);

  check(res, {
    'status 200': (r) => r.status === 200,
    'response time < 3s': (r) => r.timings.duration < 3000,
  });

  errorRate.add(res.status >= 500 ? 1 : 0);
  sleep(0.5);
}

export function handleSummary(data) {
  const summary = {
    'Total Requests': data.metrics.http_reqs.values.count,
    'Avg Response Time (ms)': Math.round(data.metrics.http_req_duration.values.avg),
    'P95 Response Time (ms)': Math.round(data.metrics.http_req_duration.values['p(95)']),
    'Max Response Time (ms)': Math.round(data.metrics.http_req_duration.values.max),
    'Requests/sec': Math.round(data.metrics.http_reqs.values.rate * 100) / 100,
  };

  console.log('\n========== SPIKE TEST SUMMARY ==========');
  for (const [key, value] of Object.entries(summary)) {
    console.log(`${key}: ${value}`);
  }
  console.log('=========================================\n');

  return {
    'spike-test-results.json': JSON.stringify(data, null, 2),
  };
}