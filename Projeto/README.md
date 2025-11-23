# 🎮 Jogo de Cartas Multiplayer Baseado em Blockchain

Projeto PBL3 - Sistema distribuído de jogo de cartas usando blockchain Ethereum privada.

## 📋 Índice

- [Requisitos](#requisitos)
- [Instalação](#instalação)
- [Configuração Inicial](#configuração-inicial)
- [Uso](#uso)
- [Estrutura do Projeto](#estrutura-do-projeto)

## 🔧 Requisitos

### Windows
- Docker Desktop
- Go 1.22 ou superior
- Git Bash (opcional, para scripts .sh)

### Linux/macOS
- Docker e Docker Compose
- Go 1.22 ou superior

## 📦 Instalação

### Windows

1. **Instale Docker Desktop**
   - Baixe em: https://www.docker.com/products/docker-desktop
   - Reinicie o computador após instalação

2. **Instale Go**
   - Baixe em: https://golang.org/dl/
   - Adicione Go ao PATH do sistema

3. **Clone o repositório**
   ```cmd
   git clone <url-do-repositorio>
   cd Problema3-Concorrencia-Conectividade/Projeto
   ```

### Linux/macOS

1. **Instale Docker e Docker Compose**
   ```bash
   # Ubuntu/Debian
   sudo apt-get update
   sudo apt-get install docker.io docker-compose
   sudo systemctl start docker
   sudo usermod -aG docker $USER
   # Faça logout e login novamente
   ```

2. **Instale Go**
   ```bash
   # Ubuntu/Debian
   sudo apt-get install golang-go
   
   # macOS (com Homebrew)
   brew install go
   ```

3. **Clone o repositório**
   ```bash
   git clone <url-do-repositorio>
   cd Problema3-Concorrencia-Conectividade/Projeto
   ```

## 🚀 Configuração Inicial

### Windows

Execute o script de configuração:

```cmd
cd scripts
setup.bat
```

Este script irá:
1. Compilar o utilitário Go `blockchain-utils`
2. Parar containers existentes
3. Remover dados antigos
4. Criar nova conta Ethereum
5. Gerar `genesis.json` automaticamente
6. Inicializar a blockchain
7. Iniciar o nó Geth

### Linux/macOS

Execute o script de configuração:

```bash
cd scripts
chmod +x setup.sh
./setup.sh
```

Este script faz as mesmas operações do script Windows.

## 📖 Uso

### Primeira Vez (Configuração Completa)

#### Windows
```cmd
cd scripts
setup.bat
```

#### Linux/macOS
```bash
cd scripts
./setup.sh
```

### Desbloquear Conta (Iniciar Clique)

Após a configuração inicial, desbloqueie a conta para que o Clique comece a selar blocos:

#### Windows
```cmd
cd scripts
unlock-account.bat
```

#### Linux/macOS
```bash
cd scripts
./unlock-account.sh
```

### Verificar Blocos

Para verificar se os blocos estão sendo criados:

#### Windows
```cmd
cd scripts
check-block.bat
```

#### Linux/macOS
```bash
cd scripts
./check-block.sh
```

Aguarde 10 segundos e execute novamente - o número do bloco deve aumentar.

### Iniciar/Parar o Nó

#### Windows
```cmd
# Iniciar
docker-compose up -d geth

# Parar
docker-compose down
```

#### Linux/macOS
```bash
# Iniciar
docker-compose up -d geth

# Parar
docker-compose down
```

### Ver Logs

```bash
docker-compose logs -f geth
```

## 🏗️ Estrutura do Projeto

```
Projeto/
├── contracts/              # Smart Contracts Solidity
│   └── GameEconomy.sol
├── cliente/                # Cliente Go
│   └── main.go
├── tools/                  # Utilitários Go cross-platform
│   ├── blockchain-utils.go
│   └── go.mod
├── scripts/                # Scripts de configuração
│   ├── setup.bat          # Windows
│   ├── setup.sh            # Linux/macOS
│   ├── unlock-account.bat  # Windows
│   ├── unlock-account.sh   # Linux/macOS
│   ├── check-block.bat     # Windows
│   └── check-block.sh      # Linux/macOS
├── data/                   # Dados da blockchain (keystore, chaindata)
├── docker-compose.yml      # Orquestração Docker
├── genesis.json            # Configuração inicial da blockchain
└── README.md               # Este arquivo
```

## 🔑 Utilitário Go (blockchain-utils)

O utilitário `blockchain-utils` é cross-platform e pode ser usado diretamente:

### Criar Conta

```bash
# Windows
tools\blockchain-utils.exe criar-conta data\keystore [senha]

# Linux/macOS
./tools/blockchain-utils criar-conta data/keystore [senha]
```

### Gerar Genesis.json

```bash
# Windows
tools\blockchain-utils.exe gerar-genesis data\keystore genesis.json

# Linux/macOS
./tools/blockchain-utils gerar-genesis data/keystore genesis.json
```

### Extrair Endereço do Keystore

```bash
# Windows
tools\blockchain-utils.exe extrair-endereco data\keystore

# Linux/macOS
./tools/blockchain-utils extrair-endereco data/keystore
```

## 🐛 Troubleshooting

### Erro: "Go não está instalado"
- Instale Go e adicione ao PATH
- Verifique com: `go version`

### Erro: "Docker não está rodando"
- Inicie Docker Desktop (Windows) ou serviço Docker (Linux)
- Verifique com: `docker ps`

### Erro: "Falha ao desbloquear conta"
- Verifique se a senha está correta (padrão: `123456`)
- Execute `setup.bat` ou `setup.sh` novamente para criar nova conta

### Blocos não estão sendo criados
- Verifique se a conta está desbloqueada: `unlock-account.bat` ou `./unlock-account.sh`
- Verifique os logs: `docker-compose logs geth`
- Aguarde alguns segundos - blocos são criados a cada 5 segundos no Clique

### Erro: "database contains incompatible genesis"
- Execute `setup.bat` ou `setup.sh` novamente para resetar tudo

## 📝 Notas

- **Senha padrão**: `123456` (use apenas para desenvolvimento/testes!)
- **Network ID**: `1337`
- **Consensus**: Clique (Proof of Authority)
- **Período de blocos**: 5 segundos
- **Saldo inicial**: 1 milhão de ETH para o signer

## 🔗 Referências

- [Ethereum Documentation](https://ethereum.org/en/developers/docs/)
- [Geth Documentation](https://geth.ethereum.org/docs/)
- [Go Ethereum](https://github.com/ethereum/go-ethereum)


