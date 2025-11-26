@echo off
REM ===================== START ALL =====================
REM Script para iniciar toda a infraestrutura (blockchain + jogo)

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set PROJECT_DIR=%SCRIPT_DIR%..
set BLOCKCHAIN_DIR=%PROJECT_DIR%\Blockchain
set JOGO_DIR=%PROJECT_DIR%\Jogo

echo ========================================
echo Iniciando Infraestrutura Completa
echo ========================================
echo.

REM Verifica se Docker está rodando
docker ps >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Docker nao esta rodando!
    echo Inicie o Docker Desktop e tente novamente.
    pause
    exit /b 1
)

REM Inicia blockchain
echo [1/2] Iniciando blockchain...
cd /d "%BLOCKCHAIN_DIR%"
docker-compose -f docker-compose-blockchain.yml up -d
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao iniciar blockchain
    pause
    exit /b 1
)
echo [OK] Blockchain iniciada
echo.

REM Aguarda blockchain estar pronta
echo Aguardando blockchain estar pronta...
:WAIT_BLOCKCHAIN
timeout /t 3 /nobreak >nul
docker exec geth-node geth attach --exec "eth.blockNumber" http://localhost:8545 >nul 2>&1
if %errorlevel% neq 0 (
    echo Aguardando blockchain...
    goto WAIT_BLOCKCHAIN
)
echo [OK] Blockchain pronta
echo.

REM Inicia jogo
echo [2/2] Iniciando jogo...
cd /d "%JOGO_DIR%"
docker-compose up -d
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao iniciar jogo
    pause
    exit /b 1
)
echo [OK] Jogo iniciado
echo.

echo ========================================
echo Infraestrutura iniciada com sucesso!
echo ========================================
echo.
echo Serviços disponíveis:
echo - Blockchain: http://localhost:8545
echo - Servidor 1: http://localhost:8080
echo - Servidor 2: http://localhost:8081
echo - Servidor 3: http://localhost:8082
echo - Broker MQTT 1: tcp://localhost:1886
echo - Broker MQTT 2: tcp://localhost:1884
echo - Broker MQTT 3: tcp://localhost:1885
echo.
echo Para parar tudo, execute: stop-all.bat
echo.
pause

