@echo off
REM ===================== VERIFICAR NÓS =====================
REM Script para verificar o status e conexão entre os nós Geth

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set PROJECT_DIR=%SCRIPT_DIR%..

echo ========================================
echo Verificando Status dos Nos
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

REM Verifica se os containers estão rodando
echo [1/5] Verificando containers...
set NODE1_RUNNING=false
set NODE2_RUNNING=false

docker ps | findstr "geth-node" >nul
if %ERRORLEVEL% EQU 0 (
    set NODE1_RUNNING=true
    echo [OK] No 1 (geth-node): RODANDO
) else (
    echo [ERRO] No 1 (geth-node): NAO ESTA RODANDO
)

docker ps | findstr "geth-peer" >nul
if %ERRORLEVEL% EQU 0 (
    set NODE2_RUNNING=true
    echo [OK] No 2 (geth-peer): RODANDO
) else (
    echo [AVISO] No 2 (geth-peer): NAO ESTA RODANDO (opcional)
)
echo.

REM Verifica se os nós estão respondendo
echo [2/5] Verificando resposta dos nos...

if "!NODE1_RUNNING!"=="true" (
    docker exec geth-node geth attach --exec "eth.blockNumber" http://localhost:8545 > "%TEMP%\block1.txt" 2>nul
    if %ERRORLEVEL% EQU 0 (
        set /p BLOCK_NODE1=<"%TEMP%\block1.txt"
        set BLOCK_NODE1=!BLOCK_NODE1: =!
        echo [OK] No 1: RESPONDENDO (Bloco: !BLOCK_NODE1!)
    ) else (
        echo [ERRO] No 1: NAO RESPONDE
        set BLOCK_NODE1=N/A
    )
    del "%TEMP%\block1.txt" 2>nul
) else (
    echo [AVISO] No 1: Container nao esta rodando
    set BLOCK_NODE1=N/A
)

if "!NODE2_RUNNING!"=="true" (
    docker exec geth-peer geth attach --exec "eth.blockNumber" http://localhost:8545 > "%TEMP%\block2.txt" 2>nul
    if %ERRORLEVEL% EQU 0 (
        set /p BLOCK_NODE2=<"%TEMP%\block2.txt"
        set BLOCK_NODE2=!BLOCK_NODE2: =!
        echo [OK] No 2: RESPONDENDO (Bloco: !BLOCK_NODE2!)
    ) else (
        echo [ERRO] No 2: NAO RESPONDE
        set BLOCK_NODE2=N/A
    )
    del "%TEMP%\block2.txt" 2>nul
) else (
    echo [AVISO] No 2: Container nao esta rodando
    set BLOCK_NODE2=N/A
)
echo.

REM Verifica sincronização de blocos
echo [3/5] Verificando sincronizacao de blocos...
if not "!BLOCK_NODE1!"=="N/A" if not "!BLOCK_NODE2!"=="N/A" (
    REM Remove "0x" se existir
    set BLOCK1_CLEAN=!BLOCK_NODE1:0x=!
    set BLOCK2_CLEAN=!BLOCK_NODE2:0x=!
    
    REM Tenta converter para números (simplificado)
    set /a BLOCK1_DEC=!BLOCK1_CLEAN! 2>nul || set BLOCK1_DEC=0
    set /a BLOCK2_DEC=!BLOCK2_CLEAN! 2>nul || set BLOCK2_DEC=0
    
    set /a DIFF=!BLOCK1_DEC! - !BLOCK2_DEC!
    if !DIFF! LSS 0 set /a DIFF=-!DIFF!
    
    if !DIFF! LEQ 2 (
        echo [OK] Blocos sincronizados (Diferenca: !DIFF!)
    ) else (
        echo [AVISO] Blocos dessincronizados (Diferenca: !DIFF!)
        echo    No 1: Bloco !BLOCK1_DEC!
        echo    No 2: Bloco !BLOCK2_DEC!
    )
) else if not "!BLOCK_NODE1!"=="N/A" (
    echo [INFO] Apenas No 1 esta ativo (Bloco: !BLOCK_NODE1!)
) else if not "!BLOCK_NODE2!"=="N/A" (
    echo [INFO] Apenas No 2 esta ativo (Bloco: !BLOCK_NODE2!)
) else (
    echo [ERRO] Nenhum no esta respondendo
)
echo.

REM Verifica conexão P2P entre os nós
echo [4/5] Verificando conexao P2P...

if "!NODE1_RUNNING!"=="true" (
    docker exec geth-node geth attach --exec "admin.peers.length" http://localhost:8545 > "%TEMP%\peers1.txt" 2>nul
    set /p PEERS_NODE1=<"%TEMP%\peers1.txt" 2>nul
    del "%TEMP%\peers1.txt" 2>nul
    if "!PEERS_NODE1!"=="" set PEERS_NODE1=0
    set PEERS_NODE1=!PEERS_NODE1: =!
    
    if !PEERS_NODE1! GTR 0 (
        echo [OK] No 1: Conectado a !PEERS_NODE1! peer(s)
    ) else (
        echo [AVISO] No 1: Sem peers conectados
    )
) else (
    echo [AVISO] No 1: Container nao esta rodando
    set PEERS_NODE1=0
)

if "!NODE2_RUNNING!"=="true" (
    docker exec geth-peer geth attach --exec "admin.peers.length" http://localhost:8545 > "%TEMP%\peers2.txt" 2>nul
    set /p PEERS_NODE2=<"%TEMP%\peers2.txt" 2>nul
    del "%TEMP%\peers2.txt" 2>nul
    if "!PEERS_NODE2!"=="" set PEERS_NODE2=0
    set PEERS_NODE2=!PEERS_NODE2: =!
    
    if !PEERS_NODE2! GTR 0 (
        echo [OK] No 2: Conectado a !PEERS_NODE2! peer(s)
    ) else (
        echo [AVISO] No 2: Sem peers conectados
    )
) else (
    echo [AVISO] No 2: Container nao esta rodando
    set PEERS_NODE2=0
)
echo.

REM Verifica se os nós se enxergam mutuamente
echo [5/5] Verificando conectividade mutua...
if "!NODE1_RUNNING!"=="true" if "!NODE2_RUNNING!"=="true" (
    if !PEERS_NODE1! GTR 0 if !PEERS_NODE2! GTR 0 (
        echo [OK] CONEXAO ESTABELECIDA: Ambos os nos se enxergam!
    ) else if !PEERS_NODE1! GTR 0 (
        echo [AVISO] CONEXAO PARCIAL: Apenas No 1 ve o No 2
    ) else if !PEERS_NODE2! GTR 0 (
        echo [AVISO] CONEXAO PARCIAL: Apenas No 2 ve o No 1
    ) else (
        echo [ERRO] SEM CONEXAO: Os nos nao estao conectados
        echo.
        echo Dicas para resolver:
        echo   1. Verifique se ambos os nos estao na mesma rede (networkid=1337)
        echo   2. Verifique se o No 2 foi iniciado com --bootnodes apontando para o No 1
        echo   3. Execute: docker exec geth-peer geth attach http://localhost:8545 --exec "admin.addPeer('ENODE_DO_NO_1')"
    )
) else if "!NODE1_RUNNING!"=="true" (
    echo [INFO] Apenas No 1 esta ativo (modo standalone)
) else (
    echo [ERRO] Nenhum no esta rodando
)
echo.

REM Resumo final
echo ========================================
echo Resumo
echo ========================================
echo.
echo Status dos Containers:
if "!NODE1_RUNNING!"=="true" (
    echo   No 1 (geth-node): [OK] RODANDO
) else (
    echo   No 1 (geth-node): [ERRO] PARADO
)
if "!NODE2_RUNNING!"=="true" (
    echo   No 2 (geth-peer): [OK] RODANDO
) else (
    echo   No 2 (geth-peer): [AVISO] PARADO (opcional)
)
echo.
echo Endpoints RPC:
echo   No 1: http://localhost:8545
echo   No 2: http://localhost:8547
echo.
echo Comandos uteis:
echo   Ver peers do No 1: docker exec geth-node geth attach http://localhost:8545 --exec "admin.peers"
echo   Ver peers do No 2: docker exec geth-peer geth attach http://localhost:8545 --exec "admin.peers"
echo   Ver blocos: docker exec geth-node geth attach http://localhost:8545 --exec "eth.blockNumber"
echo   Ver enode do No 1: docker exec geth-node geth attach http://localhost:8545 --exec "admin.nodeInfo.enode"
echo.
pause
