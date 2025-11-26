@echo off
REM ===================== BAREMA ITEM 1: ARQUITETURA =====================
REM Script para forçar criação de bloco enviando uma transação

echo ========================================
echo Forcando criacao de bloco...
echo ========================================
echo.

(
echo var account = eth.accounts[0];
echo if ^(!account^) {
echo   console.log^("ERRO: Nenhuma conta encontrada!"^);
echo   exit;
echo }
echo console.log^("Conta: " + account^);
echo var unlocked = personal.unlockAccount(account, "123456", 0^);
echo if ^(unlocked^) {
echo   console.log^("Conta desbloqueada!"^);
echo } else {
echo   console.log^("ERRO: Falha ao desbloquear!"^);
echo   exit;
echo }
echo console.log^("Enviando transacao..."^);
echo var txHash = eth.sendTransaction^({from: account, to: account, value: web3.toWei^(0.0001, "ether"^)}^);
echo console.log^("Transacao enviada: " + txHash^);
echo console.log^("Aguardando confirmacao..."^);
echo var startBlock = eth.blockNumber;
echo var maxWait = 30;
echo var waited = 0;
echo while ^(eth.blockNumber == startBlock ^&^& waited ^< maxWait^) {
echo   for ^(var i = 0; i ^< 1000; i++^) { }
echo   waited++;
echo }
echo var blockNumber = eth.blockNumber;
echo console.log^("Bloco atual: " + blockNumber^);
echo if ^(blockNumber ^> 0^) {
echo   console.log^("SUCCESS: Blocos estao sendo criados!"^);
echo } else {
echo   console.log^("AVISO: Ainda no bloco 0."^);
echo }
echo exit
) > temp_force.js

docker exec -i geth-node geth attach http://localhost:8545 < temp_force.js
del temp_force.js

echo.
pause

