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

// json summary
export function handleSummary(data) {
  const out = JSON.stringify({
    state: {
      testRunDurationMs: data.state?.testRunDurationMs,
      testRunDuration: data.state?.testRunDuration, // seconds
    },
    metrics: data.metrics
  }, null, 2);

  const summaryPath = __ENV.K6_SUMMARY_OUT || "tests/load/perf_cached.json";
  return {
    [summaryPath]: json,
    stdout: "\nSaved summary to " + summaryPath + "\n"
  };
}

