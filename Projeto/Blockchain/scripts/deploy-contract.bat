@echo off
REM ===================== DEPLOY DO CONTRATO =====================
REM Script para fazer deploy do contrato GameEconomy na blockchain

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set TOOLS_DIR=%SCRIPT_DIR%..\tools

echo ========================================
echo Deploy do Contrato GameEconomy
echo ========================================
echo.

REM Verifica se o Geth está rodando
echo [1/3] Verificando se o Geth está rodando...
docker ps | findstr "geth-node" >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [ERRO] Geth não está rodando!
    echo Execute: docker-compose up -d
    pause
    exit /b 1
)
echo [OK] Geth está rodando
echo.

REM Verifica se o bytecode existe
echo [2/3] Verificando se o contrato foi compilado...
if not exist "%SCRIPT_DIR%..\contracts\GameEconomy.bin" (
    echo [ERRO] Contrato não compilado!
    echo Execute: compile-contract.bat primeiro
    pause
    exit /b 1
)
echo [OK] Bytecode encontrado
echo.

REM Executa o deploy
echo [3/3] Executando deploy...
cd /d "%TOOLS_DIR%"
if not exist "deploy-contract.exe" (
    echo [ERRO] deploy-contract.exe não encontrado!
    echo Compilando...
    go build -o deploy-contract.exe deploy-contract.go
    if %ERRORLEVEL% NEQ 0 (
        echo [ERRO] Falha ao compilar deploy-contract.go
        pause
        exit /b 1
    )
)

.\deploy-contract.exe

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo Deploy concluído com sucesso!
    echo ========================================
    echo.
    echo Agora você pode usar o cliente do jogo!
) else (
    echo.
    echo [ERRO] Deploy falhou!
    echo Verifique os erros acima.
)

pause


