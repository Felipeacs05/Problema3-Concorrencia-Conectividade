# Guia Completo - Windows (Do Zero ao Funcionamento)

Este guia leva você do zero até ter a blockchain funcionando perfeitamente.

## 🚀 Passo a Passo Completo

### PASSO 1: Resetar Tudo (Começar do Zero)

```cmd
cd scripts
resetar-tudo.bat
cd ..
```

**O que este script faz:**
- Para todos os containers
- Remove dados antigos
- Cria nova conta
- Gera genesis.json automaticamente
- Inicializa blockchain
- Inicia o nó Geth

**Tempo estimado:** 1-2 minutos

---

### PASSO 2: Aguardar Inicialização

```cmd
REM Aguarde o script anterior terminar completamente
REM Depois aguarde mais 10 segundos
timeout /t 10 /nobreak
```

---

### PASSO 3: Iniciar Clique (Criar Blocos)

```cmd
cd scripts
iniciar-clique.bat
cd ..
```

**O que faz:** Inicia o mecanismo Clique (Proof of Authority) que cria blocos automaticamente.

**Resultado esperado:** Nenhuma mensagem de erro.

---

### PASSO 4: Verificar se Está Funcionando

```cmd
cd scripts
verificar-bloco.bat
cd ..
```

**Resultado esperado:** Um número (ex: `1`, `5`, `10`)

**Teste:** Execute novamente após 10 segundos - o número deve aumentar!

```cmd
timeout /t 10 /nobreak
cd scripts
verificar-bloco.bat
cd ..
```

---

### PASSO 5: Verificar Status Completo

```cmd
cd scripts
verificar-status.bat
cd ..
```

**O que mostra:**
- Número de blocos
- Contas disponíveis
- Saldo da primeira conta

---

## ✅ Checklist de Verificação

Execute estes comandos para verificar se tudo está OK:

```cmd
REM 1. Container está rodando?
docker ps
REM Deve mostrar "geth-node" com status "Up"

REM 2. Blocos estão sendo criados?
cd scripts
verificar-bloco.bat
cd ..
REM Execute 2 vezes com intervalo de 10 segundos
REM O número deve aumentar!

REM 3. Conta está visível?
cd scripts
obter-endereco-simples.bat
cd ..
REM Deve mostrar um endereço começando com 0x

REM 4. Logs não mostram erros?
docker-compose logs geth --tail 20
REM Não deve ter mensagens "Fatal" ou "ERROR"
```

---

## 🔧 Se Algo Der Errado

### Erro: "Container is restarting"

```cmd
REM Veja os logs para identificar o erro
docker-compose logs geth --tail 50

REM Se o erro for sobre genesis, execute:
cd scripts
resetar-tudo.bat
cd ..
```

### Erro: "can't start clique chain without signers"

```cmd
REM O genesis.json não tem signer configurado
REM Execute o reset completo:
cd scripts
resetar-tudo.bat
cd ..
```

### Erro: "database contains incompatible genesis"

```cmd
REM Há dados antigos incompatíveis
REM Execute o reset completo:
cd scripts
resetar-tudo.bat
cd ..
```

### Blocos não estão sendo criados

```cmd
REM Verifique se o Clique está iniciado
cd scripts
iniciar-clique.bat
cd ..

REM Aguarde 10 segundos e verifique novamente
timeout /t 10 /nobreak
cd scripts
verificar-bloco.bat
cd ..
```

---

## 📝 Comandos Úteis

```cmd
REM Ver logs em tempo real
docker-compose logs -f geth

REM Parar tudo
docker-compose down

REM Reiniciar
docker-compose restart geth

REM Acessar console do Geth
docker exec -it geth-node geth attach http://localhost:8545
```

---

## 🎯 Resultado Final Esperado

Quando tudo estiver funcionando:

1. ✅ Container `geth-node` rodando
2. ✅ Blocos sendo criados (número aumenta a cada 5-10 segundos)
3. ✅ Conta visível quando lista contas
4. ✅ Logs sem erros fatais
5. ✅ Clique ativo e criando blocos

---

## 🚀 Próximos Passos (Após Blockchain Funcionando)

1. **Compilar Cliente Go:**
```cmd
cd cliente
go mod download
go build -o jogo-cartas.exe main.go
```

2. **Fazer Deploy do Contrato:**
   - Use Hardhat ou Remix
   - Veja seção "Deploy do Smart Contract" no README

3. **Executar Cliente:**
```cmd
cd cliente
jogo-cartas.exe
```

---

**Última atualização:** Novembro 2024


