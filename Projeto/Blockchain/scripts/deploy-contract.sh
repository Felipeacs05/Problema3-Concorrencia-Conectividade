#!/bin/bash
# ===================== BAREMA ITEM 2: SMART CONTRACTS =====================
# Script para compilar e fazer deploy do contrato GameEconomy.sol
# Requer: solc (Solidity Compiler) instalado

echo "Compilando contrato GameEconomy.sol..."

# BAREMA ITEM 2: SMART CONTRACTS - Compila o contrato
solc --abi --bin contracts/GameEconomy.sol -o build/

echo "Contrato compilado! Arquivos em ./build/"
echo "Use o cliente Go para fazer deploy do contrato."

