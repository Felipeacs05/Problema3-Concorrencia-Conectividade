# 🚀 Guia Rápido - Windows

## Executar Tudo do Zero

### Passo 1: Verificar Pré-requisitos

Abra o PowerShell ou CMD e verifique:

```cmd
# Verificar Docker
docker --version

# Verificar Go
go version
```

Se algum não estiver instalado:
- **Docker**: https://www.docker.com/products/docker-desktop
- **Go**: https://golang.org/dl/

### Passo 2: Navegar até a Pasta do Projeto

```cmd
cd "C:\Users\bluti\OneDrive\Desktop\UEFS\5 Semestre\MI - Concorrência e Conectividade\Problema3-Concorrencia-Conectividade\Projeto"
```

### Passo 3: Executar Setup Completo

```cmd
cd scripts
setup.bat
```

Este script irá:
1. ✅ Compilar o utilitário Go `blockchain-utils`
2. ✅ Parar containers existentes
3. ✅ Remover dados antigos
4. ✅ Criar nova conta Ethereum (senha: `123456`)
5. ✅ Gerar `genesis.json` automaticamente
6. ✅ Inicializar a blockchain
7. ✅ Iniciar o nó Geth

**Aguarde a conclusão** - pode levar alguns minutos.

### Passo 4: Desbloquear Conta (Iniciar Clique)

Após o setup, desbloqueie a conta para que o Clique comece a selar blocos:

```cmd
unlock-account.bat
```

Você deve ver: `SUCCESS: Conta desbloqueada!`

### Passo 5: Verificar se Está Funcionando

```cmd
check-block.bat
```

Você deve ver o número do bloco (começa em `0`).

**Aguarde 10 segundos** e execute novamente:

```cmd
timeout /t 10 /nobreak
check-block.bat
```

O número do bloco deve ter aumentado! ✅

---

## Comandos Úteis

### Ver Logs do Geth
```cmd
cd ..
docker-compose logs -f geth
```

### Parar o Nó
```cmd
cd ..
docker-compose down
```

### Iniciar o Nó Novamente
```cmd
cd ..
docker-compose up -d geth
```

### Obter Enode (para compartilhar com outros nós)
```cmd
cd scripts
get-enode.bat
```

### Verificar Peers Conectados
```cmd
check-peers.bat
```

---

## Troubleshooting

### Erro: "Go não está instalado"
- Instale Go: https://golang.org/dl/
- Adicione ao PATH do sistema
- Reinicie o terminal

### Erro: "Docker não está rodando"
- Abra Docker Desktop
- Aguarde até aparecer "Docker is running"

### Erro: "Falha ao compilar blockchain-utils"
- Verifique se Go está instalado: `go version`
- Execute: `cd tools && go mod tidy`

### Blocos não estão sendo criados
- Verifique se desbloqueou a conta: `unlock-account.bat`
- Verifique os logs: `docker-compose logs geth`
- Aguarde alguns segundos - blocos são criados a cada 5 segundos

### Erro: "database contains incompatible genesis"
- Execute `setup.bat` novamente para resetar tudo

---

## Estrutura de Comandos

```
Projeto/
├── scripts/
│   ├── setup.bat              ← PRIMEIRA VEZ: Execute este
│   ├── unlock-account.bat     ← Depois: Desbloquear conta
│   ├── check-block.bat        ← Verificar blocos
│   ├── get-enode.bat          ← Obter enode
│   ├── connect-peer.bat       ← Conectar a outro nó
│   └── check-peers.bat        ← Verificar peers
└── docker-compose.yml          ← Configuração Docker
```

---

## Resumo Rápido

```cmd
# 1. Setup completo (primeira vez)
cd scripts
setup.bat

# 2. Desbloquear conta
unlock-account.bat

# 3. Verificar blocos
check-block.bat

# 4. Aguardar 10 segundos e verificar novamente
timeout /t 10 /nobreak
check-block.bat
```

**Pronto!** Se o número do bloco aumentou, está tudo funcionando! 🎉


