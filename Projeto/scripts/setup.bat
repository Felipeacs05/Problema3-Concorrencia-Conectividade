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
echo [1/6] Compilando utilitario blockchain-utils...
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
echo [2/6] Parando containers...
cd /d "%PROJECT_DIR%"
docker-compose down 2>nul
echo [OK] Containers parados
echo.

REM Remove dados antigos
echo [3/6] Removendo dados antigos...
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
echo [4/6] Criando nova conta...
"%TOOLS_DIR%\blockchain-utils.exe" criar-conta "%KEYSTORE_DIR%" "123456"
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao criar conta
    pause
    exit /b 1
)

REM Cria arquivo password.txt (necessário para o docker-compose)
echo 123456 > "%DATA_DIR%\password.txt"
echo.

REM Gera genesis.json
echo [5/6] Gerando genesis.json...
"%TOOLS_DIR%\blockchain-utils.exe" gerar-genesis "%KEYSTORE_DIR%" "%GENESIS_FILE%"
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao gerar genesis.json
    pause
    exit /b 1
)
echo.

REM Extrai endereço e atualiza docker-compose.yml
echo [6/6] Atualizando docker-compose.yml...
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


