@echo off
REM ===================== NOVO NÓ =====================
REM Script para criar e iniciar um segundo nó Geth no mesmo computador
REM Este script conecta automaticamente ao primeiro nó (geth-node)

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set PROJECT_DIR=%SCRIPT_DIR%..
set BLOCKCHAIN_DIR=%PROJECT_DIR%\Blockchain
set TOOLS_DIR=%BLOCKCHAIN_DIR%\tools
set DATA_DIR=%BLOCKCHAIN_DIR%\data
set DATA2_DIR=%BLOCKCHAIN_DIR%\data2
set KEYSTORE2_DIR=%DATA2_DIR%\keystore
set GENESIS_FILE=%BLOCKCHAIN_DIR%\genesis.json
set PEER_COMPOSE_FILE=%BLOCKCHAIN_DIR%\docker-compose-peer.yml

echo ========================================
echo Criando Novo No (Peer)
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

REM Verifica se o primeiro nó está rodando
echo [1/8] Verificando se o primeiro no esta rodando...
docker ps | findstr "geth-node" >nul
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: O primeiro no (geth-node) nao esta rodando!
    echo Execute 'start-all.bat' primeiro para iniciar o no principal.
    pause
    exit /b 1
)

REM Aguarda o nó 1 estar pronto
echo Aguardando no 1 estar pronto...
set WAIT_COUNT=0
:WAIT_NODE1
timeout /t 2 /nobreak >nul
docker exec geth-node geth attach --exec "eth.blockNumber" http://localhost:8545 >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [OK] No 1 esta pronto
    goto :NODE1_READY
)
set /a WAIT_COUNT+=1
if !WAIT_COUNT! GEQ 30 (
    echo ERRO: No 1 nao respondeu apos 60 segundos
    pause
    exit /b 1
)
goto :WAIT_NODE1
:NODE1_READY
echo.

REM Verifica se o nó 2 já existe
docker ps | findstr "geth-peer" >nul
if %ERRORLEVEL% EQU 0 (
    echo [AVISO] O no 2 (geth-peer) ja esta rodando!
    set /p RECRIAR="Deseja parar e recriar? (s/N): "
    if /i "!RECRIAR!"=="s" (
        echo Parando no 2 existente...
        docker stop geth-peer >nul 2>&1
        docker rm geth-peer >nul 2>&1
    ) else (
        echo Operacao cancelada.
        pause
        exit /b 0
    )
)

REM Cria diretórios para o nó 2
echo [2/8] Criando diretorios para o no 2...
if not exist "%KEYSTORE2_DIR%" mkdir "%KEYSTORE2_DIR%"
if not exist "%DATA2_DIR%" mkdir "%DATA2_DIR%"
echo [OK] Diretorios criados
echo.

REM Remove dados antigos do nó 2 (se existirem)
if exist "%DATA2_DIR%\geth" (
    echo [3/8] Removendo dados antigos do no 2...
    rmdir /s /q "%DATA2_DIR%\geth"
    echo [OK] Dados antigos removidos
) else (
    echo [3/8] Nenhum dado antigo encontrado
)
echo.

REM Verifica se Go está instalado (para criar conta)
where go >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [AVISO] Go nao esta instalado. Criando conta manualmente...
    echo 123456 > "%DATA2_DIR%\password.txt"
    docker run --rm -v "%KEYSTORE2_DIR%:/keystore" -v "%DATA2_DIR%\password.txt:/password.txt" ethereum/client-go:v1.13.15 account new --keystore /keystore --password /password.txt >nul 2>&1
    for /f "tokens=3" %%a in ('docker run --rm -v "%KEYSTORE2_DIR%:/keystore" ethereum/client-go:v1.13.15 account list --keystore /keystore ^| findstr "Account"') do set ADDRESS=%%a
    set ADDRESS=!ADDRESS:{=!
    set ADDRESS=!ADDRESS:}=!
) else (
    REM Compila o utilitário blockchain-utils se necessário
    if not exist "%TOOLS_DIR%\blockchain-utils.exe" (
        echo [3.5/8] Compilando blockchain-utils...
        cd /d "%TOOLS_DIR%"
        go mod tidy
        set GOOS=windows
        set GOARCH=amd64
        go build -o blockchain-utils.exe blockchain-utils.go
        set GOOS=
        set GOARCH=
        cd /d "%PROJECT_DIR%"
    )
    
    REM Cria conta usando blockchain-utils
    echo [4/8] Criando nova conta para o no 2...
    "%TOOLS_DIR%\blockchain-utils.exe" criar-conta "%KEYSTORE2_DIR%" "123456"
    if %ERRORLEVEL% NEQ 0 (
        echo ERRO: Falha ao criar conta
        pause
        exit /b 1
    )
    
    REM Extrai endereço da conta
    "%TOOLS_DIR%\blockchain-utils.exe" extrair-endereco "%KEYSTORE2_DIR%" > "%TEMP%\peer-address.txt"
    set /p ADDRESS=<"%TEMP%\peer-address.txt"
    del "%TEMP%\peer-address.txt" 2>nul
    
    if "!ADDRESS!"=="" (
        echo ERRO: Falha ao extrair endereco
        pause
        exit /b 1
    )
    
    REM Cria arquivo password.txt
    echo 123456 > "%DATA2_DIR%\password.txt"
    echo [OK] Conta criada: !ADDRESS!
)
echo.

REM Inicializa blockchain do nó 2 com genesis.json
echo [5/8] Inicializando blockchain do no 2...
docker run --rm -v "%DATA2_DIR%:/root/.ethereum" -v "%GENESIS_FILE%:/genesis.json" ethereum/client-go:v1.13.15 --datadir=/root/.ethereum init /genesis.json
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao inicializar blockchain do no 2
    pause
    exit /b 1
)
echo [OK] Blockchain do no 2 inicializada
echo.

REM Obtém o enode do nó 1
echo [6/8] Obtendo enode do no 1...
docker exec geth-node geth attach --exec "admin.nodeInfo.enode" http://localhost:8545 > "%TEMP%\enode.txt" 2>nul
set /p ENODE=<"%TEMP%\enode.txt"
del "%TEMP%\enode.txt" 2>nul

REM Remove aspas e espaços
set ENODE=!ENODE:"=!
set ENODE=!ENODE: =!

if "!ENODE!"=="" (
    echo ERRO: Falha ao obter enode do no 1
    pause
    exit /b 1
)

REM Tenta descobrir o IP da máquina
for /f "tokens=2 delims=:" %%a in ('ipconfig ^| findstr /i "IPv4"') do (
    set HOST_IP=%%a
    set HOST_IP=!HOST_IP: =!
    goto :IP_FOUND
)
set HOST_IP=127.0.0.1
:IP_FOUND

REM Substitui [::] pelo IP real
set ENODE=!ENODE:[::]:=[%HOST_IP%]:!
set ENODE=!ENODE:@[::]@=@%HOST_IP%@!

echo [OK] Enode obtido: !ENODE!
echo.

REM Cria o arquivo docker-compose-peer.yml
echo [7/8] Criando docker-compose-peer.yml...
(
echo # ===================== NÓ PEER (SEGUNDO NÓ) =====================
echo # Este arquivo define o segundo nó Geth que se conecta ao primeiro nó
echo.
echo services:
echo   geth-peer:
echo     image: ethereum/client-go:v1.13.15
echo     container_name: geth-peer
echo     ports:
echo       - "8547:8545"      # HTTP RPC (porta diferente do nó 1)
echo       - "8548:8546"      # WebSocket (porta diferente do nó 1)
echo       - "30304:30303/udp"  # P2P discovery (porta diferente do nó 1)
echo       - "30304:30303"    # P2P TCP (porta diferente do nó 1)
echo     volumes:
echo       - ./data2:/root/.ethereum
echo       - ./genesis.json:/genesis.json
echo       - ./password-peer.txt:/root/.ethereum/password.txt
echo     command:
echo       - --datadir=/root/.ethereum
echo       - --networkid=1337
echo       - --http
echo       - --http.addr=0.0.0.0
echo       - --http.port=8545
echo       - --http.api=eth,net,web3,personal,miner,admin
echo       - --http.corsdomain=*
echo       - --http.vhosts=*
echo       - --ws
echo       - --ws.addr=0.0.0.0
echo       - --ws.port=8546
echo       - --ws.api=eth,net,web3,personal,miner,admin
echo       - --ws.origins="*"
echo       - --allow-insecure-unlock
echo       - --unlock=!ADDRESS!
echo       - --password=/root/.ethereum/password.txt
echo       - --bootnodes=!ENODE!
echo       - --port=30303
echo     stdin_open: true
echo     tty: true
echo     restart: unless-stopped
echo     networks:
echo       - peer_network
echo.
echo networks:
echo   peer_network:
echo     driver: bridge
) > "%PEER_COMPOSE_FILE%"

REM Copia password.txt para password-peer.txt
copy "%DATA2_DIR%\password.txt" "%BLOCKCHAIN_DIR%\password-peer.txt" >nul 2>&1
if not exist "%BLOCKCHAIN_DIR%\password-peer.txt" (
    echo 123456 > "%BLOCKCHAIN_DIR%\password-peer.txt"
)

echo [OK] Arquivo docker-compose-peer.yml criado
echo.

REM Inicia o nó 2
echo [8/8] Iniciando no 2...
cd /d "%BLOCKCHAIN_DIR%"
docker-compose -f docker-compose-peer.yml up -d
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao iniciar no 2
    pause
    exit /b 1
)
echo [OK] No 2 iniciado
echo.

REM Aguarda o nó 2 estar pronto
echo Aguardando no 2 estar pronto...
timeout /t 6 /nobreak >nul
set WAIT_COUNT=0
:WAIT_NODE2
timeout /t 2 /nobreak >nul
docker exec geth-peer geth attach --exec "eth.blockNumber" http://localhost:8545 >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    goto :NODE2_READY
)
set /a WAIT_COUNT+=1
if !WAIT_COUNT! GEQ 15 (
    echo [AVISO] No 2 pode nao estar totalmente pronto ainda
    goto :NODE2_READY
)
goto :WAIT_NODE2
:NODE2_READY

REM Verifica conexão entre os nós
echo.
echo Verificando conexao entre os nos...
timeout /t 3 /nobreak >nul
docker exec geth-node geth attach --exec "admin.peers.length" http://localhost:8545 > "%TEMP%\peers1.txt" 2>nul
set /p PEERS_NODE1=<"%TEMP%\peers1.txt" 2>nul
del "%TEMP%\peers1.txt" 2>nul
if "!PEERS_NODE1!"=="" set PEERS_NODE1=0

docker exec geth-peer geth attach --exec "admin.peers.length" http://localhost:8545 > "%TEMP%\peers2.txt" 2>nul
set /p PEERS_NODE2=<"%TEMP%\peers2.txt" 2>nul
del "%TEMP%\peers2.txt" 2>nul
if "!PEERS_NODE2!"=="" set PEERS_NODE2=0

echo.
echo ========================================
echo Novo No Criado com Sucesso!
echo ========================================
echo.
echo Status da conexao:
echo - No 1 (geth-node) ve !PEERS_NODE1! peer(s)
echo - No 2 (geth-peer) ve !PEERS_NODE2! peer(s)
echo.
echo Servicos disponiveis:
echo - No 1 RPC: http://localhost:8545
echo - No 2 RPC: http://localhost:8547
echo - No 1 WebSocket: ws://localhost:8546
echo - No 2 WebSocket: ws://localhost:8548
echo.
echo Comandos uteis:
echo - Ver peers do no 1: docker exec geth-node geth attach http://localhost:8545 --exec "admin.peers"
echo - Ver peers do no 2: docker exec geth-peer geth attach http://localhost:8545 --exec "admin.peers"
echo - Parar no 2: cd %BLOCKCHAIN_DIR% ^&^& docker-compose -f docker-compose-peer.yml down
echo.
pause
