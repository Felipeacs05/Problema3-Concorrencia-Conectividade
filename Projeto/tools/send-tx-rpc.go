// ===================== BAREMA ITEM 1: ARQUITETURA =====================
// Utilitário para enviar transação usando personal_sendTransaction (RPC direto)

package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

func main() {
	// Conecta ao Geth via RPC
	rpcClient, err := rpc.Dial("http://127.0.0.1:8545")
	if err != nil {
		fmt.Printf("ERRO: Falha ao conectar RPC: %v\n", err)
		os.Exit(1)
	}
	defer rpcClient.Close()

	// Conecta também via ethclient para verificações
	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		fmt.Printf("ERRO: Falha ao conectar: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// Obtém a primeira conta do keystore
	keystorePath := filepath.Join("..", "data", "keystore")
	ks := keystore.NewKeyStore(keystorePath, keystore.StandardScryptN, keystore.StandardScryptP)

	if len(ks.Accounts()) == 0 {
		fmt.Println("ERRO: Nenhuma conta encontrada no keystore!")
		os.Exit(1)
	}

	account := ks.Accounts()[0]
	fmt.Printf("Conta: %s\n", account.Address.Hex())

	// Verifica saldo
	balance, _ := client.BalanceAt(context.Background(), account.Address, nil)
	balanceEth := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18))
	fmt.Printf("Saldo: %f ETH\n", balanceEth)

	// Desbloqueia a conta via RPC
	var unlockResult bool
	err = rpcClient.Call(&unlockResult, "personal_unlockAccount", account.Address, "123456", 0)
	if err != nil {
		fmt.Printf("ERRO: Falha ao desbloquear conta: %v\n", err)
		os.Exit(1)
	}
	if !unlockResult {
		fmt.Println("ERRO: Falha ao desbloquear conta (senha incorreta?)")
		os.Exit(1)
	}
	fmt.Println("Conta desbloqueada!")

	// Prepara transação (envia para si mesmo)
	destAddr := account.Address
	value := big.NewInt(1000000000000000) // 0.001 ETH

	// Usa personal_sendTransaction (mais direto)
	txParams := map[string]interface{}{
		"from":  account.Address.Hex(),
		"to":    destAddr.Hex(),
		"value": fmt.Sprintf("0x%x", value),
	}

	var txHash common.Hash
	err = rpcClient.Call(&txHash, "personal_sendTransaction", txParams, "123456")
	if err != nil {
		fmt.Printf("ERRO: Falha ao enviar transação: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("SUCCESS: Transação enviada: %s\n", txHash.Hex())
	fmt.Println("Aguardando confirmação...")

	// Aguarda confirmação
	startBlock, _ := client.BlockNumber(context.Background())
	waited := 0
	for waited < 30 {
		currentBlock, _ := client.BlockNumber(context.Background())
		if currentBlock > startBlock {
			fmt.Printf("SUCCESS: Bloco criado! Bloco atual: %d\n", currentBlock)
			os.Exit(0)
		}
		time.Sleep(1 * time.Second)
		waited++
	}

	currentBlock, _ := client.BlockNumber(context.Background())
	fmt.Printf("Bloco atual: %d\n", currentBlock)
	if currentBlock > 0 {
		fmt.Println("SUCCESS: Blocos estão sendo criados!")
	} else {
		fmt.Println("AVISO: Ainda no bloco 0.")
	}
}

