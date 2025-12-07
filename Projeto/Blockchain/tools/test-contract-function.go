package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const rpcURL = "http://localhost:8545"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: test-contract-function.exe <endereco_do_contrato>")
		os.Exit(1)
	}

	contractAddress := common.HexToAddress(os.Args[1])

	// Conecta ao Geth
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		fmt.Printf("[ERRO] Falha ao conectar: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

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
			fmt.Printf("[OK] ABI carregado de: %s\n", abiPath)
			break
		}
	}

	if err != nil {
		fmt.Printf("[ERRO] Não foi possível carregar ABI: %v\n", err)
		os.Exit(1)
	}

	contractABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		fmt.Printf("[ERRO] Falha ao fazer parse do ABI: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("========================================")
	fmt.Println("Teste de Funções do Contrato")
	fmt.Println("========================================")
	fmt.Printf("Contrato: %s\n", contractAddress.Hex())
	fmt.Println()

	// Verifica se registrarPartidaAdmin existe no ABI
	if _, ok := contractABI.Methods["registrarPartidaAdmin"]; ok {
		fmt.Println("✓ registrarPartidaAdmin encontrado no ABI")
	} else {
		fmt.Println("✗ registrarPartidaAdmin NÃO encontrado no ABI")
	}

	// Tenta chamar owner() para verificar se o contrato responde
	fmt.Println("\nTestando chamada owner()...")
	ownerData, err := contractABI.Pack("owner")
	if err != nil {
		fmt.Printf("✗ Erro ao preparar chamada owner: %v\n", err)
	} else {
		result, err := client.CallContract(context.Background(), ethereum.CallMsg{
			To:   &contractAddress,
			Data: ownerData,
		}, nil)
		if err != nil {
			fmt.Printf("✗ Erro ao chamar owner: %v\n", err)
		} else {
			var ownerAddr common.Address
			err = contractABI.UnpackIntoInterface(&ownerAddr, "owner", result)
			if err != nil {
				fmt.Printf("✗ Erro ao decodificar owner: %v\n", err)
			} else {
				fmt.Printf("✓ Owner do contrato: %s\n", ownerAddr.Hex())
			}
		}
	}

	// Tenta preparar uma chamada para registrarPartidaAdmin (sem enviar)
	fmt.Println("\nTestando preparação de chamada registrarPartidaAdmin...")
	testJ1 := common.HexToAddress("0xb4c5a85f6787C240EB75a722F071119ed51A9C96")
	testJ2 := common.HexToAddress("0x2D0c678F09ca6dD1692d81E1332F81783f84747D")
	testVencedor := common.HexToAddress("0x2D0c678F09ca6dD1692d81E1332F81783f84747D")

	data, err := contractABI.Pack("registrarPartidaAdmin", testJ1, testJ2, testVencedor)
	if err != nil {
		fmt.Printf("✗ Erro ao preparar chamada registrarPartidaAdmin: %v\n", err)
		fmt.Println("  Isso significa que a função não existe no ABI ou os parâmetros estão errados")
	} else {
		fmt.Printf("✓ Chamada preparada com sucesso! Dados: %s...\n", common.Bytes2Hex(data[:min(20, len(data))]))
	}

	// Lista todas as funções disponíveis no ABI
	fmt.Println("\nFunções disponíveis no ABI:")
	for name, method := range contractABI.Methods {
		if strings.Contains(name, "Partida") {
			fmt.Printf("  - %s\n", name)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
