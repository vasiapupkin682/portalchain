#!/bin/bash
set -e

CHAIN_ID="portalchain"
MONIKER=${MONIKER:-"my-validator"}
SEEDS=${SEEDS:-""}
REPO_URL=${REPO_URL:-"https://github.com/portalchain/portalchain.git"}
BUILD_DIR=${PORTALCHAIN_DIR:-"/tmp/portalchain"}

echo "=== Installing PortalChain Validator ==="

# 1. System dependencies
echo "[1/6] Installing dependencies..."
sudo apt-get update -qq
sudo apt-get install -y -qq curl git build-essential jq

# 2. Install Go
echo "[2/6] Installing Go..."
if ! command -v go &> /dev/null; then
    wget -q https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    export PATH=$PATH:/usr/local/go/bin
    rm go1.21.0.linux-amd64.tar.gz
fi
echo "Go version: $(go version)"

# 3. Install portalchaind
echo "[3/6] Installing portalchaind..."
if ! command -v portalchaind &> /dev/null; then
    if [ ! -d "$BUILD_DIR" ]; then
        git clone "$REPO_URL" "$BUILD_DIR"
    fi
    cd "$BUILD_DIR"
    go install ./cmd/portalchaind/...
    cd - > /dev/null
fi
echo "portalchaind version: $(portalchaind version 2>/dev/null || echo 'installed')"

# 4. Initialize node
echo "[4/6] Initializing node..."
if [ ! -f ~/.portalchain/config/genesis.json ]; then
    portalchaind init "$MONIKER" --chain-id $CHAIN_ID
    # Download genesis when testnet launches:
    # curl -s https://raw.githubusercontent.com/portalchain/portalchain/master/genesis.json \
    #   > ~/.portalchain/config/genesis.json
    echo "⚠️  Remember to download genesis.json when testnet launches"
fi

# 5. Configure
echo "[5/6] Configuring..."
if [ -n "$SEEDS" ]; then
    sed -i "s/seeds = \"\"/seeds = \"$SEEDS\"/" ~/.portalchain/config/config.toml
fi

# 6. Create systemd service
echo "[6/6] Creating systemd service..."
sudo tee /etc/systemd/system/portalchain.service > /dev/null << EOF
[Unit]
Description=PortalChain Validator Node
After=network-online.target

[Service]
User=$USER
ExecStart=$(which portalchaind) start
Restart=always
RestartSec=3
LimitNOFILE=65535
Environment="HOME=$HOME"

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable portalchain

echo ""
echo "✅ Validator installed!"
echo ""
echo "Next steps:"
echo "  1. Add your key:    portalchaind keys add validator --keyring-backend test"
echo "  2. Get testnet DAAI from faucet in Telegram bot"
echo "  3. Start node:      sudo systemctl start portalchain"
echo "  4. Check logs:      sudo journalctl -fu portalchain"
echo "  5. Create validator tx after node syncs"
