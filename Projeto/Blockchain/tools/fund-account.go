// ===================== BAREMA ITEM 1: ARQUITETURA =====================
// Utilitário para transferir ETH da conta do signer para outra conta

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
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

	// Obtém a conta do servidor (pode estar no docker-compose ou ser a primeira)
	keystorePath := filepath.Join("..", "data", "keystore")
	ks := keystore.NewKeyStore(keystorePath, keystore.StandardScryptN, keystore.StandardScryptP)

	if len(ks.Accounts()) == 0 {
		fmt.Println("ERRO: Nenhuma conta encontrada no keystore!")
		os.Exit(1)
	}

	// Tenta encontrar a conta que está desbloqueada no Geth (lê do docker-compose)
	// Ou usa a primeira conta disponível
	var signerAccount accounts.Account
	dockerComposePath := filepath.Join("..", "docker-compose-blockchain.yml")
	if dockerComposeBytes, err := ioutil.ReadFile(dockerComposePath); err == nil {
		// Procura por --unlock= no docker-compose
		content := string(dockerComposeBytes)
		if strings.Contains(content, "--unlock=") {
			// Extrai o endereço após --unlock=
			parts := strings.Split(content, "--unlock=")
			if len(parts) > 1 {
				addrPart := strings.Fields(parts[1])[0]
				addrPart = strings.Trim(addrPart, "\"'\n\r")
				if strings.HasPrefix(addrPart, "0x") {
					unlockAddr := common.HexToAddress(addrPart)
					// Procura essa conta no keystore
					for _, acc := range ks.Accounts() {
						if acc.Address == unlockAddr {
							signerAccount = acc
							fmt.Printf("Conta signer (do docker-compose): %s\n", signerAccount.Address.Hex())
							break
						}
					}
				}
			}
		}
	}

	// Se não encontrou, usa a primeira conta
	if signerAccount.Address == (common.Address{}) {
		signerAccount = ks.Accounts()[0]
		fmt.Printf("Conta signer (primeira disponível): %s\n", signerAccount.Address.Hex())
	}

	// Lê senha do arquivo password.txt ou tenta senhas comuns
	passwordPath := filepath.Join("..", "data", "password.txt")
	passwordsToTry := []string{"123456"} // padrão
	
	if passwordBytes, err := ioutil.ReadFile(passwordPath); err == nil {
		readPassword := strings.TrimSpace(string(passwordBytes))
		// Remove quebras de linha e espaços extras
		readPassword = strings.Trim(readPassword, "\r\n\t ")
		if readPassword != "" {
			passwordsToTry = append([]string{readPassword}, passwordsToTry...)
		}
	}

	// Tenta desbloquear com cada senha
	var password string
	unlocked := false
	for _, pwd := range passwordsToTry {
		err = ks.Unlock(signerAccount, pwd)
		if err == nil {
			password = pwd
			unlocked = true
			break
		}
	}

	if !unlocked {
		fmt.Printf("ERRO: Falha ao desbloquear conta no keystore\n")
		fmt.Printf("Tentou as seguintes senhas:\n")
		for _, pwd := range passwordsToTry {
			fmt.Printf("  - '%s'\n", pwd)
		}
		fmt.Printf("\nDica: Verifique o arquivo %s ou a senha usada ao criar a conta do servidor\n", passwordPath)
		os.Exit(1)
	}
	fmt.Println("Conta desbloqueada no keystore!")

	// Desbloqueia a conta no Geth via RPC (opcional, mas ajuda)
	var unlockResult bool
	rpcClient.Call(&unlockResult, "personal_unlockAccount", signerAccount.Address, password, 0)
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

