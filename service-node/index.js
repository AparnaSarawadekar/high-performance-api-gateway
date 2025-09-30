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

