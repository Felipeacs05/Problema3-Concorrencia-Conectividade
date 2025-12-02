@echo off
REM ===================== RUN CLIENTE =====================
REM Script para executar o cliente do jogo (Windows)

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set PROJECT_DIR=%SCRIPT_DIR%..
set JOGO_DIR=%PROJECT_DIR%\Jogo
set CLIENTE_DIR=%JOGO_DIR%\cliente

echo ========================================
echo Executando Cliente do Jogo
echo ========================================
echo.

REM Verifica se o executável existe
if exist "%CLIENTE_DIR%\cliente.exe" (
    cd /d "%CLIENTE_DIR%"
    echo Executando cliente...
    echo.
    .\cliente.exe
) else (
    echo ERRO: cliente.exe nao encontrado!
    echo Execute setup-game.bat primeiro para compilar o cliente.
    echo.
    pause
    exit /b 1
)



