#!/bin/bash
# ===================== BAREMA ITEM 1: ARQUITETURA =====================
# Script para forçar criação de bloco executando dentro do container

echo "========================================"
echo "Forcando criacao de bloco (dentro do container)..."
echo "========================================"
echo ""

# Executa dentro do container onde o Geth tem acesso direto ao keystore
docker exec geth-node sh -c '
geth attach http://localhost:8545 --exec "
var acc = eth.accounts[0];
console.log(\"Account: \" + acc);
var unlocked = personal.unlockAccount(acc, \"123456\", 0);
console.log(\"Unlocked: \" + unlocked);
if (!unlocked) {
  console.log(\"ERRO: Falha ao desbloquear!\");
  exit;
}
console.log(\"Enviando transacao...\");
var tx = eth.sendTransaction({from: acc, to: acc, value: 1000000000000000});
console.log(\"TX Hash: \" + tx);
console.log(\"Aguardando confirmacao...\");
var startBlock = eth.blockNumber;
var waited = 0;
while (eth.blockNumber == startBlock && waited < 30) {
  for (var i = 0; i < 1000; i++) { }
  waited++;
}
var blockNumber = eth.blockNumber;
console.log(\"Bloco atual: \" + blockNumber);
if (blockNumber > 0) {
  console.log(\"SUCCESS: Blocos estao sendo criados!\");
} else {
  console.log(\"AVISO: Ainda no bloco 0.\");
}
"
'

echo ""
echo "========================================"
echo "Verifique o bloco com: check-block.bat"
echo "========================================"


