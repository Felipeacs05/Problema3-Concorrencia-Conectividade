#!/bin/bash
# ===================== BAREMA ITEM 1: ARQUITETURA =====================
# Script para verificar conexões P2P com outros nós

echo "Verificando conexões P2P..."

# BAREMA ITEM 1: ARQUITETURA - Lista todos os peers conectados
docker exec -it geth-node geth attach http://localhost:8545 --exec "admin.peers"

echo ""
echo "Número de peers conectados:"
docker exec -it geth-node geth attach http://localhost:8545 --exec "net.peerCount"

