#!/bin/bash
# ===================== VERIFICAR NÓS =====================
# Script para verificar o status e conexão entre os nós Geth

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "========================================"
echo "Verificando Status dos Nós"
echo "========================================"
echo ""

# Verifica se Docker está rodando
if ! docker ps >/dev/null 2>&1; then
    echo "❌ ERRO: Docker não está rodando!"
    echo "Inicie o Docker e tente novamente."
    exit 1
fi

# Verifica se os containers estão rodando
echo "[1/5] Verificando containers..."
NODE1_RUNNING=false
NODE2_RUNNING=false

if docker ps | grep -q "geth-node"; then
    NODE1_RUNNING=true
    echo "✅ Nó 1 (geth-node): RODANDO"
else
    echo "❌ Nó 1 (geth-node): NÃO ESTÁ RODANDO"
fi

if docker ps | grep -q "geth-peer"; then
    NODE2_RUNNING=true
    echo "✅ Nó 2 (geth-peer): RODANDO"
else
    echo "⚠️  Nó 2 (geth-peer): NÃO ESTÁ RODANDO (opcional)"
fi
echo ""

# Verifica se os nós estão respondendo
echo "[2/5] Verificando resposta dos nós..."

if [ "$NODE1_RUNNING" = true ]; then
    # CORREÇÃO: Usando IPC para evitar erro de admin not defined
    if docker exec geth-node geth attach --exec "eth.blockNumber" /root/.ethereum/geth.ipc >/dev/null 2>&1; then
        BLOCK_NODE1=$(docker exec geth-node geth attach --exec "eth.blockNumber" /root/.ethereum/geth.ipc 2>/dev/null | tr -d ' \n')
        echo "✅ Nó 1: RESPONDENDO (Bloco: $BLOCK_NODE1)"
    else
        echo "❌ Nó 1: NÃO RESPONDE"
        BLOCK_NODE1="N/A"
    fi
else
    echo "⚠️  Nó 1: Container não está rodando"
    BLOCK_NODE1="N/A"
fi

if [ "$NODE2_RUNNING" = true ]; then
    if docker exec geth-peer geth attach --exec "eth.blockNumber" /root/.ethereum/geth.ipc >/dev/null 2>&1; then
        BLOCK_NODE2=$(docker exec geth-peer geth attach --exec "eth.blockNumber" /root/.ethereum/geth.ipc 2>/dev/null | tr -d ' \n')
        echo "✅ Nó 2: RESPONDENDO (Bloco: $BLOCK_NODE2)"
    else
        echo "❌ Nó 2: NÃO RESPONDE"
        BLOCK_NODE2="N/A"
    fi
else
    echo "⚠️  Nó 2: Container não está rodando"
    BLOCK_NODE2="N/A"
fi
echo ""

# Verifica sincronização de blocos
echo "[3/5] Verificando sincronização de blocos..."
if [ "$BLOCK_NODE1" != "N/A" ] && [ "$BLOCK_NODE2" != "N/A" ]; then
    # Remove "0x" se existir e converte para decimal
    BLOCK1_DEC=$(printf "%d" $BLOCK_NODE1 2>/dev/null || echo "0")
    BLOCK2_DEC=$(printf "%d" $BLOCK_NODE2 2>/dev/null || echo "0")
    
    DIFF=$((BLOCK1_DEC - BLOCK2_DEC))
    if [ $DIFF -lt 0 ]; then
        DIFF=$((-DIFF))
    fi
    
    if [ $DIFF -le 2 ]; then
        echo "✅ Blocos sincronizados (Diferença: $DIFF)"
    else
        echo "⚠️  Blocos dessincronizados (Diferença: $DIFF)"
        echo "   Nó 1: Bloco $BLOCK1_DEC"
        echo "   Nó 2: Bloco $BLOCK2_DEC"
    fi
elif [ "$BLOCK_NODE1" != "N/A" ]; then
    echo "ℹ️  Apenas Nó 1 está ativo (Bloco: $BLOCK_NODE1)"
elif [ "$BLOCK_NODE2" != "N/A" ]; then
    echo "ℹ️  Apenas Nó 2 está ativo (Bloco: $BLOCK_NODE2)"
else
    echo "❌ Nenhum nó está respondendo"
fi
echo ""

# Verifica conexão P2P entre os nós
echo "[5/6] Verificando conexão P2P..."

# Aguarda um pouco para dar tempo da conexão se estabelecer
sleep 3

if [ "$NODE1_RUNNING" = true ]; then
    # Tenta múltiplas vezes (conexão pode estar sendo estabelecida)
    PEERS_NODE1="0"
    for i in {1..3}; do
        PEERS_TMP=$(docker exec geth-node geth attach --exec "admin.peers.length" /root/.ethereum/geth.ipc 2>/dev/null | tr -d ' \n' || echo "0")
        if [ ! -z "$PEERS_TMP" ] && [ "$PEERS_TMP" != "" ] && [ "$PEERS_TMP" != "null" ]; then
            PEERS_NODE1="$PEERS_TMP"
            if [ "$PEERS_NODE1" -gt 0 ]; then
                break
            fi
        fi
        sleep 2
    done
    
    if [ "$PEERS_NODE1" -gt 0 ]; then
        echo "✅ Nó 1: Conectado a $PEERS_NODE1 peer(s)"
        
        # Mostra detalhes dos peers via IPC
        PEER_INFO=$(docker exec geth-node geth attach --exec "admin.peers" /root/.ethereum/geth.ipc 2>/dev/null)
        if [ ! -z "$PEER_INFO" ] && [ "$PEER_INFO" != "[]" ] && [ "$PEER_INFO" != "null" ]; then
            echo "   Detalhes dos peers:"
            echo "$PEER_INFO" | grep -o '"remoteAddress":"[^"]*"' | sed 's/"remoteAddress":"/   - /' | sed 's/"$//' || true
        fi
    else
        echo "⚠️  Nó 1: Sem peers conectados"
    fi
else
    echo "⚠️  Nó 1: Container não está rodando"
fi

if [ "$NODE2_RUNNING" = true ]; then
    # Tenta múltiplas vezes
    PEERS_NODE2="0"
    for i in {1..3}; do
        PEERS_TMP=$(docker exec geth-peer geth attach --exec "admin.peers.length" /root/.ethereum/geth.ipc 2>/dev/null | tr -d ' \n' || echo "0")
        if [ ! -z "$PEERS_TMP" ] && [ "$PEERS_TMP" != "" ] && [ "$PEERS_TMP" != "null" ]; then
            PEERS_NODE2="$PEERS_TMP"
            if [ "$PEERS_NODE2" -gt 0 ]; then
                break
            fi
        fi
        sleep 2
    done
    
    if [ "$PEERS_NODE2" -gt 0 ]; then
        echo "✅ Nó 2: Conectado a $PEERS_NODE2 peer(s)"
        
        # Mostra detalhes dos peers via IPC
        PEER_INFO=$(docker exec geth-peer geth attach --exec "admin.peers" /root/.ethereum/geth.ipc 2>/dev/null)
        if [ ! -z "$PEER_INFO" ] && [ "$PEER_INFO" != "[]" ] && [ "$PEER_INFO" != "null" ]; then
            echo "   Detalhes dos peers:"
            echo "$PEER_INFO" | grep -o '"remoteAddress":"[^"]*"' | sed 's/"remoteAddress":"/   - /' | sed 's/"$//' || true
        fi
    else
        echo "⚠️  Nó 2: Sem peers conectados"
    fi
else
    echo "⚠️  Nó 2: Container não está rodando"
fi
echo ""

# Verifica se os nós se enxergam mutuamente
echo "[6/6] Verificando conectividade mútua..."
if [ "$NODE1_RUNNING" = true ] && [ "$NODE2_RUNNING" = true ]; then
    if [ "$PEERS_NODE1" -gt 0 ] && [ "$PEERS_NODE2" -gt 0 ]; then
        echo "✅ CONEXÃO ESTABELECIDA: Ambos os nós se enxergam!"
    elif [ "$PEERS_NODE1" -gt 0 ] || [ "$PEERS_NODE2" -gt 0 ]; then
        echo "⚠️  CONEXÃO PARCIAL: Apenas um nó vê o outro"
    else
        echo "❌ SEM CONEXÃO: Os nós não estão conectados"
        echo ""
        echo "💡 Dicas para resolver:"
        echo "   1. Verifique se ambos os nós estão na mesma rede (networkid=1337)"
        echo "   2. Verifique se o Nó 2 foi iniciado com --bootnodes apontando para o Nó 1"
        echo "   3. Execute: ./connect-peer.sh"
    fi
elif [ "$NODE1_RUNNING" = true ]; then
    echo "ℹ️  Apenas Nó 1 está ativo (modo standalone)"
else
    echo "❌ Nenhum nó está rodando"
fi
echo ""

# Resumo final
echo "========================================"
echo "Resumo"
echo "========================================"
echo ""
echo "Status dos Containers:"
echo "  Nó 1 (geth-node): $([ "$NODE1_RUNNING" = true ] && echo "✅ RODANDO" || echo "❌ PARADO")"
echo "  Nó 2 (geth-peer): $([ "$NODE2_RUNNING" = true ] && echo "✅ RODANDO" || echo "⚠️  PARADO (opcional)")"
echo ""
echo "Endpoints RPC:"
echo "  Nó 1: http://localhost:8545"
echo "  Nó 2: http://localhost:8547"
echo ""
echo "Comandos úteis:"
echo "  Ver peers do Nó 1: docker exec geth-node geth attach /root/.ethereum/geth.ipc --exec 'admin.peers'"
echo "  Ver peers do Nó 2: docker exec geth-peer geth attach /root/.ethereum/geth.ipc --exec 'admin.peers'"
echo "  Ver blocos: docker exec geth-node geth attach /root/.ethereum/geth.ipc --exec 'eth.blockNumber'"
echo "  Ver enode do Nó 1: docker exec geth-node geth attach /root/.ethereum/geth.ipc --exec 'admin.nodeInfo.enode'"
echo ""
read -p "Pressione Enter para continuar..."