#!/bin/bash
# ===================== CONECTAR PEER =====================
# Script para conectar manualmente o Nó 2 ao Nó 1

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "========================================"
echo "Conectando Nó 2 ao Nó 1"
echo "========================================"
echo ""

# Verifica se os containers estão rodando
if ! docker ps | grep -q "geth-node"; then
    echo "❌ ERRO: Nó 1 (geth-node) não está rodando!"
    exit 1
fi

if ! docker ps | grep -q "geth-peer"; then
    echo "❌ ERRO: Nó 2 (geth-peer) não está rodando!"
    exit 1
fi

# Obtém o enode do Nó 1
echo "[1/3] Obtendo enode do Nó 1..."
ENODE=$(docker exec geth-node geth attach --exec "admin.nodeInfo.enode" /root/.ethereum/geth.ipc 2>/dev/null | tr -d '"' | tr -d ' ')

if [ -z "$ENODE" ]; then
    echo "❌ ERRO: Falha ao obter enode do Nó 1"
    exit 1
fi

# Como o Nó 2 está em network_mode: host, ele precisa se conectar ao IP do host
# O Nó 1 mapeia a porta 30303 para o host, então usamos o IP do host
HOST_IP=$(hostname -I | awk '{print $1}' 2>/dev/null || echo "127.0.0.1")

# Verifica se o Nó 2 está em host mode
NODE2_HOST_MODE=$(docker inspect -f '{{.HostConfig.NetworkMode}}' geth-peer 2>/dev/null || echo "")

if [ "$NODE2_HOST_MODE" = "host" ]; then
    # Nó 2 em host mode: usa IP do host
    ENODE=$(echo "$ENODE" | sed "s/@[^:]*:/@${HOST_IP}:/")
    echo "   Nó 2 em host mode, usando IP do host: $HOST_IP"
else
    # Nó 2 em bridge: tenta IP do container do Nó 1
    CONTAINER_IP=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' geth-node 2>/dev/null || echo "")
    if [ ! -z "$CONTAINER_IP" ] && [ "$CONTAINER_IP" != "" ]; then
        ENODE=$(echo "$ENODE" | sed "s/@[^:]*:/@${CONTAINER_IP}:/")
        echo "   Usando IP do container: $CONTAINER_IP"
    else
        ENODE=$(echo "$ENODE" | sed "s/@[^:]*:/@${HOST_IP}:/")
        echo "   Usando IP do host: $HOST_IP"
    fi
fi

echo "[OK] Enode: $ENODE"
echo ""

# Conecta o Nó 2 ao Nó 1
echo "[2/3] Conectando Nó 2 ao Nó 1..."
RESULT=$(docker exec geth-peer geth attach --exec "admin.addPeer('$ENODE')" /root/.ethereum/geth.ipc 2>/dev/null | tr -d ' \n')

if [ "$RESULT" = "true" ]; then
    echo "✅ Conexão estabelecida!"
else
    echo "⚠️  Resultado: $RESULT"
    echo "   Tentando método alternativo (localhost)..."
    
    # Método alternativo: usar localhost (se ambos estiverem no mesmo host)
    ENODE_LOCAL=$(echo "$ENODE" | sed "s/@[^:]*:/@127.0.0.1:/")
    
    RESULT2=$(docker exec geth-peer geth attach --exec "admin.addPeer('$ENODE_LOCAL')" /root/.ethereum/geth.ipc 2>/dev/null | tr -d ' \n')
    
    if [ "$RESULT2" = "true" ]; then
        echo "✅ Conexão estabelecida (via localhost)!"
    else
        echo "❌ Falha ao conectar."
        echo ""
        echo "💡 Diagnóstico:"
        echo "   - Verifique se a porta 30303 está acessível"
        echo "   - Verifique os logs: docker logs geth-peer"
        echo "   - Tente verificar manualmente:"
        echo "     docker exec geth-peer geth attach /root/.ethereum/geth.ipc --exec \"admin.peers\""
        exit 1
    fi
fi
echo ""

# Aguarda alguns segundos para a conexão se estabelecer
echo "[3/3] Aguardando sincronização..."
sleep 8

# Verifica se a conexão foi estabelecida (tenta algumas vezes)
PEERS=0
for i in {1..5}; do
    PEERS=$(docker exec geth-peer geth attach --exec "admin.peers.length" /root/.ethereum/geth.ipc 2>/dev/null | tr -d ' \n' || echo "0")
    if [ "$PEERS" -gt 0 ]; then
        break
    fi
    sleep 2
done

if [ "$PEERS" -gt 0 ]; then
    echo "✅ SUCESSO! Nó 2 está conectado a $PEERS peer(s)"
    echo ""
    
    # Mostra detalhes do peer
    PEER_INFO=$(docker exec geth-peer geth attach --exec "admin.peers" /root/.ethereum/geth.ipc 2>/dev/null)
    if [ ! -z "$PEER_INFO" ] && [ "$PEER_INFO" != "[]" ]; then
        echo "   Detalhes da conexão:"
        echo "$PEER_INFO" | grep -o '"remoteAddress":"[^"]*"' | sed 's/"remoteAddress":"/   - /' | sed 's/"$//' | head -1 || true
    fi
    
    echo ""
    echo "Execute './verificar-nos.sh' para ver o status completo"
else
    echo "⚠️  Conexão ainda não detectada após 18 segundos"
    echo ""
    echo "💡 Possíveis causas:"
    echo "   1. Firewall bloqueando porta 30303"
    echo "   2. Nós em redes Docker diferentes"
    echo "   3. Genesis blocks diferentes"
    echo ""
    echo "   Tente verificar manualmente:"
    echo "   docker exec geth-peer geth attach /root/.ethereum/geth.ipc --exec 'admin.peers'"
    echo "   docker exec geth-node geth attach /root/.ethereum/geth.ipc --exec 'admin.peers'"
fi
echo ""
