// ===================== BAREMA ITEM 1: ARQUITETURA =====================
// Utilitário para enviar transação e forçar criação de bloco no Clique

package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/core/types"
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

	// Conecta também via ethclient para operações normais
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

	// Desbloqueia a conta no keystore Go
	err = ks.Unlock(account, "123456")
	if err != nil {
		fmt.Printf("ERRO: Falha ao desbloquear conta no keystore: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Conta desbloqueada no keystore!")

	// Desbloqueia a conta no Geth via RPC (opcional)
	var unlockResult bool
	rpcClient.Call(&unlockResult, "personal_unlockAccount", account.Address, "123456", 0)
	if unlockResult {
		fmt.Println("Conta desbloqueada no Geth também!")
	}

	// Obtém nonce
	nonce, err := client.PendingNonceAt(context.Background(), account.Address)
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

	// Cria transação (envia para si mesmo)
	value := big.NewInt(1000000000000000) // 0.001 ETH
	gasLimit := uint64(21000)

	tx := types.NewTransaction(nonce, account.Address, value, gasLimit, gasPrice, nil)

	// Assina a transação localmente
	signedTx, err := ks.SignTx(account, tx, chainID)
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

