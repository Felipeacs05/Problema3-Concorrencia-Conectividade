@echo off
REM ===================== BAREMA ITEM 1: ARQUITETURA =====================
REM Script para transferir ETH da conta do signer para outra conta

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set TEMP_FILE=%SCRIPT_DIR%temp_fund.js

echo ========================================
echo Transferindo ETH para conta...
echo ========================================
echo.

if "%1"=="" (
    echo ERRO: Forneca o endereco da conta destino!
    echo.
    echo Uso: fund-account.bat ^<endereco^>
    echo.
    echo Exemplo:
    echo   fund-account.bat 0x7041e32e2E3b380368e445885b0EdBBC33F234CC
    echo.
    pause
    exit /b 1
)

set DEST_ADDR=%1
set AMOUNT=100

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
) > "%TEMP_FILE%"

docker exec -i geth-node geth attach < "%TEMP_FILE%"

del "%TEMP_FILE%" 2>nul

echo.
echo Aguarde alguns segundos e verifique o saldo no cliente.
echo.
pause

