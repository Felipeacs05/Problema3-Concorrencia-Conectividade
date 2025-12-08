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

# Ajusta o IP para o IP do container do Nó 1 (dentro da rede Docker)
# Primeiro, tenta descobrir o IP do container
CONTAINER_IP=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' geth-node 2>/dev/null || echo "")

if [ ! -z "$CONTAINER_IP" ] && [ "$CONTAINER_IP" != "" ]; then
    # Substitui o IP no enode pelo IP do container
    ENODE=$(echo "$ENODE" | sed "s/@[^:]*:/@${CONTAINER_IP}:/")
    echo "   Usando IP do container: $CONTAINER_IP"
else
    # Se não conseguir, tenta usar o IP do host na porta mapeada
    HOST_IP=$(hostname -I | awk '{print $1}' 2>/dev/null || echo "127.0.0.1")
    ENODE=$(echo "$ENODE" | sed "s/@[^:]*:/@${HOST_IP}:/")
    echo "   Usando IP do host: $HOST_IP"
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
    echo "   Tentando método alternativo..."
    
    # Método alternativo: usar o IP do host na porta mapeada
    HOST_IP=$(hostname -I | awk '{print $1}' 2>/dev/null || echo "127.0.0.1")
    ENODE_HOST=$(echo "$ENODE" | sed "s/@[^:]*:/@${HOST_IP}:/" | sed "s/:30303/:30303/")
    
    RESULT2=$(docker exec geth-peer geth attach --exec "admin.addPeer('$ENODE_HOST')" /root/.ethereum/geth.ipc 2>/dev/null | tr -d ' \n')
    
    if [ "$RESULT2" = "true" ]; then
        echo "✅ Conexão estabelecida (método alternativo)!"
    else
        echo "❌ Falha ao conectar. Verifique os logs:"
        echo "   docker logs geth-peer"
        exit 1
    fi
fi
echo ""

# Aguarda alguns segundos para a conexão se estabelecer
echo "[3/3] Aguardando sincronização..."
sleep 5

# Verifica se a conexão foi estabelecida
PEERS=$(docker exec geth-peer geth attach --exec "admin.peers.length" /root/.ethereum/geth.ipc 2>/dev/null | tr -d ' \n' || echo "0")

if [ "$PEERS" -gt 0 ]; then
    echo "✅ SUCESSO! Nó 2 está conectado a $PEERS peer(s)"
    echo ""
    echo "Execute './verificar-nos.sh' para ver o status completo"
else
    echo "⚠️  Conexão pode não ter sido estabelecida ainda"
    echo "   Aguarde alguns segundos e execute './verificar-nos.sh' novamente"
fi
echo ""
