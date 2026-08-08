import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('errors');
const responseTrend = new Trend('response_time_trend');

const BASE_URL = 'https://api-gateway-406932815122.europe-west1.run.app';

// Spike test - sudden burst of traffic
export const options = {
  scenarios: {
    spike: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
    { duration: '10s', target: 10 },      // Normal
    { duration: '5s', target: 1000 },     // SPIKE!
    { duration: '1m', target: 1000 },     // Stay at spike
    { duration: '10s', target: 10 },      // Recovery
    { duration: '30s', target: 10 },      // Normal again
],
    },
  },
};

export default function () {
  // Simple GET request to tour service through API Gateway
  const res = http.get(`${BASE_URL}/api/tours/published`);
  
  responseTrend.add(res.timings.duration);
  
  check(res, {
    'status is 200 or 401': (r) => r.status === 200 || r.status === 401,
    'response time < 3s': (r) => r.timings.duration < 3000,
  });
  
  errorRate.add(res.status >= 500 ? 1 : 0);
  
  sleep(0.5);
}