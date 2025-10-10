import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  vus: 80,
  duration: "15s",
  thresholds: {
    'checks{type:got429}': ['rate>0.05'], // expect some 429s
    http_req_failed: ['rate<0.99'], // allow up to 99% fail (429s)
  },
};

export default function () {
  const r = http.get("http://localhost:8080/limited");
  check(r, {
    "200 or 429": (res) => res.status === 200 || res.status === 429,
    "got429": (res) => res.status === 429,
  }, { type: "got429" });
  // verify limiter headers when 429
  if (r.status === 429) {
    check(r, {
      "has Retry-After": (res) => !!res.headers["Retry-After"],
      "has RateLimit-Remaining": (res) => !!res.headers["RateLimit-Remaining"],
      "scope is client/global": (res) => {
        const s = res.headers["RateLimit-Scope"];
        return s === "client" || s === "global";
      },
    }, { type: "headers" });
  }
  sleep(0.05);
}
