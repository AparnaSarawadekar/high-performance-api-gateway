import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  scenarios: {
    quick: {
      executor: 'constant-arrival-rate',
      rate: 20,          // 20 RPS
      timeUnit: '1s',
      duration: '15s',   // ~300 requests
      preAllocatedVUs: 20,
      maxVUs: 50,
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<500'],
  },
};

export default function () {
  const res = http.get('http://localhost:8080/healthz');
  check(res, { 'status 200': (r) => r.status === 200 });
  sleep(0.01);
}
