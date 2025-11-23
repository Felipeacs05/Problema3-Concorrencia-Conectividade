#!/bin/bash
# ===================== BAREMA ITEM 1: ARQUITETURA =====================
# Script cross-platform para configurar a blockchain (Linux/macOS)
# Funciona em conjunto com o utilitário Go blockchain-utils

set -e  # Para na primeira erro

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
TOOLS_DIR="$PROJECT_DIR/tools"
DATA_DIR="$PROJECT_DIR/data"
KEYSTORE_DIR="$DATA_DIR/keystore"
GENESIS_FILE="$PROJECT_DIR/genesis.json"

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
echo "[1/5] Compilando utilitário blockchain-utils..."
cd "$TOOLS_DIR"
go mod tidy
go build -o blockchain-utils blockchain-utils.go
if [ ! -f "$TOOLS_DIR/blockchain-utils" ]; then
    echo "ERRO: Falha ao compilar blockchain-utils"
    exit 1
fi
echo "✓ Utilitário compilado"
echo ""

# Para containers se estiverem rodando
echo "[2/5] Parando containers..."
cd "$PROJECT_DIR"
docker-compose down 2>/dev/null || true
echo "✓ Containers parados"
echo ""

# Remove dados antigos
echo "[3/5] Removendo dados antigos..."
if [ -d "$DATA_DIR/geth" ]; then
    rm -rf "$DATA_DIR/geth"
    echo "✓ Dados antigos removidos"
else
    echo "✓ Nenhum dado antigo encontrado"
fi

# Cria diretórios necessários
mkdir -p "$KEYSTORE_DIR"
echo ""

# Cria conta
echo "[4/5] Criando nova conta..."
"$TOOLS_DIR/blockchain-utils" criar-conta "$KEYSTORE_DIR" "123456"
if [ $? -ne 0 ]; then
    echo "ERRO: Falha ao criar conta"
    exit 1
fi
echo ""

# Gera genesis.json
echo "[5/5] Gerando genesis.json..."
"$TOOLS_DIR/blockchain-utils" gerar-genesis "$KEYSTORE_DIR" "$GENESIS_FILE"
if [ $? -ne 0 ]; then
    echo "ERRO: Falha ao gerar genesis.json"
    exit 1
fi
echo ""

# Inicializa blockchain
echo "Inicializando blockchain..."
docker-compose run --rm geth --datadir /root/.ethereum init /genesis.json
if [ $? -ne 0 ]; then
    echo "ERRO: Falha ao inicializar blockchain"
    exit 1
fi
echo ""

# Inicia nó
echo "Iniciando nó Geth..."
docker-compose up -d geth
echo "Aguardando inicialização..."
sleep 10
echo ""

echo "========================================"
echo "Configuração concluída!"
echo "========================================"
echo ""
echo "Próximos passos:"
echo "  1. Desbloquear conta: ./scripts/unlock-account.sh"
echo "  2. Verificar blocos: ./scripts/check-block.sh"
echo ""


