#!/bin/bash
# ===================== CRIAR CONTA JOGADOR =====================
# Script para criar uma nova conta (carteira) para um jogador

set -e  # Para em caso de erro

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BLOCKCHAIN_DIR="$PROJECT_DIR/Blockchain"
KEYSTORE_DIR="$BLOCKCHAIN_DIR/data/keystore"
TOOLS_DIR="$BLOCKCHAIN_DIR/tools"

echo "========================================"
echo "Criar Nova Conta de Jogador"
echo "========================================"
echo ""

# Verifica se Go está instalado
if ! command -v go &> /dev/null; then
    echo "ERRO: Go não está instalado!"
    exit 1
fi

# Solicita senha
read -sp "Digite uma senha para a nova conta: " SENHA
echo ""
if [ -z "$SENHA" ]; then
    echo "ERRO: Senha não pode ser vazia"
    exit 1
fi

# Cria conta
echo "Criando nova conta..."
"$TOOLS_DIR/blockchain-utils" criar-conta "$KEYSTORE_DIR" "$SENHA"
if [ $? -ne 0 ]; then
    echo "ERRO: Falha ao criar conta"
    exit 1
fi

echo ""
echo "========================================"
echo "Conta criada com sucesso!"
echo "========================================"
echo ""
echo "IMPORTANTE:"
echo "- Guarde o arquivo do keystore em local seguro"
echo "- Anote a senha escolhida"
echo "- O arquivo está em: $KEYSTORE_DIR"
echo ""
echo "Para usar esta conta no jogo:"
echo "1. Copie o arquivo do keystore para o diretório do cliente"
echo "2. Use o caminho do arquivo ao fazer login"
echo ""
read -p "Pressione Enter para continuar..."

