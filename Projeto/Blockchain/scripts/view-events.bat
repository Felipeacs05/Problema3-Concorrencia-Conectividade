@echo off
REM ===================== VIEW EVENTS =====================
REM Script para visualizar eventos do contrato GameEconomy na blockchain

setlocal

set SCRIPT_DIR=%~dp0
set PROJECT_DIR=%SCRIPT_DIR%..
set TOOLS_DIR=%PROJECT_DIR%\Blockchain\tools

cd /d "%TOOLS_DIR%"

if "%1"=="" (
    echo Uso: view-events.bat ^<endereco_do_contrato^> [bloco_inicial] [bloco_final]
    echo.
    echo Exemplos:
    echo   view-events.bat 0x4D82F25Ef5058fE2308135D88A386c90FEdBe99A
    echo   view-events.bat 0x4D82F25Ef5058fE2308135D88A386c90FEdBe99A 0 1000
    echo.
    echo Para obter o endereco do contrato, veja o arquivo:
    echo   Blockchain/contract-address.txt
    echo.
    pause
    exit /b 1
)

REM Compila se necessário
if not exist "view-events.exe" (
    echo Compilando view-events.exe...
    go build -o view-events.exe view-events.go
    if %ERRORLEVEL% NEQ 0 (
        echo [ERRO] Falha ao compilar view-events.go
        pause
        exit /b 1
    )
)

REM Executa
view-events.exe %*

pause

