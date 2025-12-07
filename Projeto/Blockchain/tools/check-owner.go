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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: check-owner.exe <endereco_contrato>")
		os.Exit(1)
	}

	contractAddr := common.HexToAddress(os.Args[1])

	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		fmt.Printf("Erro ao conectar: %v\n", err)
		os.Exit(1)
	}

	// Carrega ABI
	abiBytes, err := ioutil.ReadFile("../contracts/GameEconomy.abi")
	if err != nil {
		fmt.Printf("Erro ao ler ABI: %v\n", err)
		os.Exit(1)
	}

	contractABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		fmt.Printf("Erro ao parsear ABI: %v\n", err)
		os.Exit(1)
	}

	// Chama owner()
	data, _ := contractABI.Pack("owner")
	result, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contractAddr,
		Data: data,
	}, nil)

	if err != nil {
		fmt.Printf("Erro ao chamar owner(): %v\n", err)
		os.Exit(1)
	}

	var owner common.Address
	contractABI.UnpackIntoInterface(&owner, "owner", result)

	fmt.Printf("Contrato: %s\n", contractAddr.Hex())
	fmt.Printf("Owner: %s\n", owner.Hex())

	// Verifica se registrarPartidaAdmin existe
	if _, ok := contractABI.Methods["registrarPartidaAdmin"]; ok {
		fmt.Println("✓ registrarPartidaAdmin existe no ABI")
	} else {
		fmt.Println("✗ registrarPartidaAdmin NÃO existe no ABI")
	}

	// Verifica código do contrato
	code, err := client.CodeAt(context.Background(), contractAddr, nil)
	if err != nil {
		fmt.Printf("Erro ao obter código: %v\n", err)
	} else {
		fmt.Printf("Tamanho do bytecode deployado: %d bytes\n", len(code))
		if len(code) == 0 {
			fmt.Println("⚠ ERRO: Não há código neste endereço! O contrato não foi deployado corretamente.")
		}
	}
}
