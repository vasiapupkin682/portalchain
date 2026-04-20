#!/bin/bash
set -e

CHAIN_ID="portalchain"
MONIKER=${MONIKER:-"my-validator"}
SEEDS=${SEEDS:-""}
VERSION="v0.1.4-testnet"
REPO="vasiapupkin682/portalchain"

echo "=== Installing PortalChain Validator ==="

# 1. System dependencies
echo "[1/5] Installing dependencies..."
sudo apt-get update -qq
sudo apt-get install -y -qq curl wget jq

# 2. Download binary
echo "[2/5] Downloading portalchaind $VERSION..."
wget -q https://github.com/$REPO/releases/download/$VERSION/portalchaind-linux-amd64 \
  -O /tmp/portalchaind
chmod +x /tmp/portalchaind
sudo mv /tmp/portalchaind /usr/local/bin/portalchaind
echo "portalchaind version: $(portalchaind version)"

# 3. Initialize node
echo "[3/5] Initializing node..."
if [ ! -f ~/.portalchain/config/genesis.json ]; then
    portalchaind init "$MONIKER" --chain-id $CHAIN_ID
    BOOTSTRAP_PEER="fab93ae9dce6f9413ab64eee95f5c65272c789b0@72.56.114.142:26656"
    sed -i "s/persistent_peers = \"\"/persistent_peers = \"$BOOTSTRAP_PEER\"/" ~/.portalchain/config/config.toml
    echo "Downloading genesis.json from bootstrap node..."
    curl -s http://72.56.114.142:26657/genesis | jq '.result.genesis' > ~/.portalchain/config/genesis.json
    echo "✅ Genesis downloaded"
fi

# 4. Configure seeds
echo "[4/5] Configuring..."
if [ -n "$SEEDS" ]; then
    sed -i "s/seeds = \"\"/seeds = \"$SEEDS\"/" ~/.portalchain/config/config.toml
fi

# 5. Create systemd service
echo "[5/5] Creating systemd service..."
sudo tee /etc/systemd/system/portalchain.service > /dev/null << SERVICE
[Unit]
Description=PortalChain Validator Node
After=network-online.target

[Service]
User=$USER
ExecStart=/usr/local/bin/portalchaind start
Restart=always
RestartSec=3
LimitNOFILE=65535
Environment="HOME=$HOME"

[Install]
WantedBy=multi-user.target
SERVICE

sudo systemctl daemon-reload
sudo systemctl enable portalchain

echo ""
echo "✅ Validator installed!"
echo ""
echo "Next steps:"
echo "  1. Create key:     portalchaind keys add validator --keyring-backend test"
echo "  2. Get DAAI:       /faucet <address> in @daai_portal_bot"
echo "  3. Start node:     sudo systemctl start portalchain"
echo "  4. Check logs:     sudo journalctl -fu portalchain"
