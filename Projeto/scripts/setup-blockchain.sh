#!/bin/bash
# ===================== SETUP BLOCKCHAIN =====================
# Script para configurar a blockchain privada Ethereum
# Este script deve ser executado antes de iniciar o jogo

set -e  # Para em caso de erro

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BLOCKCHAIN_DIR="$PROJECT_DIR/Blockchain"
TOOLS_DIR="$BLOCKCHAIN_DIR/tools"
DATA_DIR="$BLOCKCHAIN_DIR/data"
KEYSTORE_DIR="$DATA_DIR/keystore"
GENESIS_FILE="$BLOCKCHAIN_DIR/genesis.json"

# Detecta qual comando docker compose usar (novo: "docker compose", antigo: "docker-compose")
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi

echo "========================================"
echo "Configurando Blockchain Privada"
echo "========================================"
echo ""

# Verifica se Go está instalado
if ! command -v go &> /dev/null; then
    echo "ERRO: Go não está instalado!"
    echo "Instale Go: https://golang.org/dl/"
    exit 1
fi

# Compila o utilitário blockchain-utils
echo "[1/7] Compilando utilitário blockchain-utils..."
cd "$TOOLS_DIR"
go mod tidy
go build -o blockchain-utils blockchain-utils.go
if [ ! -f "$TOOLS_DIR/blockchain-utils" ]; then
    echo "ERRO: Falha ao compilar blockchain-utils"
    exit 1
fi
echo "[OK] Utilitário compilado"
echo ""

# Para containers se estiverem rodando
echo "[2/7] Parando containers blockchain..."
cd "$BLOCKCHAIN_DIR"
$DOCKER_COMPOSE -f docker-compose-blockchain.yml down 2>/dev/null || true
echo "[OK] Containers parados"
echo ""

# Remove dados antigos
echo "[3/7] Removendo dados antigos..."
REMOVIDO=0

if [ -d "$DATA_DIR/geth" ]; then
    rm -rf "$DATA_DIR/geth"
    REMOVIDO=1
fi

if [ -d "$KEYSTORE_DIR" ]; then
    rm -rf "$KEYSTORE_DIR"
    REMOVIDO=1
fi

if [ -f "$DATA_DIR/password.txt" ]; then
    rm -f "$DATA_DIR/password.txt"
    REMOVIDO=1
fi

if [ $REMOVIDO -eq 1 ]; then
    echo "[OK] Dados antigos removidos"
else
    echo "[OK] Nenhum dado antigo encontrado"
fi

# Cria diretórios necessários
mkdir -p "$KEYSTORE_DIR"
mkdir -p "$DATA_DIR"
echo ""

# Cria conta
echo "[4/7] Criando nova conta..."
"$TOOLS_DIR/blockchain-utils" criar-conta "$KEYSTORE_DIR" "123456"
if [ $? -ne 0 ]; then
    echo "ERRO: Falha ao criar conta"
    exit 1
fi

# Cria arquivo password.txt
echo "123456" > "$DATA_DIR/password.txt"
echo ""

# Gera genesis.json
echo "[5/7] Gerando genesis.json..."
"$TOOLS_DIR/blockchain-utils" gerar-genesis "$KEYSTORE_DIR" "$GENESIS_FILE"
if [ $? -ne 0 ]; then
    echo "ERRO: Falha ao gerar genesis.json"
    exit 1
fi
echo "[OK] Genesis.json gerado"
echo ""

# Extrai endereço da conta criada
echo "[6/8] Extraindo endereco da conta..."
ADDRESS=$("$TOOLS_DIR/blockchain-utils" extrair-endereco "$KEYSTORE_DIR")
if [ -z "$ADDRESS" ]; then
    echo "ERRO: Falha ao extrair endereco"
    exit 1
fi
echo "[OK] Endereco extraido: $ADDRESS"
echo ""

# Atualiza docker-compose.yml com endereço da conta
echo "[7/9] Atualizando docker-compose.yml..."
"$TOOLS_DIR/blockchain-utils" atualizar-docker-compose "$ADDRESS" "$BLOCKCHAIN_DIR/docker-compose-blockchain.yml"
if [ $? -ne 0 ]; then
    echo "ERRO: Falha ao atualizar docker-compose.yml"
    exit 1
fi
echo "[OK] Docker-compose.yml atualizado"
echo ""

# Inicializa blockchain (geth init)
echo "[8/9] Inicializando blockchain com genesis.json..."
cd "$BLOCKCHAIN_DIR"
if [ -d "$DATA_DIR/geth" ]; then
    echo "Removendo dados antigos da blockchain..."
    rm -rf "$DATA_DIR/geth"
fi
docker run --rm -v "$DATA_DIR:/root/.ethereum" -v "$GENESIS_FILE:/genesis.json" ethereum/client-go:v1.13.15 --datadir=/root/.ethereum init /genesis.json
if [ $? -ne 0 ]; then
    echo "ERRO: Falha ao inicializar blockchain"
    exit 1
fi
echo "[OK] Blockchain inicializada"
echo ""

# Inicia containers
echo "[9/10] Iniciando containers blockchain..."
$DOCKER_COMPOSE -f docker-compose-blockchain.yml up -d
if [ $? -ne 0 ]; then
    echo "ERRO: Falha ao iniciar containers"
    exit 1
fi
echo "[OK] Containers iniciados"
echo ""

# Aguarda Geth estar pronto
echo "Aguardando Geth estar pronto..."
while true; do
    sleep 2
    if docker exec geth-node geth attach --exec "eth.blockNumber" http://localhost:8545 >/dev/null 2>&1; then
        break
    fi
    echo "Aguardando porta RPC 8545..."
done
echo "[OK] Geth está pronto"
echo ""

# Desbloqueia conta
echo "Desbloqueando conta..."
docker exec geth-node geth attach --exec "personal.unlockAccount(eth.accounts[0], '123456', 0)" http://localhost:8545 >/dev/null 2>&1 || true
echo "[OK] Conta desbloqueada"
echo ""

# Faz deploy do contrato
echo "[10/10] Fazendo deploy do contrato..."
cd "$BLOCKCHAIN_DIR/scripts"

# Verifica se o contrato foi compilado
if [ ! -f "$BLOCKCHAIN_DIR/contracts/GameEconomy.bin" ]; then
    echo "[AVISO] Contrato não compilado. Compilando..."
    chmod +x compile-contract.sh 2>/dev/null || true
    ./compile-contract.sh
    if [ $? -ne 0 ]; then
        echo "[AVISO] Falha ao compilar contrato. Execute compile-contract.sh manualmente."
        echo ""
    fi
fi

# Executa deploy
if [ -f "deploy-contract.sh" ]; then
    chmod +x deploy-contract.sh 2>/dev/null || true
    ./deploy-contract.sh
    if [ $? -eq 0 ]; then
        echo "[OK] Contrato deployado com sucesso"
    else
        echo "[AVISO] Deploy pode ter falhado. Execute deploy-contract.sh manualmente."
    fi
else
    echo "[AVISO] Script de deploy não encontrado. Execute deploy-contract.sh manualmente."
fi
echo ""

echo "========================================"
echo "Blockchain configurada com sucesso!"
echo "========================================"
echo ""
echo "Próximos passos:"
echo "1. Execute setup-game.sh para configurar o jogo"
echo "2. Execute start-all.sh para iniciar tudo"
echo ""
read -p "Pressione Enter para continuar..."

