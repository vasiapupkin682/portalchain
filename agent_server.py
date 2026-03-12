#!/usr/bin/env python3
"""
PortalChain AI Agent Server — real task execution and PoI reporting.

HTTP server that executes tasks via configurable inference providers
(Ollama, OpenAI-compatible, Anthropic), buffers results, and submits
real metrics to the PortalChain blockchain.
"""

import hashlib
import json
import logging
import os
import subprocess
import time
from abc import ABC, abstractmethod
from typing import Any, Optional

import requests
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

# ========== Inference Configuration (env vars) ==========
INFERENCE_TYPE = os.getenv("INFERENCE_TYPE", "ollama").lower()
INFERENCE_URL = os.getenv("INFERENCE_URL", "http://localhost:11434").rstrip("/")
INFERENCE_API_KEY = os.getenv("INFERENCE_API_KEY", "")
INFERENCE_MODEL = os.getenv("INFERENCE_MODEL", "llama3.2")
INFERENCE_TIMEOUT = int(os.getenv("INFERENCE_TIMEOUT", "120"))

SYSTEM_PROMPT = (
    "You are a helpful AI assistant. Always respond in the same language as "
    "the user's question. If the question is in Russian, respond in Russian. "
    "If in English, respond in English."
)


def _build_messages(history: list[dict[str, str]], prompt: str) -> list[dict[str, str]]:
    """Build message array from history + new user prompt."""
    messages = [{"role": "system", "content": SYSTEM_PROMPT}]
    for h in history:
        if isinstance(h, dict) and h.get("role") and h.get("content"):
            messages.append({"role": h["role"], "content": str(h["content"])})
    messages.append({"role": "user", "content": prompt})
    return messages


class InferenceProvider(ABC):
    """Abstract inference provider."""

    @abstractmethod
    def generate(
        self,
        prompt: str,
        history: list[dict[str, str]],
        max_tokens: int = 500,
        temperature: float = 0.7,
    ) -> tuple[str, bool]:
        """Returns (result_text, success)."""
        pass

    @abstractmethod
    def is_available(self) -> bool:
        """Check if the provider is reachable."""
        pass


class OllamaProvider(InferenceProvider):
    """Ollama API (local models)."""

    def generate(
        self,
        prompt: str,
        history: list[dict[str, str]],
        max_tokens: int = 500,
        temperature: float = 0.7,
    ) -> tuple[str, bool]:
        messages = _build_messages(history, prompt)
        try:
            resp = requests.post(
                f"{INFERENCE_URL}/api/chat",
                json={
                    "model": INFERENCE_MODEL,
                    "messages": messages,
                    "stream": False,
                    "options": {
                        "num_predict": max_tokens,
                        "temperature": temperature,
                    },
                },
                timeout=INFERENCE_TIMEOUT,
            )
            if resp.status_code != 200:
                return "", False
            text = resp.json().get("message", {}).get("content", "").strip()
            return text, True
        except Exception as e:
            logger.error("Ollama inference failed: %s", e)
            return "", False

    def is_available(self) -> bool:
        try:
            r = requests.get(f"{INFERENCE_URL}/api/tags", timeout=3)
            return r.status_code == 200
        except Exception:
            return False


class OpenAICompatibleProvider(InferenceProvider):
    """OpenAI API format (Groq, Together, OpenRouter, vLLM, LM Studio, etc.)."""

    def generate(
        self,
        prompt: str,
        history: list[dict[str, str]],
        max_tokens: int = 500,
        temperature: float = 0.7,
    ) -> tuple[str, bool]:
        messages = _build_messages(history, prompt)
        headers = {"Content-Type": "application/json"}
        if INFERENCE_API_KEY:
            headers["Authorization"] = f"Bearer {INFERENCE_API_KEY}"
        try:
            resp = requests.post(
                f"{INFERENCE_URL}/v1/chat/completions",
                headers=headers,
                json={
                    "model": INFERENCE_MODEL,
                    "messages": messages,
                    "max_tokens": max_tokens,
                    "temperature": temperature,
                },
                timeout=INFERENCE_TIMEOUT,
            )
            if resp.status_code != 200:
                return "", False
            data = resp.json()
            text = data.get("choices", [{}])[0].get("message", {}).get("content", "").strip()
            return text, True
        except Exception as e:
            logger.error("OpenAI-compatible inference failed: %s", e)
            return "", False

    def is_available(self) -> bool:
        try:
            headers = {"Content-Type": "application/json"}
            if INFERENCE_API_KEY:
                headers["Authorization"] = f"Bearer {INFERENCE_API_KEY}"
            r = requests.get(
                f"{INFERENCE_URL}/v1/models",
                headers=headers,
                timeout=5,
            )
            return r.status_code == 200
        except Exception:
            try:
                r = requests.get(INFERENCE_URL, timeout=3)
                return r.status_code < 500
            except Exception:
                return False


class AnthropicProvider(InferenceProvider):
    """Anthropic Claude API."""

    def generate(
        self,
        prompt: str,
        history: list[dict[str, str]],
        max_tokens: int = 500,
        temperature: float = 0.7,
    ) -> tuple[str, bool]:
        messages = _build_messages(history, prompt)
        # Anthropic uses "system" as separate field, not in messages
        system_msg = None
        api_messages = []
        for m in messages:
            if m.get("role") == "system":
                system_msg = m.get("content", "")
            else:
                api_messages.append(m)
        body: dict[str, Any] = {
            "model": INFERENCE_MODEL,
            "max_tokens": max_tokens,
            "messages": api_messages,
        }
        if system_msg:
            body["system"] = system_msg
        headers = {
            "Content-Type": "application/json",
            "x-api-key": INFERENCE_API_KEY,
            "anthropic-version": "2023-06-01",
        }
        try:
            resp = requests.post(
                f"{INFERENCE_URL}/v1/messages",
                headers=headers,
                json=body,
                timeout=INFERENCE_TIMEOUT,
            )
            if resp.status_code != 200:
                return "", False
            data = resp.json()
            content = data.get("content", [])
            if content and isinstance(content[0], dict):
                text = content[0].get("text", "").strip()
            else:
                text = ""
            return text, True
        except Exception as e:
            logger.error("Anthropic inference failed: %s", e)
            return "", False

    def is_available(self) -> bool:
        if not INFERENCE_API_KEY:
            return False
        try:
            headers = {"x-api-key": INFERENCE_API_KEY, "anthropic-version": "2023-06-01"}
            r = requests.get(f"{INFERENCE_URL}/v1/models", headers=headers, timeout=5)
            return r.status_code == 200
        except Exception:
            return False


def get_inference_provider() -> InferenceProvider:
    """Factory: return provider based on INFERENCE_TYPE."""
    if INFERENCE_TYPE == "ollama":
        return OllamaProvider()
    if INFERENCE_TYPE == "openai_compatible":
        return OpenAICompatibleProvider()
    if INFERENCE_TYPE == "anthropic":
        return AnthropicProvider()
    logger.warning("Unknown INFERENCE_TYPE=%s, defaulting to ollama", INFERENCE_TYPE)
    return OllamaProvider()


# ========== Logging ==========
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)
logger = logging.getLogger(__name__)

app = FastAPI(title="PortalChain Agent", version="1.0.0")
agent: Optional["PortalChainAgent"] = None
inference_provider = None


# ========== Request/Response Models ==========

class TaskRequest(BaseModel):
    prompt: str
    max_tokens: int = 500
    temperature: float = 0.7
    history: list[dict[str, str]] = []


def classify_task(prompt: str) -> str:
    """Classify task type from prompt for category-based reputation."""
    prompt_lower = prompt.lower()
    code_keywords = [
        "code", "function", "python", "javascript", "debug",
        "script", "программ", "код", "функци",
    ]
    analysis_keywords = [
        "analyze", "compare", "research", "анализ",
        "сравни", "исследуй", "объясни подробно",
    ]
    if any(w in prompt_lower for w in code_keywords):
        return "code"
    if any(w in prompt_lower for w in analysis_keywords):
        return "analysis"
    return "text"


class TaskResponse(BaseModel):
    result: str
    task_hash: str
    result_hash: str
    latency_ms: int
    agent: str
    epoch: int
    task_type: str = "general"


class StatusResponse(BaseModel):
    agent: str
    validator: str
    reputation: float
    tasks_completed: int
    tasks_failed: int
    current_epoch: int
    buffer_size: int
    inference_provider: str
    inference_model: str
    inference_available: bool


class HealthResponse(BaseModel):
    status: str
    validator: str


# ========== PortalChainAgent ==========

class PortalChainAgent:
    def __init__(self, validator_name: str, chain_id: str = "portalchain"):
        self.validator = validator_name
        self.chain_id = chain_id
        self.epoch_counter = self._load_epoch()
        self.task_buffer: list[dict] = []
        # Buffer size of 10 tasks ensures normalized_score > typical reputation.
        # With maxScore=100: 10 tasks = 0.10 normalized.
        # Reputation grows when normalized_score > current_reputation.
        # Minimum recommended buffer_size = 10.
        self.buffer_size = 10
        self.stats = {"completed": 0, "failed": 0, "total_latency": 0}

        result = subprocess.run(
            ["portalchaind", "keys", "show", self.validator, "--address"],
            capture_output=True,
            text=True,
        )
        self.address = result.stdout.strip() if result.returncode == 0 else ""

        if not self.address:
            logger.warning("Could not resolve validator address from CLI")

        logger.info(
            "PortalChainAgent initialized: validator=%s address=%s epoch=%d",
            self.validator,
            self.address[:16] + "..." if self.address else "N/A",
            self.epoch_counter,
        )

    def _epoch_file(self) -> str:
        return f"agent_{self.validator}_epoch.json"

    def _load_epoch(self) -> int:
        path = f"agent_{self.validator}_epoch.json"
        if os.path.exists(path):
            try:
                with open(path) as f:
                    data = json.load(f)
                return int(data.get("epoch", 7000))
            except Exception as e:
                logger.warning("Could not load epoch from %s: %s", path, e)
        return 7000

    def _save_epoch(self):
        path = self._epoch_file()
        try:
            with open(path, "w") as f:
                json.dump({"epoch": self.epoch_counter}, f)
        except Exception as e:
            logger.error("Could not save epoch to %s: %s", path, e)

    def get_reputation(self) -> float:
        """Query blockchain for current reputation (YAML output)."""
        try:
            result = subprocess.run(
                ["portalchaind", "q", "poi", "reputation", self.address],
                capture_output=True,
                text=True,
                check=True,
            )
            for line in result.stdout.splitlines():
                line = line.strip()
                if line.startswith("value:"):
                    val = line.split(":", 1)[1].strip().strip('"')
                    return float(val)
        except Exception as e:
            logger.debug("Could not query reputation: %s", e)
        return 0.0

    def _check_inference(self) -> bool:
        """Check if inference provider is reachable."""
        if inference_provider is None:
            return False
        return inference_provider.is_available()

    def _run_inference(
        self,
        prompt: str,
        history: list[dict[str, str]],
        max_tokens: int,
        temperature: float,
    ) -> tuple[str, bool]:
        """Returns (result_text, success)."""
        if inference_provider is None:
            return "", False
        return inference_provider.generate(prompt, history, max_tokens, temperature)

    def submit_poi_report(self):
        """Submit real metrics from task_buffer to blockchain."""
        if not self.task_buffer:
            return

        n = len(self.task_buffer)
        success_count = sum(1 for t in self.task_buffer if t.get("success", False))
        total_latency = sum(t.get("latency_ms", 0) for t in self.task_buffer)
        weighted_sum = sum(t.get("latency_ms", 0) / 10 for t in self.task_buffer)
        avg_latency = int(total_latency / n) if n else 0
        reliability = success_count / n if n else 0.0

        # Use most common task_type from buffer
        task_types = [t.get("task_type", "general") for t in self.task_buffer]
        task_type = max(set(task_types), key=task_types.count) if task_types else "general"

        cmd = [
            "portalchaind", "tx", "poi", "submit-report",
            "--epoch", str(self.epoch_counter),
            "--tasks-processed", str(n),
            "--weighted-task-sum", str(int(weighted_sum)),
            "--avg-latency", str(avg_latency),
            "--reliability", str(reliability),
            "--sampling-failures", "0",
            "--task-type", task_type,
            "--from", self.validator,
            "--chain-id", self.chain_id,
            "--yes",
            "--output", "json",
            "--broadcast-mode", "sync",
        ]

        try:
            result = subprocess.run(cmd, capture_output=True, text=True)
            logger.info(f"CLI stdout: {result.stdout[:200]}")
            if result.stderr:
                logger.error(f"CLI stderr: {result.stderr[:200]}")
            if result.returncode != 0:
                logger.error(f"CLI returncode: {result.returncode}")
            else:
                logger.info(
                    "Submitted PoI report: epoch=%d tasks=%d reliability=%.3f",
                    self.epoch_counter,
                    n,
                    reliability,
                )
                self.epoch_counter += 1
                self._save_epoch()
        except Exception as e:
            logger.error(
                "%s Error submitting PoI report: %s",
                time.strftime("%Y-%m-%d %H:%M:%S"),
                e,
            )

    def execute_task(
        self,
        prompt: str,
        history: list[dict[str, str]] | None = None,
        max_tokens: int = 500,
        temperature: float = 0.7,
    ) -> TaskResponse:
        """Run a single task and return response. Buffers for PoI report."""
        if history is None:
            history = []
        task_type = classify_task(prompt)

        start = time.perf_counter()
        result_text, success = self._run_inference(prompt, history, max_tokens, temperature)
        latency_ms = int((time.perf_counter() - start) * 1000)

        task_hash = hashlib.sha256(prompt.encode()).hexdigest()
        result_hash = hashlib.sha256(result_text.encode()).hexdigest() if result_text else ""

        if success:
            self.stats["completed"] += 1
        else:
            self.stats["failed"] += 1
        self.stats["total_latency"] += latency_ms

        self.task_buffer.append({
            "task_hash": task_hash,
            "result_hash": result_hash,
            "latency_ms": latency_ms,
            "success": success,
            "task_type": task_type,
        })

        if len(self.task_buffer) >= self.buffer_size:
            self.submit_poi_report()
            self.task_buffer.clear()

        return TaskResponse(
            result=result_text,
            task_hash=task_hash,
            result_hash=result_hash,
            latency_ms=latency_ms,
            agent=self.address,
            epoch=self.epoch_counter,
            task_type=task_type,
        )


# ========== API Endpoints ==========

@app.post("/task", response_model=TaskResponse)
def post_task(req: TaskRequest):
    if agent is None:
        raise HTTPException(status_code=503, detail="Agent not initialized")
    if not agent._check_inference():
        raise HTTPException(
            status_code=503,
            detail="inference unavailable",
        )
    try:
        resp = agent.execute_task(
            prompt=req.prompt,
            history=req.history,
            max_tokens=req.max_tokens,
            temperature=req.temperature,
        )
        return resp
    except Exception as e:
        logger.error("%s Task execution error: %s", time.strftime("%Y-%m-%d %H:%M:%S"), e)
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/status", response_model=StatusResponse)
def get_status():
    if agent is None:
        raise HTTPException(status_code=503, detail="Agent not initialized")
    return StatusResponse(
        agent=agent.address,
        validator=agent.validator,
        reputation=agent.get_reputation(),
        tasks_completed=agent.stats["completed"],
        tasks_failed=agent.stats["failed"],
        current_epoch=agent.epoch_counter,
        buffer_size=len(agent.task_buffer),
        inference_provider=INFERENCE_TYPE,
        inference_model=INFERENCE_MODEL,
        inference_available=agent._check_inference(),
    )


@app.get("/health", response_model=HealthResponse)
def get_health():
    if agent is None:
        raise HTTPException(status_code=503, detail="Agent not initialized")
    return HealthResponse(status="ok", validator=agent.validator)


def get_balance():
    """Query blockchain for DAAI balance via CLI."""
    if agent is None or not agent.address:
        return "0"
    try:
        result = subprocess.run(
            ["portalchaind", "q", "bank", "balances", agent.address, "--output", "json"],
            capture_output=True,
            text=True,
            timeout=10,
        )
        if result.returncode != 0:
            return "0"
        data = json.loads(result.stdout)
        balances = data.get("balances", [])
        daai = next((b for b in balances if b.get("denom") == "daai"), {"amount": "0"})
        return daai["amount"]
    except Exception:
        return "0"


@app.get("/balance")
def balance_endpoint():
    """Return agent's DAAI balance."""
    if agent is None:
        raise HTTPException(status_code=503, detail="Agent not initialized")
    bal = get_balance()
    return {"address": agent.address, "balance": bal, "denom": "daai"}


# ========== Startup ==========

if __name__ == "__main__":
    import argparse
    import uvicorn

    parser = argparse.ArgumentParser(description="PortalChain Agent HTTP Server")
    parser.add_argument("--from", dest="validator", default="alice", help="Validator key name")
    parser.add_argument("--port", type=int, default=8000, help="Server port")
    parser.add_argument(
        "--buffer-size",
        type=int,
        default=10,
        help="Tasks before submitting PoI report (min 10 for reputation growth)",
    )
    parser.add_argument("--chain-id", default="portalchain", help="Chain ID")
    args = parser.parse_args()

    agent = PortalChainAgent(args.validator, chain_id=args.chain_id)
    agent.buffer_size = args.buffer_size
    inference_provider = get_inference_provider()

    logger.info(
        "Starting agent server on port %d (buffer_size=%d, inference=%s, model=%s)",
        args.port, agent.buffer_size, INFERENCE_TYPE, INFERENCE_MODEL,
    )
    uvicorn.run(app, host="0.0.0.0", port=args.port)
