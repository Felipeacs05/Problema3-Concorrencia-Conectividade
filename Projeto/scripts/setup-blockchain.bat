@echo off
REM ===================== SETUP BLOCKCHAIN =====================
REM Script para configurar a blockchain privada Ethereum
REM Este script deve ser executado antes de iniciar o jogo

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set PROJECT_DIR=%SCRIPT_DIR%..
set BLOCKCHAIN_DIR=%PROJECT_DIR%\Blockchain
set TOOLS_DIR=%BLOCKCHAIN_DIR%\tools
set DATA_DIR=%BLOCKCHAIN_DIR%\data
set KEYSTORE_DIR=%DATA_DIR%\keystore
set GENESIS_FILE=%BLOCKCHAIN_DIR%\genesis.json

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
echo [1/9] Compilando utilitario blockchain-utils...
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
echo [2/9] Parando containers blockchain...
cd /d "%BLOCKCHAIN_DIR%"
docker-compose -f docker-compose-blockchain.yml down 2>nul
echo [OK] Containers parados
echo.

REM Remove dados antigos
echo [3/9] Removendo dados antigos...
set REMOVIDO=0

if exist "%DATA_DIR%\geth" (
    rmdir /s /q "%DATA_DIR%\geth"
    set REMOVIDO=1
)

if exist "%KEYSTORE_DIR%" (
    rmdir /s /q "%KEYSTORE_DIR%"
    set REMOVIDO=1
)

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
echo [4/9] Criando nova conta...
"%TOOLS_DIR%\blockchain-utils.exe" criar-conta "%KEYSTORE_DIR%" "123456"
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao criar conta
    pause
    exit /b 1
)

REM Cria arquivo password.txt
(
echo 123456
) > "%DATA_DIR%\password.txt"
echo.

REM Gera genesis.json
echo [5/9] Gerando genesis.json...
"%TOOLS_DIR%\blockchain-utils.exe" gerar-genesis "%KEYSTORE_DIR%" "%GENESIS_FILE%"
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao gerar genesis.json
    pause
    exit /b 1
)
echo [OK] Genesis.json gerado
echo.

REM Extrai endereço da conta criada
echo [6/9] Extraindo endereco da conta...
"%TOOLS_DIR%\blockchain-utils.exe" extrair-endereco "%KEYSTORE_DIR%" > "%TEMP%\blockchain-address.txt"
set /p ADDRESS=<"%TEMP%\blockchain-address.txt"
del "%TEMP%\blockchain-address.txt" 2>nul
if "%ADDRESS%"=="" (
    echo ERRO: Falha ao extrair endereco
    pause
    exit /b 1
)
echo [OK] Endereco extraido: %ADDRESS%
echo.

REM Atualiza docker-compose.yml com endereço da conta
echo [7/9] Atualizando docker-compose.yml...
"%TOOLS_DIR%\blockchain-utils.exe" atualizar-docker-compose "%ADDRESS%" "%BLOCKCHAIN_DIR%\docker-compose-blockchain.yml"
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao atualizar docker-compose.yml
    pause
    exit /b 1
)
echo [OK] Docker-compose.yml atualizado
echo.

REM Inicializa blockchain (geth init)
echo [8/9] Inicializando blockchain com genesis.json...
cd /d "%BLOCKCHAIN_DIR%"
if exist "%DATA_DIR%\geth" (
    echo Removendo dados antigos da blockchain...
    rmdir /s /q "%DATA_DIR%\geth"
)
docker run --rm -v "%DATA_DIR%:/root/.ethereum" -v "%GENESIS_FILE%:/genesis.json" ethereum/client-go:v1.13.15 --datadir=/root/.ethereum init /genesis.json
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao inicializar blockchain
    pause
    exit /b 1
)
echo [OK] Blockchain inicializada
echo.

REM Inicia containers
echo [9/10] Iniciando containers blockchain...
docker-compose -f docker-compose-blockchain.yml up -d
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao iniciar containers
    pause
    exit /b 1
)
echo [OK] Containers iniciados
echo.

REM Aguarda Geth estar pronto
echo Aguardando Geth estar pronto...
:WAIT_LOOP
timeout /t 2 /nobreak >nul
docker exec geth-node geth attach --exec "eth.blockNumber" http://localhost:8545 >nul 2>&1
if %errorlevel% neq 0 (
    echo Aguardando porta RPC 8545...
    goto WAIT_LOOP
)
echo [OK] Geth esta pronto
echo.

REM Desbloqueia conta
echo Desbloqueando conta...
docker exec geth-node geth attach --exec "personal.unlockAccount(eth.accounts[0], '123456', 0)" http://localhost:8545 >nul 2>&1
echo [OK] Conta desbloqueada
echo.

REM Faz deploy do contrato
echo [10/10] Fazendo deploy do contrato...
cd /d "%BLOCKCHAIN_DIR%\scripts"

REM Verifica se o contrato foi compilado
if not exist "%BLOCKCHAIN_DIR%\contracts\GameEconomy.bin" (
    echo [AVISO] Contrato nao compilado. Compilando...
    call compile-contract.bat
    if %ERRORLEVEL% NEQ 0 (
        echo [AVISO] Falha ao compilar contrato. Execute compile-contract.bat manualmente.
        goto :skip_deploy
    )
)

REM Executa deploy
if exist "deploy-contract.bat" (
    call deploy-contract.bat
    if %ERRORLEVEL% EQU 0 (
        echo [OK] Contrato deployado com sucesso
    ) else (
        echo [AVISO] Deploy pode ter falhado. Execute deploy-contract.bat manualmente.
    )
) else (
    echo [AVISO] Script de deploy nao encontrado. Execute deploy-contract.bat manualmente.
)
:skip_deploy
echo.

echo ========================================
echo Blockchain configurada com sucesso!
echo ========================================
echo.
echo Próximos passos:
echo 1. Execute setup-game.bat para configurar o jogo
echo 2. Execute start-all.bat para iniciar tudo
echo.
pause

