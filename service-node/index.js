import express from "express";
import morgan from "morgan";

const app = express();
const PORT = process.env.PORT || 8002;

app.use(morgan("tiny"));
app.use(express.json());

app.get("/healthz", (_req, res) => {
  res.json({ ok: true, service: "service-node" });
});

app.get("/metrics", (_req, res) => {
  res.json({
    ok: true,
    uptime_s: process.uptime(),
    service: "service-node"
  });
});

app.listen(PORT, () => {
  console.log(`service-node listening on :${PORT}`);
});

app.post("/infer", (req, res) => {
  const { prompt } = req.body || {};
  if (!prompt) return res.status(400).json({ error: "prompt required" });

  const delayMs = 30 + Math.floor(Math.random() * 21); // 30–50ms latency
  setTimeout(() => {
    res.json({
      result: `echo: ${prompt}`,
      latency_ms: delayMs,
      model: "node-dummy-v0",
    });
  }, delayMs);
});

