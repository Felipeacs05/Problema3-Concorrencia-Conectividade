#!/bin/bash
# ===================== SETUP GAME =====================
# Script para configurar o jogo distribuído
# Este script deve ser executado após setup-blockchain.sh

set -e  # Para em caso de erro

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
JOGO_DIR="$PROJECT_DIR/Jogo"

echo "========================================"
echo "Configurando Jogo Distribuído"
echo "========================================"
echo ""

# Verifica se Go está instalado
if ! command -v go &> /dev/null; then
    echo "ERRO: Go não está instalado!"
    echo "Instale Go: https://golang.org/dl/"
    exit 1
fi

# Compila servidor
echo "[1/3] Compilando servidor..."
cd "$JOGO_DIR/servidor"
go mod tidy
go build -o servidor main.go
if [ $? -ne 0 ]; then
    echo "ERRO: Falha ao compilar servidor"
    exit 1
fi
echo "[OK] Servidor compilado"
echo ""

# Compila cliente
echo "[2/3] Compilando cliente..."
cd "$JOGO_DIR/cliente"
go mod tidy
go build -o cliente main.go blockchain_client.go
if [ $? -ne 0 ]; then
    echo "ERRO: Falha ao compilar cliente"
    exit 1
fi
echo "[OK] Cliente compilado"
echo ""

# Verifica se o contrato foi deployado
echo "[3/3] Verificando contrato blockchain..."
if [ -f "$PROJECT_DIR/contract-address.txt" ]; then
    echo "[OK] Contrato encontrado"
    cat "$PROJECT_DIR/contract-address.txt"
else
    echo "[AVISO] Contrato não encontrado. Execute setup-blockchain.sh primeiro."
fi
echo ""

echo "========================================"
echo "Jogo configurado com sucesso!"
echo "========================================"
echo ""
echo "Próximos passos:"
echo "1. Execute start-all.sh para iniciar tudo"
echo ""
read -p "Pressione Enter para continuar..."

