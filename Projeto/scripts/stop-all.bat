@echo off
REM ===================== STOP ALL =====================
REM Script para parar toda a infraestrutura

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set PROJECT_DIR=%SCRIPT_DIR%..
set BLOCKCHAIN_DIR=%PROJECT_DIR%\Blockchain
set JOGO_DIR=%PROJECT_DIR%\Jogo

echo ========================================
echo Parando Infraestrutura
echo ========================================
echo.

REM Para jogo
echo [1/2] Parando jogo...
cd /d "%JOGO_DIR%"
docker-compose down
echo [OK] Jogo parado
echo.

REM Para blockchain
echo [2/2] Parando blockchain...
cd /d "%BLOCKCHAIN_DIR%"
docker-compose -f docker-compose-blockchain.yml down
echo [OK] Blockchain parada
echo.

echo ========================================
echo Infraestrutura parada com sucesso!
echo ========================================
echo.
pause

