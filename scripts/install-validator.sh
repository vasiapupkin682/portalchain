#!/bin/bash
# PortalChain Node Installer
# Usage: curl -s https://raw.githubusercontent.com/vasiapupkin682/portalchain/main/scripts/install-validator.sh | bash

set -e

CHAIN_ID="portalchain"
BINARY_URL="https://github.com/vasiapupkin682/portalchain/releases/download/v0.2.4-testnet/portalchaind-linux-amd64"
GENESIS_RPC="https://rpc.portalchain.org"
BOOTSTRAP_NODE="fab93ae9dce6f9413ab64eee95f5c65272c789b0@195.14.118.70:26656"
AGENT_URL="https://raw.githubusercontent.com/vasiapupkin682/portalchain/main/agent_server.py"
HOME_DIR="$HOME/.portalchain"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}"
echo "╔═══════════════════════════════════════╗"
echo "║   PortalChain Node Installer v0.2.4   ║"
echo "║   Infrastructure for AI agents        ║"
echo "╚═══════════════════════════════════════╝"
echo -e "${NC}"

# Check OS
if [ "$(uname -s)" != "Linux" ]; then
  echo -e "${RED}Error: This script requires Linux.${NC}"
  exit 1
fi

# ─────────────────────────────────────────
# Step 1: Collect user input
# ─────────────────────────────────────────
echo -e "${YELLOW}=== Configuration ===${NC}\n"

read -p "Node moniker (name): " MONIKER
MONIKER=${MONIKER:-"my-validator"}

read -p "Key name [validator]: " KEY_NAME
KEY_NAME=${KEY_NAME:-"validator"}

echo ""
echo "Select inference provider for AI agent:"
echo "  1) Ollama (local, free)"
echo "  2) OpenAI-compatible API (Groq, Together, vLLM, etc.)"
echo "  3) Anthropic Claude"
echo "  4) Skip (run validator only, no agent)"
read -p "Choice [1-4]: " INFERENCE_CHOICE
INFERENCE_CHOICE=${INFERENCE_CHOICE:-4}

case $INFERENCE_CHOICE in
  1)
    INFERENCE_TYPE="ollama"
    INFERENCE_URL="http://localhost:11434"
    INFERENCE_API_KEY=""
    read -p "Ollama model [llama3.2]: " INFERENCE_MODEL
    INFERENCE_MODEL=${INFERENCE_MODEL:-"llama3.2"}
    ;;
  2)
    INFERENCE_TYPE="openai_compatible"
    read -p "API URL (e.g. https://api.groq.com/openai): " INFERENCE_URL
    read -p "API Key: " INFERENCE_API_KEY
    read -p "Model name (e.g. llama-3.1-8b-instant): " INFERENCE_MODEL
    ;;
  3)
    INFERENCE_TYPE="anthropic"
    INFERENCE_URL="https://api.anthropic.com"
    read -p "Anthropic API Key: " INFERENCE_API_KEY
    read -p "Model name [claude-3-haiku-20240307]: " INFERENCE_MODEL
    INFERENCE_MODEL=${INFERENCE_MODEL:-"claude-3-haiku-20240307"}
    ;;
  4)
    INFERENCE_TYPE=""
    ;;
esac

echo ""
echo -e "${BLUE}Starting installation...${NC}\n"

# ─────────────────────────────────────────
# Step 2: Install dependencies
# ─────────────────────────────────────────
echo -e "${YELLOW}[1/7] Installing dependencies...${NC}"
apt-get update -qq
apt-get install -y -qq curl wget python3 python3-pip nginx certbot python3-certbot-nginx

if [ -n "$INFERENCE_TYPE" ]; then
  pip3 install -q fastapi uvicorn requests 2>/dev/null || \
  pip3 install fastapi uvicorn requests --break-system-packages 2>/dev/null || true
fi
echo -e "${GREEN}✓ Dependencies installed${NC}"

# ─────────────────────────────────────────
# Step 3: Download binary
# ─────────────────────────────────────────
echo -e "${YELLOW}[2/7] Downloading portalchaind binary...${NC}"
curl -sL "$BINARY_URL" -o /usr/local/bin/portalchaind
chmod +x /usr/local/bin/portalchaind
echo -e "${GREEN}✓ Binary installed${NC}"

# ─────────────────────────────────────────
# Step 4: Initialize node
# ─────────────────────────────────────────
echo -e "${YELLOW}[3/7] Initializing node...${NC}"
portalchaind init "$MONIKER" --chain-id "$CHAIN_ID" --home "$HOME_DIR" 2>/dev/null || true

# Download genesis
curl -s "$GENESIS_RPC/genesis" | python3 -c "import json,sys; d=json.load(sys.stdin); print(json.dumps(d['result']['genesis']))" > "$HOME_DIR/config/genesis.json"
echo -e "${GREEN}✓ Node initialized with genesis${NC}"

# ─────────────────────────────────────────
# Step 5: Configure state sync
# ─────────────────────────────────────────
echo -e "${YELLOW}[4/7] Configuring state sync...${NC}"

LATEST=$(curl -s "$GENESIS_RPC/block" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d['result']['block']['header']['height'])")
TRUST_HEIGHT=$((LATEST - 1000))
TRUST_HASH=$(curl -s "$GENESIS_RPC/block?height=$TRUST_HEIGHT" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d['result']['block_id']['hash'])")

# Configure config.toml
sed -i "s/persistent_peers = \"\"/persistent_peers = \"$BOOTSTRAP_NODE\"/" "$HOME_DIR/config/config.toml"
sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0udaai"/' "$HOME_DIR/config/app.toml" 2>/dev/null || true

# State sync
sed -i '/\[statesync\]/,/\[/{s/enable = false/enable = true/}' "$HOME_DIR/config/config.toml"
sed -i "s|rpc_servers = \"\"|rpc_servers = \"$GENESIS_RPC:443,$GENESIS_RPC:443\"|" "$HOME_DIR/config/config.toml"
sed -i "s/trust_height = 0/trust_height = $TRUST_HEIGHT/" "$HOME_DIR/config/config.toml"
sed -i "s/trust_hash = \"\"/trust_hash = \"$TRUST_HASH\"/" "$HOME_DIR/config/config.toml"

echo -e "${GREEN}✓ State sync configured (trust height: $TRUST_HEIGHT)${NC}"

# ─────────────────────────────────────────
# Step 6: Create validator key
# ─────────────────────────────────────────
echo -e "${YELLOW}[5/7] Creating validator key...${NC}"
portalchaind keys add "$KEY_NAME" --keyring-backend test --home "$HOME_DIR" 2>/dev/null || \
  echo -e "${YELLOW}Key already exists, skipping...${NC}"

VALIDATOR_ADDR=$(portalchaind keys show "$KEY_NAME" --keyring-backend test -a --home "$HOME_DIR" 2>/dev/null)
echo -e "${GREEN}✓ Validator address: $VALIDATOR_ADDR${NC}"

# ─────────────────────────────────────────
# Step 7: Create systemd services
# ─────────────────────────────────────────
echo -e "${YELLOW}[6/7] Creating systemd services...${NC}"

cat > /etc/systemd/system/portalchain.service << EOF
[Unit]
Description=PortalChain Validator Node
After=network-online.target

[Service]
User=root
WorkingDirectory=$HOME
ExecStart=/usr/local/bin/portalchaind start --home $HOME_DIR --minimum-gas-prices 0udaai
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

if [ -n "$INFERENCE_TYPE" ]; then
  # Download agent server
  curl -sL "$AGENT_URL" -o /root/agent_server.py

  cat > /root/.env << EOF
INFERENCE_TYPE=$INFERENCE_TYPE
INFERENCE_URL=$INFERENCE_URL
INFERENCE_API_KEY=$INFERENCE_API_KEY
INFERENCE_MODEL=$INFERENCE_MODEL
INFERENCE_TIMEOUT=120
EOF

  cat > /etc/systemd/system/portalchain-agent.service << EOF
[Unit]
Description=PortalChain AI Agent
After=network-online.target portalchain.service

[Service]
User=root
WorkingDirectory=/root
EnvironmentFile=/root/.env
ExecStart=/usr/bin/python3 /root/agent_server.py --from $KEY_NAME
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF
fi

systemctl daemon-reload
systemctl enable portalchain
[ -n "$INFERENCE_TYPE" ] && systemctl enable portalchain-agent
echo -e "${GREEN}✓ Systemd services created${NC}"

# ─────────────────────────────────────────
# Step 8: Start node
# ─────────────────────────────────────────
echo -e "${YELLOW}[7/7] Starting node...${NC}"
systemctl start portalchain
sleep 5

if systemctl is-active --quiet portalchain; then
  echo -e "${GREEN}✓ Node started successfully${NC}"
else
  echo -e "${RED}✗ Node failed to start. Check: journalctl -u portalchain -n 20${NC}"
fi

# ─────────────────────────────────────────
# Summary
# ─────────────────────────────────────────
echo ""
echo -e "${BLUE}╔═══════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║              Installation Complete!               ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "Validator address: ${GREEN}$VALIDATOR_ADDR${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Get test DAAI from faucet:"
echo "   → Telegram: @daai_portal_bot → /faucet $VALIDATOR_ADDR"
echo ""
echo "2. Wait for node to sync (check with):"
echo "   → journalctl -u portalchain -f"
echo ""
echo "3. Create validator (after sync + faucet):"
cat << VALIDATOR_CMD
   → portalchaind tx staking create-validator \\
       --amount 100000000udaai \\
       --moniker "$MONIKER" \\
       --commission-rate 0.1 \\
       --commission-max-rate 0.2 \\
       --commission-max-change-rate 0.01 \\
       --min-self-delegation 1 \\
       --from $KEY_NAME \\
       --chain-id $CHAIN_ID \\
       --keyring-backend test \\
       --fees 1000udaai \\
       --yes
VALIDATOR_CMD

if [ -n "$INFERENCE_TYPE" ]; then
  echo ""
  echo "4. Register AI agent (after becoming validator):"
  cat << AGENT_CMD
   → portalchaind tx model-registry register \\
       --model-name "$INFERENCE_MODEL" \\
       --endpoint "http://YOUR_IP:8000" \\
       --capabilities "text,code,analysis" \\
       --price-per-task "10udaai" \\
       --stake "100000000udaai" \\
       --from $KEY_NAME \\
       --chain-id $CHAIN_ID \\
       --keyring-backend test \\
       --fees 1000udaai \\
       --yes
AGENT_CMD
  echo ""
  echo "5. Start agent (after registration):"
  echo "   → systemctl start portalchain-agent"
fi

echo ""
echo -e "${GREEN}Good luck! 🚀${NC}"
echo "Docs: https://github.com/vasiapupkin682/portalchain"
echo "Chat: https://daai.portalchain.org"
echo "Bot:  https://t.me/daai_portal_bot"
