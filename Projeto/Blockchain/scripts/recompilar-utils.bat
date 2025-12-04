@echo off
REM ===================== RECOMPILAR BLOCKCHAIN-UTILS =====================
REM Script para recompilar o utilitário blockchain-utils
REM Use este script se o blockchain-utils.exe não estiver funcionando

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set TOOLS_DIR=%SCRIPT_DIR%..\tools

echo ========================================
echo Recompilando blockchain-utils
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

echo Go encontrado:
go version
echo.

cd /d "%TOOLS_DIR%"

REM Remove executável antigo
if exist "blockchain-utils.exe" (
    echo Removendo executavel antigo...
    del /q "blockchain-utils.exe"
    echo [OK] Executavel antigo removido
    echo.
)

REM Atualiza dependências
echo Atualizando dependencias...
go mod tidy
if %ERRORLEVEL% NEQ 0 (
    echo ERRO: Falha ao atualizar dependencias
    pause
    exit /b 1
)
echo [OK] Dependencias atualizadas
echo.

REM Compila explicitamente para Windows x64
echo Compilando para Windows x64...
set GOOS=windows
set GOARCH=amd64
go build -o blockchain-utils.exe blockchain-utils.go

REM Limpa variáveis de ambiente
set GOOS=
set GOARCH=

if not exist "blockchain-utils.exe" (
    echo ERRO: Falha ao compilar blockchain-utils
    pause
    exit /b 1
)

echo [OK] Compilacao concluida
echo.

REM Testa o executável
echo Testando executavel...
"%TOOLS_DIR%\blockchain-utils.exe" >nul 2>&1
if %ERRORLEVEL% EQU 1 (
    echo [OK] Executavel funcionando corretamente!
) else (
    echo [AVISO] Executavel compilado, mas teste retornou codigo %ERRORLEVEL%
    echo Isso pode ser normal se o executavel espera argumentos.
)

echo.
echo ========================================
echo Recompilacao concluida!
echo ========================================
echo.
echo Executavel criado em: %TOOLS_DIR%\blockchain-utils.exe
echo.
pause




