#!/bin/bash
set -e

AGENT_PORT=${AGENT_PORT:-8000}
INFERENCE_TYPE=${INFERENCE_TYPE:-"ollama"}
INFERENCE_MODEL=${INFERENCE_MODEL:-"llama3.2"}
REPO_URL=${REPO_URL:-"https://github.com/portalchain/portalchain.git"}
PORTALCHAIN_DIR=${PORTALCHAIN_DIR:-"$HOME/portalchain"}

echo "=== Installing PortalChain AI Operator ==="

# 0. Ensure portalchain repo exists
if [ ! -d "$PORTALCHAIN_DIR" ]; then
    echo "[0/5] Cloning portalchain repo..."
    git clone "$REPO_URL" "$PORTALCHAIN_DIR"
else
    echo "[0/5] Using existing repo at $PORTALCHAIN_DIR"
fi

# 1. System dependencies
echo "[1/5] Installing dependencies..."
sudo apt-get update -qq
sudo apt-get install -y -qq curl python3 python3-pip

# 2. Install Ollama
echo "[2/5] Installing Ollama..."
if ! command -v ollama &> /dev/null; then
    curl -fsSL https://ollama.ai/install.sh | sh
fi

# 3. Pull model
echo "[3/5] Pulling model $INFERENCE_MODEL..."
ollama pull $INFERENCE_MODEL

# 4. Install agent_server dependencies
echo "[4/5] Installing Python dependencies..."
cd "$PORTALCHAIN_DIR"
pip3 install -r requirements.txt --break-system-packages -q 2>/dev/null || pip3 install -r requirements.txt -q --user

# 5. Create systemd service for agent
echo "[5/5] Creating systemd service..."
read -p "Enter your operator key name (e.g. alice): " KEY_NAME

sudo tee /etc/systemd/system/portalchain-agent.service > /dev/null << EOF
[Unit]
Description=PortalChain AI Agent
After=network-online.target ollama.service

[Service]
User=$USER
WorkingDirectory=$PORTALCHAIN_DIR
ExecStart=/usr/bin/python3 agent_server.py --from $KEY_NAME --port $AGENT_PORT
Restart=always
RestartSec=5
Environment="HOME=$HOME"
Environment="INFERENCE_TYPE=$INFERENCE_TYPE"
Environment="INFERENCE_MODEL=$INFERENCE_MODEL"
Environment="INFERENCE_URL=http://localhost:11434"

[Install]
WantedBy=multi-user.target
EOF

# Ollama service
if [ ! -f /etc/systemd/system/ollama.service ]; then
    sudo tee /etc/systemd/system/ollama.service > /dev/null << EOF
[Unit]
Description=Ollama AI Runtime
After=network-online.target

[Service]
User=$USER
ExecStart=$(which ollama) serve
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF
fi

sudo systemctl daemon-reload
sudo systemctl enable ollama 2>/dev/null || true
sudo systemctl enable portalchain-agent

echo ""
echo "✅ AI Operator installed!"
echo ""
echo "Next steps:"
echo "  1. Add your key:    portalchaind keys add operator --keyring-backend test"
echo "  2. Get testnet DAAI from faucet in Telegram bot"
echo "  3. Register model:  portalchaind tx model-registry register \\"
echo "       --model-name \"$INFERENCE_MODEL\" \\"
echo "       --endpoint \"http://YOUR_IP:$AGENT_PORT\" \\"
echo "       --capabilities \"text,code,analysis\" \\"
echo "       --price-per-task \"10udaai\" \\"
echo "       --from operator --chain-id portalchain --yes"
echo "  4. Start services:  sudo systemctl start ollama portalchain-agent"
echo "  5. Check agent:     curl http://localhost:$AGENT_PORT/health"
