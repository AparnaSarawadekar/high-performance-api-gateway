import http from 'k6/http';
import { check, sleep } from 'k6';

// Allow port/endpoint overrides without editing the file
const BASE = __ENV.BASE_URL || 'http://localhost:8080';
const PATH = __ENV.PATH || '/healthz';

export const options = {
  scenarios: {
    warmup: {
      executor: 'constant-arrival-rate',
      rate: 50,               // 50 RPS for 30s
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 50,
      maxVUs: 100,
    },
    baseline: {
      executor: 'constant-arrival-rate',
      rate: 150,              // baseline throughput target
      timeUnit: '1s',
      duration: '2m',
      preAllocatedVUs: 200,
      maxVUs: 300,
      startTime: '30s',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],     // <1% errors
    http_req_duration: ['p(95)<250', 'p(99)<400'],
    checks: ['rate>0.99'],
  },
};

export default function () {
  const res = http.get(`${BASE}${PATH}`);
  check(res, { 'status 200': (r) => r.status === 200 });
  sleep(0.01); // small jitter
}
