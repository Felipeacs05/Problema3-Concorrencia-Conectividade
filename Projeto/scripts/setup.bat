@echo off
REM ===================== BAREMA ITEM 1: ARQUITETURA =====================
REM Script cross-platform para configurar a blockchain (Windows)
REM Funciona em conjunto com o utilitário Go blockchain-utils

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set PROJECT_DIR=%SCRIPT_DIR%..
set TOOLS_DIR=%PROJECT_DIR%\tools
set DATA_DIR=%PROJECT_DIR%\data
set KEYSTORE_DIR=%DATA_DIR%\keystore
set GENESIS_FILE=%PROJECT_DIR%\genesis.json

echo ========================================
echo Configurando Blockchain Privada
echo ========================================
echo.

REM Verifica se Go está instalado
where go >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Go nao esta instalado!
    echo Instale Go: https://golang.org/dl/
    pause
    exit /b 1
)

REM Compila o utilitário blockchain-utils
echo [1/7] Compilando utilitario blockchain-utils...
cd /d "%TOOLS_DIR%"
go mod tidy
go build -o blockchain-utils.exe blockchain-utils.go
if not exist "%TOOLS_DIR%\blockchain-utils.exe" (
    echo ERRO: Falha ao compilar blockchain-utils
    pause
    exit /b 1
)
echo [OK] Utilitario compilado
echo.

REM Para containers se estiverem rodando
echo [2/7] Parando containers...
cd /d "%PROJECT_DIR%"
docker-compose down 2>nul
echo [OK] Containers parados
echo.

REM Remove dados antigos
echo [3/7] Removendo dados antigos...
set REMOVIDO=0

REM Remove blockchain antiga
if exist "%DATA_DIR%\geth" (
    rmdir /s /q "%DATA_DIR%\geth"
    set REMOVIDO=1
)

REM Remove contas antigas do keystore
if exist "%KEYSTORE_DIR%" (
    rmdir /s /q "%KEYSTORE_DIR%"
    set REMOVIDO=1
)

REM Remove password.txt antigo (será recriado)
if exist "%DATA_DIR%\password.txt" (
    del /q "%DATA_DIR%\password.txt"
    set REMOVIDO=1
)

if %REMOVIDO%==1 (
    echo [OK] Dados antigos removidos
) else (
    echo [OK] Nenhum dado antigo encontrado
)

REM Cria diretórios necessários
if not exist "%KEYSTORE_DIR%" mkdir "%KEYSTORE_DIR%"
if not exist "%DATA_DIR%" mkdir "%DATA_DIR%"
echo.

REM Cria conta
echo [4/7] Criando nova conta...
"%TOOLS_DIR%\blockchain-utils.exe" criar-conta "%KEYSTORE_DIR%" "123456"
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao criar conta
    pause
    exit /b 1
)

REM Cria arquivo password.txt (necessário para o docker-compose)
REM Usa método confiável para criar arquivo sem quebra de linha extra
(
echo 123456
) > "%DATA_DIR%\password.txt"
echo.

REM Gera genesis.json
echo [5/7] Gerando genesis.json...
"%TOOLS_DIR%\blockchain-utils.exe" gerar-genesis "%KEYSTORE_DIR%" "%GENESIS_FILE%"
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao gerar genesis.json
    pause
    exit /b 1
)
echo.

REM Extrai endereço e atualiza docker-compose.yml
echo [6/7] Atualizando docker-compose.yml...
set ENDERECO=
cd /d "%TOOLS_DIR%"
for /f "delims=" %%a in ('blockchain-utils.exe extrair-endereco "%KEYSTORE_DIR%"') do set ENDERECO=%%a
cd /d "%PROJECT_DIR%"
if "!ENDERECO!"=="" (
    echo ERRO: Falha ao extrair endereco
    pause
    exit /b 1
)
cd /d "%TOOLS_DIR%"
blockchain-utils.exe atualizar-docker-compose "!ENDERECO!" "%PROJECT_DIR%\docker-compose.yml"
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao atualizar docker-compose.yml
    pause
    exit /b 1
)
cd /d "%PROJECT_DIR%"
echo [OK] docker-compose.yml atualizado com endereco: !ENDERECO!
echo.

REM Verifica se Docker está rodando
echo [7/7] Verificando Docker e baixando imagem do Geth...
docker ps >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [ERRO] Docker nao esta rodando ou nao esta acessivel
    echo.
    echo Solucoes:
    echo   1. Inicie o Docker Desktop
    echo   2. Verifique se o Docker esta instalado corretamente
    pause
    exit /b 1
)
echo [OK] Docker esta rodando

REM Faz pull da imagem Docker com retry
set RETRY_COUNT=0
set MAX_RETRIES=3
:PULL_IMAGE
set /a ATTEMPT_NUM=%RETRY_COUNT%+1
echo Baixando imagem ethereum/client-go:latest (tentativa !ATTEMPT_NUM!/%MAX_RETRIES%)...
docker pull ethereum/client-go:latest
if %ERRORLEVEL% EQU 0 (
    echo [OK] Imagem baixada com sucesso
    goto PULL_SUCCESS
)
set /a RETRY_COUNT+=1
if !RETRY_COUNT! LSS %MAX_RETRIES% (
    echo [AVISO] Falha ao baixar imagem (tentativa !RETRY_COUNT!/%MAX_RETRIES%). Tentando novamente em 5 segundos...
    timeout /t 5 /nobreak >nul
    goto PULL_IMAGE
) else (
    echo [ERRO] Falha ao baixar imagem Docker apos %MAX_RETRIES% tentativas
    echo.
    echo Possiveis causas:
    echo   - Problema de conexao com a internet
    echo   - Docker Hub indisponivel
    echo   - Firewall bloqueando conexao
    echo.
    echo Solucoes:
    echo   1. Verifique sua conexao com a internet
    echo   2. Tente executar manualmente: docker pull ethereum/client-go:latest
    echo   3. Verifique configuracoes de proxy/firewall
    pause
    exit /b 1
)
:PULL_SUCCESS
echo.

REM Inicializa blockchain
echo Inicializando blockchain...
docker-compose run --rm geth --datadir /root/.ethereum init /genesis.json
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao inicializar blockchain
    pause
    exit /b 1
)
echo.

REM Inicia nó
echo Iniciando no Geth...
docker-compose up -d

echo Aguardando inicializacao (pode levar alguns segundos)...
timeout /t 10 /nobreak >nul

:WAIT_LOOP
docker exec geth-node geth attach --exec "eth.blockNumber" http://localhost:8545 >nul 2>&1
if %errorlevel% neq 0 (
    echo Aguardando porta RPC 8545...
    timeout /t 2 /nobreak >nul
    goto WAIT_LOOP
)

echo.
echo ========================================
echo Configuracao concluida
echo ========================================
echo.
echo Proximos passos:
echo   1. Desbloquear conta: scripts\unlock-account.bat
echo   2. Verificar blocos: scripts\check-block.bat
echo.

pause


