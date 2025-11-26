#!/bin/bash
# ===================== STOP ALL =====================
# Script para parar toda a infraestrutura

set -e  # Para em caso de erro

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BLOCKCHAIN_DIR="$PROJECT_DIR/Blockchain"
JOGO_DIR="$PROJECT_DIR/Jogo"

echo "========================================"
echo "Parando Infraestrutura"
echo "========================================"
echo ""

# Para jogo
echo "[1/2] Parando jogo..."
cd "$JOGO_DIR"
docker-compose down
echo "[OK] Jogo parado"
echo ""

# Para blockchain
echo "[2/2] Parando blockchain..."
cd "$BLOCKCHAIN_DIR"
docker-compose -f docker-compose-blockchain.yml down
echo "[OK] Blockchain parada"
echo ""

echo "========================================"
echo "Infraestrutura parada com sucesso!"
echo "========================================"
echo ""
read -p "Pressione Enter para continuar..."

