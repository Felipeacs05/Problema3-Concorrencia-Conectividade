#!/bin/bash

# ===================== REBUILD AND START =====================
# Este script reconstrói as imagens Docker e inicia a infraestrutura do jogo e blockchain.

# CORREÇÃO DE CAMINHOS: Usa caminhos absolutos para evitar erro de "diretório não encontrado"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
JOGO_DIR="$PROJECT_DIR/Jogo"
BLOCKCHAIN_DIR="$PROJECT_DIR/Blockchain"

# --- Função de Verificação de Erro ---
check_error() {
    if [ $? -ne 0 ]; then
        echo "ERRO: $1" >&2
        exit 1
    fi
}

# --- INÍCIO DO SCRIPT ---
echo "========================================"
echo " Reconstruindo e Iniciando Infraestrutura"
echo "========================================"
echo ""
echo "Este processo vai:"
echo "1. Parar todos os containers"
echo "2. Reconstruir imagens Docker (servidores e cliente) - SEM CACHE"
echo "3. Inicializar blockchain (se dados não existirem)"
echo "4. Iniciar blockchain (Geth)"
echo "5. Iniciar jogo (Servidores e Brokers)"
echo ""
read -n 1 -s -r -p "Pressione qualquer tecla para continuar..."
echo ""

# 1. Verifica se Docker está rodando
if ! docker info > /dev/null 2>&1; then
    echo "ERRO: Docker não está rodando!" >&2
    echo "Inicie o Docker Desktop e tente novamente."
    exit 1
fi

# 2. Parar tudo primeiro
echo "[1/5] Parando containers existentes..."

# Parar Jogo
if [ -d "$JOGO_DIR" ]; then
    (cd "$JOGO_DIR" && docker-compose down > /dev/null 2>&1)
    check_error "Falha ao parar containers do Jogo"
fi

# Parar Blockchain
if [ -d "$BLOCKCHAIN_DIR" ]; then
    (cd "$BLOCKCHAIN_DIR" && docker-compose -f docker-compose-blockchain.yml down > /dev/null 2>&1)
    check_error "Falha ao parar containers da Blockchain"
fi
echo "[OK] Containers parados"
echo ""

# 3. Reconstrói as imagens do jogo (servidores e cliente)
echo "[2/5] Reconstruindo imagens Docker do jogo (servidores e cliente)..."
echo "      Isso pode demorar vários minutos, aguarde..."
if [ -d "$JOGO_DIR" ]; then
    (cd "$JOGO_DIR" && docker-compose build --no-cache)
    check_error "Falha ao reconstruir imagens do jogo"
fi
echo "[OK] Imagens do jogo reconstruídas"
echo ""

# 4. Inicia blockchain
echo "[3/5] Inicializando blockchain..."
cd "$BLOCKCHAIN_DIR"
check_error "Diretório da Blockchain não encontrado: $BLOCKCHAIN_DIR"

DATA_DIR="$PWD/data"
GENESIS_FILE="$PWD/genesis.json"
GETH_IMAGE="ethereum/client-go:v1.13.15"

# Inicializa se chaindata não existe
if [ ! -d "$DATA_DIR/geth" ]; then
    echo "[INFO] Chaindata não encontrado, inicializando blockchain com genesis.json..."
    
    # Executa o init com volumes mapeados
    docker run --rm -v "$DATA_DIR:/root/.ethereum" -v "$GENESIS_FILE:/genesis.json" "$GETH_IMAGE" --datadir=/root/.ethereum init /genesis.json
    check_error "Falha ao inicializar blockchain"
    echo "[OK] Blockchain inicializada"
    echo ""
fi

# Inicia container blockchain
docker-compose -f docker-compose-blockchain.yml up -d
check_error "Falha ao iniciar blockchain"
echo "[OK] Container blockchain iniciado (geth-node)"
echo ""

# 5. Aguarda blockchain estar pronta
echo "[4/5] Aguardando blockchain estar pronta (8 segundos)..."
sleep 8

# Verifica se o container Geth está rodando
if ! docker ps | grep -q "geth-node"; then
    echo "[AVISO] Container Geth pode não estar rodando. Continuando..."
fi

echo "[OK] Blockchain pronta (ou iniciando em background)"
echo ""

# 6. Inicia jogo
echo "[5/5] Iniciando jogo (servidores e brokers)..."
cd "$JOGO_DIR"
check_error "Diretório do Jogo não encontrado: $JOGO_DIR"

docker-compose up -d
check_error "Falha ao iniciar jogo"
echo "[OK] Jogo iniciado"
echo ""

# --- FIM DO SCRIPT ---
echo "========================================"
echo " Reconstrução e Inicialização Concluídas!"
echo "========================================"
echo ""
echo "Serviços disponíveis (Verifique logs se houver erro!):"
echo "- Blockchain: http://localhost:8545 (ou host.docker.internal:8545)"
echo "- Servidor 1: http://localhost:8080"
echo "- Servidor 2: http://localhost:8081"
echo "- Servidor 3: http://localhost:8082"
echo "- Broker MQTT 1: tcp://localhost:1886"
echo ""
echo "Para parar tudo, execute: $PROJECT_DIR/scripts/stop-all.sh"
echo ""