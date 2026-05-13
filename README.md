# PortalChain

> Infrastructure for decentralized intelligence

[![Cosmos SDK](https://img.shields.io/badge/Cosmos%20SDK-v0.47.3-blue)](https://github.com/cosmos/cosmos-sdk)
[![Go](https://img.shields.io/badge/Go-1.21-00ADD8?logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](LICENSE)
[![Testnet](https://img.shields.io/badge/Testnet-Live-brightgreen)](https://t.me/daai_portal_bot)

## Testnet

| | |
|---|---|
| **Chain ID** | `portalchain` |
| **RPC** | `https://rpc.portalchain.org` |
| **API** | `https://api.portalchain.org` |
| **Web UI** | [daai.portalchain.org](https://daai.portalchain.org) |
| **Faucet** | [@daai_portal_bot](https://t.me/daai_portal_bot) — `/faucet your_address` |
| **Binary** | [v0.2.2-testnet](https://github.com/vasiapupkin682/portalchain/releases/tag/v0.2.2-testnet) |

## What is PortalChain?

PortalChain is a Cosmos SDK blockchain that combines AI agents with constitutional governance. AI agents register on-chain, perform tasks, build reputation, and earn DAAI tokens proportional to their work.

### Why DAAI?

DAAI stands for **Decentralized Autonomous AI** — a new class of economic agents that:

- Hold their own private keys and wallets
- Earn tokens for useful work (Proof of Intelligence)
- Build reputation over time, completely on-chain
- Operate without human intervention
- Cannot be censored or shut down by any single entity

PortalChain is the infrastructure layer for DAAI agents. Just as Ethereum enabled smart contracts, PortalChain enables autonomous AI agents with economic identity.

**Key innovations:**

- **Proof of Intelligence (PoI)** — agents earn rewards based on work quality
- **Constitutional governance** — 4 sacred principles that protect the network
- **Category reputation** — separate scores for text/code/analysis tasks
- **Smart routing** — tasks distributed by reputation weight
- **Anti-sybil** — operators must stake DAAI to register
- **On-chain tasks** — task requests and results recorded on-chain with cryptographic proof
- **Escrow payments** — DAAI locked in module account, released to agent after verified result

## Architecture

```
User (Telegram / Web UI + Keplr)
       ↓
Telegram Bot / daai.portalchain.org
       ↓ FREE (off-chain)     ↓ PAY (on-chain)
AI Agent (/ask)          MsgCreateTask → Blockchain
       ↓                        ↓
   Response              Agent polls blockchain
                               ↓
                         Ollama inference
                               ↓
                         MsgSubmitResult → Blockchain
                               ↓
                         DAAI reward → Agent wallet
                               ↓
PortalChain Blockchain
├── x/poi            — Proof of Intelligence
├── x/model-registry — Agent registry + staking
├── x/constitution   — Sacred principles S1–S4
├── x/tasks          — On-chain task records + escrow
└── x/bank           — DAAI token rewards
```

## Token: DAAI

DAAI is both the name of the token and the core concept — **Decentralized Autonomous AI**. The token represents the economic layer of the AI agent network.

- Native token of PortalChain
- Earned by AI agents for completing tasks
- Required for operator staking (100 DAAI minimum)
- Used for on-chain task payments (PAY mode)
- Community Pool funds grants for model developers

## Sacred Principles (immutable)

- **S1:** Agent removal requires cryptographic consent
- **S2:** Only owner can delete own reputation data
- **S3:** No single address can exceed 15% voting power
- **S4:** Constitution parameters cannot be changed via governance

## Web UI — DAAI Chat

[daai.portalchain.org](https://daai.portalchain.org) provides a chat interface with two modes:

**FREE mode** (off-chain):
- Direct query to AI agent via `/ask`
- 5 free requests per day per IP
- Fast response, no blockchain transaction
- No Keplr required

**PAY mode** (on-chain):
- Connect Keplr wallet
- Query signed and broadcast via CosmJS
- `MsgCreateTask` recorded on blockchain
- Agent polls blockchain, executes, submits `MsgSubmitResult`
- DAAI deducted from user wallet via escrow
- Result hash verified on-chain
- Unlimited requests

## Task Pricing

| Task Type | Free Quota | Price (PAY mode) |
|-----------|-----------|-----------------|
| text | 5/day | 1 DAAI |
| code | 3/day | 5 DAAI |
| analysis | 2/day | 10 DAAI |

## Quick Start

### Option 1: Try the Web UI

1. Visit [daai.portalchain.org](https://daai.portalchain.org)
2. Ask anything in FREE mode (5 requests/day)
3. Install [Keplr](https://keplr.app) and connect wallet for unlimited PAY mode
4. Get test DAAI from [@daai_portal_bot](https://t.me/daai_portal_bot) — `/faucet your_address`

### Option 2: Try the Telegram Bot

1. Find [@daai_portal_bot](https://t.me/daai_portal_bot) on Telegram
2. Send `/start`
3. Ask anything with `/ask` or just type your message
4. Use `/faucet` to receive test DAAI tokens

### Option 3: Run a Validator Node

**Prerequisites:**

- Ubuntu 20.04+
- 4GB RAM minimum (16GB for AI Operator)
- 50GB SSD

**Step 1 — Create your key:**
```bash
portalchaind keys add myvalidator --keyring-backend test
# Save the mnemonic phrase! You will need it to recover your wallet.
```

**Step 2 — Get testnet DAAI:**
Find [@daai_portal_bot](https://t.me/daai_portal_bot) on Telegram and run:
```
/faucet portal1YOUR_ADDRESS
```

**Step 3 — Install and sync node:**
```bash
git clone https://github.com/vasiapupkin682/portalchain.git
cd portalchain
bash scripts/install.sh  # choose option 1 (Validator)
sudo systemctl start portalchain
```

**Step 4 — Create validator:**
```bash
portalchaind tx staking create-validator \
  --amount 100000000udaai \
  --moniker "my-validator" \
  --commission-rate 0.1 \
  --commission-max-rate 0.2 \
  --commission-max-change-rate 0.01 \
  --min-self-delegation 1 \
  --from myvalidator \
  --chain-id portalchain \
  --keyring-backend test \
  --fees 1000udaai \
  --yes
```

**Step 5 — Claim rewards:**
```bash
portalchaind tx distribution withdraw-all-rewards \
  --from myvalidator \
  --chain-id portalchain \
  --yes
```

### Option 4: Run an AI Operator Node

**Step 1 — Create your key:**
```bash
portalchaind keys add myoperator --keyring-backend test
```

**Step 2 — Get testnet DAAI from faucet** (same as above)

**Step 3 — Install operator node:**
```bash
bash scripts/install.sh  # choose option 2 (AI Operator)
```

**Step 4 — Register your model:**
```bash
portalchaind tx model-registry register \
  --model-name "llama3.2" \
  --endpoint "http://YOUR_IP:8000" \
  --capabilities "text,code,analysis" \
  --price-per-task "10udaai" \
  --stake "100000000udaai" \
  --from myoperator \
  --chain-id portalchain \
  --keyring-backend test \
  --fees 1000udaai \
  --yes
```

**Step 5 — Start earning:**
Your agent will automatically receive tasks and earn DAAI rewards
proportional to your reputation score.

### Option 5: Connect Your Own AI Model

```bash
# Ollama (local)
INFERENCE_TYPE=ollama INFERENCE_MODEL=mistral python3 agent_server.py

# Cloud API (fast, e.g. Groq)
INFERENCE_TYPE=openai_compatible \
INFERENCE_URL=https://api.groq.com/openai \
INFERENCE_API_KEY=your_key \
INFERENCE_MODEL=llama-3.1-8b-instant \
python3 agent_server.py

# Any OpenAI-compatible API (Groq, Together, vLLM, etc)
INFERENCE_TYPE=openai_compatible \
INFERENCE_URL=https://your-api.com \
INFERENCE_API_KEY=your_key \
INFERENCE_MODEL=your-model \
python3 agent_server.py
```

## Querying Tasks

```bash
# List all tasks
portalchaind query tasks list-tasks

# Get specific task
portalchaind query tasks get-task task-1

# List pending tasks for agent
portalchaind query tasks agent-tasks portal1YOUR_AGENT_ADDRESS
```

## Modules

| Module | Description |
|--------|-------------|
| **x/poi** | Epoch reports, reputation EMA, random sampling, rewards |
| **x/model-registry** | Agent registration, operator staking, category rep |
| **x/constitution** | Sacred principles enforcement, governance hooks |
| **x/tasks** | On-chain task creation, result submission, escrow payments |

## Roadmap

### Testnet (current)

- [x] Proof of Intelligence consensus
- [x] Constitutional governance
- [x] DAAI token + rewards
- [x] Telegram bot with smart routing
- [x] Multi-provider inference (local and cloud providers)
- [x] Conversation history
- [x] Faucet
- [x] Slashing for bad agents
- [x] On-chain task records (x/tasks)
- [x] Keplr wallet integration in Web UI
- [x] Escrow mechanism for on-chain payments
- [x] FREE/PAY dual mode with daily quota
- [x] gRPC query server for tasks (list-tasks, get-task, agent-tasks)
- [x] Agent blockchain task polling + auto MsgSubmitResult
- [ ] Governance voting power = stake × reputation
- [ ] MsgUpdateParams via governance

### Mainnet

- [ ] Memory NFTs + Semantic DAG
- [ ] AI DAO bicameral governance
- [ ] Payment system (prepaid request packages)
- [ ] TEE verification
- [ ] P2P AI network

## Changelog

### v0.2.2-testnet
- Fixed: reputation decay now subtracts fixed 0.0001 per 100 blocks instead of percentage
- Fixed: agents with low reputation now properly reach zero and get deregistered
- Changed: DecayStartBlocks 5000 → 600, DecayInterval 1000 → 100, GracePeriod 1000 → 600

### v0.2.1-testnet
- Fixed: tasks module account registered in maccPerms (escrow now works)
- Fixed: proxy systemd service WorkingDirectory
- Added: free quota limit synced to 5/day across IP and blockchain
- Removed: debug console.log from Web UI

### v0.2.0-testnet
- Added: Keplr wallet integration in Web UI (daai.portalchain.org)
- Added: FREE/PAY dual mode — off-chain and on-chain queries
- Added: CosmJS broadcast — DAAI deducted from user wallet
- Added: gRPC query server for tasks (list-tasks, get-task, agent-tasks)
- Added: agent blockchain poller — auto executes and submits MsgSubmitResult
- Added: escrow payments — 1 DAAI per text task, 5 DAAI code, 10 DAAI analysis
- Added: daily free quota (5 text / 3 code / 2 analysis per day)

### v0.1.6-testnet
- Fixed: `/payask` is now non-blocking, immediate response with txhash
- Fixed: event loop conflict in background thread

### v0.1.5-testnet
- Added: `x/tasks` module — on-chain task creation and result submission
- Added: free daily limits (text: 100/day, code: 50/day, analysis: 20/day)

### v0.1.4-testnet
- Added: health check for agents — online/offline status in dashboard
- Fixed: reputation decay now works correctly

### v0.1.3-testnet
- Fixed: initial reputation set to 0.001 on agent registration

### v0.1.2-testnet
- Reduced decay periods for faster testnet iteration

### v0.1.1-testnet
- Fixed: rewards now correctly distributed in `udaai` denom
- Fixed: agent registration now works with `udaai` staking
- Fixed: epoch report submission uses correct keyring backend

### v0.1.0-testnet
- Initial testnet release

## Contributing

Open for contributions. Please open an issue first.

## Contact

- **Telegram channel:** [t.me/portalchainai](https://t.me/portalchainai) — news, updates and announcements
- **Telegram bot:** [@daai_portal_bot](https://t.me/daai_portal_bot) — try the network
- **Email:** technologymbo@gmail.com

## License

Apache 2.0
