package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const rpcURL = "http://localhost:8545"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: view-transactions.exe <endereco_da_conta> [bloco_inicial] [bloco_final]")
		fmt.Println("\nExemplos:")
		fmt.Println("  view-transactions.exe 0x2d50FD74Cc3eB554b16013667045804D28Bc54a4")
		fmt.Println("  view-transactions.exe 0x2d50FD74Cc3eB554b16013667045804D28Bc54a4 0 1000")
		fmt.Println("\nSe bloco_inicial e bloco_final não forem especificados, busca os últimos 1000 blocos.")
		os.Exit(1)
	}

	addressStr := os.Args[1]
	address := common.HexToAddress(addressStr)

	// Conecta ao Geth
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		fmt.Printf("[ERRO] Falha ao conectar ao Geth: %v\n", err)
		fmt.Println("Certifique-se de que o Geth está rodando (docker ps)")
		os.Exit(1)
	}
	defer client.Close()

	// Obtém o número do bloco atual
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		fmt.Printf("[ERRO] Falha ao obter bloco atual: %v\n", err)
		os.Exit(1)
	}
	currentBlock := header.Number.Uint64()

	// Define o range de blocos
	var fromBlock, toBlock uint64
	if len(os.Args) >= 3 {
		fmt.Sscanf(os.Args[2], "%d", &fromBlock)
	}
	if len(os.Args) >= 4 {
		fmt.Sscanf(os.Args[3], "%d", &toBlock)
	} else {
		toBlock = currentBlock
	}
	if fromBlock == 0 {
		if currentBlock > 1000 {
			fromBlock = currentBlock - 1000
		} else {
			fromBlock = 0
		}
	}

	fmt.Println("========================================")
	fmt.Println("Visualizador de Transações Blockchain")
	fmt.Println("========================================")
	fmt.Printf("Conta: %s\n", address.Hex())
	fmt.Printf("Buscando transações dos blocos %d a %d...\n", fromBlock, toBlock)
	fmt.Println()

	// Itera pelos blocos para encontrar transações
	var txCount int
	for i := fromBlock; i <= toBlock && i <= currentBlock; i++ {
		block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(i)))
		if err != nil {
			continue
		}

		for _, tx := range block.Transactions() {
			from, err := types.Sender(types.NewEIP155Signer(tx.ChainId()), tx)
			if err != nil {
				continue
			}

			// Verifica se é uma transação da conta
			if from == address || (tx.To() != nil && *tx.To() == address) {
				txCount++
				receipt, _ := client.TransactionReceipt(context.Background(), tx.Hash())

				fmt.Printf("═══════════════════════════════════════════════════════════\n")
				fmt.Printf("Transação #%d\n", txCount)
				fmt.Printf("Hash: %s\n", tx.Hash().Hex())
				fmt.Printf("Bloco: %d\n", i)
				fmt.Printf("De: %s\n", from.Hex())
				if tx.To() != nil {
					fmt.Printf("Para: %s\n", tx.To().Hex())
				} else {
					fmt.Printf("Para: [Criação de Contrato]\n")
				}
				fmt.Printf("Valor: %s ETH\n", weiToEther(tx.Value()))
				fmt.Printf("Gas usado: %d\n", tx.Gas())
				if receipt != nil {
					fmt.Printf("Gas usado (receipt): %d\n", receipt.GasUsed)
					fmt.Printf("Status: %s\n", statusString(receipt.Status))
				}
				fmt.Printf("Timestamp: %s\n", time.Unix(int64(block.Time()), 0).Format("2006-01-02 15:04:05"))
				if len(tx.Data()) > 0 {
					fmt.Printf("Dados: %s...\n", common.Bytes2Hex(tx.Data()[:min(20, len(tx.Data()))]))
				}
				fmt.Println()
			}
		}
	}

	if txCount == 0 {
		fmt.Println("Nenhuma transação encontrada neste intervalo.")
	} else {
		fmt.Printf("Total de transações encontradas: %d\n", txCount)
	}
}

func weiToEther(wei *big.Int) string {
	ether := new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(1e18))
	return ether.Text('f', 18)
}

func statusString(status uint64) string {
	if status == 1 {
		return "✓ Sucesso"
	}
	return "✗ Falhou"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

