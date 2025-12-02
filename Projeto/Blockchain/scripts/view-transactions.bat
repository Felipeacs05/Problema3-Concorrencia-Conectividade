@echo off
REM ===================== VIEW TRANSACTIONS =====================
REM Script para visualizar transações de uma conta na blockchain

setlocal

set SCRIPT_DIR=%~dp0
set TOOLS_DIR=%SCRIPT_DIR%..\tools

cd /d "%TOOLS_DIR%"

if "%1"=="" (
    echo Uso: view-transactions.bat ^<endereco_da_conta^> [bloco_inicial] [bloco_final]
    echo.
    echo Exemplos:
    echo   view-transactions.bat 0x2d50FD74Cc3eB554b16013667045804D28Bc54a4
    echo   view-transactions.bat 0x2d50FD74Cc3eB554b16013667045804D28Bc54a4 0 1000
    echo.
    pause
    exit /b 1
)

REM Compila se necessário
if not exist "view-transactions.exe" (
    echo Compilando view-transactions.exe...
    go build -o view-transactions.exe view-transactions.go
    if %ERRORLEVEL% NEQ 0 (
        echo [ERRO] Falha ao compilar view-transactions.go
        pause
        exit /b 1
    )
)

REM Executa
view-transactions.exe %*

pause

