// tests/load/cache_test.js
import http from "k6/http";
import { check, sleep } from "k6";
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.1/index.js';
import { json } from 'k6/encoding';

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
  const out = JSON.stringify(
    {
      state: {
        testRunDurationMs: data.state?.testRunDurationMs,
        testRunDuration: data.state?.testRunDuration,
      },
      metrics: data.metrics,
    },
    null,
    2
  );

  const summaryPath = __ENV.K6_SUMMARY_OUT || "tests/load/perf_cache_v2_redis_summary.json";

  return {
    [summaryPath]: out, //write plain JSON string (no json() wrapper)
    stdout: `\nSaved summary to ${summaryPath}\n`,
  };
}