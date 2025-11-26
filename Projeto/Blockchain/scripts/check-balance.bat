@echo off
REM ===================== BAREMA ITEM 1: ARQUITETURA =====================
REM Script para verificar saldo da conta

echo Verificando saldo da conta...
echo.

(
echo var account = eth.accounts[0];
echo if ^(!account^) {
echo   console.log^("ERRO: Nenhuma conta encontrada!"^);
echo   exit;
echo }
echo console.log^("Conta: " + account^);
echo var balance = eth.getBalance^(account^);
echo console.log^("Saldo: " + web3.fromWei^(balance, "ether"^) + " ETH"^);
echo console.log^("Saldo (Wei): " + balance^);
echo exit
) > temp_balance.js

docker exec -i geth-node geth attach http://localhost:8545 < temp_balance.js
del temp_balance.js

pause


