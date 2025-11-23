#!/bin/bash
# ===================== BAREMA ITEM 1: ARQUITETURA =====================
# Script para obter o enode deste nó (Linux/macOS)
# Útil para compartilhar com outros participantes da rede

echo "Obtendo enode do nó local..."
echo ""

echo "admin.nodeInfo.enode" | docker exec -i geth-node geth attach http://localhost:8545

echo ""
echo "Compartilhe este enode com outros participantes para que eles possam se conectar."
echo ""


