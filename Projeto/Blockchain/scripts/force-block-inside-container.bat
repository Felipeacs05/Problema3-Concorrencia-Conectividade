@echo off
REM ===================== BAREMA ITEM 1: ARQUITETURA =====================
REM Script para forçar criação de bloco executando dentro do container

echo ========================================
echo Forcando criacao de bloco (dentro do container)...
echo ========================================
echo.

REM Executa script bash dentro do container
docker exec geth-node bash /scripts/force-block-inside-container.sh

echo.
echo ========================================
echo Verifique o bloco com: check-block.bat
echo ========================================
echo.

pause


