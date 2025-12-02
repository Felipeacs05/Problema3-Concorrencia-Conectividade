package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const rpcURL = "http://localhost:8545"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: view-events.exe <endereco_do_contrato> [bloco_inicial] [bloco_final]")
		fmt.Println("\nExemplos:")
		fmt.Println("  view-events.exe 0x4D82F25Ef5058fE2308135D88A386c90FEdBe99A")
		fmt.Println("  view-events.exe 0x4D82F25Ef5058fE2308135D88A386c90FEdBe99A 0 1000")
		fmt.Println("\nSe bloco_inicial e bloco_final não forem especificados, busca os últimos 1000 blocos.")
		os.Exit(1)
	}

	contractAddressStr := os.Args[1]
	contractAddress := common.HexToAddress(contractAddressStr)

	// Conecta ao Geth
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		fmt.Printf("[ERRO] Falha ao conectar ao Geth: %v\n", err)
		fmt.Println("Certifique-se de que o Geth está rodando (docker ps)")
		os.Exit(1)
	}
	defer client.Close()

	// Carrega ABI do contrato
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
		fmt.Printf("[ERRO] Falha ao carregar ABI: %v\n", err)
		fmt.Println("Certifique-se de que GameEconomy.abi existe em Blockchain/contracts/")
		os.Exit(1)
	}

	contractABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		fmt.Printf("[ERRO] Falha ao fazer parse do ABI: %v\n", err)
		os.Exit(1)
	}

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
	fmt.Println("Visualizador de Eventos do Contrato")
	fmt.Println("========================================")
	fmt.Printf("Contrato: %s\n", contractAddress.Hex())
	fmt.Printf("Buscando eventos dos blocos %d a %d...\n", fromBlock, toBlock)
	fmt.Println()

	// Busca logs do contrato
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
		Addresses: []common.Address{contractAddress},
	}

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		fmt.Printf("[ERRO] Falha ao buscar logs: %v\n", err)
		os.Exit(1)
	}

	if len(logs) == 0 {
		fmt.Println("Nenhum evento encontrado neste intervalo.")
		return
	}

	fmt.Printf("Total de eventos encontrados: %d\n\n", len(logs))

	// Processa cada log
	for i, log := range logs {
		block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(log.BlockNumber)))
		if err != nil {
			continue
		}

		fmt.Printf("═══════════════════════════════════════════════════════════\n")
		fmt.Printf("Evento #%d\n", i+1)
		fmt.Printf("Bloco: %d\n", log.BlockNumber)
		fmt.Printf("Hash da transação: %s\n", log.TxHash.Hex())
		fmt.Printf("Timestamp: %s\n", time.Unix(int64(block.Time()), 0).Format("2006-01-02 15:04:05"))

		// Tenta identificar o evento pelo tópico
		if len(log.Topics) > 0 {
			eventSig := log.Topics[0].Hex()
			fmt.Printf("Assinatura do evento: %s\n", eventSig)

			// Tenta fazer decode do evento
			for _, event := range contractABI.Events {
				eventSigHash := common.BytesToHash(contractABI.Events[event.Name].ID.Bytes())
				if eventSigHash == log.Topics[0] {
					fmt.Printf("Tipo: %s\n", event.Name)

					// Faz decode dos dados
					unpacked, err := contractABI.Unpack(event.Name, log.Data)
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
								fmt.Printf("%s=%s", event.Inputs[j-1].Name, log.Topics[j].Hex())
							}
						}
						fmt.Println()
					}
					break
				}
			}
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
	default:
		jsonBytes, _ := json.Marshal(v)
		return string(jsonBytes)
	}
}

