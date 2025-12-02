@echo off
REM ===================== SETUP GAME =====================
REM Script para configurar o jogo distribuído
REM Este script deve ser executado após setup-blockchain.bat

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set PROJECT_DIR=%SCRIPT_DIR%..
set JOGO_DIR=%PROJECT_DIR%\Jogo

echo ========================================
echo Configurando Jogo Distribuído
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

REM Compila servidor
echo [1/3] Compilando servidor...
cd /d "%JOGO_DIR%\servidor"
go mod tidy
go build -o servidor.exe main.go
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao compilar servidor
    pause
    exit /b 1
)
echo [OK] Servidor compilado
echo.

REM Compila cliente (Windows)
echo [2/3] Compilando cliente (Windows)...
cd /d "%JOGO_DIR%\cliente"
go mod tidy
go build -o cliente.exe main.go blockchain_client.go
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao compilar cliente
    pause
    exit /b 1
)
echo [OK] Cliente compilado (Windows: cliente.exe)
echo.

REM Verifica se o contrato foi deployado
echo [3/3] Verificando contrato blockchain...
if exist "%PROJECT_DIR%\contract-address.txt" (
    echo [OK] Contrato encontrado
    type "%PROJECT_DIR%\contract-address.txt"
) else (
    echo [AVISO] Contrato nao encontrado. Execute setup-blockchain.bat primeiro.
)
echo.

echo ========================================
echo Jogo configurado com sucesso!
echo ========================================
echo.
echo Próximos passos:
echo 1. Execute start-all.bat para iniciar tudo
echo.
pause

