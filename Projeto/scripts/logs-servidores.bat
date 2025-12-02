@echo off
REM ===================== LOGS SERVIDORES =====================
REM Script para visualizar logs dos servidores de jogo

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set PROJECT_DIR=%SCRIPT_DIR%..
set JOGO_DIR=%PROJECT_DIR%\Jogo

echo ========================================
echo Logs dos Servidores de Jogo
echo ========================================
echo.
echo Escolha qual servidor ver os logs:
echo 1. Servidor 1
echo 2. Servidor 2
echo 3. Servidor 3
echo 4. Todos os servidores
echo 5. Broker 1
echo 6. Broker 2
echo 7. Broker 3
echo.
set /p OPCAO="Opção: "

cd /d "%JOGO_DIR%"

if "%OPCAO%"=="1" (
    echo.
    echo === Logs do Servidor 1 ===
    docker logs -f servidor1
) else if "%OPCAO%"=="2" (
    echo.
    echo === Logs do Servidor 2 ===
    docker logs -f servidor2
) else if "%OPCAO%"=="3" (
    echo.
    echo === Logs do Servidor 3 ===
    docker logs -f servidor3
) else if "%OPCAO%"=="4" (
    echo.
    echo === Logs de Todos os Servidores ===
    echo Pressione Ctrl+C para sair
    start "Servidor 1" cmd /k "docker logs -f servidor1"
    start "Servidor 2" cmd /k "docker logs -f servidor2"
    start "Servidor 3" cmd /k "docker logs -f servidor3"
) else if "%OPCAO%"=="5" (
    echo.
    echo === Logs do Broker 1 ===
    docker logs -f broker1
) else if "%OPCAO%"=="6" (
    echo.
    echo === Logs do Broker 2 ===
    docker logs -f broker2
) else if "%OPCAO%"=="7" (
    echo.
    echo === Logs do Broker 3 ===
    docker logs -f broker3
) else (
    echo Opção inválida!
    pause
    exit /b 1
)



