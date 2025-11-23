@echo off
REM Script para iniciar o miner no Clique

echo ========================================
echo Iniciando miner no Clique...
echo ========================================
echo.

docker exec -i geth-node geth attach --exec "miner.start()"

timeout /t 3 /nobreak >nul

docker exec -i geth-node geth attach --exec "eth.mining"
docker exec -i geth-node geth attach --exec "eth.blockNumber"

echo.
echo Se eth.mining = true e blockNumber > 0, o Clique esta funcionando!
echo.
pause

