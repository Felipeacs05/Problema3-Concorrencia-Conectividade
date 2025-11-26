@echo off
REM ===================== BAREMA ITEM 1: ARQUITETURA =====================
REM Script para enviar transação de teste e forçar criação do primeiro bloco

echo ========================================
echo Enviando transacao de teste...
echo ========================================
echo.
echo Este script envia uma transacao simples para forcar
echo a criacao do primeiro bloco no Clique (PoA).
echo.

cd ..\tools
send-tx.exe

if %errorlevel% neq 0 (
    echo.
    echo [ERRO] Falha ao enviar transacao!
    pause
    exit /b 1
)

echo.
echo ========================================
echo Transacao enviada com sucesso!
echo ========================================
echo.
echo Verifique os blocos com: check-block.bat
echo.
pause

