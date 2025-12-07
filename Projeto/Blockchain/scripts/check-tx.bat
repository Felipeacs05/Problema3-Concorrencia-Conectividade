@echo off
REM ===================== CHECK TRANSACTION =====================
REM Script para verificar eventos de uma transação específica

setlocal

set SCRIPT_DIR=%~dp0
set TOOLS_DIR=%SCRIPT_DIR%..\tools

cd /d "%TOOLS_DIR%"

if "%1"=="" (
    echo Uso: check-tx.bat ^<hash_da_transacao^>
    echo.
    echo Exemplo:
    echo   check-tx.bat 0x80faebe1ac84d2c4fb86ac26d84987436fc2599ed057d99b6628c5f9d5c2f51e
    echo.
    pause
    exit /b 1
)

REM Compila se necessário
if not exist "check-tx.exe" (
    echo Compilando check-tx.exe...
    go build -o check-tx.exe check-tx.go
    if %ERRORLEVEL% NEQ 0 (
        echo [ERRO] Falha ao compilar check-tx.go
        pause
        exit /b 1
    )
)

REM Executa
check-tx.exe %*

pause
