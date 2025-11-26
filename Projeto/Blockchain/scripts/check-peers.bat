@echo off
REM ===================== BAREMA ITEM 1: ARQUITETURA =====================
REM Script para verificar conexões P2P com outros nós - Windows

echo Verificando conexoes P2P...
echo.

echo admin.peers > temp_peers.js
echo exit >> temp_peers.js

docker exec -i geth-node geth attach http://localhost:8545 < temp_peers.js

del temp_peers.js

echo.
echo Numero de peers conectados:
echo net.peerCount > temp_count.js
echo exit >> temp_count.js

docker exec -i geth-node geth attach http://localhost:8545 < temp_count.js

del temp_count.js

pause


