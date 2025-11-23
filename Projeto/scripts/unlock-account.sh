#!/bin/bash
# ===================== BAREMA ITEM 1: ARQUITETURA =====================
# Script para desbloquear conta no Clique (Linux/macOS)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "========================================"
echo "Desbloqueando conta do signer..."
echo "========================================"
echo ""

# Desbloqueia a conta usando senha padrão
(
cat <<EOF
var accounts = eth.accounts;
if (accounts.length == 0) {
  console.log("ERRO: Nenhuma conta encontrada!");
  exit;
}
var account = accounts[0];
console.log("Conta: " + account);
var result = personal.unlockAccount(account, "123456", 0);
if (result) {
  console.log("SUCCESS: Conta desbloqueada!");
} else {
  console.log("ERRO: Falha ao desbloquear conta!");
}
exit
EOF
) | docker exec -i geth-node geth attach http://localhost:8545

echo ""
echo "========================================"
echo "Clique deve começar a selar blocos automaticamente"
echo "========================================"
echo ""


