#!/bin/bash
# ===================== VIEW EVENTS =====================
# Script para visualizar eventos do contrato GameEconomy na blockchain

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TOOLS_DIR="$SCRIPT_DIR/../tools"

cd "$TOOLS_DIR"

if [ -z "$1" ]; then
    echo "Uso: view-events.sh <endereco_do_contrato> [bloco_inicial] [bloco_final]"
    echo ""
    echo "Exemplos:"
    echo "  ./view-events.sh 0x4D82F25Ef5058fE2308135D88A386c90FEdBe99A"
    echo "  ./view-events.sh 0x4D82F25Ef5058fE2308135D88A386c90FEdBe99A 0 1000"
    echo ""
    echo "Para obter o endereco do contrato, veja o arquivo:"
    echo "  Blockchain/contract-address.txt"
    echo ""
    exit 1
fi

# Compila se necessário
if [ ! -f "view-events" ]; then
    echo "Compilando view-events..."
    go build -o view-events view-events.go
    if [ $? -ne 0 ]; then
        echo "[ERRO] Falha ao compilar view-events.go"
        exit 1
    fi
fi

# Executa
./view-events "$@"

