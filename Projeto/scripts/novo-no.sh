#!/bin/bash
# ===================== NOVO NÓ =====================
# Script para criar e iniciar um segundo nó Geth no mesmo computador
# Este script conecta automaticamente ao primeiro nó (geth-node)

set -e  # Para em caso de erro

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BLOCKCHAIN_DIR="$PROJECT_DIR/Blockchain"
TOOLS_DIR="$BLOCKCHAIN_DIR/tools"
DATA_DIR="$BLOCKCHAIN_DIR/data"
DATA2_DIR="$BLOCKCHAIN_DIR/data2"
KEYSTORE2_DIR="$DATA2_DIR/keystore"
GENESIS_FILE="$BLOCKCHAIN_DIR/genesis.json"
PEER_COMPOSE_FILE="$BLOCKCHAIN_DIR/docker-compose-peer.yml"

# Detecta qual comando docker compose usar
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi

echo "========================================"
echo "Criando Novo Nó (Peer)"
echo "========================================"
echo ""

# Verifica se Docker está rodando
if ! docker ps >/dev/null 2>&1; then
    echo "ERRO: Docker não está rodando!"
    echo "Inicie o Docker e tente novamente."
    exit 1
fi

# Verifica se o primeiro nó está rodando
echo "[1/8] Verificando se o primeiro nó está rodando..."
if ! docker ps | grep -q "geth-node"; then
    echo "ERRO: O primeiro nó (geth-node) não está rodando!"
    echo "Execute 'start-all.sh' primeiro para iniciar o nó principal."
    exit 1
fi

# Aguarda o nó 1 estar pronto
echo "Aguardando nó 1 estar pronto..."
for i in {1..30}; do
    # CORREÇÃO 1: Usa IPC em vez de HTTP para garantir acesso a APIs
    if docker exec geth-node geth attach --exec "eth.blockNumber" /root/.ethereum/geth.ipc >/dev/null 2>&1; then
        break
    fi
    if [ $i -eq 30 ]; then
        echo "ERRO: Nó 1 não respondeu após 60 segundos"
        exit 1
    fi
    sleep 2
done
echo "[OK] Nó 1 está pronto"
echo ""

# Verifica se o nó 2 já existe
if docker ps | grep -q "geth-peer"; then
    echo "[AVISO] O nó 2 (geth-peer) já está rodando!"
    read -p "Deseja parar e recriar? (s/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Ss]$ ]]; then
        echo "Parando nó 2 existente..."
        cd "$BLOCKCHAIN_DIR"
        $DOCKER_COMPOSE -f docker-compose-peer.yml down 2>/dev/null || true
    else
        echo "Operação cancelada."
        exit 0
    fi
fi

# Cria diretórios para o nó 2
echo "[2/8] Criando diretórios para o nó 2..."
mkdir -p "$KEYSTORE2_DIR"
mkdir -p "$DATA2_DIR"
echo "[OK] Diretórios criados"
echo ""

# Remove dados antigos do nó 2 (se existirem)
if [ -d "$DATA2_DIR/geth" ]; then
    echo "[3/8] Removendo dados antigos do nó 2..."
    rm -rf "$DATA2_DIR/geth"
    echo "[OK] Dados antigos removidos"
else
    echo "[3/8] Nenhum dado antigo encontrado"
fi
echo ""

# Verifica se Go está instalado (para criar conta)
if ! command -v go &> /dev/null; then
    echo "[AVISO] Go não está instalado. Criando conta manualmente..."
    # Cria conta usando docker
    echo "123456" > "$DATA2_DIR/password.txt"
    docker run --rm -v "$KEYSTORE2_DIR:/keystore" -v "$DATA2_DIR/password.txt:/password.txt" ethereum/client-go:v1.13.15 account new --keystore /keystore --password /password.txt >/dev/null 2>&1
    ADDRESS=$(docker run --rm -v "$KEYSTORE2_DIR:/keystore" ethereum/client-go:v1.13.15 account list --keystore /keystore | head -1 | awk '{print $3}' | tr -d '{}')
else
    # Compila o utilitário blockchain-utils se necessário
    if [ ! -f "$TOOLS_DIR/blockchain-utils" ]; then
        echo "[3.5/8] Compilando blockchain-utils..."
        cd "$TOOLS_DIR"
        go mod tidy
        go build -o blockchain-utils blockchain-utils.go
        cd "$PROJECT_DIR"
    fi
    
    # Cria conta usando blockchain-utils
    echo "[4/8] Criando nova conta para o nó 2..."
    "$TOOLS_DIR/blockchain-utils" criar-conta "$KEYSTORE2_DIR" "123456"
    if [ $? -ne 0 ]; then
        echo "ERRO: Falha ao criar conta"
        exit 1
    fi
    
    # Extrai endereço da conta
    ADDRESS=$("$TOOLS_DIR/blockchain-utils" extrair-endereco "$KEYSTORE2_DIR")
    if [ -z "$ADDRESS" ]; then
        echo "ERRO: Falha ao extrair endereço"
        exit 1
    fi
    
    # Cria arquivo password.txt
    echo "123456" > "$DATA2_DIR/password.txt"
    echo "[OK] Conta criada: $ADDRESS"
fi
echo ""

# Inicializa blockchain do nó 2 com genesis.json
echo "[5/8] Inicializando blockchain do nó 2..."
docker run --rm -v "$DATA2_DIR:/root/.ethereum" -v "$GENESIS_FILE:/genesis.json" ethereum/client-go:v1.13.15 --datadir=/root/.ethereum init /genesis.json
if [ $? -ne 0 ]; then
    echo "ERRO: Falha ao inicializar blockchain do nó 2"
    exit 1
fi
echo "[OK] Blockchain do nó 2 inicializada"
echo ""

# Obtém o enode do nó 1
echo "[6/8] Obtendo enode do nó 1..."
# CORREÇÃO 1: Usa IPC aqui também para ter acesso ao admin.nodeInfo
ENODE=$(docker exec geth-node geth attach --exec "admin.nodeInfo.enode" /root/.ethereum/geth.ipc 2>/dev/null | tr -d '"' | tr -d ' ')

if [ -z "$ENODE" ]; then
    echo "ERRO: Falha ao obter enode do nó 1 (Retorno vazio)"
    exit 1
fi

# Verifica se o retorno parece um erro de JS
if [[ "$ENODE" == *"ReferenceError"* ]]; then
    echo "ERRO: Falha ao obter enode: $ENODE"
    exit 1
fi

# Substitui [::] pelo IP localhost ou IP da máquina
HOST_IP=$(hostname -I | awk '{print $1}' 2>/dev/null || echo "127.0.0.1")
if [ -z "$HOST_IP" ] || [ "$HOST_IP" == "" ]; then
    HOST_IP="127.0.0.1"
fi

# Substitui [::] pelo IP real
ENODE=$(echo "$ENODE" | sed "s/@\[::\]:/@${HOST_IP}:/g" | sed "s/@\[::\]/@${HOST_IP}/g")

echo "[OK] Enode obtido: $ENODE"
echo ""

# Cria o arquivo docker-compose-peer.yml
echo "[7/8] Criando docker-compose-peer.yml..."

# CORREÇÃO 2: Garante indentação com ESPAÇOS e não TABS
cat > "$PEER_COMPOSE_FILE" << EOF
# ===================== NÓ PEER (SEGUNDO NÓ) =====================
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
      - --ws.origins="*"
      - --allow-insecure-unlock
      - --unlock=${ADDRESS}
      - --password=/root/.ethereum/password.txt
      - --bootnodes=${ENODE}
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

# Copia password.txt para password-peer.txt (necessário para o volume)
cp "$DATA2_DIR/password.txt" "$BLOCKCHAIN_DIR/password-peer.txt" 2>/dev/null || echo "123456" > "$BLOCKCHAIN_DIR/password-peer.txt"

echo "[OK] Arquivo docker-compose-peer.yml criado"
echo ""

# Inicia o nó 2
echo "[8/8] Iniciando nó 2..."
cd "$BLOCKCHAIN_DIR"
$DOCKER_COMPOSE -f docker-compose-peer.yml up -d
if [ $? -ne 0 ]; then
    echo "ERRO: Falha ao iniciar nó 2"
    exit 1
fi
echo "[OK] Nó 2 iniciado"
echo ""

# Aguarda o nó 2 estar pronto
echo "Aguardando nó 2 estar pronto..."
for i in {1..30}; do
    # CORREÇÃO 1: Usa IPC no segundo nó também
    if docker exec geth-peer geth attach --exec "eth.blockNumber" /root/.ethereum/geth.ipc >/dev/null 2>&1; then
        break
    fi
    if [ $i -eq 30 ]; then
        echo "[AVISO] Nó 2 pode não estar totalmente pronto ainda"
    fi
    sleep 2
done

# Verifica conexão entre os nós
echo ""
echo "Verificando conexão entre os nós..."
sleep 3
# CORREÇÃO 1: IPC para verificação
PEERS_NODE1=$(docker exec geth-node geth attach --exec "admin.peers.length" /root/.ethereum/geth.ipc 2>/dev/null | tr -d ' \n' || echo "0")
PEERS_NODE2=$(docker exec geth-peer geth attach --exec "admin.peers.length" /root/.ethereum/geth.ipc 2>/dev/null | tr -d ' \n' || echo "0")

echo ""
echo "========================================"
echo "Novo Nó Criado com Sucesso!"
echo "========================================"
echo ""
echo "Status da conexão:"
echo "- Nó 1 (geth-node) vê $PEERS_NODE1 peer(s)"
echo "- Nó 2 (geth-peer) vê $PEERS_NODE2 peer(s)"
echo ""
echo "Serviços disponíveis:"
echo "- Nó 1 RPC: http://localhost:8545"
echo "- Nó 2 RPC: http://localhost:8547"
echo "- Nó 1 WebSocket: ws://localhost:8546"
echo "- Nó 2 WebSocket: ws://localhost:8548"
echo ""
echo "Comandos úteis:"
echo "- Ver peers do nó 1: docker exec geth-node geth attach /root/.ethereum/geth.ipc --exec 'admin.peers'"
echo "- Ver peers do nó 2: docker exec geth-peer geth attach /root/.ethereum/geth.ipc --exec 'admin.peers'"
echo "- Parar nó 2: cd $BLOCKCHAIN_DIR && $DOCKER_COMPOSE -f docker-compose-peer.yml down"
echo ""
read -p "Pressione Enter para continuar..."