@echo off
REM ===================== BAREMA ITEM 1: ARQUITETURA =====================
REM Script para testar o cliente Go principal

echo ========================================
echo Testando cliente Go principal...
echo ========================================
echo.
echo NOTA: O cliente Go acessa o keystore diretamente
echo       e deve conseguir enviar transacoes corretamente.
echo.
echo Pressione qualquer tecla para continuar...
pause >nul

cd ..\cliente
if not exist cliente.exe (
    echo [ERRO] cliente.exe nao encontrado!
    echo Compile primeiro com: go build -o cliente.exe main.go
    pause
    exit /b 1
)
cliente.exe

pause


