#!/bin/bash
# ===================== DIAGNÓSTICO DE REDE =====================
# Script para diagnosticar problemas de conexão entre nós

echo "========================================"
echo "Diagnóstico de Rede dos Nós"
echo "========================================"
echo ""

# 1. Verifica containers
echo "[1/6] Verificando containers..."
if docker ps | grep -q "geth-node"; then
    echo "✅ geth-node está rodando"
    NODE1_IP=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' geth-node 2>/dev/null || echo "N/A")
    NODE1_NETWORK=$(docker inspect -f '{{.HostConfig.NetworkMode}}' geth-node 2>/dev/null || echo "N/A")
    echo "   IP: $NODE1_IP"
    echo "   Network Mode: $NODE1_NETWORK"
else
    echo "❌ geth-node NÃO está rodando"
    NODE1_IP="N/A"
    NODE1_NETWORK="N/A"
fi

if docker ps | grep -q "geth-peer"; then
    echo "✅ geth-peer está rodando"
    NODE2_IP=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' geth-peer 2>/dev/null || echo "N/A")
    NODE2_NETWORK=$(docker inspect -f '{{.HostConfig.NetworkMode}}' geth-peer 2>/dev/null || echo "N/A")
    echo "   IP: $NODE2_IP"
    echo "   Network Mode: $NODE2_NETWORK"
else
    echo "❌ geth-peer NÃO está rodando"
    NODE2_IP="N/A"
    NODE2_NETWORK="N/A"
fi
echo ""

# 2. Verifica enodes
echo "[2/6] Verificando enodes..."
if docker ps | grep -q "geth-node"; then
    ENODE1=$(docker exec geth-node geth attach --exec "admin.nodeInfo.enode" /root/.ethereum/geth.ipc 2>/dev/null | tr -d '"' | tr -d ' ')
    if [ ! -z "$ENODE1" ]; then
        echo "✅ Enode do Nó 1 obtido"
        echo "   $ENODE1"
    else
        echo "❌ Falha ao obter enode do Nó 1"
    fi
fi

if docker ps | grep -q "geth-peer"; then
    ENODE2=$(docker exec geth-peer geth attach --exec "admin.nodeInfo.enode" /root/.ethereum/geth.ipc 2>/dev/null | tr -d '"' | tr -d ' ')
    if [ ! -z "$ENODE2" ]; then
        echo "✅ Enode do Nó 2 obtido"
        echo "   $ENODE2"
    else
        echo "❌ Falha ao obter enode do Nó 2"
    fi
fi
echo ""

# 3. Verifica networkid
echo "[3/6] Verificando networkid..."
if docker ps | grep -q "geth-node"; then
    NETID1=$(docker exec geth-node geth attach --exec "net.version" /root/.ethereum/geth.ipc 2>/dev/null | tr -d '"' | tr -d ' ')
    echo "   Nó 1 networkid: $NETID1"
fi

if docker ps | grep -q "geth-peer"; then
    NETID2=$(docker exec geth-peer geth attach --exec "net.version" /root/.ethereum/geth.ipc 2>/dev/null | tr -d '"' | tr -d ' ')
    echo "   Nó 2 networkid: $NETID2"
fi
echo ""

# 4. Verifica peers
echo "[4/6] Verificando peers conectados..."
if docker ps | grep -q "geth-node"; then
    PEERS1=$(docker exec geth-node geth attach --exec "admin.peers.length" /root/.ethereum/geth.ipc 2>/dev/null | tr -d ' \n' || echo "0")
    echo "   Nó 1 tem $PEERS1 peer(s)"
    if [ "$PEERS1" -gt 0 ]; then
        PEER_INFO1=$(docker exec geth-node geth attach --exec "admin.peers" /root/.ethereum/geth.ipc 2>/dev/null)
        echo "$PEER_INFO1" | grep -o '"remoteAddress":"[^"]*"' | sed 's/"remoteAddress":"/     - /' | sed 's/"$//' || true
    fi
fi

if docker ps | grep -q "geth-peer"; then
    PEERS2=$(docker exec geth-peer geth attach --exec "admin.peers.length" /root/.ethereum/geth.ipc 2>/dev/null | tr -d ' \n' || echo "0")
    echo "   Nó 2 tem $PEERS2 peer(s)"
    if [ "$PEERS2" -gt 0 ]; then
        PEER_INFO2=$(docker exec geth-peer geth attach --exec "admin.peers" /root/.ethereum/geth.ipc 2>/dev/null)
        echo "$PEER_INFO2" | grep -o '"remoteAddress":"[^"]*"' | sed 's/"remoteAddress":"/     - /' | sed 's/"$//' || true
    fi
fi
echo ""

# 5. Verifica portas
echo "[5/6] Verificando portas..."
HOST_IP=$(hostname -I | awk '{print $1}' 2>/dev/null || echo "127.0.0.1")
echo "   IP do Host: $HOST_IP"
echo "   Porta 30303 (Nó 1): $(netstat -tuln 2>/dev/null | grep ':30303' || echo 'Não detectada')"
echo "   Porta 30304 (Nó 2): $(netstat -tuln 2>/dev/null | grep ':30304' || echo 'Não detectada')"
echo ""

# 6. Recomendações
echo "[6/6] Recomendações..."
echo ""

if [ "$NODE1_NETWORK" != "host" ] && [ "$NODE2_NETWORK" = "host" ]; then
    echo "⚠️  PROBLEMA DETECTADO:"
    echo "   - Nó 1 está em rede bridge ($NODE1_IP)"
    echo "   - Nó 2 está em host mode"
    echo "   - Eles não podem se comunicar diretamente"
    echo ""
    echo "💡 SOLUÇÃO:"
    echo "   1. Reinicie o Nó 1 com network_mode: host, OU"
    echo "   2. Use o IP do host ($HOST_IP) na porta 30303 para conectar"
    echo "   3. Execute: ./connect-peer.sh"
elif [ "$PEERS1" = "0" ] && [ "$PEERS2" = "0" ]; then
    echo "⚠️  Nenhum peer conectado"
    echo ""
    echo "💡 SOLUÇÃO:"
    echo "   1. Execute: ./connect-peer.sh"
    echo "   2. Aguarde 10-15 segundos"
    echo "   3. Execute: ./verificar-nos.sh novamente"
else
    echo "✅ Configuração parece correta"
    echo "   Se ainda não conecta, verifique firewall"
fi
echo ""
