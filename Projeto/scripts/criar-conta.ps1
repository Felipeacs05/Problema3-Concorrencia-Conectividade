# ===================== BAREMA ITEM 1: ARQUITETURA =====================
# Script PowerShell para criar uma nova conta Ethereum
# Resolve o problema de entrada de senha no Docker

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Criando nova conta Ethereum..." -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Solicita senha do usuário (oculta)
$senha = Read-Host "Digite uma senha para a nova conta" -AsSecureString
$senhaPlain = [Runtime.InteropServices.Marshal]::PtrToStringAuto([Runtime.InteropServices.Marshal]::SecureStringToBSTR($senha))

# Cria arquivo temporário com a senha
$tempFile = Join-Path $PSScriptRoot "temp_password.txt"
$senhaPlain | Out-File -FilePath $tempFile -Encoding ASCII -NoNewline

# Obtém o caminho absoluto do arquivo
$tempFileAbs = Resolve-Path $tempFile

# Cria a conta usando o arquivo de senha
docker-compose run --rm -v "${tempFileAbs}:/password.txt" geth --datadir /root/.ethereum account new --password /password.txt

# Remove arquivo temporário
Remove-Item $tempFile -Force

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "Conta criada com sucesso!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "IMPORTANTE: Anote o endereco da conta acima!" -ForegroundColor Yellow
Write-Host ""

