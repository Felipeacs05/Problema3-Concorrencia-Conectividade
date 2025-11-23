@echo off
REM ===================== BAREMA ITEM 1: ARQUITETURA =====================
REM Script para verificar se tudo está configurado corretamente

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set TEMP_FILE=%SCRIPT_DIR%temp_verify.js

echo ========================================
echo Verificando configuracao...
echo ========================================
echo.

REM Cria arquivo de verificação
(
echo console.log^("=== VERIFICACAO COMPLETA ==="^);
echo console.log^(""^);
echo var accounts = eth.accounts;
echo console.log^("Contas encontradas: " + accounts.length^);
echo for ^(var i = 0; i ^< accounts.length; i++^) {
echo   console.log^("  [" + i + "] " + accounts[i]^);
echo }
echo console.log^(""^);
echo if ^(accounts.length ^> 0^) {
echo   var account = accounts[0];
echo   var balance = eth.getBalance^(account^);
echo   console.log^("Saldo da conta: " + web3.fromWei^(balance, "ether"^) + " ETH"^);
echo   console.log^("Bloco atual: " + eth.blockNumber^);
echo   var wallets = personal.listWallets^();
echo   console.log^("Wallets desbloqueadas: " + wallets.length^);
echo   var block = eth.getBlock^(0^);
echo   if ^(block^) {
echo     console.log^("Genesis extraData: " + block.extraData.substring^(0, 100^) + "..."^);
echo   }
echo }
echo console.log^(""^);
echo console.log^("=== FIM ==="^);
echo exit
) > "%TEMP_FILE%"

docker exec -i geth-node geth attach http://localhost:8545 < "%TEMP_FILE%"

if exist "%TEMP_FILE%" del "%TEMP_FILE%"

echo.
pause


