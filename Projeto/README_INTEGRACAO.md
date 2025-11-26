# Integração Problema 2 + Problema 3

Este projeto integra o jogo distribuído (Problema 2) com a blockchain Ethereum (Problema 3).

## Estrutura do Projeto

```
Projeto/
├── Blockchain/          # Infraestrutura blockchain (Geth, contratos, scripts)
│   ├── contracts/       # Smart contracts Solidity
│   ├── data/           # Dados do Geth (chaindata, keystore)
│   ├── scripts/        # Scripts de setup da blockchain
│   └── tools/          # Utilitários Go para blockchain
│
├── Jogo/               # Aplicação do jogo distribuído
│   ├── servidor/       # Servidores de jogo (com suporte blockchain)
│   ├── cliente/        # Cliente do jogo (com carteira integrada)
│   ├── protocolo/      # Protocolo de comunicação
│   └── mosquitto/      # Configuração dos brokers MQTT
│
└── scripts/            # Scripts unificados
    ├── setup-blockchain.bat
    ├── setup-game.bat
    ├── start-all.bat
    ├── stop-all.bat
    └── criar-conta-jogador.bat
```

## Como Funciona a Integração

### 1. **Login e Autenticação**
- O cliente usa uma carteira Ethereum (keystore) para autenticação
- O servidor verifica a assinatura do cliente para garantir identidade
- Cada jogador é identificado pelo endereço da sua carteira

### 2. **Compra de Cartas**
- O cliente compra pacotes diretamente na blockchain (via smart contract)
- O servidor consulta a blockchain para obter o inventário atualizado
- As cartas são NFTs únicos na blockchain

### 3. **Troca de Cartas**
- As trocas são processadas 100% na blockchain
- O smart contract garante atomicidade (ou troca tudo ou não troca nada)
- O servidor sincroniza o estado local com a blockchain

### 4. **Partidas**
- A lógica de jogo (turnos, ataques, defesas) roda no servidor (rápido)
- O resultado final é registrado na blockchain (persistente)
- O servidor valida que o jogador possui as cartas que está usando

## Setup Inicial

### Passo 1: Configurar Blockchain
```bash
cd scripts
setup-blockchain.bat
```

Este script:
- Cria a rede privada Ethereum
- Gera o genesis block
- Inicia o nó Geth
- Faz deploy do smart contract
- Salva o endereço do contrato em `contract-address.txt`

### Passo 2: Configurar Jogo
```bash
cd scripts
setup-game.bat
```

Este script:
- Compila o servidor e cliente
- Verifica se o contrato foi deployado

### Passo 3: Criar Contas de Jogadores
```bash
cd scripts
criar-conta-jogador.bat
```

Execute este script para cada jogador que vai participar.

### Passo 4: Iniciar Tudo
```bash
cd scripts
start-all.bat
```

Este script:
- Inicia a blockchain
- Inicia os servidores de jogo
- Inicia os brokers MQTT

## Uso

### Para Jogadores

1. **Primeira vez:**
   - Execute `criar-conta-jogador.bat` para criar sua carteira
   - Guarde o arquivo keystore e a senha em local seguro

2. **Ao iniciar o jogo:**
   - Execute o cliente: `Jogo/cliente/cliente.exe`
   - Selecione seu arquivo keystore
   - Digite sua senha
   - O cliente se conecta ao servidor via MQTT

3. **Comprar cartas:**
   - Use o comando `/comprar` no cliente
   - O cliente envia transação para a blockchain
   - O servidor sincroniza seu inventário

4. **Jogar:**
   - Use `/jogar` para entrar na fila de matchmaking
   - Quando encontrar oponente, a partida inicia
   - Use `/jogar_carta <id>` para jogar uma carta
   - O servidor valida que você possui a carta na blockchain

5. **Trocar cartas:**
   - Use `/trocar` para iniciar uma troca
   - A troca é processada na blockchain
   - Ambos os jogadores recebem as cartas trocadas

## Variáveis de Ambiente

O servidor usa as seguintes variáveis de ambiente (configuradas no docker-compose.yml):

- `BLOCKCHAIN_RPC_URL`: URL do nó Geth (ex: `http://geth:8545`)
- `CONTRACT_ADDRESS`: Endereço do smart contract deployado
- `KEYSTORE_PATH`: Caminho para o keystore (dentro do container)
- `SERVER_PASSWORD`: Senha da conta do servidor (para registrar partidas)

## Arquitetura

### Comunicação
- **Cliente ↔ Servidor**: MQTT (Pub/Sub)
- **Servidor ↔ Servidor**: REST API (HTTP)
- **Cliente ↔ Blockchain**: JSON-RPC (HTTP)
- **Servidor ↔ Blockchain**: JSON-RPC (HTTP)

### Persistência
- **Cartas e Inventários**: Blockchain (imutável, verificável)
- **Estado de Partidas**: Servidor (rápido, temporário)
- **Resultados de Partidas**: Blockchain (registro permanente)

## Troubleshooting

### Blockchain não inicia
- Verifique se Docker está rodando
- Verifique se a porta 8545 está livre
- Execute `docker logs geth-node` para ver erros

### Servidor não conecta à blockchain
- Verifique se `CONTRACT_ADDRESS` está definido
- Verifique se o contrato foi deployado
- Verifique se o Geth está acessível na rede Docker

### Cliente não encontra carteira
- Verifique se o arquivo keystore existe
- Verifique se o caminho está correto
- Tente criar uma nova conta com `criar-conta-jogador.bat`

## Próximos Passos

- [ ] Implementar sincronização automática de inventário
- [ ] Adicionar eventos de blockchain para notificações em tempo real
- [ ] Implementar sistema de recompensas baseado em partidas
- [ ] Adicionar interface web para visualização de cartas

