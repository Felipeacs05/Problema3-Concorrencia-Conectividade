@echo off
REM ===================== BAREMA ITEM 2: SMART CONTRACTS =====================
REM Script para compilar o contrato Solidity usando Docker

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set PROJECT_DIR=%SCRIPT_DIR%..
set CONTRACTS_DIR=%PROJECT_DIR%\contracts
set CONTRACT_FILE=%CONTRACTS_DIR%\GameEconomy.sol

echo ========================================
echo Compilando Contrato Solidity
echo ========================================
echo.

REM Verifica se o arquivo do contrato existe
if not exist "%CONTRACT_FILE%" (
    echo [ERRO] Arquivo do contrato nao encontrado: %CONTRACT_FILE%
    pause
    exit /b 1
)

echo [1/3] Verificando Docker...
docker --version >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [ERRO] Docker nao esta instalado ou nao esta rodando!
    pause
    exit /b 1
)
echo [OK] Docker encontrado
echo.

echo [2/3] Compilando GameEconomy.sol...
echo.

REM Compila o contrato usando Docker (gera .bin e .abi)
docker run --rm ^
    -v "%CONTRACTS_DIR%:/contracts" ^
    -w /contracts ^
    ethereum/solc:0.8.20 ^
    --bin --abi ^
    --optimize --optimize-runs 200 ^
    -o /contracts ^
    --overwrite ^
    GameEconomy.sol

if %ERRORLEVEL% NEQ 0 (
    echo [ERRO] Falha ao compilar contrato!
    pause
    exit /b 1
)

echo.
echo [3/3] Verificando arquivos gerados...

REM Aguarda um pouco para garantir que os arquivos foram escritos
timeout /t 2 /nobreak >nul 2>&1

REM Verifica se os arquivos foram criados
set BIN_OK=0
set ABI_OK=0

if exist "%CONTRACTS_DIR%\GameEconomy.bin" (
    echo [OK] Bytecode gerado: contracts\GameEconomy.bin
    set BIN_OK=1
)

if exist "%CONTRACTS_DIR%\GameEconomy.abi" (
    echo [OK] ABI gerado: contracts\GameEconomy.abi
    set ABI_OK=1
)

if %BIN_OK%==1 if %ABI_OK%==1 (
    echo.
    echo ========================================
    echo Compilacao concluida com sucesso!
    echo ========================================
    echo.
    echo Proximo passo: Fazer deploy do contrato
    echo   - Execute o cliente Go
    echo   - Escolha opcao 6 (Deploy do Contrato)
    echo.
    pause
    exit /b 0
) else (
    echo.
    echo [ERRO] Alguns arquivos nao foram gerados!
    if %BIN_OK%==0 echo   - GameEconomy.bin: FALTANDO
    if %ABI_OK%==0 echo   - GameEconomy.abi: FALTANDO
    echo.
    echo Verifique os erros de compilacao acima.
    pause
    exit /b 1
)
