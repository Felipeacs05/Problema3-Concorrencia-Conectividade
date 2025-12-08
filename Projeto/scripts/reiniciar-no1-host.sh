#!/bin/bash
# ===================== REINICIAR NÓ 1 COM HOST MODE =====================
# Script para reiniciar o Nó 1 com network_mode: host

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BLOCKCHAIN_DIR="$PROJECT_DIR/Blockchain"

# Detecta docker compose
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi

echo "========================================"
echo "Reiniciando Nó 1 com Host Mode"
echo "========================================"
echo ""

# Para o Nó 1
echo "[1/3] Parando Nó 1..."
cd "$BLOCKCHAIN_DIR"
$DOCKER_COMPOSE -f docker-compose-blockchain.yml down
echo "[OK] Nó 1 parado"
echo ""

# Aguarda um pouco
sleep 2

# Inicia o Nó 1 com a nova configuração
echo "[2/3] Iniciando Nó 1 com network_mode: host..."
$DOCKER_COMPOSE -f docker-compose-blockchain.yml up -d
echo "[OK] Nó 1 iniciado"
echo ""

# Aguarda o Nó 1 estar pronto
echo "[3/3] Aguardando Nó 1 estar pronto..."
for i in {1..30}; do
    if docker exec geth-node geth attach --exec "eth.blockNumber" /root/.ethereum/geth.ipc >/dev/null 2>&1; then
        echo "[OK] Nó 1 está pronto"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "⚠️  Nó 1 pode não estar totalmente pronto ainda"
    fi
    sleep 2
done
echo ""

echo "========================================"
echo "✅ Nó 1 reiniciado com sucesso!"
echo "========================================"
echo ""
echo "Agora execute:"
echo "  bash connect-peer.sh"
echo "  bash verificar-nos.sh"
echo ""
