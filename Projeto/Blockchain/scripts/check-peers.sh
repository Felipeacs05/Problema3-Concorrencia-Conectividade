#!/bin/bash
# ===================== BAREMA ITEM 1: ARQUITETURA =====================
# Script para verificar conexões P2P com outros nós (Linux/macOS)

echo "Verificando conexões P2P..."
echo ""

echo "admin.peers" | docker exec -i geth-node geth attach http://localhost:8545

echo ""
echo "Número de peers conectados:"
echo "net.peerCount" | docker exec -i geth-node geth attach http://localhost:8545

echo ""


