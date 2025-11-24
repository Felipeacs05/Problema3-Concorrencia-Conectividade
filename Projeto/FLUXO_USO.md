# Fluxo de Uso da Blockchain

## Quando usar cada script?

### 🆕 PRIMEIRA VEZ (Setup Completo)

**Quando?** Nunca executou o projeto antes OU quer começar do zero (limpar tudo).

**O que faz:**
1. Para containers Docker antigos
2. **DELETA** toda a blockchain existente (`data/geth/chaindata/`)
3. **DELETA** todas as contas (`data/keystore/`)
4. Cria nova conta de minerador (signer)
5. Gera novo arquivo `genesis.json` com a conta
6. Inicializa blockchain do zero
7. Inicia container Geth com mineração ativada

**Comando:**
```bash
cd scripts
setup.bat          # Windows
./setup.sh         # Linux
```

**Resultado:**
- Blockchain vazia (bloco 0)
- 1 conta com 1.000.000 ETH
- Container rodando e minerando

---

### 🔄 USO NORMAL (Dia a Dia)

**Quando?** Blockchain já existe e você quer continuar usando.

**O que faz:**
1. Apenas inicia o container Docker
2. Carrega a blockchain existente
3. Continua minerando de onde parou

**Comando:**
```bash
docker-compose up -d
```

**Resultado:**
- Blockchain continua do último bloco
- Todas as contas e contratos preservados
- Saldo mantido

---

### 🎮 Jogar (Cliente)

**Quando?** Container já está rodando.

**Comando:**
```bash
cd scripts
test-client.bat    # Windows
```

**Fluxo:**
1. Conecta ao Geth (porta 8545)
2. Carrega conta existente (com senha)
3. Exibe menu do jogo

---

## Comparação: Setup vs Uso Normal

| Item | Setup (1ª vez) | Uso Normal |
|------|----------------|------------|
| **Blockchain** | Cria nova (bloco 0) | Continua existente |
| **Contas** | Deleta e cria nova | Mantém todas |
| **Saldo** | Reset para 1M ETH | Mantém saldo real |
| **Contratos** | Nenhum deployado | Mantém deployados |
| **Quando usar** | 1ª vez ou resetar | Sempre que já existe |
| **Tempo** | ~30 segundos | ~5 segundos |

---

## Fluxo Completo de Uso

### Primeira Vez
```bash
# 1. Setup inicial (UMA VEZ)
cd scripts
setup.bat

# 2. Anotar a SENHA que você criou (ex: 123456)

# 3. Jogar
test-client.bat
# Digite a senha quando solicitado
# Opção 6: Deploy do contrato
# Opção 7: Configurar endereço (copie o endereço que apareceu)
# Opção 2: Comprar pacote
# Opção 1: Ver suas cartas
```

### Próximas Vezes
```bash
# 1. Verificar se container está rodando
docker ps

# Se NÃO estiver rodando:
docker-compose up -d

# 2. Jogar
cd scripts
test-client.bat
# Digite a senha (a mesma de antes)
# O contrato e suas cartas estarão lá!
```

---

## Verificações Úteis

### Container rodando?
```bash
docker ps
# Deve mostrar: geth-node
```

### Blockchain minerando?
```bash
cd scripts
check-block.bat
# Se o número aumentar, está minerando
```

### Ver saldo da conta?
```bash
cd scripts
check-balance.bat
```

---

## Troubleshooting

### "Contrato não configurado"
- Você precisa fazer deploy (opção 6) na primeira vez
- Depois configure o endereço (opção 7)

### "Timeout aguardando confirmação"
- A mineração pode ter parado
- Execute: `unlock-account.bat` para reativar

### "Senha incorreta"
- Você precisa usar a MESMA senha do setup
- Se esqueceu, execute setup.bat novamente (perde tudo)

### "Transação falhou (status: 0)"
- Gas insuficiente (corrigido na última versão)
- OU contrato muito grande (otimizado)

---

## Resumo Rápido

**Setup:** Apenas na primeira vez ou para resetar.
**Normal:** `docker-compose up -d` + `test-client.bat`
**Senha:** Anote! Não tem como recuperar.

