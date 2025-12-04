#!/bin/bash
# ===================== BAREMA ITEM 2: SMART CONTRACTS =====================
# Script para compilar o contrato Solidity usando Docker

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
CONTRACTS_DIR="$PROJECT_DIR/contracts"
CONTRACT_FILE="$CONTRACTS_DIR/GameEconomy.sol"

echo "========================================"
echo "Compilando Contrato Solidity"
echo "========================================"
echo ""

# Verifica se o arquivo do contrato existe
if [ ! -f "$CONTRACT_FILE" ]; then
    echo "[ERRO] Arquivo do contrato não encontrado: $CONTRACT_FILE"
    exit 1
fi

echo "[1/3] Verificando Docker..."
if ! docker --version > /dev/null 2>&1; then
    echo "[ERRO] Docker não está instalado ou não está rodando!"
    exit 1
fi
echo "[OK] Docker encontrado"
echo ""

echo "[2/3] Compilando GameEconomy.sol..."
echo ""

# Compila o contrato usando Docker (gera .bin e .abi)
docker run --rm \
    -v "$CONTRACTS_DIR:/contracts" \
    -w /contracts \
    ethereum/solc:0.8.20 \
    --bin --abi \
    --evm-version paris \
    --optimize --optimize-runs 200 \
    -o /contracts \
    --overwrite \
    GameEconomy.sol

echo ""
echo "[3/3] Verificando arquivos gerados..."

# Aguarda um pouco para garantir que os arquivos foram escritos
sleep 2

# Verifica se os arquivos foram criados
BIN_OK=0
ABI_OK=0

if [ -f "$CONTRACTS_DIR/GameEconomy.bin" ]; then
    echo "[OK] Bytecode gerado: contracts/GameEconomy.bin"
    BIN_OK=1
fi

if [ -f "$CONTRACTS_DIR/GameEconomy.abi" ]; then
    echo "[OK] ABI gerado: contracts/GameEconomy.abi"
    ABI_OK=1
fi

if [ $BIN_OK -eq 1 ] && [ $ABI_OK -eq 1 ]; then
    echo ""
    echo "========================================"
    echo "Compilação concluída com sucesso!"
    echo "========================================"
    echo ""
    echo "Próximo passo: Fazer deploy do contrato"
    echo "  - Execute o cliente Go"
    echo "  - Escolha opção 6 (Deploy do Contrato)"
    echo ""
    exit 0
else
    echo ""
    echo "[ERRO] Alguns arquivos não foram gerados!"
    if [ $BIN_OK -eq 0 ]; then
        echo "  - GameEconomy.bin: FALTANDO"
    fi
    if [ $ABI_OK -eq 0 ]; then
        echo "  - GameEconomy.abi: FALTANDO"
    fi
    echo ""
    echo "Verifique os erros de compilação acima."
    exit 1
fi




