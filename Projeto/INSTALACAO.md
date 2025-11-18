# Guia de Instalação Passo a Passo

Este guia detalha cada passo necessário para configurar e executar o sistema completo.

## Pré-requisitos

### 1. Instalar Docker e Docker Compose

**Windows:**
- Baixe e instale [Docker Desktop](https://www.docker.com/products/docker-desktop)

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get update
sudo apt-get install docker.io docker-compose
sudo usermod -aG docker $USER
# Faça logout e login novamente
```

**macOS:**
- Baixe e instale [Docker Desktop](https://www.docker.com/products/docker-desktop)

### 2. Instalar Go (Golang)

**Windows:**
- Baixe de https://golang.org/dl/
- Execute o instalador

**Linux:**
```bash
sudo apt-get install golang-go
```

**macOS:**
```bash
brew install go
```

### 3. Instalar Solidity Compiler (Opcional, para compilar contratos)

**Linux:**
```bash
sudo add-apt-repository ppa:ethereum/ethereum
sudo apt-get update
sudo apt-get install solc
```

**macOS:**
```bash
brew install solidity
```

## Configuração Inicial

### Passo 1: Clone o Repositório

```bash
git clone <url-do-repositorio>
cd Problema3-Concorrencia-Conectividade/Projeto/
```

### Passo 2: Inicializar a Blockchain (Primeira Vez)

```bash
# Cria diretório de dados
mkdir -p data

# Inicializa blockchain com genesis.json
docker-compose run --rm geth --datadir /root/.ethereum init /genesis.json
```

### Passo 3: Criar Conta Inicial

```bash
# Cria uma nova conta (você será solicitado a criar uma senha)
docker-compose run --rm geth --datadir /root/.ethereum account new
```

**IMPORTANTE:** Anote o endereço da conta criada! Ele será algo como `0x1234...5678`

### Passo 4: Obter Endereço da Conta para Mineração

Você precisa do índice da conta (geralmente 0 se for a primeira):

```bash
# Lista todas as contas
docker-compose run --rm geth --datadir /root/.ethereum account list
```

### Passo 5: Iniciar o Nó Geth

```bash
# Inicia o nó (mineração automática habilitada)
docker-compose up geth
```

O nó estará disponível em:
- **RPC:** http://localhost:8545
- **WebSocket:** ws://localhost:8546

### Passo 6: Verificar se o Nó Está Funcionando

Em outro terminal:

```bash
# Acessa o console do Geth
docker exec -it geth-node geth attach http://localhost:8545

# No console, teste:
> eth.accounts
> eth.blockNumber
> miner.start(1)  # Inicia mineração se não estiver automática
```

## Conectar Múltiplos Nós (Rede P2P)

### No Primeiro Nó (Bootnode):

1. Inicie o nó normalmente:
```bash
docker-compose up geth
```

2. Obtenha o enode:
```bash
docker exec -it geth-node geth attach http://localhost:8545 --exec "admin.nodeInfo.enode"
```

O resultado será algo como:
```
"enode://a979fb575495b8d6db44f750317d0f4622bf4c2aa3365d6af7c284339968eef29b69ad0dce72a4d8db5ebb4968de0e3bec910127f134779fbcb0cb6d333116fc@[::]:30303"
```

**Substitua `[::]` pelo IP real da máquina!** Exemplo:
```
enode://a979fb575495b8d6db44f750317d0f4622bf4c2aa3365d6af7c284339968eef29b69ad0dce72a4d8db5ebb4968de0e3bec910127f134779fbcb0cb6d333116fc@192.168.1.100:30303
```

### No Segundo Nó (Peer):

1. No mesmo repositório, inicialize a blockchain:
```bash
mkdir -p data
docker-compose run --rm geth --datadir /root/.ethereum init /genesis.json
docker-compose run --rm geth --datadir /root/.ethereum account new
```

2. Inicie o nó conectando ao bootnode:
```bash
# Use o script helper
chmod +x scripts/conectar-peer.sh
./scripts/conectar-peer.sh "enode://...@192.168.1.100:30303"

# Ou manualmente, edite docker-compose.yml e adicione ao command:
# --bootnodes=enode://...@192.168.1.100:30303
docker-compose up geth
```

3. Verifique a conexão:
```bash
docker exec -it geth-node geth attach http://localhost:8545 --exec "admin.peers"
```

## Compilar e Implantar o Smart Contract

### Opção 1: Usando Solc (Compilador Solidity)

```bash
# Compila o contrato
solc --abi --bin contracts/GameEconomy.sol -o build/

# Os arquivos gerados estarão em:
# - build/GameEconomy.abi
# - build/GameEconomy.bin
```

### Opção 2: Usando Hardhat (Recomendado)

```bash
# Instale Hardhat
npm install --save-dev hardhat

# Crie um projeto Hardhat
npx hardhat init

# Copie o contrato para contracts/
# Configure hardhat.config.js
# Compile e faça deploy
npx hardhat compile
npx hardhat run scripts/deploy.js --network localhost
```

### Opção 3: Deploy Manual via Console Geth

```bash
# Acesse o console
docker exec -it geth-node geth attach http://localhost:8545

# No console (exemplo simplificado):
> var bytecode = "0x<cole-o-bytecode-aqui>"
> var abi = [<cole-o-abi-aqui>]
> var contract = eth.contract(abi)
> var deployed = contract.new({from: eth.accounts[0], data: bytecode, gas: 3000000})
> // Aguarde alguns blocos...
> deployed.address  // Este é o endereço do contrato!
```

## Compilar e Executar o Cliente Go

### Passo 1: Instalar Dependências

```bash
cd cliente/
go mod download
```

### Passo 2: Compilar

```bash
go build -o jogo-cartas main.go
```

### Passo 3: Executar

```bash
./jogo-cartas
```

### Passo 4: Usar Docker (Alternativa)

```bash
# Compila e executa em container
docker-compose build cliente
docker-compose run cliente
```

## Obter ETH para Testes

Em uma rede privada, você precisa de ETH para pagar gas. Opções:

### Opção 1: Mineração

```bash
# Inicia mineração (receberá recompensas)
docker exec -it geth-node geth attach http://localhost:8545 --exec "miner.start(1)"

# Aguarde alguns blocos serem minerados
# Verifique seu saldo:
docker exec -it geth-node geth attach http://localhost:8545 --exec "eth.getBalance(eth.accounts[0])"
```

### Opção 2: Saldo Inicial no Genesis

Edite `genesis.json` e adicione seu endereço em `alloc`:

```json
"alloc": {
  "0xSEU_ENDERECO_AQUI": {
    "balance": "0x2000000000000000000000000000000000000000000000000000000000000"
  }
}
```

Depois, reinicialize a blockchain (isso apaga todos os dados!):
```bash
rm -rf data/
docker-compose run --rm geth --datadir /root/.ethereum init /genesis.json
```

## Troubleshooting

### Erro: "port already in use"

**Solução:** Pare outros processos usando as portas 8545, 8546 ou 30303:
```bash
# Linux
sudo lsof -i :8545
sudo kill -9 <PID>

# Windows
netstat -ano | findstr :8545
taskkill /PID <PID> /F
```

### Erro: "cannot connect to docker daemon"

**Solução:** Certifique-se de que o Docker está rodando:
```bash
sudo systemctl start docker  # Linux
# Ou inicie Docker Desktop no Windows/macOS
```

### Erro: "insufficient funds for gas"

**Solução:** Você precisa de ETH. Minerar alguns blocos:
```bash
docker exec -it geth-node geth attach http://localhost:8545 --exec "miner.start(1)"
```

### Erro: "peer connection failed"

**Soluções:**
1. Verifique se ambos os nós estão na mesma rede
2. Verifique firewall (porta 30303 deve estar aberta)
3. Use o IP correto (não localhost) no enode
4. Certifique-se de que `network_mode: host` está no docker-compose.yml

## Próximos Passos

Após a instalação bem-sucedida:

1. ✅ Nó Geth rodando e minerando
2. ✅ Conta criada e com saldo
3. ✅ Contrato GameEconomy implantado
4. ✅ Cliente Go compilado e executando
5. ✅ Múltiplos nós conectados (se aplicável)

Agora você pode começar a usar o sistema! Consulte o README.md principal para mais informações sobre como usar o cliente.

