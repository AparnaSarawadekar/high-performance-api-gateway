// tests/load/cache_test.js
import http from "k6/http";
import { check, sleep } from "k6";

export const options = { vus: 40, duration: "20s" };

export default function () {
  const ip = `10.0.0.${__VU}`;
  const r = http.get("http://localhost:8080/slow", { headers: { "X-Forwarded-For": ip }});
  check(r, {
    "status 200": (res) => res.status === 200,
    "has X-Cache": (res) => !!res.headers["X-Cache"],
  });
  sleep(0.05);
}
