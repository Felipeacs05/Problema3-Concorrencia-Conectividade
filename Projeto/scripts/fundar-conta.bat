@echo off
REM ===================== FUNDAR CONTA =====================
REM Script para transferir ETH para uma conta existente

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set PROJECT_DIR=%SCRIPT_DIR%..
set BLOCKCHAIN_DIR=%PROJECT_DIR%\Blockchain
set TOOLS_DIR=%BLOCKCHAIN_DIR%\tools

echo ========================================
echo Transferir ETH para Conta
echo ========================================
echo.

if "%1"=="" (
    echo Uso: fundar-conta.bat ^<endereco_da_conta^> [quantidade_em_ETH]
    echo.
    echo Exemplo:
    echo   fundar-conta.bat 0x0a8b6dd6F3A38D8Cc5Dbb872A83F1a24B7e49E1f
    echo   fundar-conta.bat 0x0a8b6dd6F3A38D8Cc5Dbb872A83F1a24B7e49E1f 100
    echo.
    echo Se a quantidade nao for especificada, sera transferido 100 ETH.
    echo.
    pause
    exit /b 1
)

set ENDERECO=%1
set QUANTIDADE=100
if not "%2"=="" set QUANTIDADE=%2

REM Verifica se fund-account-rpc.exe existe
cd /d "%TOOLS_DIR%"
if exist "fund-account-rpc.exe" (
    echo [INFO] Usando metodo RPC - conta desbloqueada no Geth...
    fund-account-rpc.exe %ENDERECO% %QUANTIDADE%
    goto :end
)

REM Se não existe, tenta compilar
echo [AVISO] fund-account-rpc.exe nao encontrado, compilando...
go build -o fund-account-rpc.exe fund-account-rpc.go
if %ERRORLEVEL% EQU 0 (
    if exist "fund-account-rpc.exe" (
        echo [INFO] Usando metodo RPC - conta desbloqueada no Geth...
        fund-account-rpc.exe %ENDERECO% %QUANTIDADE%
        goto :end
    )
)

REM Fallback para fund-account.exe
echo [AVISO] Usando metodo keystore...
if not exist "fund-account.exe" (
    echo Compilando fund-account.exe...
    go build -o fund-account.exe fund-account.go
    if %ERRORLEVEL% NEQ 0 (
        echo [ERRO] Falha ao compilar fund-account.go
        pause
        exit /b 1
    )
)

echo Transferindo %QUANTIDADE% ETH para: %ENDERECO%
echo.
fund-account.exe %ENDERECO% %QUANTIDADE%

:end
if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo Transferencia concluida com sucesso!
    echo ========================================
) else (
    echo.
    echo [ERRO] Falha na transferencia!
    echo Verifique se:
    echo - O Geth esta rodando (docker ps)
    echo - O endereco esta correto
    echo - A conta do servidor tem ETH suficiente
)

pause
