@echo off
REM ===================== BAREMA ITEM 1: ARQUITETURA =====================
REM Script para verificar número de blocos (Windows)

echo eth.blockNumber > temp_cmd.js
echo exit >> temp_cmd.js

docker exec -i geth-node geth attach http://localhost:8545 < temp_cmd.js

del temp_cmd.js


