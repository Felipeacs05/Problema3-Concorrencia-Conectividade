# Scripts de Setup

Este diretório contém scripts para configurar e gerenciar a infraestrutura do projeto.

## Scripts Disponíveis

### Windows (.bat)
- `setup-blockchain.bat` - Configura a blockchain privada
- `setup-game.bat` - Compila servidor e cliente
- `criar-conta-jogador.bat` - Cria nova carteira para jogador
- `start-all.bat` - Inicia toda a infraestrutura
- `stop-all.bat` - Para toda a infraestrutura

### Linux/Mac (.sh)
- `setup-blockchain.sh` - Configura a blockchain privada
- `setup-game.sh` - Compila servidor e cliente
- `criar-conta-jogador.sh` - Cria nova carteira para jogador
- `start-all.sh` - Inicia toda a infraestrutura
- `stop-all.sh` - Para toda a infraestrutura

## Como Usar no Linux

### 1. Dar permissão de execução aos scripts

```bash
cd scripts
chmod +x *.sh
```

### 2. Executar os scripts na ordem:

```bash
# 1. Configurar blockchain
./setup-blockchain.sh

# 2. Configurar jogo
./setup-game.sh

# 3. Criar conta de jogador (opcional, para cada jogador)
./criar-conta-jogador.sh

# 4. Iniciar tudo
./start-all.sh

# Para parar tudo
./stop-all.sh
```

## Requisitos

- Go 1.22+ instalado
- Docker e Docker Compose instalados e rodando
- Acesso à internet (para baixar imagens Docker)

## Notas

- Os scripts `.bat` são para Windows
- Os scripts `.sh` são para Linux/Mac
- Ambos fazem a mesma coisa, apenas adaptados para cada sistema operacional
- No Linux, certifique-se de que os scripts `.sh` têm permissão de execução (`chmod +x`)

