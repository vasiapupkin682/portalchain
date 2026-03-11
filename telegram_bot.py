#!/usr/bin/env python3
"""
Telegram bot for PortalChain AI agent.
Connects users to the agent_server running on localhost:8000.
"""

import asyncio
import os

import requests
from telegram import Update
from telegram.ext import Application, CommandHandler, MessageHandler, filters

AGENT_URL = os.getenv("AGENT_URL", "http://localhost:8000")
BOT_TOKEN = os.getenv("TELEGRAM_BOT_TOKEN")


async def start(update: Update, context):
    await update.message.reply_text(
        "🤖 Welcome to PortalChain AI Agent!\n\n"
        "I'm powered by Llama 3.2 running on a decentralized network.\n"
        "Every task I complete is recorded on the PortalChain blockchain.\n\n"
        "Commands:\n"
        "/ask <question> — Ask me anything\n"
        "/status — Check agent status\n"
        "/reputation — Check my blockchain reputation\n"
        "/help — Show this message"
    )


async def ask(update: Update, context):
    if not context.args:
        await update.message.reply_text("Usage: /ask <your question>")
        return

    question = " ".join(context.args)

    # Show typing indicator
    await update.message.reply_text("🤔 Thinking...")

    try:
        response = requests.post(
            f"{AGENT_URL}/task",
            json={"prompt": question, "max_tokens": 500},
            timeout=120,
        )

        if response.status_code == 200:
            data = response.json()
            result = data["result"]
            latency_ms = data["latency_ms"]
            epoch = data["epoch"]
            task_type = data.get("task_type", "general")
            reply = (
                f"{result}\n\n"
                f"─────────────────\n"
                f"⛓ Recorded on PortalChain\n"
                f"📊 Epoch: {epoch} | ⚡ {latency_ms}ms | 🏷 {task_type}"
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
        response = requests.get(f"{AGENT_URL}/status", timeout=10)
        data = response.json()

        reply = (
            f"🤖 Agent Status\n"
            f"─────────────────\n"
            f"📍 Validator: {data['validator']}\n"
            f"⭐ Reputation: {data['reputation']:.4f}\n"
            f"✅ Tasks completed: {data['tasks_completed']}\n"
            f"❌ Tasks failed: {data['tasks_failed']}\n"
            f"📊 Current epoch: {data['current_epoch']}\n"
            f"🔄 Buffer: {data['buffer_size']} tasks\n"
            f"🧠 Ollama: {'✅' if data['ollama_available'] else '❌'}"
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

    app = Application.builder().token(BOT_TOKEN).build()

    app.add_handler(CommandHandler("start", start))
    app.add_handler(CommandHandler("help", start))
    app.add_handler(CommandHandler("ask", ask))
    app.add_handler(CommandHandler("status", status))
    app.add_handler(CommandHandler("reputation", reputation))
    app.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, handle_message))

    print("🤖 PortalChain Telegram bot started")
    app.run_polling()


if __name__ == "__main__":
    main()
