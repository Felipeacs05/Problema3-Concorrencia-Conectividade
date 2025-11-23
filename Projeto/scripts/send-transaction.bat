@echo off
REM ===================== BAREMA ITEM 1: ARQUITETURA =====================
REM Script para enviar transação e forçar criação de bloco

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set TEMP_FILE=%SCRIPT_DIR%temp_tx.js

echo ========================================
echo Enviando transacao para forcar bloco...
echo ========================================
echo.

REM Remove arquivo antigo se existir
if exist "%TEMP_FILE%" del "%TEMP_FILE%"

REM Cria arquivo JavaScript linha por linha
(
echo var account = eth.accounts[0];
echo if ^(!account^) {
echo   console.log^("ERRO: Nenhuma conta encontrada!"^);
echo   exit;
echo }
echo console.log^("Conta: " + account^);
echo var unlocked = personal.unlockAccount(account, "123456", 0^);
echo console.log^("Desbloqueado: " + unlocked^);
echo try {
echo   var txHash = eth.sendTransaction^({from: account, to: account, value: 1000000000000000}^);
echo   console.log^("SUCCESS: Transacao enviada: " + txHash^);
echo   console.log^("Aguardando confirmacao..."^);
echo   var startBlock = eth.blockNumber;
echo   var waited = 0;
echo   while ^(eth.blockNumber == startBlock ^&^& waited ^< 30^) {
echo     for ^(var i = 0; i ^< 1000; i++^) { }
echo     waited++;
echo   }
echo   var blockNumber = eth.blockNumber;
echo   console.log^("Bloco atual: " + blockNumber^);
echo   if ^(blockNumber ^> 0^) {
echo     console.log^("SUCCESS: Blocos estao sendo criados!"^);
echo   }
echo } catch ^(e^) {
echo   console.log^("ERRO: " + e.message^);
echo }
echo exit
) > "%TEMP_FILE%"

REM Verifica se arquivo foi criado
if not exist "%TEMP_FILE%" (
    echo ERRO: Falha ao criar arquivo temporario!
    pause
    exit /b 1
)

REM Executa
docker exec -i geth-node geth attach http://localhost:8545 < "%TEMP_FILE%"

REM Remove arquivo
if exist "%TEMP_FILE%" del "%TEMP_FILE%"

echo.
pause
