from fastapi import FastAPI, Body
from pydantic import BaseModel
import random, asyncio, time

app = FastAPI(title="service-python")

class InferRequest(BaseModel):
    prompt: str

@app.get("/healthz")
async def healthz():
    return {"ok": True, "service": "service-python"}

@app.post("/infer")
async def infer(req: InferRequest = Body(...)):
    # simulating 30–50ms "AI" latency
    delay_ms = random.randint(30, 50)
    await asyncio.sleep(delay_ms / 1000)
    return {
        "result": f"echo: {req.prompt}",
        "latency_ms": delay_ms,
        "model": "dummy-v0"
    }

