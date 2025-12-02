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

REM Verifica se chaindata existe, se não, inicializa
set DATA_DIR=%BLOCKCHAIN_DIR%\data
set GENESIS_FILE=%BLOCKCHAIN_DIR%\genesis.json

if not exist "%DATA_DIR%\geth" (
    echo [INFO] Chaindata nao encontrado, inicializando blockchain...
    docker run --rm -v "%DATA_DIR%:/root/.ethereum" -v "%GENESIS_FILE%:/genesis.json" ethereum/client-go:v1.13.15 --datadir=/root/.ethereum init /genesis.json
    if %ERRORLEVEL% NEQ 0 (
        echo ERRO: Falha ao inicializar blockchain
        pause
        exit /b 1
    )
    echo [OK] Blockchain inicializada
    echo.
)

REM Inicia container
docker-compose -f docker-compose-blockchain.yml up -d
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao iniciar blockchain
    pause
    exit /b 1
)
echo [OK] Container blockchain iniciado
echo.

REM Aguarda blockchain estar pronta - VERSÃO ULTRA SIMPLIFICADA
echo Aguardando blockchain estar pronta...
timeout /t 8 /nobreak >nul

REM Verifica uma vez se está funcionando
docker ps 2>nul | findstr "geth-node" >nul
if %ERRORLEVEL% NEQ 0 (
    echo [AVISO] Container Geth pode nao estar rodando
    echo Continuando mesmo assim...
    echo.
) else (
    echo [OK] Container verificado
)

netstat -an 2>nul | findstr ":8545" | findstr "LISTENING" >nul
if %ERRORLEVEL% NEQ 0 (
    echo [AVISO] Porta 8545 pode nao estar aberta ainda
    echo Continuando mesmo assim...
    echo.
) else (
    echo [OK] Porta RPC verificada
)

echo [OK] Blockchain pronta (ou iniciando em background)
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
echo Servicos disponiveis:
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
