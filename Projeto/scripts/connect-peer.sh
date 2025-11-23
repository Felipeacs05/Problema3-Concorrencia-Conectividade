#!/bin/bash
# ===================== BAREMA ITEM 1: ARQUITETURA =====================
# Script para conectar este nó a um peer (bootnode) - Linux/macOS
# Uso: ./connect-peer.sh <enode-do-bootnode>

if [ -z "$1" ]; then
    echo "Uso: $0 <enode-do-bootnode>"
    echo "Exemplo: $0 enode://abc123...@192.168.1.100:30303"
    exit 1
fi

BOOTNODE_ENODE="$1"

echo "Conectando ao bootnode: $BOOTNODE_ENODE"
echo ""

# BAREMA ITEM 1: ARQUITETURA - Inicia o nó com bootnode especificado
export BOOTNODE_ENODE
docker-compose up geth


