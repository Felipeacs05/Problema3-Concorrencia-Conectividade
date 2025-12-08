@echo off
REM ===================== CONECTAR PEER =====================
REM Script para conectar manualmente o Nó 2 ao Nó 1

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set PROJECT_DIR=%SCRIPT_DIR%..

echo ========================================
echo Conectando No 2 ao No 1
echo ========================================
echo.

REM Verifica se os containers estão rodando
docker ps | findstr "geth-node" >nul
if %ERRORLEVEL% NEQ 0 (
    echo [ERRO] No 1 (geth-node) nao esta rodando!
    pause
    exit /b 1
)

docker ps | findstr "geth-peer" >nul
if %ERRORLEVEL% NEQ 0 (
    echo [ERRO] No 2 (geth-peer) nao esta rodando!
    pause
    exit /b 1
)

REM Obtém o enode do Nó 1
echo [1/3] Obtendo enode do No 1...
docker exec geth-node geth attach --exec "admin.nodeInfo.enode" /root/.ethereum/geth.ipc > "%TEMP%\enode.txt" 2>nul
set /p ENODE=<"%TEMP%\enode.txt"
del "%TEMP%\enode.txt" 2>nul

REM Remove aspas e espaços
set ENODE=!ENODE:"=!
set ENODE=!ENODE: =!

if "!ENODE!"=="" (
    echo [ERRO] Falha ao obter enode do No 1
    pause
    exit /b 1
)

REM Tenta descobrir o IP do container
for /f "tokens=*" %%a in ('docker inspect -f "{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}" geth-node 2^>nul') do set CONTAINER_IP=%%a

if not "!CONTAINER_IP!"=="" (
    REM Substitui o IP no enode
    set ENODE=!ENODE:@*=@!CONTAINER_IP!:
    echo    Usando IP do container: !CONTAINER_IP!
) else (
    REM Usa IP do host
    for /f "tokens=2 delims=:" %%a in ('ipconfig ^| findstr /i "IPv4"') do (
        set HOST_IP=%%a
        set HOST_IP=!HOST_IP: =!
        goto :IP_FOUND
    )
    set HOST_IP=127.0.0.1
    :IP_FOUND
    set ENODE=!ENODE:@*=@!HOST_IP!:
    echo    Usando IP do host: !HOST_IP!
)

echo [OK] Enode: !ENODE!
echo.

REM Conecta o Nó 2 ao Nó 1
echo [2/3] Conectando No 2 ao No 1...
docker exec geth-peer geth attach --exec "admin.addPeer('!ENODE!')" /root/.ethereum/geth.ipc > "%TEMP%\result.txt" 2>nul
set /p RESULT=<"%TEMP%\result.txt"
del "%TEMP%\result.txt" 2>nul
set RESULT=!RESULT: =!

if "!RESULT!"=="true" (
    echo [OK] Conexao estabelecida!
) else (
    echo [AVISO] Resultado: !RESULT!
    echo    Tentando metodo alternativo...
    
    REM Método alternativo
    docker exec geth-peer geth attach --exec "admin.addPeer('!ENODE!')" /root/.ethereum/geth.ipc > "%TEMP%\result2.txt" 2>nul
    set /p RESULT2=<"%TEMP%\result2.txt"
    del "%TEMP%\result2.txt" 2>nul
    set RESULT2=!RESULT2: =!
    
    if "!RESULT2!"=="true" (
        echo [OK] Conexao estabelecida (metodo alternativo)!
    ) else (
        echo [ERRO] Falha ao conectar. Verifique os logs:
        echo    docker logs geth-peer
        pause
        exit /b 1
    )
)
echo.

REM Aguarda sincronização
echo [3/3] Aguardando sincronizacao...
timeout /t 5 /nobreak >nul

REM Verifica conexão
docker exec geth-peer geth attach --exec "admin.peers.length" /root/.ethereum/geth.ipc > "%TEMP%\peers.txt" 2>nul
set /p PEERS=<"%TEMP%\peers.txt"
del "%TEMP%\peers.txt" 2>nul
set PEERS=!PEERS: =!

if !PEERS! GTR 0 (
    echo [OK] SUCESSO! No 2 esta conectado a !PEERS! peer(s)
    echo.
    echo Execute 'verificar-nos.bat' para ver o status completo
) else (
    echo [AVISO] Conexao pode nao ter sido estabelecida ainda
    echo    Aguarde alguns segundos e execute 'verificar-nos.bat' novamente
)
echo.
pause
