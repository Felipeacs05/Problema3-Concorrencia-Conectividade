@echo off
cd /d "%~dp0..\tools"
if not exist "test-contract-function.exe" (
    echo Compilando test-contract-function.exe...
    go build -o test-contract-function.exe test-contract-function.go
)
test-contract-function.exe %*
pause
