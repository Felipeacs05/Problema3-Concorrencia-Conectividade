#!/bin/bash
# ===================== START ALL =====================
# Script para iniciar toda a infraestrutura (blockchain + jogo)

set -e  # Para em caso de erro

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BLOCKCHAIN_DIR="$PROJECT_DIR/Blockchain"
JOGO_DIR="$PROJECT_DIR/Jogo"

echo "========================================"
echo "Iniciando Infraestrutura Completa"
echo "========================================"
echo ""

# Verifica se Docker está rodando
if ! docker ps >/dev/null 2>&1; then
    echo "ERRO: Docker não está rodando!"
    echo "Inicie o Docker e tente novamente."
    exit 1
fi

# Inicia blockchain
echo "[1/2] Iniciando blockchain..."
cd "$BLOCKCHAIN_DIR"
docker-compose -f docker-compose-blockchain.yml up -d
if [ $? -ne 0 ]; then
    echo "ERRO: Falha ao iniciar blockchain"
    exit 1
fi
echo "[OK] Blockchain iniciada"
echo ""

# Aguarda blockchain estar pronta
echo "Aguardando blockchain estar pronta..."
while true; do
    sleep 3
    if docker exec geth-node geth attach --exec "eth.blockNumber" http://localhost:8545 >/dev/null 2>&1; then
        break
    fi
    echo "Aguardando blockchain..."
done
echo "[OK] Blockchain pronta"
echo ""

# Inicia jogo
echo "[2/2] Iniciando jogo..."
cd "$JOGO_DIR"
docker-compose up -d
if [ $? -ne 0 ]; then
    echo "ERRO: Falha ao iniciar jogo"
    exit 1
fi
echo "[OK] Jogo iniciado"
echo ""

echo "========================================"
echo "Infraestrutura iniciada com sucesso!"
echo "========================================"
echo ""
echo "Serviços disponíveis:"
echo "- Blockchain: http://localhost:8545"
echo "- Servidor 1: http://localhost:8080"
echo "- Servidor 2: http://localhost:8081"
echo "- Servidor 3: http://localhost:8082"
echo "- Broker MQTT 1: tcp://localhost:1886"
echo "- Broker MQTT 2: tcp://localhost:1884"
echo "- Broker MQTT 3: tcp://localhost:1885"
echo ""
echo "Para parar tudo, execute: stop-all.sh"
echo ""
read -p "Pressione Enter para continuar..."

