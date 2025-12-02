package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// Versão alternativa que usa o Geth diretamente via RPC
// A conta já está desbloqueada no Geth, então não precisa de senha

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: fund-account-rpc.exe <endereco_destino> [quantidade_em_ETH]")
		fmt.Println("Exemplo: fund-account-rpc.exe 0x7041e32e2E3b380368e445885b0EdBBC33F234CC 100")
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

	// Obtém a primeira conta do Geth (que está desbloqueada)
	var accounts []common.Address
	err = rpcClient.Call(&accounts, "eth_accounts")
	if err != nil || len(accounts) == 0 {
		fmt.Println("ERRO: Nenhuma conta desbloqueada encontrada no Geth!")
		fmt.Println("Certifique-se de que o Geth está rodando e a conta está desbloqueada")
		os.Exit(1)
	}

	signerAccount := accounts[0]
	fmt.Printf("Conta signer (do Geth): %s\n", signerAccount.Hex())

	// Converte ETH para Wei
	amountWei := new(big.Float).Mul(big.NewFloat(amountEth), big.NewFloat(1e18))
	amountWeiInt, _ := amountWei.Int(nil)

	fmt.Printf("Transferindo %f ETH para: %s\n", amountEth, destAddr.Hex())

	// Obtém gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		fmt.Printf("ERRO: Falha ao obter gas price: %v\n", err)
		os.Exit(1)
	}

	// Envia transação diretamente via personal_sendTransaction (conta já está desbloqueada)
	// Lê a senha do arquivo password.txt ou usa padrão
	password := "123456" // Senha padrão
	if passwordBytes, err := os.ReadFile("../data/password.txt"); err == nil {
		password = strings.TrimSpace(string(passwordBytes))
	} else if passwordBytes, err := os.ReadFile("../../data/password.txt"); err == nil {
		password = strings.TrimSpace(string(passwordBytes))
	}
	
	var txHash common.Hash
	err = rpcClient.Call(&txHash, "personal_sendTransaction", map[string]interface{}{
		"from":     signerAccount.Hex(),
		"to":       destAddr.Hex(),
		"value":    fmt.Sprintf("0x%x", amountWeiInt),
		"gas":      "0x5208", // 21000
		"gasPrice": fmt.Sprintf("0x%x", gasPrice),
	}, password)

	if err != nil {
		fmt.Printf("ERRO: Falha ao enviar transação: %v\n", err)
		fmt.Println("\nDica: Certifique-se de que a conta está desbloqueada no Geth")
		fmt.Printf("Execute: docker exec -it geth-node geth attach --exec \"personal.unlockAccount(eth.accounts[0], '', 0)\"\n")
		os.Exit(1)
	}

	fmt.Printf("SUCCESS: Transação enviada: %s\n", txHash.Hex())
	fmt.Println("Aguarde alguns segundos para confirmação...")

	// Aguarda confirmação
	for i := 0; i < 30; i++ {
		receipt, err := client.TransactionReceipt(context.Background(), txHash)
		if err == nil && receipt != nil {
			fmt.Println("SUCCESS: Transação confirmada!")
			
			// Verifica saldo da conta destino
			balance, _ := client.BalanceAt(context.Background(), destAddr, nil)
			balanceEth := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18))
			fmt.Printf("Novo saldo da conta destino: %f ETH\n", balanceEth)
			os.Exit(0)
		}
		time.Sleep(1 * time.Second)
	}

	fmt.Println("AVISO: Transação ainda pendente. Verifique mais tarde.")
}

