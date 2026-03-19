#!/bin/bash
set -e

AGENT_PORT=${AGENT_PORT:-8000}
INFERENCE_TYPE=${INFERENCE_TYPE:-"ollama"}
INFERENCE_MODEL=${INFERENCE_MODEL:-"llama3.2"}
VERSION="v0.1.0-testnet"
REPO="vasiapupkin682/portalchain"
PORTALCHAIN_DIR=${PORTALCHAIN_DIR:-"$HOME/portalchain"}

echo "=== Installing PortalChain AI Operator ==="

# 1. System dependencies
echo "[1/6] Installing dependencies..."
sudo apt-get update -qq
sudo apt-get install -y -qq curl wget python3 python3-pip

# 2. Download binary
echo "[2/6] Downloading portalchaind $VERSION..."
if ! command -v portalchaind &> /dev/null; then
    wget -q https://github.com/$REPO/releases/download/$VERSION/portalchaind-linux-amd64 \
      -O /tmp/portalchaind
    chmod +x /tmp/portalchaind
    sudo mv /tmp/portalchaind /usr/local/bin/portalchaind
fi

# 3. Clone repo for agent_server.py
echo "[3/6] Downloading agent server..."
if [ ! -d "$PORTALCHAIN_DIR" ]; then
    git clone https://github.com/$REPO.git "$PORTALCHAIN_DIR"
fi

# 4. Install Ollama
echo "[4/6] Installing Ollama..."
if ! command -v ollama &> /dev/null; then
    curl -fsSL https://ollama.ai/install.sh | sh
fi
ollama pull $INFERENCE_MODEL

# 5. Install Python dependencies
echo "[5/6] Installing Python dependencies..."
pip3 install -r $PORTALCHAIN_DIR/requirements.txt --break-system-packages -q

# 6. Create systemd services
echo "[6/6] Creating systemd services..."
read -p "Enter your operator key name (e.g. alice): " KEY_NAME

sudo tee /etc/systemd/system/ollama.service > /dev/null << SERVICE
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
SERVICE

sudo tee /etc/systemd/system/portalchain-agent.service > /dev/null << SERVICE
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
SERVICE

sudo systemctl daemon-reload
sudo systemctl enable ollama
sudo systemctl enable portalchain-agent

echo ""
echo "✅ AI Operator installed!"
echo ""
echo "Next steps:"
echo "  1. Create key:     portalchaind keys add operator --keyring-backend test"
echo "  2. Get DAAI:       /faucet <address> in @daai_portal_bot"
echo "  3. Start services: sudo systemctl start ollama portalchain-agent"
echo "  4. Register model: portalchaind tx model-registry register \\"
echo "       --model-name \"$INFERENCE_MODEL\" \\"
echo "       --endpoint \"http://YOUR_IP:$AGENT_PORT\" \\"
echo "       --capabilities \"text,code,analysis\" \\"
echo "       --price-per-task \"10udaai\" \\"
echo "       --from operator --chain-id portalchain --yes"
echo "  5. Check agent:    curl http://localhost:$AGENT_PORT/health"
