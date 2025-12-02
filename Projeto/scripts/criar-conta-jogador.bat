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

REM Cria conta e captura o endereço
echo Criando nova conta...
set TEMP_FILE=%TEMP%\endereco_%RANDOM%.txt
"%TOOLS_DIR%\blockchain-utils.exe" criar-conta "%KEYSTORE_DIR%" "%SENHA%" > "%TEMP_FILE%" 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao criar conta
    del "%TEMP_FILE%" >nul 2>&1
    pause
    exit /b 1
)

REM Extrai o endereço da saída (procura por "Endereço: 0x...")
set ENDERECO=
for /f "usebackq tokens=*" %%a in ("%TEMP_FILE%") do (
    echo %%a | findstr /C:"Endereço:" >nul 2>&1
    if !ERRORLEVEL! EQU 0 (
        REM Extrai o endereço (remove "Endereço: " e espaços)
        for /f "tokens=2" %%b in ("%%a") do set ENDERECO=%%b
    )
)
del "%TEMP_FILE%" >nul 2>&1

REM Remove espaços do endereço
set ENDERECO=%ENDERECO: =%

if "%ENDERECO%"=="" (
    echo [AVISO] Endereco vazio, pulando transferencia de ETH
    goto :skip_fund
)

REM Verifica se fund-account.exe existe
if not exist "%TOOLS_DIR%\fund-account.exe" (
    echo [AVISO] fund-account.exe nao encontrado, compilando...
    cd /d "%TOOLS_DIR%"
    go build -o fund-account.exe fund-account.go
    if %ERRORLEVEL% NEQ 0 (
        echo [AVISO] Falha ao compilar fund-account.go
        goto :skip_fund
    )
)

REM Transfere ETH para a nova conta (100 ETH)
echo.
echo Transferindo 100 ETH para a nova conta %ENDERECO%...
cd /d "%TOOLS_DIR%"
"%TOOLS_DIR%\fund-account.exe" %ENDERECO% 100
if %ERRORLEVEL% EQU 0 (
    echo [OK] ETH transferido com sucesso!
) else (
    echo [AVISO] Falha ao transferir ETH automaticamente.
    echo.
    echo Para transferir manualmente, execute:
    echo   cd %TOOLS_DIR%
    echo   .\fund-account.exe %ENDERECO% 100
)

:skip_fund
echo.
echo ========================================
echo Conta criada com sucesso!
echo ========================================
if not "%ENDERECO%"=="" (
    echo.
    echo Endereco: %ENDERECO%
    echo Senha: %SENHA%
    echo.
)
echo IMPORTANTE:
echo - Guarde o arquivo do keystore em local seguro
echo - Anote a senha escolhida
echo - O arquivo esta em: %KEYSTORE_DIR%
echo.
echo Para usar esta conta no jogo:
echo 1. Execute run-cliente.bat
echo 2. Quando perguntar, digite 's' para conectar carteira
echo 3. Digite a senha: %SENHA%
echo.
pause

