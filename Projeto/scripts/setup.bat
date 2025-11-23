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
echo [1/5] Compilando utilitario blockchain-utils...
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
echo [2/5] Parando containers...
cd /d "%PROJECT_DIR%"
docker-compose down 2>nul
echo [OK] Containers parados
echo.

REM Remove dados antigos
echo [3/5] Removendo dados antigos...
if exist "%DATA_DIR%\geth" (
    rmdir /s /q "%DATA_DIR%\geth"
    echo [OK] Dados antigos removidos
) else (
    echo [OK] Nenhum dado antigo encontrado
)

REM Cria diretórios necessários
if not exist "%KEYSTORE_DIR%" mkdir "%KEYSTORE_DIR%"
echo.

REM Cria conta
echo [4/5] Criando nova conta...
"%TOOLS_DIR%\blockchain-utils.exe" criar-conta "%KEYSTORE_DIR%" "123456"
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao criar conta
    pause
    exit /b 1
)
echo.

REM Gera genesis.json
echo [5/5] Gerando genesis.json...
"%TOOLS_DIR%\blockchain-utils.exe" gerar-genesis "%KEYSTORE_DIR%" "%GENESIS_FILE%"
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao gerar genesis.json
    pause
    exit /b 1
)
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
docker-compose up -d geth
echo Aguardando inicializacao...
timeout /t 10 /nobreak >nul
echo.

echo ========================================
echo Configuracao concluida!
echo ========================================
echo.
echo Proximos passos:
echo   1. Desbloquear conta: scripts\unlock-account.bat
echo   2. Verificar blocos: scripts\check-block.bat
echo.

pause


