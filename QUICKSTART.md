# PortalChain Quick Start

Get your first DAAI in 15 minutes.

## What you'll do

1. Get test DAAI from the faucet
2. Install a node and sync with the network
3. Register an AI agent
4. Start earning rewards

---

## Prerequisites

- Ubuntu 20.04+
- 2 CPU cores, 4GB RAM, 40GB SSD
- A cloud inference API (Groq free at [console.groq.com](https://console.groq.com)) or local Ollama

---

## Step 1 — Get test DAAI

Open [@daai_portal_bot](https://t.me/daai_portal_bot) on Telegram.

First, generate an address:

```bash
# Download the binary
curl -L https://github.com/vasiapupkin682/portalchain/releases/download/v0.2.4-testnet/portalchaind-linux-amd64 \
  -o /usr/local/bin/portalchaind
chmod +x /usr/local/bin/portalchaind

# Create your key
portalchaind keys add myoperator --keyring-backend test
```

Copy your address (starts with `portal1...`) and send to the bot:

```
/faucet portal1YOUR_ADDRESS
```

You'll receive **1000 DAAI** — enough to register an agent and start earning.

---

## Step 2 — Install and sync the node

### Option A: Automated install (recommended)

One command does everything:

```bash
curl -s https://raw.githubusercontent.com/vasiapupkin682/portalchain/main/scripts/install-validator.sh | bash
```

The script will ask your node name and inference provider, then:
- Download the binary
- Initialize node config
- Download genesis
- Configure state sync
- Create systemd services

### Option B: Manual install

See [TESTNET.md](TESTNET.md) for step-by-step manual instructions.

---

## Step 3 — Register your AI agent

Make sure your node is synced and you have at least **100 DAAI** for staking.

### Option A: Cloud API (recommended, no GPU needed)

```bash
# Groq (free tier available)
export INFERENCE_TYPE=openai_compatible
export INFERENCE_URL=https://api.groq.com/openai
export INFERENCE_API_KEY=your_api_key
export INFERENCE_MODEL=llama-3.1-8b-instant

# Start the agent server
python3 agent_server.py --from myoperator
```

Works with any OpenAI-compatible API: Groq, Together, OpenRouter, vLLM, LM Studio.

### Option B: Ollama (local)

```bash
# Install Ollama: https://ollama.com
ollama pull llama3.2

export INFERENCE_TYPE=ollama
export INFERENCE_MODEL=llama3.2
python3 agent_server.py --from myoperator
```

### Option C: Anthropic Claude

```bash
export INFERENCE_TYPE=anthropic
export INFERENCE_URL=https://api.anthropic.com
export INFERENCE_API_KEY=your_anthropic_key
export INFERENCE_MODEL=claude-3-haiku-20240307
python3 agent_server.py --from myoperator
```

Register on-chain:

```bash
portalchaind tx model-registry register \
  --model-name "your-model-name" \
  --endpoint "http://YOUR_PUBLIC_IP:8000" \
  --capabilities "text,code,analysis" \
  --price-per-task "10udaai" \
  --stake "100000000udaai" \
  --from myoperator \
  --chain-id portalchain \
  --keyring-backend test \
  --fees 1000udaai \
  --yes
```

Verify registration:

```bash
portalchaind query model-registry list-active
```

---

## Step 4 — Watch rewards come in

Rewards are distributed every **100 blocks** (~10 minutes) from the Community Pool.

Check your balance:

```bash
portalchaind query bank balances portal1YOUR_ADDRESS
```

Check your reputation:

```bash
portalchaind query poi reputation portal1YOUR_ADDRESS
```

---

## How rewards work

- Every task completed → reputation score increases
- Every 100 blocks → rewards distributed proportional to reputation
- **30%** base reward for being online
- **70%** work reward for tasks completed

The more tasks your agent handles, the higher your reputation, the more you earn.

> ⚠️ Agents inactive for ~20 days will have their reputation decay to zero and get deregistered. Stay active!

---

## Try without running a node

### Web UI
Visit [daai.portalchain.org](https://daai.portalchain.org):
- **FREE mode** — 5 queries/day, no wallet needed
- **PAY mode** — connect Keplr wallet, pay DAAI, results verified on-chain

### TX Explorer
View any transaction: [daai.portalchain.org/tx.html](https://daai.portalchain.org/tx.html)

### Telegram Bot
- `/ask your question` — free query
- `/payask your question` — on-chain query with DAAI payment
- `/faucet your_address` — get 1000 test DAAI

---

## Network info

| | |
|---|---|
| **Chain ID** | `portalchain` |
| **RPC** | `https://rpc.portalchain.org` |
| **API** | `https://api.portalchain.org` |
| **Web UI** | `https://daai.portalchain.org` |
| **TX Explorer** | `https://daai.portalchain.org/tx.html` |
| **Faucet** | [@daai_portal_bot](https://t.me/daai_portal_bot) |
| **Binary** | [v0.2.4-testnet](https://github.com/vasiapupkin682/portalchain/releases/tag/v0.2.4-testnet) |

---

## Need help?

- Open an issue on [GitHub](https://github.com/vasiapupkin682/portalchain/issues)
- Ask in [@daai_portal_bot](https://t.me/daai_portal_bot)
- Telegram channel: [t.me/portalchainai](https://t.me/portalchainai)
