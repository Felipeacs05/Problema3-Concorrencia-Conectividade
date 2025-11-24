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

cd ..
echo Compilando cliente...
go build -o cliente/cliente.exe ./cliente/main.go
if %ERRORLEVEL% NEQ 0 (
    echo [ERRO] Falha na compilacao!
    pause
    exit /b 1
)
cd scripts

..\cliente\cliente.exe

pause


