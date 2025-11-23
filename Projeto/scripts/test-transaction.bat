@echo off
REM ===================== BAREMA ITEM 1: ARQUITETURA =====================
REM Script para testar envio de transação de forma mais simples

echo Testando envio de transacao...
echo.

(
echo var account = eth.accounts[0];
echo console.log^("Conta: " + account^);
echo.
echo REM Verifica se conta esta desbloqueada
echo var wallets = personal.listWallets^();
echo console.log^("Wallets desbloqueadas: " + wallets.length^);
echo.
echo REM Desbloqueia novamente
echo var unlocked = personal.unlockAccount(account, "123456", 0^);
echo console.log^("Desbloqueado: " + unlocked^);
echo.
echo REM Tenta enviar transacao
echo try {
echo   var tx = eth.sendTransaction^({from: account, to: account, value: 1000000000000000^});
echo   console.log^("SUCCESS: Transacao enviada: " + tx^);
echo } catch ^(e^) {
echo   console.log^("ERRO: " + e.message^);
echo }
echo exit
) > temp_test.js

docker exec geth-node geth attach http://localhost:8545 < temp_test.js
del temp_test.js

pause


