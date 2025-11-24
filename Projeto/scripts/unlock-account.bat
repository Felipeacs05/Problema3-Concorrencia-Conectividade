@echo off
REM ===================== BAREMA ITEM 1: ARQUITETURA =====================
REM Script para desbloquear conta no Clique (Windows)

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set TEMP_FILE=%SCRIPT_DIR%temp_unlock.js

echo ========================================
echo Desbloqueando conta do signer...
echo ========================================
echo.

REM Cria arquivo temporário
(
echo var accounts = eth.accounts;
echo if ^(accounts.length == 0^) {
echo   console.log^("ERRO: Nenhuma conta encontrada!"^);
echo   exit;
echo }
echo var account = accounts[0];
echo console.log^("Conta: " + account^);
echo var result = personal.unlockAccount(account, "123456", 0^);
echo if ^(result^) {
echo   console.log^("SUCCESS: Conta desbloqueada permanentemente!"^);
echo   miner.start^();
echo   console.log^("Minerador iniciado (miner.start)"^);
echo } else {
echo   console.log^("ERRO: Falha ao desbloquear conta!"^);
echo   console.log^("Verifique se a senha esta correta (padrao: 123456)"^);
echo }
echo exit
) > "%TEMP_FILE%"

REM Verifica se o arquivo foi criado
if not exist "%TEMP_FILE%" (
    echo ERRO: Falha ao criar arquivo temporario!
    pause
    exit /b 1
)

REM Executa o comando
docker exec -i geth-node geth attach http://localhost:8545 < "%TEMP_FILE%"

REM Remove arquivo temporário
if exist "%TEMP_FILE%" del "%TEMP_FILE%"

echo.
echo ========================================
echo Clique deve comecar a selar blocos automaticamente
echo ========================================
echo.

pause

