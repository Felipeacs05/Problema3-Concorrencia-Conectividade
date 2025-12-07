@echo off
cd /d "%~dp0..\tools"
if not exist "check-owner.exe" (
    echo Compilando...
    go build -o check-owner.exe check-owner.go
)
check-owner.exe %*
pause
