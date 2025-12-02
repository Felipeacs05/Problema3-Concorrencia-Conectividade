#!/bin/bash
# ===================== VIEW TRANSACTIONS =====================
# Script para visualizar transações de uma conta na blockchain

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
TOOLS_DIR="$PROJECT_DIR/Blockchain/tools"

cd "$TOOLS_DIR"

if [ -z "$1" ]; then
    echo "Uso: view-transactions.sh <endereco_da_conta> [bloco_inicial] [bloco_final]"
    echo ""
    echo "Exemplos:"
    echo "  ./view-transactions.sh 0x2d50FD74Cc3eB554b16013667045804D28Bc54a4"
    echo "  ./view-transactions.sh 0x2d50FD74Cc3eB554b16013667045804D28Bc54a4 0 1000"
    echo ""
    exit 1
fi

# Compila se necessário
if [ ! -f "view-transactions" ]; then
    echo "Compilando view-transactions..."
    go build -o view-transactions view-transactions.go
    if [ $? -ne 0 ]; then
        echo "[ERRO] Falha ao compilar view-transactions.go"
        exit 1
    fi
fi

# Executa
./view-transactions "$@"

