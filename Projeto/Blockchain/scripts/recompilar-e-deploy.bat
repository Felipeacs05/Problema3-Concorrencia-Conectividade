@echo off
REM ===================== RECOMPILAR E REDEPLOY =====================
REM Script que recompila o contrato e faz deploy automaticamente

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set PROJECT_DIR=%SCRIPT_DIR%..

echo ========================================
echo RECOMPILAR E REDEPLOY DO CONTRATO
echo ========================================
echo.
echo Este script vai:
echo   1. Recompilar o contrato GameEconomy.sol
echo   2. Fazer deploy do novo contrato
echo   3. Atualizar contract-address.txt automaticamente
echo.
echo ATENCAO: O endereco do contrato vai mudar!
echo   - Voce precisara atualizar o servidor com o novo endereco
echo   - Ou reiniciar o servidor para ele ler o novo contract-address.txt
echo.
pause

REM Passo 1: Compilar
echo.
echo ========================================
echo [PASSO 1/2] Compilando contrato...
echo ========================================
call "%SCRIPT_DIR%compile-contract.bat"
if %ERRORLEVEL% NEQ 0 (
    echo.
    echo [ERRO] Compilacao falhou!
    pause
    exit /b 1
)

echo.
echo ========================================
echo [PASSO 2/2] Fazendo deploy...
echo ========================================
call "%SCRIPT_DIR%deploy-contract.bat"
if %ERRORLEVEL% NEQ 0 (
    echo.
    echo [ERRO] Deploy falhou!
    pause
    exit /b 1
)

echo.
echo ========================================
echo CONCLUIDO!
echo ========================================
echo.
echo O contrato foi recompilado e redeployado com sucesso!
echo.
echo PROXIMOS PASSOS:
echo   1. Verifique o novo endereco em: contract-address.txt
echo   2. Reinicie o servidor do jogo para ele usar o novo contrato
echo   3. Teste novamente o registro de partidas
echo.
pause
