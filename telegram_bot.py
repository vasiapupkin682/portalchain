#!/usr/bin/env python3
"""
Telegram bot for PortalChain AI agent.
Queries x/model-registry on-chain to find the best agent for each task,
or falls back to AGENT_URL when no agents are registered.
"""

import json
import os
import random
import subprocess
import time

import requests
from telegram import Update
from telegram.ext import Application, CommandHandler, MessageHandler, filters
from telegram.request import HTTPXRequest

AGENT_URL = os.getenv("AGENT_URL", "http://localhost:8000")  # fallback
CHAIN_RPC = os.getenv("CHAIN_RPC", "http://localhost:26657")
BOT_TOKEN = os.getenv("TELEGRAM_BOT_TOKEN")
MAX_HISTORY = 20

# conversation_history: chat_id -> list of {"role": "user"|"assistant", "content": str}
conversation_history: dict[int, list[dict[str, str]]] = {}

FAUCET_FILE = "faucet_history.json"
FAUCET_AMOUNT = "1000daai"
FAUCET_COOLDOWN_HOURS = 24


def load_faucet_history() -> dict:
    try:
        with open(FAUCET_FILE) as f:
            return json.load(f)
    except Exception:
        return {}


def save_faucet_history(history: dict):
    with open(FAUCET_FILE, "w") as f:
        json.dump(history, f, indent=2)


def can_receive_faucet(address: str) -> tuple[bool, int]:
    """Returns (can_receive, hours_until_next)"""
    history = load_faucet_history()
    if address not in history:
        return True, 0
    last_time = history[address]
    elapsed = time.time() - last_time
    cooldown = FAUCET_COOLDOWN_HOURS * 3600
    if elapsed >= cooldown:
        return True, 0
    hours_left = int((cooldown - elapsed) / 3600) + 1
    return False, hours_left


def get_active_models() -> list:
    """Query model registry via CLI and return list of active models."""
    try:
        result = subprocess.run(
            ["portalchaind", "q", "model-registry", "list-active", "--output", "json"],
            capture_output=True,
            text=True,
            timeout=10,
        )
        if result.returncode != 0:
            return []
        data = json.loads(result.stdout)
        if isinstance(data, list):
            return data
        return data.get("models", [])
    except Exception:
        return []


def classify_task(prompt: str) -> str:
    """Classify task into category."""
    prompt_lower = prompt.lower()
    code_keywords = [
        "code", "function", "python", "javascript",
        "debug", "script", "программ", "код", "функци",
    ]
    analysis_keywords = [
        "analyze", "compare", "research", "анализ",
        "сравни", "исследуй", "объясни подробно",
    ]
    if any(w in prompt_lower for w in code_keywords):
        return "code"
    elif any(w in prompt_lower for w in analysis_keywords):
        return "analysis"
    else:
        return "text"


def get_rep_for_category(model: dict, category: str) -> float:
    """Get reputation for specific category from model record."""
    mapping = {
        "code": "rep_code",
        "analysis": "rep_analysis",
        "text": "rep_text",
        "general": "rep_general",
    }
    field = mapping.get(category, "rep_general")
    try:
        return float(model.get(field, "0.0"))
    except (ValueError, TypeError):
        return 0.0


def select_agent(models: list, task_type: str) -> str:
    """
    Weighted random selection based on category reputation.

    Algorithm:
    1. Get max reputation in network for relative threshold
    2. Filter: only models where rep > max_rep * 0.2 (relative threshold)
       If no models pass threshold, use all active models (cold start)
    3. Weighted random: probability proportional to reputation
       Add small base weight (0.1) so agents with 0.0 rep still get some chance
    4. Return endpoint of selected agent
    """
    if not models:
        return AGENT_URL  # fallback

    # Get category reputation for each model
    reps = [(m, get_rep_for_category(m, task_type)) for m in models]

    # Relative threshold
    max_rep = max(r for _, r in reps) if reps else 0
    threshold = max_rep * 0.2

    # Filter by threshold
    eligible = [(m, r) for m, r in reps if r >= threshold]
    if not eligible:
        eligible = reps  # cold start: use all

    # Weighted random with base weight
    base_weight = 0.1
    weights = [r + base_weight for _, r in eligible]
    total = sum(weights)

    rand = random.uniform(0, total)
    cumulative = 0
    for (model, _), weight in zip(eligible, weights):
        cumulative += weight
        if rand <= cumulative:
            return model.get("endpoint", AGENT_URL)

    return eligible[-1][0].get("endpoint", AGENT_URL)


async def start(update: Update, context):
    await update.message.reply_text(
        "🤖 Welcome to PortalChain AI Agent!\n\n"
        "I'm powered by Llama 3.2 running on a decentralized network.\n"
        "Every task I complete is recorded on the PortalChain blockchain.\n"
        "I route your questions to the best available agent.\n\n"
        "Commands:\n"
        "/ask <question> — Ask me anything\n"
        "/forget — Clear conversation history\n"
        "/status — Check agent status\n"
        "/reputation — Check my blockchain reputation\n"
        "/agents — List active AI agents in network\n"
        "/faucet <address> — Get 1000 test DAAI tokens\n"
        "/help — Show this message"
    )


def get_history_for_chat(chat_id: int) -> list[dict[str, str]]:
    """Get conversation history for a chat, trimmed to MAX_HISTORY."""
    history = conversation_history.get(chat_id, [])
    if len(history) > MAX_HISTORY:
        history = history[-MAX_HISTORY:]
        conversation_history[chat_id] = history
    return history


def append_to_history(chat_id: int, role: str, content: str):
    """Append a message to conversation history."""
    if chat_id not in conversation_history:
        conversation_history[chat_id] = []
    conversation_history[chat_id].append({"role": role, "content": content})
    if len(conversation_history[chat_id]) > MAX_HISTORY:
        conversation_history[chat_id] = conversation_history[chat_id][-MAX_HISTORY:]


async def ask(update: Update, context):
    if not context.args:
        await update.message.reply_text("Usage: /ask <your question>")
        return

    question = " ".join(context.args)
    chat_id = update.effective_chat.id if update.effective_chat else 0

    # Show typing indicator
    await update.message.reply_text("🤔 Thinking...")

    task_type = classify_task(question)
    models = get_active_models()
    endpoint = select_agent(models, task_type)
    history = get_history_for_chat(chat_id)

    try:
        response = requests.post(
            f"{endpoint}/task",
            json={"prompt": question, "max_tokens": 500, "history": history},
            timeout=120,
        )

        if response.status_code == 200:
            data = response.json()
            result = data["result"]
            latency_ms = data["latency_ms"]
            epoch = data["epoch"]
            task_type_resp = data.get("task_type", task_type)

            # Update conversation history
            append_to_history(chat_id, "user", question)
            append_to_history(chat_id, "assistant", result)

            if len(models) > 1:
                reply = (
                    f"{result}\n\n"
                    f"─────────────────\n"
                    f"🤖 Agent: {endpoint}\n"
                    f"⛓ Recorded on PortalChain\n"
                    f"📊 Epoch: {epoch} | ⚡ {latency_ms}ms | 🏷 {task_type_resp}"
                )
            else:
                reply = (
                    f"{result}\n\n"
                    f"─────────────────\n"
                    f"⛓ Recorded on PortalChain\n"
                    f"📊 Epoch: {epoch} | ⚡ {latency_ms}ms | 🏷 {task_type_resp}"
                )
            await update.message.reply_text(reply)
        else:
            await update.message.reply_text("❌ Agent unavailable, try again later")

    except requests.exceptions.Timeout:
        await update.message.reply_text(
            "⏰ The agent is thinking too long. Try a simpler question."
        )
    except Exception as e:
        await update.message.reply_text(f"❌ Error: {str(e)[:100]}")


async def status(update: Update, context):
    try:
        active_models = get_active_models()

        response = requests.get(f"{AGENT_URL}/status", timeout=10)
        data = response.json()

        balance = "0"
        try:
            bal_resp = requests.get(f"{AGENT_URL}/balance", timeout=5)
            if bal_resp.status_code == 200:
                balance = bal_resp.json().get("balance", "0")
        except Exception:
            pass

        inference = data.get("inference_available", data.get("ollama_available", False))
        reply = (
            f"🤖 Agent Status\n"
            f"─────────────────\n"
            f"🌐 Active agents: {len(active_models)}\n"
            f"📍 Validator: {data['validator']}\n"
            f"⭐ Reputation: {data['reputation']:.4f}\n"
            f"💰 Balance: {balance} DAAI\n"
            f"✅ Tasks completed: {data['tasks_completed']}\n"
            f"❌ Tasks failed: {data['tasks_failed']}\n"
            f"📊 Current epoch: {data['current_epoch']}\n"
            f"🔄 Buffer: {data['buffer_size']} tasks\n"
            f"🧠 Inference: {data.get('inference_provider', 'ollama')}/{data.get('inference_model', 'n/a')} {'✅' if inference else '❌'}"
        )
        await update.message.reply_text(reply)
    except Exception as e:
        await update.message.reply_text(f"❌ Cannot reach agent: {str(e)[:100]}")


async def reputation(update: Update, context):
    try:
        response = requests.get(f"{AGENT_URL}/status", timeout=10)
        data = response.json()
        rep = data["reputation"]

        # Visual reputation bar
        filled = int(rep * 20)
        bar = "█" * filled + "░" * (20 - filled)

        reply = (
            f"⭐ Blockchain Reputation\n"
            f"─────────────────\n"
            f"[{bar}]\n"
            f"{rep:.6f} / 1.000000\n\n"
            f"📍 Recorded on PortalChain\n"
            f"🔗 Validator: {data['validator']}"
        )
        await update.message.reply_text(reply)
    except Exception as e:
        await update.message.reply_text(f"❌ Error: {str(e)[:100]}")


async def agents(update: Update, context):
    models = get_active_models()
    if not models:
        await update.message.reply_text("No active agents found")
        return

    lines = [f"🌐 Active Agents: {len(models)}\n"]
    for i, m in enumerate(models, 1):
        try:
            rep_text = float(m.get("rep_text", "0.0"))
        except (ValueError, TypeError):
            rep_text = 0.0
        try:
            rep_code = float(m.get("rep_code", "0.0"))
        except (ValueError, TypeError):
            rep_code = 0.0
        lines.append(
            f"[{i}] {m.get('model_name', 'unknown')}\n"
            f"    📍 {m.get('endpoint', 'N/A')}\n"
            f"    ⭐ text:{rep_text:.3f} code:{rep_code:.3f}\n"
        )

    await update.message.reply_text("\n".join(lines))


async def forget(update: Update, context):
    chat_id = update.effective_chat.id if update.effective_chat else 0
    if chat_id in conversation_history:
        conversation_history[chat_id] = []
    await update.message.reply_text("🧹 Conversation history cleared.")


async def faucet(update: Update, context):
    if not context.args:
        await update.message.reply_text(
            "Usage: /faucet <your_portal_address>\n"
            "Example: /faucet portal1abc...xyz\n\n"
            "You can get 1000 DAAI once every 24 hours."
        )
        return

    address = context.args[0].strip()

    # Validate address format
    if not address.startswith("portal1") or len(address) < 20:
        await update.message.reply_text("❌ Invalid address. Must start with portal1...")
        return

    # Check cooldown
    can_receive, hours_left = can_receive_faucet(address)
    if not can_receive:
        await update.message.reply_text(
            f"⏰ Already received tokens.\n"
            f"Come back in {hours_left} hours."
        )
        return

    await update.message.reply_text("⏳ Sending tokens...")

    try:
        # Send tokens via CLI
        result = subprocess.run(
            [
                "portalchaind",
                "tx",
                "bank",
                "send",
                "alice",  # faucet account
                address,
                FAUCET_AMOUNT,
                "--chain-id",
                "portalchain",
                "--keyring-backend",
                "test",
                "--yes",
                "--output",
                "json",
            ],
            capture_output=True,
            text=True,
            timeout=30,
        )

        if result.returncode == 0:
            data = json.loads(result.stdout)
            txhash = data.get("txhash", "unknown")

            # Save to history
            history = load_faucet_history()
            history[address] = time.time()
            save_faucet_history(history)

            await update.message.reply_text(
                f"✅ Sent 1000 DAAI!\n\n"
                f"📍 To: {address[:20]}...\n"
                f"🔗 TX: {txhash[:20]}...\n\n"
                f"⛓ PortalChain Testnet"
            )
        else:
            await update.message.reply_text(
                "❌ Transaction failed. Try again later."
            )
    except subprocess.TimeoutExpired:
        await update.message.reply_text("⏰ Timeout. Node may be busy, try again.")
    except Exception as e:
        await update.message.reply_text(f"❌ Error: {str(e)[:100]}")


async def handle_message(update: Update, context):
    # Handle plain text messages (not commands) as /ask
    text = update.message.text
    context.args = text.split()
    await ask(update, context)


def main():
    if not BOT_TOKEN:
        print("❌ Set TELEGRAM_BOT_TOKEN environment variable")
        print("   export TELEGRAM_BOT_TOKEN=your_token_here")
        return

    request = HTTPXRequest(
        connection_pool_size=8,
        read_timeout=120,
        write_timeout=120,
        connect_timeout=30,
    )
    app = Application.builder().token(BOT_TOKEN).request(request).build()

    app.add_handler(CommandHandler("start", start))
    app.add_handler(CommandHandler("help", start))
    app.add_handler(CommandHandler("ask", ask))
    app.add_handler(CommandHandler("status", status))
    app.add_handler(CommandHandler("reputation", reputation))
    app.add_handler(CommandHandler("agents", agents))
    app.add_handler(CommandHandler("faucet", faucet))
    app.add_handler(CommandHandler("forget", forget))
    app.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, handle_message))

    print("🤖 PortalChain Telegram bot started")
    app.run_polling()


if __name__ == "__main__":
    main()
