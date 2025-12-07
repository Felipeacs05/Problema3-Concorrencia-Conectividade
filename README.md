# Jogo de Cartas Multiplayer com Blockchain (PBL3 - TEC502)

Este repositório contém a implementação do **Problema 3** da disciplina TEC502 - Concorrência e Conectividade, oferecida pela Universidade Estadual de Feira de Santana (UEFS).

## 📝 Descrição do Projeto

O projeto consiste na evolução do "Jogo de Cartas Multiplayer" para uma arquitetura **descentralizada baseada em blockchain**. A solução utiliza uma **blockchain privada Ethereum** rodando localmente via Geth, onde todas as transações cruciais do jogo (posse de cartas, compra de pacotes, trocas e resultados de partidas) são registradas de forma **imutável, transparente e auditável**.

O sistema foi projetado para rodar em uma rede local (laboratório) via switch, permitindo que múltiplos computadores formem uma rede P2P sem necessidade de servidores centralizados.

## ✨ Funcionalidades Implementadas

* **Blockchain Privada Ethereum:** Rede local usando Geth (Go Ethereum) com configuração Proof of Work (PoW) com dificuldade baixa para mineração rápida.
* **Smart Contracts (Solidity):** Contrato `GameEconomy.sol` que gerencia:
  * **NFTs de Cartas:** Cada carta é um token não-fungível (ERC-721 simplificado) com propriedade única e verificável.
  * **Compra de Pacotes:** Sistema que previne duplo gasto através de atomicidade de transações, usando block hash como fonte de aleatoriedade.
  * **Trocas de Cartas:** Sistema de propostas e aceitação para trocas atômicas entre jogadores.
  * **Registro de Partidas:** Eventos permanentes na blockchain para auditabilidade completa.
* **Aplicação Cliente (Go CLI):** Interface de terminal que permite:
  * Gerenciamento de carteira (criação/importação de contas)
  * Visualização de saldo e cartas
  * Compra de pacotes
  * Criação e aceitação de propostas de troca
  * Registro de resultados de partidas
  * Escuta de eventos em tempo real
* **Infraestrutura Docker:** Containerização completa com `network_mode: host` para comunicação P2P entre máquinas no laboratório.

## 🛠️ Arquitetura e Tecnologias

* **Blockchain:** Ethereum (Geth) - Rede Privada
* **Smart Contracts:** Solidity ^0.8.20
* **Cliente:** Go (Golang) 1.22+
* **Comunicação:** RPC/HTTP (porta 8545), WebSocket (porta 8546)
* **Containerização:** Docker & Docker Compose
* **Bibliotecas Go:**
  * `github.com/ethereum/go-ethereum` - Cliente Ethereum
  * `github.com/fatih/color` - Interface colorida no terminal
  * `github.com/manifoldco/promptui` - Prompts interativos

## 📋 Pré-requisitos

Antes de começar, certifique-se de ter instalado:

* **Docker** (>= 20.10)
* **Docker Compose**
* **Go** (>= 1.22) - Para compilar o cliente
* **Git**

## 🚀 Como Executar o Projeto

### 📋 Comandos Rápidos por Sistema Operacional

---

## 🪟 WINDOWS

### Primeira Vez (Configuração Inicial)

```cmd
REM 1. Navegue até a pasta do projeto
cd "C:\Program Files\Pbl-Redes\Problema3-Concorrencia-Conectividade\Projeto"

REM 2. Execute os comandos na sequência:
scripts\setup-blockchain.bat
scripts\setup-game.bat
scripts\start-all.bat
scripts\criar-conta-jogador.bat
Blockchain\scripts\fund-account.bat <endereco_da_conta>
scripts\run-cliente.bat

REM 3. Quando finalizar, pare todos os serviços:
scripts\stop-all.bat
```

**Nota:** Substitua `<endereco_da_conta>` pelo endereço da conta criada no passo `criar-conta-jogador.bat`.

### Outras Vezes (Uso Diário)

```cmd
REM 1. Navegue até a pasta do projeto
cd "C:\Program Files\Pbl-Redes\Problema3-Concorrencia-Conectividade\Projeto"

REM 2. Execute apenas:
scripts\start-all.bat
scripts\run-cliente.bat

REM 3. Quando finalizar:
scripts\stop-all.bat
```

### Comandos Úteis (Windows)

```cmd
REM Ver logs do servidor
scripts\logs-servidores.bat

REM Ver eventos/transferências do contrato
Blockchain\scripts\view-events.bat <endereco_do_contrato>

REM Parar todos os serviços
scripts\stop-all.bat

REM Acessar console do Geth
docker exec -it geth-node geth attach http://localhost:8545

REM Ver saldo
docker exec geth-node geth attach http://localhost:8545 --exec "eth.getBalance(eth.accounts[0])"

REM Ver número de blocos
docker exec geth-node geth attach http://localhost:8545 --exec "eth.blockNumber"

REM Listar contas
docker exec geth-node geth attach http://localhost:8545 --exec "eth.accounts"
```

---

## 🐧 LINUX (Ubuntu/Debian)

### Primeira Vez (Configuração Inicial)

```bash
# 1. Navegue até a pasta do projeto
cd ~/Problema3-Concorrencia-Conectividade/Projeto/

# 2. Execute os comandos na sequência:
./scripts/setup-blockchain.sh
./scripts/setup-game.sh
./scripts/start-all.sh
./scripts/criar-conta-jogador.sh
./Blockchain/scripts/transfer-eth.sh <endereco_da_conta>
./scripts/run-cliente.sh

# 3. Quando finalizar, pare todos os serviços:
./scripts/stop-all.sh
```

**Nota:** Substitua `<endereco_da_conta>` pelo endereço da conta criada no passo `criar-conta-jogador.sh`.

### Outras Vezes (Uso Diário)

```bash
# 1. Navegue até a pasta do projeto
cd ~/Problema3-Concorrencia-Conectividade/Projeto/

# 2. Execute apenas:
./scripts/start-all.sh
./scripts/run-cliente.sh

# 3. Quando finalizar:
./scripts/stop-all.sh
```

### Comandos Úteis (Linux)

```bash
# Ver logs do servidor
./scripts/logs-servidores.sh

# Ver eventos/transferências do contrato
./Blockchain/scripts/view-events.sh <endereco_do_contrato>

# Parar todos os serviços
./scripts/stop-all.sh

# Acessar console do Geth
docker exec -it geth-node geth attach http://localhost:8545

# Ver saldo
docker exec geth-node geth attach http://localhost:8545 --exec "eth.getBalance(eth.accounts[0])"

# Ver número de blocos
docker exec geth-node geth attach http://localhost:8545 --exec "eth.blockNumber"

# Listar contas
docker exec geth-node geth attach http://localhost:8545 --exec "eth.accounts"

# Obter enode (para conectar outros nós)
docker exec geth-node geth attach http://localhost:8545 --exec "admin.nodeInfo.enode"

# Ver peers conectados
docker exec geth-node geth attach http://localhost:8545 --exec "admin.peers"
```

---

## 🔗 Conectar Múltiplos Nós (Rede P2P)

### No Primeiro Computador (Bootnode)

**Windows:**
```cmd
REM 1. Inicia o nó normalmente
docker-compose up -d geth

REM 2. Obtenha seu IP local
ipconfig
REM Procure por "IPv4 Address" (ex: 192.168.1.100)

REM 3. Obtenha o enode
docker exec geth-node geth attach http://localhost:8545 --exec "admin.nodeInfo.enode"

REM 4. Substitua [::] pelo seu IP real no enode
REM Exemplo: enode://abc...@192.168.1.100:30303
```

**Linux:**
```bash
# 1. Inicia o nó normalmente
docker-compose up -d geth

# 2. Obtenha seu IP local
hostname -I
# Ou: ip addr show

# 3. Obtenha o enode
docker exec geth-node geth attach http://localhost:8545 --exec "admin.nodeInfo.enode"

# 4. Substitua [::] pelo seu IP real no enode
# Exemplo: enode://abc...@192.168.1.100:30303
```

### No Segundo Computador (Peer)

**Windows:**
```cmd
REM 1. Inicialize normalmente (passos 1-4 da "Primeira Vez")

REM 2. Edite docker-compose.yml e adicione ao command:
REM --bootnodes=enode://abc...@192.168.1.100:30303

REM 3. Inicia o nó
docker-compose up -d geth

REM 4. Verifica conexão
docker exec geth-node geth attach http://localhost:8545 --exec "admin.peers"
```

**Linux:**
```bash
# 1. Inicialize normalmente (passos 1-4 da "Primeira Vez")

# 2. Use variável de ambiente ou edite docker-compose.yml
export BOOTNODE_ENODE="enode://abc...@192.168.1.100:30303"
docker-compose up -d geth

# 3. Verifica conexão
docker exec geth-node geth attach http://localhost:8545 --exec "admin.peers"
```

---

## 📦 Deploy do Smart Contract

### Opção 1: Usando Hardhat (Recomendado)

**Windows:**
```cmd
REM 1. Instale Node.js: https://nodejs.org/

REM 2. Instale Hardhat
npm install --save-dev hardhat

REM 3. Crie projeto
npx hardhat init
REM Escolha: "Create a JavaScript project"

REM 4. Copie o contrato
copy contracts\GameEconomy.sol hardhat-project\contracts\

REM 5. Compile
cd hardhat-project
npx hardhat compile

REM 6. Configure hardhat.config.js:
REM networks: {
REM   localhost: { url: "http://localhost:8545" }
REM }

REM 7. Crie scripts/deploy.js e execute:
npx hardhat run scripts/deploy.js --network localhost
```

**Linux:**
```bash
# 1. Instale Node.js
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# 2. Instale Hardhat
npm install --save-dev hardhat

# 3. Crie projeto
npx hardhat init

# 4. Copie o contrato
cp contracts/GameEconomy.sol hardhat-project/contracts/

# 5. Compile
cd hardhat-project
npx hardhat compile

# 6. Configure hardhat.config.js e faça deploy
npx hardhat run scripts/deploy.js --network localhost
```

### Opção 2: Usando Remix (Online)

1. Acesse: https://remix.ethereum.org/
2. Cole o código do contrato `GameEconomy.sol`
3. Compile
4. Conecte à rede local (Web3 Provider: http://localhost:8545)
5. Faça deploy

---

## 🚀 Como Executar o Projeto (Guia Detalhado)

### Passo 1: Clone o Repositório

```bash
git clone <url-do-repositorio>
cd Problema3-Concorrencia-Conectividade/Projeto/
```

### Passo 2: Gerar o Genesis Block

O arquivo `genesis.json` já está incluído no projeto. Ele configura uma rede privada com:

* **Chain ID:** 1337
* **Consenso:** Proof of Work (PoW) - não requer signers pré-configurados
* **Dificuldade:** Baixa (0x400) para mineração rápida em PCs de laboratório
* **Gas Limit:** Alto (0x8000000) para suportar contratos complexos

**Nota:** O genesis.json foi configurado para PoW (não Clique) para facilitar a configuração inicial. Qualquer um pode minerar sem necessidade de configurar signers.

Se precisar personalizar, edite `genesis.json` antes de continuar.

### Passo 3: Inicializar o Primeiro Nó (Bootnode)

O primeiro nó da rede será o "bootnode" (nó inicial que outros nós podem conectar).

#### 3.1. Inicializar a Blockchain

```bash
# Cria diretório de dados
mkdir -p data

# Inicializa blockchain com genesis.json
docker-compose run --rm geth --datadir /root/.ethereum init /genesis.json

# Cria conta inicial (senha será solicitada)
docker-compose run --rm geth --datadir /root/.ethereum account new
```

**Anote o endereço da conta criada!** Você precisará dele para minerar e receber recompensas.

#### 3.2. Obter o Enode do Bootnode

Após iniciar o nó, você precisará obter o **enode** (endereço P2P) para que outros nós possam se conectar:

```bash
# Inicia o nó
docker-compose up geth

# Em outro terminal, acesse o console do Geth
docker exec -it geth-node geth attach http://localhost:8545

# No console do Geth, execute:
> admin.nodeInfo.enode
```

O resultado será algo como:
```
"enode://<hash>@<seu-ip>:30303"
```

**Anote este enode completo!** Você precisará compartilhá-lo com outros participantes.

#### 3.3. Iniciar o Nó com Mineração

Para iniciar o nó com mineração automática:

```bash
# Edite docker-compose.yml e adicione --mine e --miner.etherbase=<endereco-da-conta>
# Ou inicie manualmente:
docker-compose up geth
```

### Passo 4: Conectar o Segundo Nó (Peer)

Em outra máquina na mesma rede (conectada via switch):

#### 4.1. Clone e Configure

```bash
# Clone o mesmo repositório
git clone <url-do-repositorio>
cd Problema3-Concorrencia-Conectividade/Projeto/

# Inicialize a blockchain (mesmo genesis.json)
mkdir -p data
docker-compose run --rm geth --datadir /root/.ethereum init /genesis.json
docker-compose run --rm geth --datadir /root/.ethereum account new
```

#### 4.2. Conecte ao Bootnode

Antes de iniciar o nó, configure a variável de ambiente com o enode do bootnode:

```bash
# Substitua <enode-do-bootnode> pelo enode obtido no Passo 3.2
export BOOTNODE_ENODE="enode://<hash>@<ip-do-bootnode>:30303"

# Inicie o nó conectando ao bootnode
docker-compose up geth
```

**Nota:** Certifique-se de que o firewall permite comunicação na porta 30303 (P2P) e 8545 (RPC).

### Passo 5: Compilar e Implantar o Smart Contract

#### 5.1. Compilar o Contrato

Você precisará do compilador Solidity (`solc`):

```bash
# Instale o solc (exemplo para Ubuntu/Debian)
sudo add-apt-repository ppa:ethereum/ethereum
sudo apt-get update
sudo apt-get install solc

# Compile o contrato
solc --abi --bin contracts/GameEconomy.sol -o build/
```

Isso gerará:
* `build/GameEconomy.abi` - Interface do contrato (ABI)
* `build/GameEconomy.bin` - Bytecode do contrato

#### 5.2. Fazer Deploy do Contrato

Você pode usar o cliente Go (após implementar a função de deploy) ou usar o console do Geth:

```bash
# Acesse o console do Geth
docker exec -it geth-node geth attach http://localhost:8545

# No console, faça o deploy (exemplo simplificado):
> var bytecode = "0x<bytecode-do-contrato>"
> var abi = [<abi-do-contrato>]
> var contract = eth.contract(abi)
> var deployed = contract.new({from: eth.accounts[0], data: bytecode, gas: 3000000})
```

**Anote o endereço do contrato implantado!** Todos os clientes precisarão deste endereço.

### Passo 6: Compilar e Executar o Cliente Go

#### 6.1. Instalar Dependências

```bash
cd cliente/
go mod download
```

#### 6.2. Compilar

```bash
go build -o jogo-cartas main.go
```

#### 6.3. Executar

```bash
./jogo-cartas
```

O cliente irá:
1. Conectar ao nó Geth local (http://localhost:8545)
2. Solicitar criação ou seleção de conta
3. Permitir configurar o endereço do contrato
4. Exibir o menu principal

## 📖 Comandos do Cliente

O cliente oferece um menu interativo com as seguintes opções:

1. **Ver Saldo e Cartas:** Exibe seu saldo de ETH e lista todas as cartas que você possui.
2. **Comprar Pacote:** Compra um pacote de 5 cartas aleatórias do contrato.
3. **Trocar Carta:** Cria uma proposta de troca com outro jogador.
4. **Ver Propostas de Troca Pendentes:** Lista propostas que você recebeu ou enviou.
5. **Registrar Vitória de Partida:** Registra o resultado de uma partida na blockchain.
6. **Ver Histórico de Partidas:** Exibe todas as partidas registradas.
7. **Configurar Endereço do Contrato:** Define o endereço do contrato GameEconomy.

## 🔧 Configuração Avançada

### Personalizar Genesis Block

Edite `genesis.json` para alterar:
* **Chain ID:** Altere `chainId` para um valor único
* **Dificuldade:** Altere `difficulty` (valores menores = mineração mais rápida)
* **Saldo Inicial:** Adicione endereços em `alloc` com saldos iniciais

### Conectar Múltiplos Peers

Para conectar a mais de um peer, você pode:

1. **Usar múltiplos bootnodes:**
```bash
export BOOTNODE_ENODE="enode://...@ip1:30303,enode://...@ip2:30303"
```

2. **Adicionar peers manualmente via console:**
```bash
docker exec -it geth-node geth attach http://localhost:8545
> admin.addPeer("enode://...@ip:30303")
```

### Persistência de Dados

Os dados da blockchain são salvos em `./data/`. Para resetar a blockchain:

```bash
# CUIDADO: Isso apaga toda a blockchain local!
rm -rf data/
docker-compose run --rm geth --datadir /root/.ethereum init /genesis.json
```

## 🧪 Testando o Sistema

### Teste Básico de Conectividade

1. Inicie o bootnode na máquina A
2. Obtenha o enode
3. Inicie o peer na máquina B conectando ao bootnode
4. Verifique se estão conectados:
```bash
docker exec -it geth-node geth attach http://localhost:8545
> admin.peers
```

### Teste de Compra de Pacote

1. Execute o cliente Go
2. Crie uma conta
3. Configure o endereço do contrato
4. Use a opção "Comprar Pacote"
5. Verifique se as cartas foram criadas na blockchain

### Teste de Troca

1. Jogador A cria proposta de troca
2. Jogador B aceita a proposta
3. Verifique se as cartas foram transferidas corretamente

## 📚 Estrutura do Projeto

```
Projeto/
├── contracts/
│   └── GameEconomy.sol          # Smart contract principal
├── cliente/
│   └── main.go                  # Aplicação cliente Go
├── scripts/
│   ├── init.sh                  # Script de inicialização
│   ├── deploy-contract.sh       # Script de deploy
│   └── password.txt             # Senha para desbloqueio
├── data/                        # Dados da blockchain (gerado)
├── build/                       # Arquivos compilados (gerado)
├── keystore/                    # Carteiras locais (gerado)
├── docker-compose.yml           # Orquestração Docker
├── genesis.json                 # Configuração do genesis block
├── go.mod                       # Dependências Go
└── README.md                    # Este arquivo
```

## 🔐 Segurança

**⚠️ IMPORTANTE:** Este projeto é para fins educacionais e de laboratório. **NÃO use em produção** sem as seguintes melhorias:

1. **Senhas Fortes:** Não use senhas vazias ou fracas em produção
2. **Keystore Seguro:** Proteja o diretório `keystore/` com permissões adequadas
3. **Rede Isolada:** A rede privada não deve estar acessível da internet
4. **Validação de Contratos:** Sempre valide contratos antes de fazer deploy
5. **Gas Limits:** Configure limites de gas apropriados para prevenir ataques

## 🐛 Troubleshooting

### Erro: "Cannot connect to Geth"

**Solução:** Verifique se o container está rodando:
```bash
docker ps
docker-compose logs geth
```

### Erro: "Insufficient funds"

**Solução:** Você precisa de ETH para pagar gas. Em uma rede privada, você pode:
1. Minerar blocos (receberá recompensas)
2. Transferir ETH de outra conta
3. Modificar o genesis.json para dar saldo inicial

### Erro: "Peer connection failed"

**Solução:**
1. Verifique se ambos os nós estão na mesma rede
2. Verifique firewall (porta 30303 deve estar aberta)
3. Verifique se o enode está correto
4. Use `network_mode: host` no docker-compose

### Erro: "Contract not deployed"

**Solução:** Certifique-se de:
1. Ter feito deploy do contrato
2. Ter configurado o endereço correto no cliente
3. Estar usando a mesma rede (mesmo chainId)

## 📝 Notas de Implementação

### Limitações Atuais

O cliente Go atual é uma **versão base** que demonstra a estrutura. Para funcionalidade completa, você precisará:

1. **Carregar o ABI do contrato:** Use `go-ethereum/accounts/abi` para fazer chamadas ao contrato
2. **Implementar escuta de eventos:** Use `client.SubscribeFilterLogs()` para escutar eventos em tempo real
3. **Melhorar tratamento de erros:** Adicione validações e mensagens de erro mais descritivas
4. **Adicionar testes:** Crie testes unitários e de integração

### Melhorias Futuras

* Interface web (React/Vue) em vez de CLI
* Sistema de matchmaking on-chain
* Lógica de jogo completa no smart contract
* Sistema de recompensas por vitórias
* Marketplace de cartas

## 📄 Licença

Este projeto é parte de um trabalho acadêmico da UEFS.

## 👥 Autores

Desenvolvido como parte da disciplina TEC502 - Concorrência e Conectividade.

---

**Última atualização:** Dezembro 2024

