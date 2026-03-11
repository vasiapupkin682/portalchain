#!/bin/bash
set -e
MONIKER="mynode"
CHAIN_ID="portalchain"
DENOM="daai"

portalchaind tendermint unsafe-reset-all
portalchaind init $MONIKER --chain-id $CHAIN_ID --overwrite
rm -rf ~/.portalchain/config/gentx/

sed -i "s/\"stake\"/\"$DENOM\"/g" ~/.portalchain/config/genesis.json

portalchaind keys show alice --keyring-backend test > /dev/null 2>&1 || \
  portalchaind keys add alice --keyring-backend test

portalchaind add-genesis-account alice 200000000$DENOM --keyring-backend test
portalchaind gentx alice 100000000$DENOM \
  --chain-id $CHAIN_ID \
  --moniker $MONIKER \
  --keyring-backend test
portalchaind collect-gentxs

echo "✅ Genesis initialized with $DENOM token"
