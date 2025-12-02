# 🎮 Guia de Execução de Múltiplos Clientes

## 📋 Pré-requisitos

1. ✅ Blockchain rodando (`start-all.bat` já executado)
2. ✅ Servidores de jogo rodando (`start-all.bat` já executado)
3. ✅ Cliente compilado (`setup-game.bat` já executado)

---

## 🚀 Como Conectar Múltiplos Clientes

### **Windows:**

#### **Cliente 1:**
1. Abra um terminal (PowerShell ou CMD)
2. Navegue até a pasta de scripts:
   ```powershell
   cd "C:\Users\bluti\OneDrive\Desktop\UEFS\5 Semestre\MI - Concorrência e Conectividade\Problema3-Concorrencia-Conectividade\Projeto\scripts"
   ```
3. Execute o cliente:
   ```powershell
   .\run-cliente.bat
   ```
4. Digite seu nome (ex: `felipe`)
5. Escolha o servidor (1, 2 ou 3)
6. **Mantenha este terminal aberto!**

#### **Cliente 2:**
1. **Abra um NOVO terminal** (não feche o primeiro!)
2. Navegue até a mesma pasta de scripts:
   ```powershell
   cd "C:\Users\bluti\OneDrive\Desktop\UEFS\5 Semestre\MI - Concorrência e Conectividade\Problema3-Concorrencia-Conectividade\Projeto\scripts"
   ```
3. Execute o cliente novamente:
   ```powershell
   .\run-cliente.bat
   ```
4. Digite um nome diferente (ex: `maria`)
5. Escolha o mesmo servidor ou outro (recomendado: mesmo servidor para jogar juntos)
6. **Mantenha este terminal aberto também!**

#### **Cliente 3, 4, 5... (quantos quiser):**
- Repita os passos acima em **novos terminais**
- Cada cliente precisa de seu próprio terminal
- Cada cliente pode ter um nome diferente
- Todos podem se conectar ao mesmo servidor ou servidores diferentes

---

### **Linux/Mac:**

#### **Cliente 1:**
1. Abra um terminal
2. Navegue até a pasta de scripts:
   ```bash
   cd "/caminho/para/Projeto/scripts"
   ```
3. Execute o cliente:
   ```bash
   ./run-cliente.sh
   ```
4. Digite seu nome
5. Escolha o servidor
6. **Mantenha este terminal aberto!**

#### **Cliente 2:**
1. **Abra um NOVO terminal** (ou nova aba)
2. Navegue até a mesma pasta:
   ```bash
   cd "/caminho/para/Projeto/scripts"
   ```
3. Execute:
   ```bash
   ./run-cliente.sh
   ```
4. Digite um nome diferente
5. Escolha o servidor
6. **Mantenha este terminal aberto!**

---

## 📝 Resumo Visual

```
Terminal 1                    Terminal 2                    Terminal 3
┌─────────────────┐          ┌─────────────────┐          ┌─────────────────┐
│ Cliente 1       │          │ Cliente 2       │          │ Cliente 3       │
│ Nome: felipe    │          │ Nome: maria     │          │ Nome: joão      │
│ Servidor: 1     │          │ Servidor: 1     │          │ Servidor: 2     │
│                 │          │                 │          │                 │
│ > /jogar        │          │ > /jogar        │          │ > /jogar        │
│ Aguardando...   │          │ Aguardando...   │          │ Aguardando...   │
│                 │          │                 │          │                 │
│ Partida!        │          │ Partida!        │          │ Partida!        │
└─────────────────┘          └─────────────────┘          └─────────────────┘
         │                            │                            │
         └────────────────────────────┼────────────────────────────┘
                                      │
                         ┌────────────▼────────────┐
                         │   Servidor de Jogo      │
                         │   (Docker Container)    │
                         └─────────────────────────┘
```

---

## ⚠️ Pontos Importantes

1. **Cada cliente precisa de seu próprio terminal**
   - Não tente rodar dois clientes no mesmo terminal
   - Cada terminal é uma instância independente

2. **Nomes diferentes**
   - Cada cliente deve ter um nome único
   - Ex: `felipe`, `maria`, `joão`, etc.

3. **Mesmo servidor para jogar juntos**
   - Se quiser que os clientes joguem entre si, conecte todos ao mesmo servidor
   - Ex: Todos escolhem "1" (Servidor 1)

4. **Servidores diferentes para testar distribuição**
   - Se quiser testar a distribuição entre servidores, conecte clientes a servidores diferentes
   - Os servidores se comunicam via REST API

5. **Comandos disponíveis em cada cliente:**
   - `/jogar` - Entra na fila de matchmaking
   - `/inventario` - Mostra suas cartas
   - `/comprar` - Compra um pacote de cartas
   - `/ajuda` - Mostra todos os comandos

---

## 🔍 Verificando se Está Funcionando

### Ver logs dos servidores:
```powershell
# Windows
.\logs-servidores.bat

# Linux/Mac
./logs-servidores.sh
```

### Verificar containers rodando:
```powershell
docker ps
```

Você deve ver:
- `geth-node` (blockchain)
- `servidor1`, `servidor2`, `servidor3` (servidores de jogo)
- `broker1`, `broker2`, `broker3` (brokers MQTT)

---

## 🎯 Exemplo Prático: Jogar uma Partida

1. **Terminal 1** - Execute `run-cliente.bat`:
   - Nome: `felipe`
   - Servidor: `1`
   - Comando: `/jogar`

2. **Terminal 2** - Execute `run-cliente.bat` (novo terminal):
   - Nome: `maria`
   - Servidor: `1` (mesmo servidor!)
   - Comando: `/jogar`

3. **Resultado:**
   - Ambos entram na fila
   - O servidor faz o matchmaking
   - Uma partida é criada automaticamente
   - Os dois clientes começam a jogar!

---

## ❓ Problemas Comuns

### "Erro ao conectar ao MQTT"
- Verifique se os servidores estão rodando: `docker ps`
- Verifique se a porta está correta (1886, 1884, 1885)
- Execute `start-all.bat` novamente se necessário

### "Cliente não encontra o servidor"
- Certifique-se de que escolheu o mesmo servidor em ambos os clientes
- Verifique os logs: `logs-servidores.bat`

### "Não consigo ver o outro jogador"
- Ambos devem estar conectados ao mesmo servidor
- Ambos devem ter executado `/jogar`
- Aguarde alguns segundos para o matchmaking

---

## 📚 Comandos do Jogo

Uma vez conectado, você pode usar:

- `/jogar` - Entra na fila para encontrar um oponente
- `/inventario` - Mostra suas cartas
- `/comprar` - Compra um pacote de cartas (requer blockchain)
- `/trocar <carta1> <carta2>` - Troca cartas com outro jogador
- `/ajuda` - Mostra todos os comandos disponíveis

---

**Dica:** Para testar rapidamente, abra 2 terminais lado a lado e execute `run-cliente.bat` em cada um!



