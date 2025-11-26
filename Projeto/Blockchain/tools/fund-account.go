// ===================== BAREMA ITEM 1: ARQUITETURA =====================
// Utilitário para transferir ETH da conta do signer para outra conta

package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: fund-account.exe <endereco_destino> [quantidade_em_ETH]")
		fmt.Println("Exemplo: fund-account.exe 0x7041e32e2E3b380368e445885b0EdBBC33F234CC 100")
		os.Exit(1)
	}

	destAddr := common.HexToAddress(os.Args[1])
	amountEth := float64(100) // padrão: 100 ETH

	if len(os.Args) >= 3 {
		parsed, err := strconv.ParseFloat(os.Args[2], 64)
		if err == nil {
			amountEth = parsed
		}
	}

	// Conecta ao Geth via RPC
	rpcClient, err := rpc.Dial("http://127.0.0.1:8545")
	if err != nil {
		fmt.Printf("ERRO: Falha ao conectar RPC: %v\n", err)
		os.Exit(1)
	}
	defer rpcClient.Close()

	// Conecta também via ethclient
	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		fmt.Printf("ERRO: Falha ao conectar: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// Obtém a primeira conta do keystore (signer)
	keystorePath := filepath.Join("..", "data", "keystore")
	ks := keystore.NewKeyStore(keystorePath, keystore.StandardScryptN, keystore.StandardScryptP)

	if len(ks.Accounts()) == 0 {
		fmt.Println("ERRO: Nenhuma conta encontrada no keystore!")
		os.Exit(1)
	}

	signerAccount := ks.Accounts()[0]
	fmt.Printf("Conta signer: %s\n", signerAccount.Address.Hex())

	// Desbloqueia a conta no keystore Go
	err = ks.Unlock(signerAccount, "123456")
	if err != nil {
		fmt.Printf("ERRO: Falha ao desbloquear conta no keystore: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Conta desbloqueada no keystore!")

	// Desbloqueia a conta no Geth via RPC (opcional, mas ajuda)
	var unlockResult bool
	rpcClient.Call(&unlockResult, "personal_unlockAccount", signerAccount.Address, "123456", 0)
	if unlockResult {
		fmt.Println("Conta desbloqueada no Geth também!")
	}

	// Converte ETH para Wei
	amountWei := new(big.Float).Mul(big.NewFloat(amountEth), big.NewFloat(1e18))
	amountWeiInt, _ := amountWei.Int(nil)

	fmt.Printf("Transferindo %f ETH para: %s\n", amountEth, destAddr.Hex())

	// Obtém nonce
	nonce, err := client.PendingNonceAt(context.Background(), signerAccount.Address)
	if err != nil {
		fmt.Printf("ERRO: Falha ao obter nonce: %v\n", err)
		os.Exit(1)
	}

	// Obtém chain ID
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		fmt.Printf("ERRO: Falha ao obter chain ID: %v\n", err)
		os.Exit(1)
	}

	// Obtém gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		fmt.Printf("ERRO: Falha ao obter gas price: %v\n", err)
		os.Exit(1)
	}

	// Cria transação
	tx := types.NewTransaction(nonce, destAddr, amountWeiInt, 21000, gasPrice, nil)

	// Assina a transação localmente
	signedTx, err := ks.SignTx(signerAccount, tx, chainID)
	if err != nil {
		fmt.Printf("ERRO: Falha ao assinar transação: %v\n", err)
		os.Exit(1)
	}

	// Envia a transação
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		fmt.Printf("ERRO: Falha ao enviar transação: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("SUCCESS: Transação enviada: %s\n", signedTx.Hash().Hex())
	fmt.Println("Aguarde alguns segundos para confirmação...")

	// Aguarda confirmação
	for i := 0; i < 10; i++ {
		receipt, err := client.TransactionReceipt(context.Background(), signedTx.Hash())
		if err == nil && receipt != nil {
			fmt.Println("SUCCESS: Transação confirmada!")
			
			// Verifica saldo da conta destino
			balance, _ := client.BalanceAt(context.Background(), destAddr, nil)
			balanceEth := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18))
			fmt.Printf("Novo saldo da conta destino: %f ETH\n", balanceEth)
			os.Exit(0)
		}
		// Aguarda 1 segundo
		select {
		case <-context.Background().Done():
			return
		default:
		}
	}

	fmt.Println("AVISO: Transação ainda pendente. Verifique mais tarde.")
}

