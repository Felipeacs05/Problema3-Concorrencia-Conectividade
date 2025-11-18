# Estrutura do Projeto

Este documento descreve a estrutura completa do projeto e o propósito de cada arquivo.

## 📁 Estrutura de Diretórios

```
Projeto/
├── contracts/                    # Smart Contracts Solidity
│   └── GameEconomy.sol          # Contrato principal do jogo
│
├── cliente/                      # Aplicação Cliente Go
│   ├── main.go                  # Código principal do cliente CLI
│   └── Dockerfile              # Container para o cliente
│
├── scripts/                      # Scripts auxiliares
│   ├── init.sh                 # Inicialização da blockchain
│   ├── deploy-contract.sh      # Deploy do contrato
│   ├── conectar-peer.sh        # Conectar a outro nó
│   ├── obter-enode.sh          # Obter endereço P2P
│   ├── minerar.sh              # Iniciar mineração
│   ├── verificar-conexoes.sh   # Verificar peers conectados
│   └── password.txt            # Senha para desbloqueio automático
│
├── data/                        # Dados da blockchain (gerado)
│   └── [chaindata do Geth]
│
├── build/                       # Arquivos compilados (gerado)
│   ├── GameEconomy.abi         # ABI do contrato (após compilação)
│   └── GameEconomy.bin         # Bytecode do contrato (após compilação)
│
├── keystore/                    # Carteiras locais (gerado)
│   └── [arquivos de keystore]
│
├── docker-compose.yml           # Orquestração Docker
├── genesis.json                 # Configuração do bloco gênese
├── go.mod                       # Dependências Go
├── .gitignore                   # Arquivos ignorados pelo Git
├── README.md                    # Documentação principal
├── INSTALACAO.md                # Guia de instalação detalhado
├── QUICKSTART.md                # Guia rápido de início
└── ESTRUTURA.md                 # Este arquivo
```

## 📄 Descrição dos Arquivos

### Smart Contracts

#### `contracts/GameEconomy.sol`
- **Propósito:** Contrato inteligente principal que gerencia toda a economia do jogo
- **Funcionalidades:**
  - Criação de cartas como NFTs (ERC-721 simplificado)
  - Sistema de compra de pacotes com prevenção de duplo gasto
  - Sistema de trocas atômicas entre jogadores
  - Registro permanente de partidas
- **Eventos:** Emite eventos para todas as ações importantes (compra, troca, partidas)

### Cliente

#### `cliente/main.go`
- **Propósito:** Aplicação CLI em Go para interagir com a blockchain
- **Funcionalidades:**
  - Gerenciamento de carteira (criação/importação)
  - Conexão com nó Geth via RPC
  - Interface de menu interativa
  - Chamadas ao smart contract (estrutura base)
- **Nota:** Requer implementação completa de ABI para funcionalidade total

#### `cliente/Dockerfile`
- **Propósito:** Containerização do cliente Go
- **Uso:** `docker-compose build cliente && docker-compose run cliente`

### Scripts

#### `scripts/init.sh`
- **Propósito:** Inicializa a blockchain pela primeira vez
- **Uso:** `bash scripts/init.sh`

#### `scripts/deploy-contract.sh`
- **Propósito:** Compila o contrato Solidity
- **Uso:** `bash scripts/deploy-contract.sh`
- **Requisitos:** Solidity compiler (`solc`) instalado

#### `scripts/conectar-peer.sh`
- **Propósito:** Conecta este nó a um bootnode
- **Uso:** `bash scripts/conectar-peer.sh <enode>`

#### `scripts/obter-enode.sh`
- **Propósito:** Obtém o endereço P2P deste nó para compartilhar
- **Uso:** `bash scripts/obter-enode.sh`

#### `scripts/minerar.sh`
- **Propósito:** Inicia mineração manualmente
- **Uso:** `bash scripts/minerar.sh`

#### `scripts/verificar-conexoes.sh`
- **Propósito:** Lista todos os peers conectados
- **Uso:** `bash scripts/verificar-conexoes.sh`

#### `scripts/password.txt`
- **Propósito:** Senha para desbloqueio automático da conta (vazia para desenvolvimento)
- **⚠️ ATENÇÃO:** Em produção, use senha forte!

### Configuração

#### `docker-compose.yml`
- **Propósito:** Orquestração de containers Docker
- **Serviços:**
  - `geth`: Nó Ethereum privado
  - `cliente`: Aplicação cliente (opcional, pode rodar nativamente)

#### `genesis.json`
- **Propósito:** Configuração do bloco gênese da blockchain privada
- **Configurações:**
  - Chain ID: 1337
  - Consenso: Proof of Authority (Clique)
  - Período: 5 segundos
  - Dificuldade: Muito baixa

#### `go.mod`
- **Propósito:** Gerenciamento de dependências Go
- **Dependências principais:**
  - `github.com/ethereum/go-ethereum` - Cliente Ethereum
  - `github.com/fatih/color` - Cores no terminal
  - `github.com/manifoldco/promptui` - Prompts interativos

#### `.gitignore`
- **Propósito:** Arquivos que não devem ser versionados
- **Inclui:** Dados da blockchain, keystore, arquivos compilados

### Documentação

#### `README.md`
- **Propósito:** Documentação principal do projeto
- **Conteúdo:** Descrição completa, instruções de uso, troubleshooting

#### `INSTALACAO.md`
- **Propósito:** Guia passo a passo de instalação
- **Conteúdo:** Instruções detalhadas para cada etapa

#### `QUICKSTART.md`
- **Propósito:** Guia rápido para começar em 5 minutos
- **Conteúdo:** Comandos essenciais e início rápido

#### `ESTRUTURA.md`
- **Propósito:** Este arquivo - documenta a estrutura do projeto

## 🔄 Fluxo de Dados

```
Cliente Go (CLI)
    │
    │ RPC/HTTP (porta 8545)
    ▼
Nó Geth (Docker)
    │
    │ P2P (porta 30303)
    ▼
Rede Blockchain Privada
    │
    │ Smart Contract
    ▼
GameEconomy.sol
```

## 🎯 Pontos de Entrada

### Para Desenvolvedores

1. **Modificar Smart Contract:** `contracts/GameEconomy.sol`
2. **Modificar Cliente:** `cliente/main.go`
3. **Ajustar Configuração:** `docker-compose.yml`, `genesis.json`

### Para Usuários

1. **Iniciar Sistema:** `QUICKSTART.md`
2. **Instalação Completa:** `INSTALACAO.md`
3. **Referência Completa:** `README.md`

## 📦 Arquivos Gerados (Não Versionados)

Estes diretórios são criados durante a execução:

- `data/` - Dados da blockchain (chaindata do Geth)
- `build/` - Arquivos compilados do contrato (ABI, bytecode)
- `keystore/` - Carteiras locais dos usuários

**⚠️ IMPORTANTE:** Nunca versione o diretório `keystore/` - contém chaves privadas!

## 🔧 Personalização

### Alterar Configuração da Rede

Edite `genesis.json`:
- `chainId`: ID único da rede
- `clique.period`: Período de mineração (segundos)
- `alloc`: Saldos iniciais de contas

### Alterar Configuração do Nó

Edite `docker-compose.yml`:
- Portas RPC/WebSocket
- APIs habilitadas
- Configurações de mineração

### Adicionar Novos Scripts

Crie novos scripts em `scripts/` e adicione ao `.gitignore` se necessário.

## 📚 Próximos Passos

Após entender a estrutura:

1. Leia `README.md` para visão geral
2. Siga `QUICKSTART.md` para começar rapidamente
3. Consulte `INSTALACAO.md` para configuração completa
4. Explore o código em `contracts/` e `cliente/`

