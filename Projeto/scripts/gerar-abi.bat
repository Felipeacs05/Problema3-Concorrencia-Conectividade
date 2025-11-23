@echo off
REM Script para gerar apenas o ABI do contrato

set SCRIPT_DIR=%~dp0
set PROJECT_DIR=%SCRIPT_DIR%..
set CONTRACTS_DIR=%PROJECT_DIR%\contracts

echo Gerando ABI do GameEconomy.sol...

docker run --rm -v "%CONTRACTS_DIR%":/contracts ethereum/solc:0.8.20 --abi --optimize --optimize-runs 200 -o /contracts --overwrite /contracts/GameEconomy.sol

if exist "%CONTRACTS_DIR%\GameEconomy.abi" (
    echo [OK] ABI gerado com sucesso!
) else (
    echo [ERRO] Falha ao gerar ABI
)

pause

