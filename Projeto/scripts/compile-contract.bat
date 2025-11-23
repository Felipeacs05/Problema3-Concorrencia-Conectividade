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

REM Compila o contrato usando Docker
docker run --rm ^
    -v "%CONTRACTS_DIR%:/contracts" ^
    -w /contracts ^
    ethereum/solc:0.8.20 ^
    --bin GameEconomy.sol ^
    -o /contracts

if %ERRORLEVEL% NEQ 0 (
    echo [ERRO] Falha ao compilar contrato!
    pause
    exit /b 1
)

echo.
echo [3/3] Verificando arquivo gerado...

REM Aguarda um pouco para garantir que o arquivo foi escrito
timeout /t 2 /nobreak >nul 2>&1

REM Verifica se o arquivo .bin foi criado
if exist "%CONTRACTS_DIR%\GameEconomy.bin" (
    echo [OK] Bytecode gerado: contracts\GameEconomy.bin
    echo.
    echo ========================================
    echo Compilacao concluida com sucesso!
    echo ========================================
    echo.
    echo Proximo passo: Fazer deploy do contrato
    echo   - Execute o cliente Go
    echo   - Escolha opcao 6 (Deploy do Contrato)
    echo.
) else (
    echo [AVISO] Verificando arquivos gerados...
    dir "%CONTRACTS_DIR%\*.bin" 2>nul
    if exist "%CONTRACTS_DIR%\GameEconomy.bin" (
        echo [OK] Arquivo encontrado apos verificacao!
    ) else (
        echo [ERRO] Arquivo GameEconomy.bin nao foi gerado!
        echo Verifique os erros de compilacao acima.
        pause
        exit /b 1
    )
)

pause

