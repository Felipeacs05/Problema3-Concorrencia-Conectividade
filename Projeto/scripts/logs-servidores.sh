#!/bin/bash
# ===================== LOGS SERVIDORES =====================
# Script para visualizar logs dos servidores de jogo (Linux/Mac)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
JOGO_DIR="$PROJECT_DIR/Jogo"

echo "========================================"
echo "Logs dos Servidores de Jogo"
echo "========================================"
echo ""
echo "Escolha qual servidor ver os logs:"
echo "1. Servidor 1"
echo "2. Servidor 2"
echo "3. Servidor 3"
echo "4. Todos os servidores (em paralelo)"
echo "5. Broker 1"
echo "6. Broker 2"
echo "7. Broker 3"
echo ""
read -p "Opção: " OPCAO

cd "$JOGO_DIR"

case "$OPCAO" in
    1)
        echo ""
        echo "=== Logs do Servidor 1 ==="
        docker logs -f servidor1
        ;;
    2)
        echo ""
        echo "=== Logs do Servidor 2 ==="
        docker logs -f servidor2
        ;;
    3)
        echo ""
        echo "=== Logs do Servidor 3 ==="
        docker logs -f servidor3
        ;;
    4)
        echo ""
        echo "=== Logs de Todos os Servidores ==="
        echo "Pressione Ctrl+C para sair"
        docker logs -f servidor1 &
        docker logs -f servidor2 &
        docker logs -f servidor3 &
        wait
        ;;
    5)
        echo ""
        echo "=== Logs do Broker 1 ==="
        docker logs -f broker1
        ;;
    6)
        echo ""
        echo "=== Logs do Broker 2 ==="
        docker logs -f broker2
        ;;
    7)
        echo ""
        echo "=== Logs do Broker 3 ==="
        docker logs -f broker3
        ;;
    *)
        echo "Opção inválida!"
        exit 1
        ;;
esac



