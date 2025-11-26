@echo off
REM Script para debugar o Clique

set SCRIPT_DIR=%~dp0
set TEMP_FILE=%SCRIPT_DIR%temp_debug_clique.js

echo ========================================
echo Debugando Clique...
echo ========================================
echo.

(
echo var accounts = eth.accounts;
echo var account = accounts[0];
echo console.log^("Conta: " + account^);
echo.
echo // Verifica se a conta esta desbloqueada
echo var isUnlocked = false;
echo try {
echo   isUnlocked = personal.unlockAccount^(account, "123456", 0^);
echo } catch^(e^) {
echo   console.log^("Erro ao verificar desbloqueio: " + e^);
echo }
echo console.log^("Conta desbloqueada: " + isUnlocked^);
echo.
echo // Verifica blocos
echo var blockNumber = eth.blockNumber;
echo console.log^("Bloco atual: " + blockNumber^);
echo.
echo // Verifica transacoes pendentes
echo var pending = txpool.content.pending;
echo var pendingCount = 0;
echo for ^(var addr in pending^) {
echo   if ^(pending[addr]^) {
echo     pendingCount += pending[addr].length;
echo   }
echo }
echo console.log^("Transacoes pendentes: " + pendingCount^);
echo.
echo // Verifica se o clique esta ativo
echo try {
echo   var signers = clique.getSigners^(eth.getBlock^(blockNumber^).hash^);
echo   console.log^("Signers autorizados: " + signers.length^);
echo   for ^(var i = 0; i < signers.length; i++^) {
echo     console.log^("  - " + signers[i]^);
echo   }
echo } catch^(e^) {
echo   console.log^("Erro ao obter signers: " + e^);
echo }
echo.
echo // Verifica se a conta atual e signer
echo var isSigner = false;
echo try {
echo   var signers = clique.getSigners^(eth.getBlock^(blockNumber^).hash^);
echo   for ^(var i = 0; i < signers.length; i++^) {
echo     if ^(signers[i].toLowerCase^(^) == account.toLowerCase^(^)^) {
echo       isSigner = true;
echo       break;
echo     }
echo   }
echo } catch^(e^) {
echo   console.log^("Erro ao verificar se e signer: " + e^);
echo }
echo console.log^("Conta e signer: " + isSigner^);
) > "%TEMP_FILE%"

docker exec -i geth-node geth attach < "%TEMP_FILE%"

del "%TEMP_FILE%" 2>nul

pause

