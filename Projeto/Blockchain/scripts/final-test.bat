@echo off
REM ===================== BAREMA ITEM 1: ARQUITETURA =====================
REM Teste final: desbloqueia e envia transação de forma mais simples

setlocal enabledelayedexpansion
set TEMP_FILE=%~dp0temp_final.js

echo ========================================
echo Teste Final - Desbloquear e Enviar TX
echo ========================================
echo.

if exist "%TEMP_FILE%" del "%TEMP_FILE%"

(
echo var acc = eth.accounts[0];
echo console.log("Account:", acc);
echo var result = personal.unlockAccount(acc, "123456", 0);
echo console.log("Unlocked:", result);
echo var txObj = {from: acc, to: acc, value: 1000000000000000};
echo var tx = eth.sendTransaction(txObj);
echo console.log("TX Hash:", tx);
echo exit
) > "%TEMP_FILE%"

docker exec -i geth-node geth attach http://localhost:8545 < "%TEMP_FILE%"

if exist "%TEMP_FILE%" del "%TEMP_FILE%"

echo.
echo Verificando bloco em 5 segundos...
timeout /t 5 /nobreak >nul

(
echo console.log("Block:", eth.blockNumber);
echo exit
) > "%TEMP_FILE%"

docker exec -i geth-node geth attach http://localhost:8545 < "%TEMP_FILE%"

if exist "%TEMP_FILE%" del "%TEMP_FILE%"

pause


