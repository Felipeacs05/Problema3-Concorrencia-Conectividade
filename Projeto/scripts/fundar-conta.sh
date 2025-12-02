#!/bin/bash
# ===================== FUNDAR CONTA =====================
# Script para transferir ETH para uma conta existente

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$SCRIPT_DIR/.."
BLOCKCHAIN_DIR="$PROJECT_DIR/Blockchain"
TOOLS_DIR="$BLOCKCHAIN_DIR/tools"

echo "========================================"
echo "Transferir ETH para Conta"
echo "========================================"
echo

if [ -z "$1" ]; then
    echo "Uso: ./fundar-conta.sh <endereco_da_conta> [quantidade_em_ETH]"
    echo
    echo "Exemplo:"
    echo "  ./fundar-conta.sh 0x0a8b6dd6F3A38D8Cc5Dbb872A83F1a24B7e49E1f"
    echo "  ./fundar-conta.sh 0x0a8b6dd6F3A38D8Cc5Dbb872A83F1a24B7e49E1f 100"
    echo
    echo "Se a quantidade não for especificada, será transferido 100 ETH."
    echo
    exit 1
fi

ENDERECO="$1"
QUANTIDADE="${2:-100}"

# Verifica se fund-account existe
if [ ! -f "$TOOLS_DIR/fund-account" ] && [ ! -f "$TOOLS_DIR/fund-account.exe" ]; then
    echo "[AVISO] fund-account não encontrado, compilando..."
    cd "$TOOLS_DIR"
    go build -o fund-account fund-account.go
    if [ $? -ne 0 ]; then
        echo "[ERRO] Falha ao compilar fund-account.go"
        exit 1
    fi
fi

echo "Transferindo $QUANTIDADE ETH para: $ENDERECO"
echo

cd "$TOOLS_DIR"
if [ -f "fund-account" ]; then
    ./fund-account "$ENDERECO" "$QUANTIDADE"
elif [ -f "fund-account.exe" ]; then
    ./fund-account.exe "$ENDERECO" "$QUANTIDADE"
else
    echo "[ERRO] fund-account não encontrado após compilação!"
    exit 1
fi

if [ $? -eq 0 ]; then
    echo
    echo "========================================"
    echo "Transferência concluída com sucesso!"
    echo "========================================"
else
    echo
    echo "[ERRO] Falha na transferência!"
    echo "Verifique se:"
    echo "- O Geth está rodando (docker ps)"
    echo "- O endereço está correto"
    echo "- A conta do servidor tem ETH suficiente"
    exit 1
fi


