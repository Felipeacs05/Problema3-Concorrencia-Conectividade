@echo off
REM ===================== BAREMA ITEM 1: ARQUITETURA =====================
REM Script para transferir ETH da conta do signer para outra conta

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set TEMP_FILE=%SCRIPT_DIR%temp_transfer.js

echo ========================================
echo Transferir ETH entre contas
echo ========================================
echo.

if "%1"=="" (
    echo Uso: transfer-eth.bat ^<endereco_destino^> [quantidade_em_ETH]
    echo.
    echo Exemplo:
    echo   transfer-eth.bat 0x7041e32e2E3b380368e445885b0EdBBC33F234CC 10
    echo   (transfere 10 ETH para a conta especificada)
    echo.
    echo Se quantidade nao for especificada, transfere 100 ETH por padrao.
    pause
    exit /b 1
)

set DEST_ADDR=%1
set AMOUNT=%2

if "%AMOUNT%"=="" set AMOUNT=100

echo Transferindo %AMOUNT% ETH para: %DEST_ADDR%
echo.

REM Cria script JavaScript temporário
(
echo var accounts = eth.accounts;
echo if ^(accounts.length == 0^) {
echo   console.log^("ERRO: Nenhuma conta encontrada!"^);
echo   exit;
echo }
echo var signer = accounts[0];
echo console.log^("Conta signer: " + signer^);
echo.
echo var dest = "%DEST_ADDR%";
echo var amount = web3.toWei^(%AMOUNT%, "ether"^);
echo.
echo var unlocked = personal.unlockAccount^(signer, "123456", 0^);
echo if ^(!unlocked^) {
echo   console.log^("ERRO: Falha ao desbloquear conta signer!"^);
echo   exit;
echo }
echo.
echo console.log^("Enviando " + %AMOUNT% + " ETH..."^);
echo var tx = eth.sendTransaction^({from: signer, to: dest, value: amount}^);
echo console.log^("Transacao: " + tx^);
echo console.log^("SUCCESS: Transferencia enviada!"^);
echo console.log^("Aguarde alguns segundos para confirmacao."^);
) > "%TEMP_FILE%"

docker exec -i geth-node geth attach < "%TEMP_FILE%"

del "%TEMP_FILE%" 2>nul

echo.
pause

