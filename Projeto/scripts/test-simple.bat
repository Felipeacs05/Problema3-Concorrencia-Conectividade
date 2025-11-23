@echo off
REM Teste simples para verificar se conseguimos enviar transação

setlocal enabledelayedexpansion
set TEMP_FILE=%~dp0temp_simple.js

echo var account = eth.accounts[0]; > "%TEMP_FILE%"
echo console.log("Account: " + account); >> "%TEMP_FILE%"
echo var unlocked = personal.unlockAccount(account, "123456", 0); >> "%TEMP_FILE%"
echo console.log("Unlocked: " + unlocked); >> "%TEMP_FILE%"
echo var balance = eth.getBalance(account); >> "%TEMP_FILE%"
echo console.log("Balance: " + web3.fromWei(balance, "ether") + " ETH"); >> "%TEMP_FILE%"
echo var tx = {from: account, to: account, value: 1000000000000000}; >> "%TEMP_FILE%"
echo console.log("Sending transaction..."); >> "%TEMP_FILE%"
echo var txHash = eth.sendTransaction(tx); >> "%TEMP_FILE%"
echo console.log("TX Hash: " + txHash); >> "%TEMP_FILE%"
echo exit >> "%TEMP_FILE%"

docker exec -i geth-node geth attach http://localhost:8545 < "%TEMP_FILE%"

if exist "%TEMP_FILE%" del "%TEMP_FILE%"

pause


