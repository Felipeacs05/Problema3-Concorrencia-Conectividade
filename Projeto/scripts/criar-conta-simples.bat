@echo off
REM ===================== BAREMA ITEM 1: ARQUITETURA =====================
REM Script simplificado para criar conta com senha padrão "123456"
REM Use apenas para desenvolvimento/testes!

echo ========================================
echo Criando conta com senha padrao: 123456
echo (Use apenas para testes!)
echo ========================================
echo.

REM Cria arquivo com senha padrão
echo 123456 > temp_password.txt

REM Cria a conta
docker-compose run --rm -v "%CD%\temp_password.txt:/password.txt" geth --datadir /root/.ethereum account new --password /password.txt

REM Remove arquivo temporário
del temp_password.txt

echo.
echo ========================================
echo Conta criada!
echo Senha: 123456
echo ========================================
echo.

pause

