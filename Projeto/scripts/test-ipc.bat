@echo off
REM Teste usando IPC em vez de HTTP

setlocal enabledelayedexpansion
set TEMP_FILE=%~dp0temp_ipc.js

(
echo var acc = eth.accounts[0];
echo console.log("Account:", acc);
echo var result = personal.unlockAccount(acc, "123456", 0);
echo console.log("Unlocked:", result);
echo var tx = eth.sendTransaction({from: acc, to: acc, value: 1000000000000000});
echo console.log("TX:", tx);
echo exit
) > "%TEMP_FILE%"

docker exec -i geth-node geth attach /root/.ethereum/geth.ipc < "%TEMP_FILE%"

if exist "%TEMP_FILE%" del "%TEMP_FILE%"

pause


