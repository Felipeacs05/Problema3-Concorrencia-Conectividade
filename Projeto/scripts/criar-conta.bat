@echo off
REM ===================== BAREMA ITEM 1: ARQUITETURA =====================
REM Script para criar uma nova conta Ethereum
REM Resolve o problema de entrada de senha no Docker

echo ========================================
echo Criando nova conta Ethereum...
echo ========================================
echo.

REM Solicita senha do usuário
set /p SENHA="Digite uma senha para a nova conta: "

REM Cria arquivo temporário com a senha
echo %SENHA% > temp_password.txt

REM Cria a conta usando o arquivo de senha
docker-compose run --rm -v "%CD%\temp_password.txt:/password.txt" geth --datadir /root/.ethereum account new --password /password.txt

REM Remove arquivo temporário
del temp_password.txt

echo.
echo ========================================
echo Conta criada com sucesso!
echo ========================================
echo.
echo IMPORTANTE: Anote o endereco da conta acima!
echo.

pause

