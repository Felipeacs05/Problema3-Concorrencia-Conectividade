@echo off
REM ===================== FORCE DEPLOY =====================
REM Script para forcar recompilacao e deploy do contrato

setlocal enabledelayedexpansion

set SCRIPT_DIR=%~dp0
set TOOLS_DIR=%SCRIPT_DIR%..\tools
set CONTRACTS_DIR=%SCRIPT_DIR%..\contracts

echo ========================================
echo FORCE DEPLOY - Recompila e Deploya
echo ========================================
echo.

REM Apaga arquivos antigos
echo [1/5] Removendo arquivos antigos...
if exist "%CONTRACTS_DIR%\GameEconomy.bin" del "%CONTRACTS_DIR%\GameEconomy.bin"
if exist "%CONTRACTS_DIR%\GameEconomy.abi" del "%CONTRACTS_DIR%\GameEconomy.abi"
if exist "%TOOLS_DIR%\deploy-contract.exe" del "%TOOLS_DIR%\deploy-contract.exe"
if exist "%TOOLS_DIR%\deploy-contract" del "%TOOLS_DIR%\deploy-contract"
echo [OK] Arquivos antigos removidos
echo.

REM Recompila o contrato
echo [2/5] Recompilando contrato...
docker run --rm -v "%CONTRACTS_DIR%:/contracts" -w /contracts ethereum/solc:0.8.20 --bin --abi --evm-version paris --optimize --optimize-runs 200 -o /contracts --overwrite GameEconomy.sol
if %ERRORLEVEL% NEQ 0 (
    echo [ERRO] Falha na compilacao!
    pause
    exit /b 1
)
echo [OK] Contrato compilado
echo.

REM Verifica se os arquivos foram gerados
echo [3/5] Verificando arquivos gerados...
timeout /t 2 /nobreak >nul
if not exist "%CONTRACTS_DIR%\GameEconomy.bin" (
    echo [ERRO] GameEconomy.bin nao foi gerado!
    pause
    exit /b 1
)
if not exist "%CONTRACTS_DIR%\GameEconomy.abi" (
    echo [ERRO] GameEconomy.abi nao foi gerado!
    pause
    exit /b 1
)
echo [OK] Arquivos gerados com sucesso
echo.

REM Verifica se registrarPartidaAdmin esta no ABI
echo [4/5] Verificando se registrarPartidaAdmin existe...
findstr /C:"registrarPartidaAdmin" "%CONTRACTS_DIR%\GameEconomy.abi" >nul
if %ERRORLEVEL% NEQ 0 (
    echo [ERRO] registrarPartidaAdmin NAO encontrado no ABI!
    echo A compilacao pode estar usando um arquivo .sol antigo.
    pause
    exit /b 1
)
echo [OK] registrarPartidaAdmin encontrado no ABI
echo.

REM Faz o deploy
echo [5/5] Fazendo deploy...
cd /d "%TOOLS_DIR%"
echo Compilando deploy-contract.exe...
go build -o deploy-contract.exe deploy-contract.go
if %ERRORLEVEL% NEQ 0 (
    echo [ERRO] Falha ao compilar deploy-contract.go
    pause
    exit /b 1
)

echo Executando deploy...
.\deploy-contract.exe

echo.
echo ========================================
echo CONCLUIDO!
echo ========================================
echo.
echo Agora reinicie os servidores com rebuild-and-start.bat
echo.
pause
