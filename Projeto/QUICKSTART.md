# Guia Rápido de Início

Este é um guia rápido para começar a usar o sistema em poucos minutos.

## Início Rápido (5 minutos)

### 1. Inicializar Blockchain

```bash
mkdir -p data
docker-compose run --rm geth --datadir /root/.ethereum init /genesis.json
docker-compose run --rm geth --datadir /root/.ethereum account new
# Digite uma senha quando solicitado
```

### 2. Iniciar Nó Geth

```bash
docker-compose up -d geth
```

Aguarde alguns segundos para o nó iniciar e começar a minerar.

### 3. Verificar se Está Funcionando

```bash
docker exec -it geth-node geth attach http://localhost:8545 --exec "eth.blockNumber"
```

Se retornar um número (ex: `5`), está funcionando!

### 4. Obter ETH (para pagar transações)

```bash
# Inicia mineração (se não estiver automática)
docker exec -it geth-node geth attach http://localhost:8545 --exec "miner.start(1)"

# Aguarde 10-20 segundos, depois verifique saldo:
docker exec -it geth-node geth attach http://localhost:8545 --exec "eth.getBalance(eth.accounts[0])"
```

### 5. Compilar e Executar Cliente

```bash
cd cliente/
go mod download
go build -o jogo-cartas main.go
./jogo-cartas
```

## Próximos Passos

1. **Fazer Deploy do Contrato:** Use Hardhat, Truffle ou console do Geth
2. **Configurar Contrato no Cliente:** Use a opção 7 do menu
3. **Comprar Primeiro Pacote:** Use a opção 2 do menu

## Comandos Úteis

```bash
# Ver logs do Geth
docker-compose logs -f geth

# Parar o nó
docker-compose down

# Reiniciar o nó
docker-compose restart geth

# Acessar console do Geth
docker exec -it geth-node geth attach http://localhost:8545
```

## Problemas Comuns

**"Cannot connect to Geth"**
→ Verifique se o container está rodando: `docker ps`

**"Insufficient funds"**
→ Minerar alguns blocos: `miner.start(1)` no console do Geth

**"Contract not deployed"**
→ Você precisa fazer deploy do contrato primeiro (veja INSTALACAO.md)

Para mais detalhes, consulte o README.md completo.

