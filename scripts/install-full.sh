#!/bin/bash
set -e

REPO_URL=${REPO_URL:-"https://github.com/vasiapupkin682/portalchain.git"}
PORTALCHAIN_DIR=${PORTALCHAIN_DIR:-"$HOME/portalchain"}

echo "=== Installing PortalChain Full Node (Validator + Operator) ==="

# Ensure repo exists for operator
if [ ! -d "$PORTALCHAIN_DIR" ]; then
    echo "Cloning portalchain repo to $PORTALCHAIN_DIR..."
    git clone "$REPO_URL" "$PORTALCHAIN_DIR"
fi

export PORTALCHAIN_DIR

# Validator builds from PORTALCHAIN_DIR when set
bash "$(dirname "$0")/install-validator.sh"
bash "$(dirname "$0")/install-operator.sh"

echo ""
echo "✅ Full node installed!"
