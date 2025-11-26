package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"jogodistribuido/protocolo"
)

const (
	rpcURL      = "http://localhost:8545"
	keystorePath = "../Blockchain/data/keystore"
	gasLimit    = uint64(80000000)
)

var (
	blockchainClient *ethclient.Client
	blockchainRPC    *rpc.Client
	contaBlockchain  common.Address
	chavePrivada     *keystore.Key
	contractAddress  common.Address
	contractABI      abi.ABI
	senhaConta       string
	blockchainEnabled bool
)

// inicializarBlockchain inicializa a conexão com a blockchain
func inicializarBlockchain() error {
	// Tenta conectar
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return fmt.Errorf("blockchain não disponível: %v", err)
	}

	rpcClient, err := rpc.Dial(rpcURL)
	if err != nil {
		// Não é crítico
	}

	blockchainClient = client
	blockchainRPC = rpcClient

	// Tenta carregar o endereço do contrato
	contractAddrBytes, err := ioutil.ReadFile("../contract-address.txt")
	if err == nil {
		contractAddress = common.HexToAddress(strings.TrimSpace(string(contractAddrBytes)))
		
		// Carrega ABI
		abiPath := "../Blockchain/contracts/GameEconomy.abi"
		abiBytes, err := ioutil.ReadFile(abiPath)
		if err == nil {
			parsedABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
			if err == nil {
				contractABI = parsedABI
				blockchainEnabled = true
				return nil
			}
		}
	}

	return fmt.Errorf("blockchain configurada mas contrato não encontrado")
}

// carregarCarteira carrega ou cria uma carteira
func carregarCarteira() error {
	if !blockchainEnabled {
		return fmt.Errorf("blockchain não está habilitada")
	}

	// Lista arquivos do keystore
	files, err := ioutil.ReadDir(keystorePath)
	if err != nil {
		return fmt.Errorf("erro ao ler keystore: %v", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("nenhuma carteira encontrada. Execute criar-conta-jogador.bat primeiro")
	}

	// Usa o primeiro arquivo encontrado (ou pode pedir para escolher)
	keyFile := files[0].Name()
	keyPath := keystorePath + "/" + keyFile

	// Solicita senha
	fmt.Print("Digite a senha da sua carteira: ")
	var senha string
	fmt.Scanln(&senha)

	jsonBytes, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo da carteira: %v", err)
	}

	key, err := keystore.DecryptKey(jsonBytes, senha)
	if err != nil {
		return fmt.Errorf("senha incorreta: %v", err)
	}

	chavePrivada = key
	contaBlockchain = key.Address
	senhaConta = senha

	fmt.Printf("✓ Carteira carregada: %s\n", contaBlockchain.Hex())
	return nil
}

// comprarPacoteBlockchain compra um pacote de cartas na blockchain
func comprarPacoteBlockchain() error {
	if !blockchainEnabled {
		return fmt.Errorf("blockchain não está habilitada")
	}

	// Prepara a chamada
	data, err := contractABI.Pack("comprarPacote")
	if err != nil {
		return fmt.Errorf("erro ao preparar chamada: %v", err)
	}

	// Valor: 1 ETH
	valor := big.NewInt(1000000000000000000) // 1 ETH em wei

	// Envia transação
	tx, err := enviarTransacaoBlockchain(data, valor)
	if err != nil {
		return fmt.Errorf("erro ao enviar transação: %v", err)
	}

	fmt.Printf("Transação enviada: %s\n", tx.Hash().Hex())
	fmt.Println("Aguardando confirmação...")

	// Aguarda confirmação
	receipt, err := aguardarConfirmacaoBlockchain(tx.Hash())
	if err != nil {
		return err
	}

	if receipt.Status == 0 {
		return fmt.Errorf("transação falhou")
	}

	fmt.Println("✓ Pacote comprado com sucesso!")
	return nil
}

// obterInventarioBlockchain obtém o inventário da blockchain
func obterInventarioBlockchain() ([]protocolo.Carta, error) {
	if !blockchainEnabled {
		return nil, fmt.Errorf("blockchain não está habilitada")
	}

	// Prepara chamada
	data, err := contractABI.Pack("obterInventario", contaBlockchain)
	if err != nil {
		return nil, fmt.Errorf("erro ao preparar chamada: %v", err)
	}

	// Faz chamada
	msg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}

	result, err := blockchainClient.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao chamar contrato: %v", err)
	}

	// Desempacota IDs
	var ids []*big.Int
	err = contractABI.UnpackIntoInterface(&ids, "obterInventario", result)
	if err != nil {
		return nil, fmt.Errorf("erro ao desempacotar: %v", err)
	}

	// Para cada ID, obtém a carta
	cartas := make([]protocolo.Carta, 0, len(ids))
	for _, id := range ids {
		carta, err := obterCartaBlockchain(id)
		if err == nil {
			cartas = append(cartas, carta)
		}
	}

	return cartas, nil
}

// obterCartaBlockchain obtém os dados de uma carta
func obterCartaBlockchain(cartaID *big.Int) (protocolo.Carta, error) {
	data, err := contractABI.Pack("obterCarta", cartaID)
	if err != nil {
		return protocolo.Carta{}, err
	}

	msg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}

	result, err := blockchainClient.CallContract(context.Background(), msg, nil)
	if err != nil {
		return protocolo.Carta{}, err
	}

	var cartaData struct {
		Id        *big.Int
		Nome      string
		Naipe     string
		Valor     *big.Int
		Raridade  string
		Timestamp *big.Int
	}

	err = contractABI.UnpackIntoInterface(&cartaData, "obterCarta", result)
	if err != nil {
		return protocolo.Carta{}, err
	}

	return protocolo.Carta{
		ID:       cartaData.Id.String(),
		Nome:     cartaData.Nome,
		Naipe:    cartaData.Naipe,
		Valor:    int(cartaData.Valor.Int64()),
		Raridade: cartaData.Raridade,
	}, nil
}

// enviarTransacaoBlockchain envia uma transação
func enviarTransacaoBlockchain(data []byte, valor *big.Int) (*types.Transaction, error) {
	nonce, err := blockchainClient.PendingNonceAt(context.Background(), contaBlockchain)
	if err != nil {
		return nil, err
	}

	gasPrice, err := blockchainClient.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	chainID, err := blockchainClient.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}

	tx := types.NewTransaction(nonce, contractAddress, valor, gasLimit, gasPrice, data)
	txAssinada, err := types.SignTx(tx, types.NewEIP155Signer(chainID), chavePrivada.PrivateKey)
	if err != nil {
		return nil, err
	}

	err = blockchainClient.SendTransaction(context.Background(), txAssinada)
	if err != nil {
		return nil, err
	}

	return txAssinada, nil
}

// aguardarConfirmacaoBlockchain aguarda confirmação
func aguardarConfirmacaoBlockchain(txHash common.Hash) (*types.Receipt, error) {
	for i := 0; i < 30; i++ {
		receipt, err := blockchainClient.TransactionReceipt(context.Background(), txHash)
		if err == nil {
			return receipt, nil
		}
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("timeout aguardando confirmação")
}

