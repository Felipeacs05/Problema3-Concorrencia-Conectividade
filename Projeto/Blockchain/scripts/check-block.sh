#!/bin/bash
# ===================== BAREMA ITEM 1: ARQUITETURA =====================
# Script para verificar número de blocos (Linux/macOS)

echo "Bloco atual:"
echo "eth.blockNumber" | docker exec -i geth-node geth attach http://localhost:8545 | grep -E "^[0-9]+$"


