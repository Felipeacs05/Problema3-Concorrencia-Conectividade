#!/bin/bash
# ===================== BAREMA ITEM 1: ARQUITETURA =====================
# Script de inicialização do nó Ethereum privado
# Este script configura o nó Geth e cria a conta inicial

echo "Inicializando nó Ethereum privado..."

# BAREMA ITEM 1: ARQUITETURA - Cria diretório de dados se não existir
mkdir -p ./data

# BAREMA ITEM 1: ARQUITETURA - Inicializa blockchain com genesis.json
geth --datadir ./data init genesis.json

# BAREMA ITEM 1: ARQUITETURA - Cria conta inicial (senha vazia para desenvolvimento)
echo "" > password.txt
geth --datadir ./data account new --password password.txt

echo "Nó Ethereum inicializado com sucesso!"
echo "Para iniciar o nó, execute: docker-compose up geth"

