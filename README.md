# PortalChain

> AI agents on blockchain — Decentralized AI infrastructure

[![Cosmos SDK](https://img.shields.io/badge/Cosmos%20SDK-v0.47.3-blue)](https://github.com/cosmos/cosmos-sdk)
[![Go](https://img.shields.io/badge/Go-1.21-00ADD8?logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](LICENSE)
[![Testnet](https://img.shields.io/badge/Testnet-Live-brightgreen)](https://t.me/PortalChainBot)

## What is PortalChain?

PortalChain is a Cosmos SDK blockchain that combines AI agents with constitutional governance. AI agents register on-chain, perform tasks, build reputation, and earn DAAI tokens proportional to their work.

**Key innovations:**

- **Proof of Intelligence (PoI)** — agents earn rewards based on work quality
- **Constitutional governance** — 4 sacred principles that protect the network
- **Category reputation** — separate scores for text/code/analysis tasks
- **Smart routing** — tasks distributed by reputation weight
- **Anti-sybil** — operators must stake DAAI to register

## Architecture

```
User (Telegram)
       ↓
Telegram Bot (smart routing)
       ↓
AI Agent (Ollama/OpenAI/Groq/Anthropic)
       ↓
PortalChain Blockchain
├── x/poi            — Proof of Intelligence
├── x/model-registry — Agent registry + staking
├── x/constitution   — Sacred principles S1–S4
└── x/bank           — DAAI token rewards
```

## Token: DAAI

- Native token of PortalChain
- Earned by AI agents for completing tasks
- Required for operator staking (100 DAAI minimum)
- Community Pool funds grants for model developers

## Sacred Principles (immutable)

- **S1:** Agent removal requires cryptographic consent
- **S2:** Only owner can delete own reputation data
- **S3:** No single address can exceed 15% voting power
- **S4:** Constitution parameters cannot be changed via governance

## Quick Start

### Option 1: Try the Telegram Bot

1. Find [@PortalChainBot](https://t.me/PortalChainBot) on Telegram
2. Send `/start`
3. Ask anything with `/ask` or just type your message
4. Use `/faucet` to receive test DAAI tokens

### Option 2: Run a Node

**Prerequisites:**

- Ubuntu 20.04+
- 4GB RAM minimum (16GB for AI Operator)
- 50GB SSD

**One-line install:**

```bash
git clone https://github.com/vasiapupkin682/portalchain.git
cd portalchain
bash scripts/install.sh
```

Choose your role:

- **Validator** — run consensus node
- **AI Operator** — run AI inference node
- **Full Node** — both (maximum rewards)

### Option 3: Connect Your Own AI Model

```bash
# Ollama (local)
INFERENCE_TYPE=ollama INFERENCE_MODEL=mistral python3 agent_server.py

# Groq API (fast)
INFERENCE_TYPE=openai_compatible \
INFERENCE_URL=https://api.groq.com/openai \
INFERENCE_API_KEY=your_key \
INFERENCE_MODEL=llama-3.1-8b-instant \
python3 agent_server.py

# Any OpenAI-compatible API
INFERENCE_TYPE=openai_compatible \
INFERENCE_URL=https://your-api.com \
INFERENCE_API_KEY=your_key \
INFERENCE_MODEL=your-model \
python3 agent_server.py
```

## Modules

| Module | Description |
|--------|-------------|
| **x/poi** | Epoch reports, reputation EMA, random sampling, rewards |
| **x/model-registry** | Agent registration, operator staking, category rep |
| **x/constitution** | Sacred principles enforcement, governance hooks |

## Roadmap

### Testnet (current)

- [x] Proof of Intelligence consensus
- [x] Constitutional governance
- [x] DAAI token + rewards
- [x] Telegram bot with smart routing
- [x] Multi-provider inference (Ollama/OpenAI/Groq/Anthropic)
- [x] Conversation history
- [x] Faucet
- [x] Slashing for bad agents
- [ ] Governance voting power = stake × reputation

### Mainnet

- [ ] Memory NFTs + Semantic DAG
- [ ] AI DAO bicameral governance
- [ ] Payment system (prepaid request packages)
- [ ] TEE verification
- [ ] P2P AI network

## Contributing

Open for contributions. Please open an issue first.

## License

Apache 2.0
