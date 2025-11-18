#!/bin/bash
# ===================== BAREMA ITEM 1: ARQUITETURA =====================
# Script para obter o enode deste nó
# Útil para compartilhar com outros participantes da rede

echo "Obtendo enode do nó local..."

# BAREMA ITEM 1: ARQUITETURA - Acessa o console do Geth e obtém informações do nó
docker exec -it geth-node geth attach http://localhost:8545 --exec "admin.nodeInfo.enode"

echo ""
echo "Compartilhe este enode com outros participantes para que eles possam se conectar."

