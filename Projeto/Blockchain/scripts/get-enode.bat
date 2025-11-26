@echo off
REM ===================== BAREMA ITEM 1: ARQUITETURA =====================
REM Script para obter o enode deste nó (Windows)
REM Útil para compartilhar com outros participantes da rede

echo Obtendo enode do no local...
echo.

echo admin.nodeInfo.enode > temp_enode.js
echo exit >> temp_enode.js

docker exec -i geth-node geth attach http://localhost:8545 < temp_enode.js

del temp_enode.js

echo.
echo Compartilhe este enode com outros participantes para que eles possam se conectar.
echo.

pause


