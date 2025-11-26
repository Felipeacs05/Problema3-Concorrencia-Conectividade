@echo off
REM ===================== BAREMA ITEM 1: ARQUITETURA =====================
REM Script que desbloqueia conta E envia transação na mesma sessão

setlocal enabledelayedexpansion
set TEMP_FILE=%~dp0temp_unlock_send.js

echo ========================================
echo Desbloqueando e enviando transacao...
echo ========================================
echo.

REM Remove arquivo antigo
if exist "%TEMP_FILE%" del "%TEMP_FILE%"

REM Cria arquivo JavaScript
(
echo var account = eth.accounts[0];
echo if ^(!account^) {
echo   console.log^("ERRO: Nenhuma conta encontrada!"^);
echo   exit;
echo }
echo console.log^("Conta: " + account^);
echo var unlocked = personal.unlockAccount(account, "123456", 0^);
echo console.log^("Desbloqueado: " + unlocked^);
echo if ^(!unlocked^) {
echo   console.log^("ERRO: Falha ao desbloquear!"^);
echo   exit;
echo }
echo console.log^("Enviando transacao..."^);
echo var txHash = eth.sendTransaction^({from: account, to: account, value: 1000000000000000}^);
echo console.log^("Transacao enviada: " + txHash^);
echo console.log^("Aguardando confirmacao..."^);
echo var startBlock = eth.blockNumber;
echo var waited = 0;
echo while ^(eth.blockNumber == startBlock ^&^& waited ^< 30^) {
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
) > "%TEMP_FILE%"

REM Executa
docker exec -i geth-node geth attach http://localhost:8545 < "%TEMP_FILE%"

REM Remove arquivo
if exist "%TEMP_FILE%" del "%TEMP_FILE%"

echo.
pause


