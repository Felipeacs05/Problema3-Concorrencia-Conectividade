#!/bin/bash
# ===================== CONECTAR PEER (Manual) =====================
# Script para forçar a conexão entre o Nó 2 e o Nó 1 via IPC
# Resolve problemas de IP (localhost vs IP da rede)

set -e

echo "========================================"
echo "Conectando Nó 2 ao Nó 1..."
echo "========================================"
echo ""

# 1. Obtém o Enode do Nó 1 via IPC
echo "[1/4] Obtendo Enode do Nó 1..."
ENODE_RAW=$(docker exec geth-node geth attach --exec "admin.nodeInfo.enode" /root/.ethereum/geth.ipc | tr -d '"')

if [ -z "$ENODE_RAW" ]; then
    echo "ERRO: Não foi possível obter o enode do Nó 1."
    exit 1
fi
echo "   Enode Original: $ENODE_RAW"

# 2. Descobre o IP real da máquina (Host)
# O container precisa acessar o host pela porta 30303
HOST_IP=$(hostname -I | awk '{print $1}')

if [ -z "$HOST_IP" ]; then
    HOST_IP="127.0.0.1"
    echo "   AVISO: Não foi possível detectar IP da rede, usando 127.0.0.1"
fi
echo "   IP do Host: $HOST_IP"

# 3. Substitui o IP no Enode
# Troca [::] ou 127.0.0.1 pelo IP real da rede local
ENODE_FIXED=$(echo "$ENODE_RAW" | sed "s/@\[::\]:/@${HOST_IP}:/g" | sed "s/@127.0.0.1:/@${HOST_IP}:/g")
echo "   Enode Corrigido: $ENODE_FIXED"
echo ""

# 4. Adiciona o Peer no Nó 2 via IPC
echo "[2/4] Adicionando Peer no Nó 2..."
RESULT=$(docker exec geth-peer geth attach --exec "admin.addPeer('$ENODE_FIXED')" /root/.ethereum/geth.ipc)

if [ "$RESULT" == "true" ]; then
    echo "   Comando enviado com sucesso (admin.addPeer retornou true)."
else
    echo "   AVISO: admin.addPeer retornou: $RESULT"
fi
echo ""

# 5. Adiciona o Peer no Nó 1 também (Bidirecional ajuda)
echo "[3/4] Adicionando Nó 2 no Nó 1 (Caminho reverso)..."
# Pega enode do nó 2
ENODE2_RAW=$(docker exec geth-peer geth attach --exec "admin.nodeInfo.enode" /root/.ethereum/geth.ipc | tr -d '"')
# Ajusta para porta 30304 (porta mapeada do peer) e IP do host
ENODE2_FIXED=$(echo "$ENODE2_RAW" | sed "s/@\[::\]:/@${HOST_IP}:/g" | sed "s/@127.0.0.1:/@${HOST_IP}:/g" | sed "s/:30303/:30304/g")

# O Nó 2 roda internamente na 30303, mas no docker-compose-peer.yml mapeamos 30304:30303.
# O Enode interno do Nó 2 vai mostrar 30303. Precisamos garantir que o Nó 1 conecte na 30304 do Host.
# Mas o geth reporta a porta de escuta interna (30303). 
# TRUQUE: O Nó 1 precisa bater no IP_HOST:30304 para chegar no container 2.
ENODE2_FINAL=$(echo "$ENODE2_FIXED" | sed "s/:30303?/:30304?/g") # Tenta corrigir porta se estiver errada

echo "   Enode Nó 2 (Ajustado): $ENODE2_FINAL"
docker exec geth-node geth attach --exec "admin.addPeer('$ENODE2_FINAL')" /root/.ethereum/geth.ipc > /dev/null
echo "   Conexão reversa solicitada."
echo ""

# 6. Verificação
echo "[4/4] Verificando conexão..."
sleep 3
PEERS=$(docker exec geth-peer geth attach --exec "admin.peers.length" /root/.ethereum/geth.ipc)
echo "   Nó 2 agora tem $PEERS peer(s)."

if [ "$PEERS" -gt 0 ]; then
    echo "✅ SUCESSO: Os nós estão conectados!"
else
    echo "❌ FALHA: Ainda sem peers. Verifique firewall ou configurações de rede."
fi