#!/bin/bash
# ===================== BAREMA ITEM 1: ARQUITETURA =====================
# Script para iniciar mineração manualmente
# Útil para obter ETH para pagar transações

echo "Iniciando mineração..."

# BAREMA ITEM 1: ARQUITETURA - Acessa o console do Geth e inicia mineração
docker exec -it geth-node geth attach http://localhost:8545 --exec "miner.start(1)"

echo "Mineração iniciada!"
echo "Para parar: docker exec -it geth-node geth attach http://localhost:8545 --exec 'miner.stop()'"

