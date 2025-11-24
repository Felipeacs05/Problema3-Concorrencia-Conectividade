@echo off
REM ===================== BAREMA ITEM 1: ARQUITETURA =====================
REM Script para desbloquear conta no Clique (Windows)

setlocal enabledelayedexpansion

echo ========================================
echo Desbloqueando conta do signer...
echo ========================================
echo.

REM Aguarda o Geth estar pronto
echo Aguardando Geth estar pronto...
set MAX_TENTATIVAS=30
set TENTATIVA=0

:WAIT_LOOP
docker exec geth-node geth attach --exec "eth.blockNumber" http://localhost:8545 >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [OK] Geth esta pronto!
    goto UNLOCK
)

set /a TENTATIVA+=1
if %TENTATIVA% GEQ %MAX_TENTATIVAS% (
    echo.
    echo [ERRO] Timeout aguardando Geth estar pronto!
    echo Verifique se o container esta rodando: docker ps
    pause
    exit /b 1
)

echo Aguardando... (%TENTATIVA%/%MAX_TENTATIVAS%)
timeout /t 2 /nobreak >nul
goto WAIT_LOOP

:UNLOCK
echo.
echo Desbloqueando conta...
docker exec geth-node geth attach --exec "personal.unlockAccount(eth.accounts[0], '123456', 0)" http://localhost:8545

if %ERRORLEVEL% EQU 0 (
    echo.
    echo [OK] Comando enviado com sucesso.
    echo Se retornou 'true', a conta esta desbloqueada.
) else (
    echo.
    echo [ERRO] Falha ao desbloquear conta.
)

echo.
echo ========================================
echo Clique pronto para selar blocos
echo ========================================
echo.

pause