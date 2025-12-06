@echo off
REM ===================== REBUILD AND START =====================
REM Script para reconstruir imagens Docker e iniciar toda a infraestrutura
REM Este script recompila 100% tudo (servidores e cliente) e depois inicia como start-all.bat

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set PROJECT_DIR=%SCRIPT_DIR%..
set BLOCKCHAIN_DIR=%PROJECT_DIR%\Blockchain
set JOGO_DIR=%PROJECT_DIR%\Jogo

echo ========================================
echo Reconstruindo e Iniciando Infraestrutura
echo ========================================
echo.
echo Este processo vai:
echo 1. Parar todos os containers
echo 2. Reconstruir imagens Docker (servidores e cliente) - SEM CACHE
echo 3. Iniciar blockchain
echo 4. Iniciar jogo
echo.
echo ATENCAO: A reconstrucao pode demorar varios minutos!
echo.
pause

REM Verifica se Docker está rodando
docker ps >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Docker nao esta rodando!
    echo Inicie o Docker Desktop e tente novamente.
    pause
    exit /b 1
)

REM Para tudo primeiro
echo [1/5] Parando containers existentes...
cd /d "%JOGO_DIR%"
docker-compose down >nul 2>&1
cd /d "%BLOCKCHAIN_DIR%"
docker-compose -f docker-compose-blockchain.yml down >nul 2>&1
echo [OK] Containers parados
echo.

REM Reconstrói as imagens do jogo (servidores e cliente)
echo [2/5] Reconstruindo imagens Docker do jogo (servidores e cliente)...
echo       Isso pode demorar varios minutos, aguarde...
cd /d "%JOGO_DIR%"
docker-compose build --no-cache
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao reconstruir imagens do jogo
    pause
    exit /b 1
)
echo [OK] Imagens do jogo reconstruidas
echo.

REM Inicia blockchain (igual ao start-all.bat)
echo [3/5] Iniciando blockchain...
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

REM Inicia container blockchain
docker-compose -f docker-compose-blockchain.yml up -d
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao iniciar blockchain
    pause
    exit /b 1
)
echo [OK] Container blockchain iniciado
echo.

REM Aguarda blockchain estar pronta
echo [4/5] Aguardando blockchain estar pronta...
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

REM Inicia jogo (igual ao start-all.bat)
echo [5/5] Iniciando jogo...
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
echo Reconstrucao e Inicializacao Concluidas!
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

