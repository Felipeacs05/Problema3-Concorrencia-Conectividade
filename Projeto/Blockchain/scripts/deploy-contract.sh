#!/bin/bash
# ===================== DEPLOY DO CONTRATO =====================
# Script para fazer deploy do contrato GameEconomy na blockchain

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TOOLS_DIR="$SCRIPT_DIR/../tools"

echo "========================================"
echo "Deploy do Contrato GameEconomy"
echo "========================================"
echo

# Verifica se o Geth está rodando
echo "[1/3] Verificando se o Geth está rodando..."
if ! docker ps | grep -q "geth-node"; then
    echo "[ERRO] Geth não está rodando!"
    echo "Execute: docker compose up -d (ou docker-compose up -d)"
    exit 1
fi
echo "[OK] Geth está rodando"
echo

# Verifica se o bytecode existe
echo "[2/3] Verificando se o contrato foi compilado..."
if [ ! -f "$SCRIPT_DIR/../contracts/GameEconomy.bin" ]; then
    echo "[ERRO] Contrato não compilado!"
    echo "Execute: ./compile-contract.sh primeiro"
    exit 1
fi
echo "[OK] Bytecode encontrado"
echo

# Executa o deploy
echo "[3/3] Executando deploy..."
cd "$TOOLS_DIR"
if [ ! -f "deploy-contract" ] && [ ! -f "deploy-contract.exe" ]; then
    echo "[AVISO] deploy-contract não encontrado!"
    echo "Compilando..."
    go build -o deploy-contract deploy-contract.go
    if [ $? -ne 0 ]; then
        echo "[ERRO] Falha ao compilar deploy-contract.go"
        exit 1
    fi
fi

if [ -f "deploy-contract" ]; then
    ./deploy-contract
elif [ -f "deploy-contract.exe" ]; then
    ./deploy-contract.exe
else
    echo "[ERRO] deploy-contract não encontrado após compilação!"
    exit 1
fi

if [ $? -eq 0 ]; then
    echo
    echo "========================================"
    echo "Deploy concluído com sucesso!"
    echo "========================================"
    echo
    echo "Agora você pode usar o cliente do jogo!"
else
    echo
    echo "[ERRO] Deploy falhou!"
    echo "Verifique os erros acima."
    exit 1
fi
