@echo off
REM ===================== BAREMA ITEM 1: ARQUITETURA =====================
REM Script para forçar criação do primeiro bloco no Clique

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set TEMP_FILE=%SCRIPT_DIR%temp_force_block.js

echo ========================================
echo Forcando criacao do primeiro bloco...
echo ========================================
echo.

(
echo var accounts = eth.accounts;
echo if ^(accounts.length == 0^) {
echo   console.log^("ERRO: Nenhuma conta encontrada!"^);
echo   exit;
echo }
echo var account = accounts[0];
echo console.log^("Conta: " + account^);
echo.
echo // Desbloqueia a conta
echo var unlocked = personal.unlockAccount^(account, "123456", 0^);
echo if ^(!unlocked^) {
echo   console.log^("ERRO: Falha ao desbloquear conta!"^);
echo   exit;
echo }
echo console.log^("Conta desbloqueada!"^);
echo.
echo // Verifica se ha transacoes pendentes
echo var pending = txpool.content.pending;
echo var hasPending = false;
echo for ^(var addr in pending^) {
echo   if ^(pending[addr].length > 0^) {
echo     hasPending = true;
echo     break;
echo   }
echo }
echo.
echo if ^(hasPending^) {
echo   console.log^("Transacoes pendentes encontradas!"^);
echo   console.log^("O Clique deve selar blocos automaticamente."^);
echo   console.log^("Aguardando 10 segundos..."^);
echo   admin.sleep^(10^);
echo   var blockNumber = eth.blockNumber;
echo   console.log^("Bloco atual: " + blockNumber^);
echo   if ^(blockNumber > 0^) {
echo     console.log^("SUCCESS: Blocos estao sendo criados!"^);
echo   } else {
echo     console.log^("AVISO: Ainda no bloco 0. O Clique pode precisar de mais tempo."^);
echo   }
echo } else {
echo   console.log^("Nenhuma transacao pendente. Enviando uma transacao de teste..."^);
echo   var tx = eth.sendTransaction^({from: account, to: account, value: web3.toWei^(0.001, "ether"^)^}^);
echo   console.log^("Transacao enviada: " + tx^);
echo   console.log^("Aguardando 10 segundos..."^);
echo   admin.sleep^(10^);
echo   var blockNumber = eth.blockNumber;
echo   console.log^("Bloco atual: " + blockNumber^);
echo   if ^(blockNumber > 0^) {
echo     console.log^("SUCCESS: Primeiro bloco criado!"^);
echo   } else {
echo     console.log^("AVISO: Ainda no bloco 0."^);
echo   }
echo }
) > "%TEMP_FILE%"

docker exec -i geth-node geth attach < "%TEMP_FILE%"

del "%TEMP_FILE%" 2>nul

echo.
pause

