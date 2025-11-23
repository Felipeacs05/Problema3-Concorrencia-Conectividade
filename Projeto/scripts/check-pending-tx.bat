@echo off
REM Script para verificar transações pendentes

set SCRIPT_DIR=%~dp0
set TEMP_FILE=%SCRIPT_DIR%temp_check_tx.js

echo ========================================
echo Verificando transacoes pendentes...
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
echo var pending = txpool.content.pending;
echo console.log^("Transacoes pendentes:"^);
echo console.log^("Total: " + JSON.stringify^(pending^).length + " caracteres"^);
echo.
echo var queued = txpool.content.queued;
echo console.log^("Transacoes na fila:"^);
echo console.log^("Total: " + JSON.stringify^(queued^).length + " caracteres"^);
echo.
echo var blockNumber = eth.blockNumber;
echo console.log^("Bloco atual: " + blockNumber^);
) > "%TEMP_FILE%"

docker exec -i geth-node geth attach < "%TEMP_FILE%"

del "%TEMP_FILE%" 2>nul

pause

