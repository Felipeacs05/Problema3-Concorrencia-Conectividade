# 📋 Como Verificar Transações e Eventos na Blockchain

## 🔍 Diferença entre Transferência de ETH e Registro de Ações do Jogo

### 1. **Transferência de ETH** 💰
- **O que é:** Movimentação de moeda (Ether) entre contas
- **Exemplo:** Quando você executa `fundar-conta.bat`, está transferindo ETH da conta do servidor para a conta do jogador
- **Propósito:** Fornecer "gas" (combustível) para pagar as taxas das transações
- **Como verificar:** Use `view-transactions.bat` com o endereço da conta

### 2. **Registro de Ações do Jogo** 🎮
- **O que é:** Chamadas de funções do contrato inteligente que modificam o estado do jogo
- **Exemplos:**
  - **Comprar Pacote:** Chama `comprarPacote()` → cria 5 NFTs (cartas) e atribui ao jogador
  - **Trocar Cartas:** Chama `criarPropostaTroca()` e `aceitarPropostaTroca()` → transfere NFTs entre jogadores
  - **Registrar Partida:** Chama `registrarPartida()` → salva o resultado da partida na blockchain
- **Propósito:** Garantir propriedade verificável das cartas (NFTs) e transparência total
- **Como verificar:** Use `view-events.bat` com o endereço do contrato

### 3. **Resumo Visual**

```
┌─────────────────────────────────────────────────────────┐
│  TRANSFERÊNCIA DE ETH                                   │
│  ────────────────────────                                │
│  De: 0xServidor → Para: 0xJogador                      │
│  Valor: 100 ETH                                         │
│  Tipo: Transação simples de moeda                       │
│  Resultado: Jogador tem ETH para pagar gas             │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│  REGISTRO DE AÇÃO DO JOGO (Compra de Pacote)            │
│  ───────────────────────────────────────────            │
│  De: 0xJogador → Para: 0xContrato                      │
│  Função: comprarPacote()                                │
│  Valor: 1 ETH (preço do pacote)                         │
│  Resultado:                                              │
│    - 5 NFTs (cartas) criados                            │
│    - Propriedade atribuída ao jogador                   │
│    - Eventos emitidos na blockchain                     │
└─────────────────────────────────────────────────────────┘
```

## 🛠️ Ferramentas para Visualizar

### **1. Visualizar Transações de uma Conta**

```bash
# Windows
cd scripts
.\view-transactions.bat 0x2d50FD74Cc3eB554b16013667045804D28Bc54a4

# Linux/Mac
cd scripts
./view-transactions.sh 0x2d50FD74Cc3eB554b16013667045804D28Bc54a4
```

**O que você verá:**
- Todas as transações enviadas e recebidas pela conta
- Valor em ETH transferido
- Gas usado
- Status (sucesso/falha)
- Hash da transação
- Bloco e timestamp

**Exemplo de saída:**
```
═══════════════════════════════════════════════════════════
Transação #1
Hash: 0xabc123...
Bloco: 42
De: 0xServidor
Para: 0xJogador
Valor: 100.000000000000000000 ETH
Gas usado: 21000
Status: ✓ Sucesso
Timestamp: 2024-01-15 14:30:25
```

### **2. Visualizar Eventos do Contrato**

Primeiro, obtenha o endereço do contrato:
```bash
# Windows
type ..\Blockchain\contract-address.txt

# Linux/Mac
cat ../Blockchain/contract-address.txt
```

Depois, visualize os eventos:
```bash
# Windows
cd scripts
.\view-events.bat 0x4D82F25Ef5058fE2308135D88A386c90FEdBe99A

# Linux/Mac
cd scripts
./view-events.sh 0x4D82F25Ef5058fE2308135D88A386c90FEdBe99A
```

**O que você verá:**
- Eventos emitidos pelo contrato (compras, trocas, partidas)
- Detalhes de cada evento (quem comprou, quais cartas foram criadas, etc.)
- Hash da transação que gerou o evento
- Bloco e timestamp

**Tipos de eventos que você pode ver:**
- `CartaCriada` - Quando uma carta NFT é criada
- `PacoteComprado` - Quando um jogador compra um pacote
- `PropostaTrocaCriada` - Quando uma troca é proposta
- `TrocaExecutada` - Quando uma troca é concluída
- `PartidaRegistrada` - Quando uma partida é registrada

**Exemplo de saída:**
```
═══════════════════════════════════════════════════════════
Evento #1
Bloco: 45
Hash da transação: 0xdef456...
Timestamp: 2024-01-15 14:35:10
Tipo: PacoteComprado
Dados: Jogador=0xJogador, Quantidade=5
Parâmetros indexados: comprador=0xJogador
```

## 📊 Fluxo Completo de uma Compra

1. **Jogador executa `/comprar` no cliente**
   - Cliente chama `comprarPacoteBlockchain()`
   - Assina transação com a chave privada da carteira

2. **Transação enviada para a blockchain**
   - Hash da transação: `0xabc123...`
   - Status: Pendente

3. **Geth processa a transação**
   - Valida a assinatura
   - Executa `comprarPacote()` no contrato
   - Cria 5 NFTs (cartas)
   - Emite eventos `CartaCriada` (5x) e `PacoteComprado` (1x)

4. **Transação confirmada**
   - Status: Sucesso
   - Bloco: 42
   - Gas usado: 800000

5. **Verificação:**
   ```bash
   # Ver a transação
   .\view-transactions.bat 0xJogador
   
   # Ver os eventos (compras, cartas criadas)
   .\view-events.bat 0xContrato
   ```

## 🔐 Segurança e Transparência

- **Todas as ações são imutáveis:** Uma vez registradas na blockchain, não podem ser alteradas
- **Verificação pública:** Qualquer pessoa pode verificar todas as transações e eventos
- **Propriedade verificável:** A blockchain prova quem é o dono de cada carta (NFT)
- **Prevenção de fraude:** Impossível duplicar cartas ou falsificar propriedade

## ❓ Perguntas Frequentes

**Q: Uma transferência de ETH é um registro?**
R: Sim, mas são tipos diferentes:
- **Transferência de ETH:** Registro de movimentação de moeda
- **Registro de ação do jogo:** Registro de mudança no estado do jogo (criação de NFTs, trocas, etc.)

**Q: Como sei se minha compra foi registrada?**
R: Use `view-events.bat` com o endereço do contrato e procure pelo evento `PacoteComprado` com seu endereço.

**Q: Onde vejo os logs em tempo real?**
R: Você pode usar:
```bash
# Logs do Geth
docker logs -f geth

# Ou use as ferramentas para buscar eventos recentes
.\view-events.bat 0xContrato $(($(docker exec geth geth attach --exec eth.blockNumber | tr -d '"') - 100)) $(docker exec geth geth attach --exec eth.blockNumber | tr -d '"')
```

**Q: Posso ver transações de outros jogadores?**
R: Sim! A blockchain é pública. Use `view-transactions.bat` com o endereço de qualquer conta.

