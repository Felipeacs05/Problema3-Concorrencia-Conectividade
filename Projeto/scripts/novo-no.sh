#!/bin/bash
# ===================== NOVO NÓ (SEM SUDO) =====================
# Script para criar e iniciar um segundo nó Geth
# Solução para: Admin API, Erro de Tabs no YAML e Permissão de Arquivos

set -e

# Definição de Caminhos
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BLOCKCHAIN_DIR="$PROJECT_DIR/Blockchain"
TOOLS_DIR="$BLOCKCHAIN_DIR/tools"
DATA2_DIR="$BLOCKCHAIN_DIR/data2"
KEYSTORE2_DIR="$DATA2_DIR/keystore"
GENESIS_FILE="$BLOCKCHAIN_DIR/genesis.json"
PEER_COMPOSE_FILE="$BLOCKCHAIN_DIR/docker-compose-peer.yml"

# Detecta docker compose
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi

echo "========================================"
echo "Criando Novo Nó (Peer) - Modo Seguro"
echo "========================================"
echo ""

# 1. Verifica Docker
if ! docker ps >/dev/null 2>&1; then
    echo "ERRO: Docker não está rodando!"
    exit 1
fi

# 2. Verifica Nó 1
echo "[1/8] Verificando Nó 1..."
if ! docker ps | grep -q "geth-node"; then
    echo "ERRO: O primeiro nó (geth-node) não está rodando!"
    exit 1
fi

# Aguarda IPC do Nó 1
echo "Aguardando IPC do Nó 1..."
for i in {1..30}; do
    if docker exec geth-node test -S /root/.ethereum/geth.ipc; then
        break
    fi
    sleep 2
done

# 3. Limpeza sem SUDO (Usa Docker para apagar arquivos do Docker)
echo "[2/8] Limpando dados antigos..."
# Remove container antigo se existir
docker rm -f geth-peer 2>/dev/null || true

# Apaga a pasta data2 usando um container auxiliar (bypass de permissão root)
if [ -d "$DATA2_DIR" ]; then
    echo "   Removendo pasta data2 (via docker)..."
    docker run --rm -v "$BLOCKCHAIN_DIR":/blockchain alpine rm -rf /blockchain/data2
fi

# Recria diretórios com permissão do usuário atual
mkdir -p "$KEYSTORE2_DIR"
mkdir -p "$DATA2_DIR"
echo "[OK] Limpeza concluída"
echo ""

# 4. Criação de Conta
echo "[3/8] Criando conta para o Nó 2..."
# Cria arquivo de senha
echo "123456" > "$DATA2_DIR/password.txt"

# Cria conta usando imagem do geth (garante compatibilidade)
docker run --rm \
  -v "$KEYSTORE2_DIR":/keystore \
  -v "$DATA2_DIR/password.txt":/password.txt \
  ethereum/client-go:v1.13.15 account new --keystore /keystore --password /password.txt >/dev/null 2>&1

# Pega o endereço da conta criada
ADDRESS=$(ls "$KEYSTORE2_DIR" | grep UTC | head -n 1 | awk -F'--' '{print $3}')
# Adiciona 0x se não tiver
if [[ "$ADDRESS" != 0x* ]]; then ADDRESS="0x$ADDRESS"; fi

echo "[OK] Conta criada: $ADDRESS"
echo ""

# 5. Inicialização (Genesis)
echo "[4/8] Inicializando blockchain do Nó 2..."
docker run --rm \
  -v "$DATA2_DIR":/root/.ethereum \
  -v "$GENESIS_FILE":/genesis.json \
  ethereum/client-go:v1.13.15 --datadir=/root/.ethereum init /genesis.json
echo "[OK] Inicializado"
echo ""

# 6. Obter Enode do Nó 1 (Via IPC para corrigir erro de admin)
echo "[5/8] Obtendo Enode do Nó 1..."
ENODE=$(docker exec geth-node geth attach --exec "admin.nodeInfo.enode" /root/.ethereum/geth.ipc | tr -d '"')

if [[ -z "$ENODE" ]] || [[ "$ENODE" == *"ReferenceError"* ]]; then
    echo "ERRO CRÍTICO: Falha ao obter enode via IPC."
    exit 1
fi

# Ajusta IP (localhost -> IP real ou host.docker.internal para Mac/Windows, mas IP é mais seguro no Linux)
HOST_IP=$(hostname -I | awk '{print $1}' 2>/dev/null || echo "127.0.0.1")
ENODE=$(echo "$ENODE" | sed "s/@\[::\]:/@${HOST_IP}:/g" | sed "s/@127.0.0.1:/@${HOST_IP}:/g")

echo "[OK] Enode: $ENODE"
echo ""

# 7. Gerar Docker Compose do Peer (Corrigindo TABs)
echo "[6/8] Criando docker-compose-peer.yml..."

# Usa printf para garantir espaços e não tabs
cat > "$PEER_COMPOSE_FILE" <<EOF
services:
  geth-peer:
    image: ethereum/client-go:v1.13.15
    container_name: geth-peer
    ports:
      - "8547:8545"
      - "8548:8546"
      - "30304:30303/udp"
      - "30304:30303"
    volumes:
      - ./data2:/root/.ethereum
      - ./genesis.json:/genesis.json
      - ./password-peer.txt:/root/.ethereum/password.txt
    command:
      - --datadir=/root/.ethereum
      - --networkid=1337
      - --http
      - --http.addr=0.0.0.0
      - --http.port=8545
      - --http.api=eth,net,web3,personal,miner,admin
      - --http.corsdomain=*
      - --http.vhosts=*
      - --ws
      - --ws.addr=0.0.0.0
      - --ws.port=8546
      - --ws.api=eth,net,web3,personal,miner,admin
      - --ws.origins=*
      - --allow-insecure-unlock
      - --unlock=$ADDRESS
      - --password=/root/.ethereum/password.txt
      - --bootnodes=$ENODE
      - --port=30303
    stdin_open: true
    tty: true
    restart: unless-stopped
    networks:
      - peer_network

networks:
  peer_network:
    driver: bridge
EOF

# Copia senha para raiz onde o compose espera
cp "$DATA2_DIR/password.txt" "$BLOCKCHAIN_DIR/password-peer.txt"

echo "[OK] Arquivo gerado com sucesso"
echo ""

# 8. Iniciar Nó 2
echo "[7/8] Iniciando Nó 2..."
cd "$BLOCKCHAIN_DIR"
$DOCKER_COMPOSE -f docker-compose-peer.yml up -d

echo "[8/8] Aguardando inicialização..."
sleep 5

# Verificação Final
if docker ps | grep -q "geth-peer"; then
    echo ""
    echo "========================================"
    echo "✅ SUCESSO: Nó 2 rodando!"
    echo "========================================"
    echo "Execute './verificar-nos.sh' para confirmar a conexão."
else
    echo "❌ ERRO: O container geth-peer falhou ao iniciar."
    docker logs geth-peer
fi