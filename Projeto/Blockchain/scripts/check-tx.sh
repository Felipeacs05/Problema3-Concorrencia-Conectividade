#!/bin/bash
# ===================== CHECK TRANSACTION =====================
# Script para verificar eventos de uma transação específica

# Obtém o diretório onde o script está
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TOOLS_DIR="$SCRIPT_DIR/../tools"

cd "$TOOLS_DIR" || exit 1

if [ -z "$1" ]; then
    echo "Uso: ./check-tx.sh <hash_da_transacao>"
    echo ""
    echo "Exemplo:"
    echo "  ./check-tx.sh 0x80faebe1ac84d2c4fb86ac26d84987436fc2599ed057d99b6628c5f9d5c2f51e"
    echo ""
    exit 1
fi

# Compila se necessário
if [ ! -f "check-tx" ]; then
    echo "Compilando check-tx.go..."
    go build -o check-tx check-tx.go
    if [ $? -ne 0 ]; then
        echo "[ERRO] Falha ao compilar check-tx.go"
        exit 1
    fi
fi

# Executa
./check-tx "$@"

