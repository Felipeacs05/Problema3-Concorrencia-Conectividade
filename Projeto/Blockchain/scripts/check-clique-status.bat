@echo off
REM ===================== BAREMA ITEM 1: ARQUITETURA =====================
REM Script para verificar status do Clique

echo Verificando status do Clique...
echo.

(
echo var account = eth.accounts[0];
echo console.log^("Conta: " + account^);
echo var balance = eth.getBalance^(account^);
echo console.log^("Saldo: " + web3.fromWei^(balance, "ether"^) + " ETH"^);
echo console.log^("Bloco atual: " + eth.blockNumber^);
echo var block = eth.getBlock^(0^);
echo if ^(block^) {
echo   console.log^("Genesis extraData: " + block.extraData^);
echo }
echo var wallets = personal.listWallets^();
echo console.log^("Contas desbloqueadas: " + wallets.length^);
echo exit
) > temp_status.js

docker exec -i geth-node geth attach http://localhost:8545 < temp_status.js
del temp_status.js

pause


