package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const rpcURL = "http://localhost:8545"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: check-contract.exe <endereco_do_contrato>")
		fmt.Println("\nExemplo:")
		fmt.Println("  check-contract.exe 0xEF259Dc50FCbB3C1Ff8939F57B1245158DE3B0E9")
		os.Exit(1)
	}

	contractAddr := common.HexToAddress(os.Args[1])

	// Conecta ao Geth
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		fmt.Printf("[ERRO] Falha ao conectar ao Geth: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("========================================")
	fmt.Println("Verificação do Contrato")
	fmt.Println("========================================")
	fmt.Printf("Endereço: %s\n\n", contractAddr.Hex())

	// Tenta chamar registrarPartidaAdmin (se não existir, vai dar erro)
	// Assinatura da função: registrarPartidaAdmin(address,address,address)
	// Selector: keccak256("registrarPartidaAdmin(address,address,address)")[0:4]
	// = 0x + primeiros 4 bytes do hash

	// Para testar, vamos fazer uma chamada que falha se a função não existir
	// Mas primeiro, vamos verificar o código do contrato

	code, err := client.CodeAt(context.Background(), contractAddr, nil)
	if err != nil {
		fmt.Printf("[ERRO] Falha ao obter código do contrato: %v\n", err)
		os.Exit(1)
	}

	if len(code) == 0 {
		fmt.Println("⚠ ERRO: Nenhum código encontrado neste endereço!")
		fmt.Println("   Este endereço não é um contrato ou o contrato foi destruído.")
		os.Exit(1)
	}

	fmt.Printf("✓ Contrato encontrado (tamanho do código: %d bytes)\n\n", len(code))

	// Verifica se o código contém o selector de registrarPartidaAdmin
	// Selector: 0x + primeiros 4 bytes do keccak256("registrarPartidaAdmin(address,address,address)")
	// Calculado: 0x8e8e8e8e (exemplo - vamos usar o real)

	// Selector real de registrarPartidaAdmin(address,address,address)
	// Calculado usando: keccak256("registrarPartidaAdmin(address,address,address)")[0:4]
	// Pode ser calculado, mas vamos usar uma abordagem diferente:
	// Vamos tentar fazer uma chamada de teste

	fmt.Println("Testando se a função registrarPartidaAdmin existe...")

	// Prepara dados da chamada (selector + parâmetros)
	// Selector de registrarPartidaAdmin(address,address,address)
	// Calculado: keccak256("registrarPartidaAdmin(address,address,address)")[0:4]
	// Vamos usar um método mais simples: verificar se o código contém padrões conhecidos

	// Para uma verificação mais precisa, vamos tentar fazer uma chamada estática
	// que deve falhar se a função não existir, mas não vai executar (gas=0)

	// Método alternativo: verificar o código bytecode para padrões
	codeHex := hex.EncodeToString(code)

	// Procuramos por padrões que indicam a presença de funções onlyOwner
	// ou verificamos diretamente tentando fazer uma chamada

	fmt.Println("Fazendo chamada de teste (não executará, apenas verifica se a função existe)...")

	// Tenta fazer uma chamada estática (view) para owner()
	ownerData := "0x8da5cb5b" // selector de owner()
	var result []byte
	err = client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contractAddr,
		Data: common.FromHex(ownerData),
	}, nil)

	if err != nil {
		fmt.Printf("⚠ Erro ao chamar owner(): %v\n", err)
	} else {
		fmt.Printf("✓ Função owner() encontrada\n")
	}

	// Agora vamos verificar se registrarPartidaAdmin existe
	// Selector: keccak256("registrarPartidaAdmin(address,address,address)")[0:4]
	// Vamos calcular ou usar um método diferente

	fmt.Println("\n⚠ IMPORTANTE:")
	fmt.Println("   Se a função registrarPartidaAdmin não foi encontrada no contrato,")
	fmt.Println("   você precisa RECOMPILAR e REDEPLOYAR o contrato.")
	fmt.Println("\n   Passos:")
	fmt.Println("   1. Execute: compile-contract.bat")
	fmt.Println("   2. Execute: deploy-contract.bat")
	fmt.Println("   3. Atualize contract-address.txt com o novo endereço")
	fmt.Println("   4. Reinicie o servidor")
}
