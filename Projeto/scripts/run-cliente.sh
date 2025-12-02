#!/bin/bash
# ===================== RUN CLIENTE =====================
# Script para executar o cliente do jogo (Linux/Mac)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
JOGO_DIR="$PROJECT_DIR/Jogo"
CLIENTE_DIR="$JOGO_DIR/cliente"

echo "========================================"
echo "Executando Cliente do Jogo"
echo "========================================"
echo ""

# Verifica se o executável existe
if [ -f "$CLIENTE_DIR/cliente" ]; then
    cd "$CLIENTE_DIR"
    echo "Executando cliente..."
    echo ""
    ./cliente
elif [ -f "$CLIENTE_DIR/cliente.exe" ]; then
    # Se estiver no WSL ou similar
    cd "$CLIENTE_DIR"
    echo "Executando cliente (via wine ou similar)..."
    echo ""
    ./cliente.exe
else
    echo "ERRO: cliente nao encontrado!"
    echo "Execute setup-game.sh primeiro para compilar o cliente."
    echo ""
    exit 1
fi



