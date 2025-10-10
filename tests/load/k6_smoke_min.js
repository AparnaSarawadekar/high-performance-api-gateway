import http from 'k6/http';
import { check } from 'k6';

// Single simple GET; k6 built-in VU/duration flags will drive the load.
export default function () {
  const res = http.get('http://localhost:8080/healthz');
  check(res, { 'status 200': (r) => r.status === 200 });
}
