@echo off
REM ===================== CRIAR CONTA JOGADOR =====================
REM Script para criar uma nova conta (carteira) para um jogador

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set PROJECT_DIR=%SCRIPT_DIR%..
set BLOCKCHAIN_DIR=%PROJECT_DIR%\Blockchain
set KEYSTORE_DIR=%BLOCKCHAIN_DIR%\data\keystore
set TOOLS_DIR=%BLOCKCHAIN_DIR%\tools

echo ========================================
echo Criar Nova Conta de Jogador
echo ========================================
echo.

REM Verifica se Go está instalado
where go >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Go nao esta instalado!
    pause
    exit /b 1
)

REM Solicita senha
set /p SENHA="Digite uma senha para a nova conta: "
if "%SENHA%"=="" (
    echo ERRO: Senha nao pode ser vazia
    pause
    exit /b 1
)

REM Cria conta
echo Criando nova conta...
"%TOOLS_DIR%\blockchain-utils.exe" criar-conta "%KEYSTORE_DIR%" "%SENHA%"
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao criar conta
    pause
    exit /b 1
)

echo.
echo ========================================
echo Conta criada com sucesso!
echo ========================================
echo.
echo IMPORTANTE:
echo - Guarde o arquivo do keystore em local seguro
echo - Anote a senha escolhida
echo - O arquivo está em: %KEYSTORE_DIR%
echo.
echo Para usar esta conta no jogo:
echo 1. Copie o arquivo do keystore para o diretório do cliente
echo 2. Use o caminho do arquivo ao fazer login
echo.
pause

