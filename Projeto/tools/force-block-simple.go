// ===================== BAREMA ITEM 1: ARQUITETURA =====================
// Utilitário simples para forçar criação de bloco enviando transação

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
)

func main() {
	// Conecta ao Geth (tenta localhost primeiro, depois 127.0.0.1)
	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		// Tenta localhost
		client, err = ethclient.Dial("http://localhost:8545")
		if err != nil {
			fmt.Printf("ERRO: Falha ao conectar ao Geth: %v\n", err)
			fmt.Println("Verifique se o Geth está rodando: docker-compose ps")
			os.Exit(1)
		}
	}
	defer client.Close()

	fmt.Println("✓ Conectado ao Geth")

	// Obtém a primeira conta do keystore
	keystorePath := filepath.Join("..", "data", "keystore")
	ks := keystore.NewKeyStore(keystorePath, keystore.StandardScryptN, keystore.StandardScryptP)
	
	accounts := ks.Accounts()
	if len(accounts) == 0 {
		fmt.Println("ERRO: Nenhuma conta encontrada no keystore!")
		fmt.Printf("Keystore path: %s\n", keystorePath)
		os.Exit(1)
	}

	account := accounts[0]
	fmt.Printf("✓ Conta encontrada: %s\n", account.Address.Hex())

	// Desbloqueia a conta
	err = ks.Unlock(account, "123456")
	if err != nil {
		fmt.Printf("ERRO: Falha ao desbloquear conta: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Conta desbloqueada")

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

	fmt.Printf("✓ Nonce: %d, Chain ID: %s\n", nonce, chainID.String())

	// Cria transação (envia para si mesmo)
	value := big.NewInt(1000000000000000) // 0.001 ETH
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(1000000000) // 1 Gwei

	tx := types.NewTransaction(nonce, account.Address, value, gasLimit, gasPrice, nil)

	// Assina a transação
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

	fmt.Printf("✓ Transação enviada: %s\n", signedTx.Hash().Hex())
	fmt.Println("Aguardando confirmação (máx 30 segundos)...")

	// Aguarda confirmação
	startBlock, _ := client.BlockNumber(context.Background())
	waited := 0
	for waited < 30 {
		currentBlock, _ := client.BlockNumber(context.Background())
		if currentBlock > startBlock {
			fmt.Printf("✓ SUCCESS: Bloco criado! Bloco atual: %d\n", currentBlock)
			os.Exit(0)
		}
		time.Sleep(1 * time.Second)
		waited++
		if waited%5 == 0 {
			fmt.Printf("  Aguardando... (%d/%d segundos)\n", waited, 30)
		}
	}

	currentBlock, _ := client.BlockNumber(context.Background())
	fmt.Printf("Bloco atual: %d\n", currentBlock)
	if currentBlock > 0 {
		fmt.Println("✓ SUCCESS: Blocos estão sendo criados!")
	} else {
		fmt.Println("⚠ AVISO: Ainda no bloco 0.")
		fmt.Println("O Clique pode não estar selando blocos automaticamente.")
	}
}


