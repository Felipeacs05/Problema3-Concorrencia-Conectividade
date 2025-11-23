@echo off
REM ===================== BAREMA ITEM 1: ARQUITETURA =====================
REM Script para conectar este nó a um peer (bootnode) - Windows
REM Uso: connect-peer.bat <enode-do-bootnode>

if "%1"=="" (
    echo Uso: %0 ^<enode-do-bootnode^>
    echo Exemplo: %0 enode://abc123...@192.168.1.100:30303
    pause
    exit /b 1
)

set BOOTNODE_ENODE=%1

echo Conectando ao bootnode: %BOOTNODE_ENODE%
echo.

REM BAREMA ITEM 1: ARQUITETURA - Inicia o nó com bootnode especificado
set BOOTNODE_ENODE=%BOOTNODE_ENODE%
docker-compose up geth


