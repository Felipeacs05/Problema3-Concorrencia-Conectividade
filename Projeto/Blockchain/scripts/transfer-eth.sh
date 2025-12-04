#!/bin/bash
# ===================== BAREMA ITEM 1: ARQUITETURA =====================
# Script para transferir ETH da conta do signer para outra conta

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "========================================"
echo "Transferir ETH entre contas"
echo "========================================"
echo ""

if [ -z "$1" ]; then
    echo "Uso: $0 <endereco_destino> [quantidade_em_ETH]"
    echo ""
    echo "Exemplo:"
    echo "  $0 0x7041e32e2E3b380368e445885b0EdBBC33F234CC 10"
    echo "  (transfere 10 ETH para a conta especificada)"
    echo ""
    echo "Se quantidade não for especificada, transfere 100 ETH por padrão."
    exit 1
fi

DEST_ADDR="$1"
AMOUNT="${2:-100}"

echo "Transferindo $AMOUNT ETH para: $DEST_ADDR"
echo ""

# Executa a transferência via geth attach
docker exec geth-node geth attach --exec "
var accounts = eth.accounts;
if (accounts.length == 0) {
    console.log('ERRO: Nenhuma conta encontrada!');
} else {
    var signer = accounts[0];
    console.log('Conta signer: ' + signer);
    
    var dest = '$DEST_ADDR';
    var amount = web3.toWei($AMOUNT, 'ether');
    
    var unlocked = personal.unlockAccount(signer, '123456', 0);
    if (!unlocked) {
        console.log('ERRO: Falha ao desbloquear conta signer!');
    } else {
        console.log('Enviando $AMOUNT ETH...');
        var tx = eth.sendTransaction({from: signer, to: dest, value: amount});
        console.log('Transacao: ' + tx);
        console.log('SUCCESS: Transferencia enviada!');
        console.log('Aguarde alguns segundos para confirmacao.');
    }
}
" http://localhost:8545

echo ""
echo "Verificando saldo da conta destino..."
sleep 3

docker exec geth-node geth attach --exec "
var balance = eth.getBalance('$DEST_ADDR');
console.log('Saldo de $DEST_ADDR: ' + web3.fromWei(balance, 'ether') + ' ETH');
" http://localhost:8545

echo ""
echo "Transferência concluída!"


