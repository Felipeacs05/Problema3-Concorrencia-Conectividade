package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const rpcURL = "http://localhost:8545"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: check-tx.exe <hash_da_transacao>")
		fmt.Println("\nExemplo:")
		fmt.Println("  check-tx.exe 0x80faebe1ac84d2c4fb86ac26d84987436fc2599ed057d99b6628c5f9d5c2f51e")
		os.Exit(1)
	}

	txHash := common.HexToHash(os.Args[1])

	// Conecta ao Geth
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		fmt.Printf("[ERRO] Falha ao conectar ao Geth: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// Obtém o receipt da transação
	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		fmt.Printf("[ERRO] Falha ao obter receipt: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("========================================")
	fmt.Println("Detalhes da Transação")
	fmt.Println("========================================")
	fmt.Printf("Hash: %s\n", txHash.Hex())
	fmt.Printf("Bloco: %d\n", receipt.BlockNumber.Uint64())
	fmt.Printf("Status: %d (1=sucesso, 0=falha)\n", receipt.Status)
	fmt.Printf("Gas usado: %d\n", receipt.GasUsed)
	fmt.Printf("Total de logs: %d\n\n", len(receipt.Logs))

	// Carrega ABI
	abiPaths := []string{
		"../contracts/GameEconomy.abi",
		"../../contracts/GameEconomy.abi",
		"../../../contracts/GameEconomy.abi",
		"../Projeto/Blockchain/contracts/GameEconomy.abi",
		"./GameEconomy.abi",
	}

	var abiBytes []byte
	for _, abiPath := range abiPaths {
		abiBytes, err = ioutil.ReadFile(abiPath)
		if err == nil {
			break
		}
	}

	if err != nil {
		fmt.Printf("[AVISO] Não foi possível carregar ABI: %v\n", err)
		fmt.Println("\nLogs brutos (sem decodificação):")
		for i, log := range receipt.Logs {
			fmt.Printf("\nLog #%d:\n", i+1)
			fmt.Printf("  Endereço: %s\n", log.Address.Hex())
			fmt.Printf("  Tópicos: %d\n", len(log.Topics))
			for j, topic := range log.Topics {
				fmt.Printf("    Topic[%d]: %s\n", j, topic.Hex())
			}
			fmt.Printf("  Dados: %s\n", common.Bytes2Hex(log.Data))
		}
		return
	}

	contractABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		fmt.Printf("[ERRO] Falha ao fazer parse do ABI: %v\n", err)
		os.Exit(1)
	}

	// Processa cada log
	if len(receipt.Logs) == 0 {
		fmt.Println("⚠ Nenhum evento (log) encontrado nesta transação!")
		return
	}

	fmt.Println("Eventos encontrados:")
	fmt.Println()

	for i, log := range receipt.Logs {
		fmt.Printf("═══════════════════════════════════════════════════════════\n")
		fmt.Printf("Evento #%d\n", i+1)
		fmt.Printf("Endereço do contrato: %s\n", log.Address.Hex())

		if len(log.Topics) == 0 {
			fmt.Println("⚠ Log sem tópicos (não é um evento padrão)")
			continue
		}

		eventSig := log.Topics[0]
		fmt.Printf("Assinatura do evento: %s\n", eventSig.Hex())

		// Tenta identificar o evento
		eventFound := false
		for eventName, event := range contractABI.Events {
			eventSigHash := common.BytesToHash(event.ID.Bytes())
			if eventSigHash == eventSig {
				eventFound = true
				fmt.Printf("Tipo: %s\n", eventName)

				// Decodifica dados
				unpacked, err := contractABI.Unpack(eventName, log.Data)
				if err == nil {
					fmt.Printf("Dados: ")
					for j, value := range unpacked {
						if j > 0 {
							fmt.Print(", ")
						}
						fmt.Print(formatValue(value))
					}
					fmt.Println()
				}

				// Processa tópicos indexados
				if len(log.Topics) > 1 {
					fmt.Printf("Parâmetros indexados: ")
					for j := 1; j < len(log.Topics) && j-1 < len(event.Inputs); j++ {
						if event.Inputs[j-1].Indexed {
							if j > 1 {
								fmt.Print(", ")
							}
							paramName := event.Inputs[j-1].Name
							paramValue := log.Topics[j].Hex()
							// Se for address, mostra como address
							if event.Inputs[j-1].Type.T == abi.AddressTy {
								addr := common.BytesToAddress(log.Topics[j].Bytes())
								paramValue = addr.Hex()
							}
							fmt.Printf("%s=%s", paramName, paramValue)
						}
					}
					fmt.Println()
				}
				break
			}
		}

		if !eventFound {
			fmt.Println("⚠ Evento não reconhecido no ABI")
			fmt.Printf("Dados brutos: %s\n", common.Bytes2Hex(log.Data))
		}
		fmt.Println()
	}
}

func formatValue(v interface{}) string {
	switch val := v.(type) {
	case common.Address:
		return val.Hex()
	case *big.Int:
		return val.String()
	case string:
		return val
	case []byte:
		return common.Bytes2Hex(val)
	case []*big.Int:
		result := "["
		for i, item := range val {
			if i > 0 {
				result += ", "
			}
			result += item.String()
		}
		result += "]"
		return result
	default:
		jsonBytes, _ := json.Marshal(v)
		return string(jsonBytes)
	}
}
