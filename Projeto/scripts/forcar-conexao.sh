#!/bin/bash
# ===================== FORÇAR CONEXÃO =====================
# Script para forçar e manter a conexão entre os nós

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "========================================"
echo "Forçando Conexão entre Nós"
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

# 1. Verifica genesis hash
echo "[1/4] Verificando genesis hash..."
GENESIS1=$(docker exec geth-node geth attach --exec "eth.getBlock(0).hash" /root/.ethereum/geth.ipc 2>/dev/null | tr -d '"' | tr -d ' ' || echo "")
GENESIS2=$(docker exec geth-peer geth attach --exec "eth.getBlock(0).hash" /root/.ethereum/geth.ipc 2>/dev/null | tr -d '"' | tr -d ' ' || echo "")

if [ -z "$GENESIS1" ] || [ -z "$GENESIS2" ]; then
    echo "⚠️  Não foi possível verificar genesis hash"
else
    if [ "$GENESIS1" = "$GENESIS2" ]; then
        echo "✅ Genesis hash idêntico: ${GENESIS1:0:16}..."
    else
        echo "❌ ERRO CRÍTICO: Genesis hash DIFERENTE!"
        echo "   Nó 1: ${GENESIS1:0:20}..."
        echo "   Nó 2: ${GENESIS2:0:20}..."
        echo ""
        echo "💡 SOLUÇÃO:"
        echo "   Os nós precisam usar o MESMO genesis.json"
        echo "   Execute: bash novo-no.sh novamente (ele usa o mesmo genesis do Nó 1)"
        exit 1
    fi
fi
echo ""

# 2. Obtém enode do Nó 1
echo "[2/4] Obtendo enode do Nó 1..."
ENODE=$(docker exec geth-node geth attach --exec "admin.nodeInfo.enode" /root/.ethereum/geth.ipc 2>/dev/null | tr -d '"' | tr -d ' ')

if [ -z "$ENODE" ]; then
    echo "❌ ERRO: Falha ao obter enode do Nó 1"
    exit 1
fi

# Ajusta IP para o host (Nó 2 está em host mode)
HOST_IP=$(hostname -I | awk '{print $1}' 2>/dev/null || echo "127.0.0.1")
ENODE=$(echo "$ENODE" | sed "s/@[^:]*:/@${HOST_IP}:/")

echo "[OK] Enode: ${ENODE:0:60}...@${HOST_IP}:30303"
echo ""

# 3. Remove peers antigos e adiciona novo
echo "[3/4] Removendo peers antigos..."
docker exec geth-peer geth attach --exec "admin.removePeer('$ENODE')" /root/.ethereum/geth.ipc >/dev/null 2>&1 || true
sleep 1

echo "   Adicionando peer..."
RESULT=$(docker exec geth-peer geth attach --exec "admin.addPeer('$ENODE')" /root/.ethereum/geth.ipc 2>/dev/null | tr -d ' \n' || echo "false")

if [ "$RESULT" = "true" ]; then
    echo "✅ Peer adicionado com sucesso"
else
    echo "⚠️  Resultado: $RESULT"
    echo "   Tentando via localhost..."
    ENODE_LOCAL=$(echo "$ENODE" | sed "s/@[^:]*:/@127.0.0.1:/")
    RESULT2=$(docker exec geth-peer geth attach --exec "admin.addPeer('$ENODE_LOCAL')" /root/.ethereum/geth.ipc 2>/dev/null | tr -d ' \n' || echo "false")
    
    if [ "$RESULT2" = "true" ]; then
        echo "✅ Peer adicionado via localhost"
    else
        echo "❌ Falha ao adicionar peer"
        exit 1
    fi
fi
echo ""

# 4. Aguarda e verifica
echo "[4/4] Aguardando conexão se estabelecer..."
for i in {1..10}; do
    PEERS=$(docker exec geth-peer geth attach --exec "admin.peers.length" /root/.ethereum/geth.ipc 2>/dev/null | tr -d ' \n' || echo "0")
    
    if [ "$PEERS" -gt 0 ]; then
        echo "✅ CONEXÃO ESTABELECIDA! ($PEERS peer(s))"
        
        # Mostra detalhes
        PEER_INFO=$(docker exec geth-peer geth attach --exec "admin.peers" /root/.ethereum/geth.ipc 2>/dev/null)
        if [ ! -z "$PEER_INFO" ] && [ "$PEER_INFO" != "[]" ] && [ "$PEER_INFO" != "null" ]; then
            echo ""
            echo "Detalhes da conexão:"
            echo "$PEER_INFO" | python3 -m json.tool 2>/dev/null || echo "$PEER_INFO"
        fi
        
        echo ""
        echo "Execute './verificar-nos.sh' para ver o status completo"
        exit 0
    fi
    
    echo "   Aguardando... ($i/10)"
    sleep 2
done

echo "⚠️  Conexão não detectada após 20 segundos"
echo ""
echo "💡 Diagnóstico:"
echo "   Execute: bash diagnostico-rede.sh"
exit 1
