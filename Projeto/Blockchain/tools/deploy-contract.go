package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("Deploy do Contrato GameEconomy")
	fmt.Println("========================================")
	fmt.Println()

	// Conecta ao Geth
	rpcURL := "http://localhost:8545"
	fmt.Printf("[1/5] Conectando ao Geth em %s...\n", rpcURL)
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		fmt.Printf("ERRO: Falha ao conectar: %v\n", err)
		fmt.Println("Certifique-se de que o Geth está rodando (docker-compose up -d)")
		os.Exit(1)
	}
	defer client.Close()
	fmt.Println("[OK] Conectado ao Geth")

	// Carrega keystore (conta do servidor)
	keystorePath := filepath.Join("..", "data", "keystore")
	fmt.Printf("[2/5] Carregando keystore de %s...\n", keystorePath)
	ks := keystore.NewKeyStore(keystorePath, keystore.StandardScryptN, keystore.StandardScryptP)
	
	if len(ks.Accounts()) == 0 {
		fmt.Println("ERRO: Nenhuma conta encontrada no keystore!")
		fmt.Println("Execute setup-blockchain.bat primeiro")
		os.Exit(1)
	}

	account := ks.Accounts()[0]
	fmt.Printf("[OK] Conta encontrada: %s\n", account.Address.Hex())

	// Lê senha
	passwordPath := filepath.Join("..", "data", "password.txt")
	fmt.Printf("[3/5] Lendo senha de %s...\n", passwordPath)
	passwordBytes, err := ioutil.ReadFile(passwordPath)
	if err != nil {
		fmt.Printf("[AVISO] Não foi possível ler password.txt: %v\n", err)
		fmt.Print("Digite a senha da conta: ")
		var password string
		fmt.Scanln(&password)
		passwordBytes = []byte(password)
	}
	password := strings.TrimSpace(string(passwordBytes))
	fmt.Println("[OK] Senha carregada")

	// Desbloqueia conta
	fmt.Println("[4/5] Desbloqueando conta...")
	err = ks.Unlock(account, password)
	if err != nil {
		fmt.Printf("ERRO: Falha ao desbloquear conta: %v\n", err)
		fmt.Println("Verifique se a senha está correta")
		os.Exit(1)
	}
	fmt.Println("[OK] Conta desbloqueada")

	// Lê bytecode
	binPath := filepath.Join("..", "contracts", "GameEconomy.bin")
	fmt.Printf("[5/5] Lendo bytecode de %s...\n", binPath)
	bytecodeHex, err := ioutil.ReadFile(binPath)
	if err != nil {
		fmt.Printf("ERRO: Falha ao ler bytecode: %v\n", err)
		fmt.Println("Execute compile-contract.bat primeiro!")
		os.Exit(1)
	}

	bytecode := strings.TrimSpace(string(bytecodeHex))
	if !strings.HasPrefix(bytecode, "0x") {
		bytecode = "0x" + bytecode
	}
	fmt.Printf("[OK] Bytecode carregado (%d bytes)\n", len(bytecode)/2)

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("Enviando transação de deploy...")
	fmt.Println("========================================")

	// Obtém nonce
	nonce, err := client.PendingNonceAt(context.Background(), account.Address)
	if err != nil {
		fmt.Printf("ERRO: Falha ao obter nonce: %v\n", err)
		os.Exit(1)
	}

	// Obtém gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		gasPrice = big.NewInt(1000000000) // 1 gwei como fallback
	}

	// Cria transação de deploy
	gasLimit := uint64(8000000) // 8M gas (ajustado para blockchain privada)
	tx := types.NewContractCreation(nonce, big.NewInt(0), gasLimit, gasPrice, common.FromHex(bytecode))

	// Obtém chain ID dinamicamente
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		// Fallback para 1337 se não conseguir obter
		chainID = big.NewInt(1337)
		fmt.Printf("[AVISO] Não foi possível obter chain ID, usando 1337\n")
	}
	fmt.Printf("Chain ID: %s\n", chainID.String())

	// Assina transação
	signedTx, err := ks.SignTx(account, tx, chainID)
	if err != nil {
		fmt.Printf("ERRO: Falha ao assinar transação: %v\n", err)
		os.Exit(1)
	}

	// Envia transação
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		fmt.Printf("ERRO: Falha ao enviar transação: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Transação enviada: %s\n", signedTx.Hash().Hex())
	fmt.Println("Aguardando confirmação (pode levar alguns segundos)...")

	// Aguarda confirmação
	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var receipt *types.Receipt
	for {
		select {
		case <-timeout:
			fmt.Println("ERRO: Timeout aguardando confirmação")
			fmt.Printf("Transação: %s\n", signedTx.Hash().Hex())
			fmt.Println("Verifique manualmente se o contrato foi deployado")
			os.Exit(1)
		case <-ticker.C:
			receipt, err = client.TransactionReceipt(context.Background(), signedTx.Hash())
			if err == nil && receipt != nil {
				contractAddress := receipt.ContractAddress
				fmt.Println()
				fmt.Println("========================================")
				fmt.Printf("✓ Contrato deployado com sucesso!\n")
				fmt.Printf("Endereço: %s\n", contractAddress.Hex())
				fmt.Println("========================================")

				// Salva endereço
				projectDir := filepath.Join("..", "..")
				addressFile := filepath.Join(projectDir, "contract-address.txt")
				err = ioutil.WriteFile(addressFile, []byte(contractAddress.Hex()+"\n"), 0644)
				if err != nil {
					fmt.Printf("[AVISO] Falha ao salvar endereço: %v\n", err)
					fmt.Printf("Salve manualmente em: %s\n", addressFile)
				} else {
					fmt.Printf("✓ Endereço salvo em: %s\n", addressFile)
				}

				fmt.Println()
				fmt.Println("Agora você pode usar o cliente do jogo!")
				os.Exit(0)
			}
		}
	}
}

