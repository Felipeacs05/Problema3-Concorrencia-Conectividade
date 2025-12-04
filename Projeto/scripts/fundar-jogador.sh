#!/bin/bash
# ===================== FUNDAR CONTA DE JOGADOR =====================
# Script para enviar ETH para uma conta de jogador

echo "========================================"
echo "Fundar Conta de Jogador"
echo "========================================"
echo ""

if [ -z "$1" ]; then
    echo "Uso: $0 <endereco_do_jogador> [quantidade_em_ETH]"
    echo ""
    echo "Exemplo:"
    echo "  $0 0x4b86c5ccccef4149732166a938f4400d7e65903e"
    echo "  $0 0x4b86c5ccccef4149732166a938f4400d7e65903e 50"
    echo ""
    echo "Se quantidade não for especificada, envia 100 ETH por padrão."
    echo ""
    echo "Para encontrar o endereço do jogador:"
    echo "  - Veja no cliente quando ele seleciona a carteira"
    echo "  - Ou liste as contas com: docker exec geth-node geth attach --exec 'eth.accounts' http://localhost:8545"
    exit 1
fi

DEST_ADDR="$1"
AMOUNT="${2:-100}"

echo "Enviando $AMOUNT ETH para: $DEST_ADDR"
echo ""

# Executa a transferência
docker exec geth-node geth attach --exec "
var signer = eth.accounts[0];
console.log('Conta signer (origem): ' + signer);
console.log('Conta destino: ' + '$DEST_ADDR');
console.log('');

// Verifica saldo do signer
var signerBalance = web3.fromWei(eth.getBalance(signer), 'ether');
console.log('Saldo do signer: ' + signerBalance + ' ETH');

// Verifica saldo atual do destino
var destBalanceBefore = web3.fromWei(eth.getBalance('$DEST_ADDR'), 'ether');
console.log('Saldo atual do destino: ' + destBalanceBefore + ' ETH');
console.log('');

// Desbloqueia conta
personal.unlockAccount(signer, '123456', 0);

// Envia transação
console.log('Enviando $AMOUNT ETH...');
var tx = eth.sendTransaction({
    from: signer, 
    to: '$DEST_ADDR', 
    value: web3.toWei($AMOUNT, 'ether')
});
console.log('Transacao enviada: ' + tx);
" http://localhost:8545

echo ""
echo "Aguardando confirmação..."
sleep 5

# Verifica saldo final
docker exec geth-node geth attach --exec "
var balance = web3.fromWei(eth.getBalance('$DEST_ADDR'), 'ether');
console.log('Saldo final de $DEST_ADDR: ' + balance + ' ETH');
" http://localhost:8545

echo ""
echo "✓ Concluído! O jogador agora pode comprar pacotes."

