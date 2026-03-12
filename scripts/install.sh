#!/bin/bash
set -e

echo "╔═══════════════════════════════════════╗"
echo "║     PortalChain Node Installer        ║"
echo "║     Testnet v0.1                      ║"
echo "╚═══════════════════════════════════════╝"
echo ""
echo "Select your role:"
echo "  1) Validator     — consensus node (2CPU, 4GB RAM, 50GB SSD)"
echo "  2) AI Operator   — inference node (8CPU, 16GB RAM, GPU optional)"
echo "  3) Full Node     — validator + operator (max rewards)"
echo ""
read -p "Enter choice [1-3]: " ROLE

case $ROLE in
  1) bash "$(dirname "$0")/install-validator.sh" ;;
  2) bash "$(dirname "$0")/install-operator.sh" ;;
  3) bash "$(dirname "$0")/install-full.sh" ;;
  *) echo "Invalid choice"; exit 1 ;;
esac
