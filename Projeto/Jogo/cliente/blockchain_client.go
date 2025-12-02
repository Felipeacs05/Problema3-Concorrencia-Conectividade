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
	gasLimit    = uint64(5000000) // Reduzido de 80M para 5M para caber no limite do bloco
)

// getKeystorePath retorna o caminho do keystore (tenta vários caminhos possíveis)
func getKeystorePath() string {
	// Tenta vários caminhos possíveis
	paths := []string{
		"../Blockchain/data/keystore",
		"../../Blockchain/data/keystore",
		"../../../Blockchain/data/keystore",
		"./Blockchain/data/keystore",
		"../Projeto/Blockchain/data/keystore",
	}
	
	for _, path := range paths {
		if files, err := ioutil.ReadDir(path); err == nil && len(files) > 0 {
			return path
		}
	}
	
	// Retorna o padrão se nenhum funcionar
	return "../Blockchain/data/keystore"
}

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

	// Tenta carregar o endereço do contrato (tenta vários caminhos)
	contractAddrPaths := []string{
		"../contract-address.txt",
		"../../contract-address.txt",
		"../../../contract-address.txt",
		"../Projeto/contract-address.txt",
		"./contract-address.txt",
	}
	
	var contractAddrBytes []byte
	for _, path := range contractAddrPaths {
		contractAddrBytes, err = ioutil.ReadFile(path)
		if err == nil {
			break
		}
	}
	
	if err != nil {
		return fmt.Errorf("contrato não encontrado. Execute setup-blockchain.bat primeiro")
	}
	
	contractAddress = common.HexToAddress(strings.TrimSpace(string(contractAddrBytes)))
	
	// Carrega ABI (tenta vários caminhos)
	abiPaths := []string{
		"../Blockchain/contracts/GameEconomy.abi",
		"../../Blockchain/contracts/GameEconomy.abi",
		"../../../Blockchain/contracts/GameEconomy.abi",
		"../Projeto/Blockchain/contracts/GameEconomy.abi",
		"./Blockchain/contracts/GameEconomy.abi",
	}
	
	var abiBytes []byte
	for _, abiPath := range abiPaths {
		abiBytes, err = ioutil.ReadFile(abiPath)
		if err == nil {
			break
		}
	}
	
	if err != nil {
		return fmt.Errorf("ABI do contrato não encontrado")
	}
	
	parsedABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		return fmt.Errorf("erro ao fazer parse do ABI: %v", err)
	}
	
	contractABI = parsedABI
	// NÃO define blockchainEnabled = true aqui
	// Só será true quando a carteira também for carregada
	fmt.Printf("✓ Conexão com blockchain estabelecida (Contrato: %s)\n", contractAddress.Hex())
	fmt.Printf("  [Aguardando carregamento da carteira...]\n")
	return nil
}

// carregarCarteira carrega ou cria uma carteira
func carregarCarteira() error {
	// Removido verificação de blockchainEnabled aqui, pois ela só será true ao final desta função

	keystorePath := getKeystorePath()
	
	// Lista arquivos do keystore
	files, err := ioutil.ReadDir(keystorePath)
	if err != nil {
		return fmt.Errorf("erro ao ler keystore (%s): %v\nExecute criar-conta-jogador.bat primeiro", keystorePath, err)
	}

	if len(files) == 0 {
		return fmt.Errorf("nenhuma carteira encontrada em %s\nExecute criar-conta-jogador.bat primeiro", keystorePath)
	}

	// Se houver múltiplas carteiras, permite escolher
	var keyFile string
	if len(files) > 1 {
		fmt.Println("\nMúltiplas carteiras encontradas:")
		for i, file := range files {
			// Extrai endereço do nome do arquivo
			parts := strings.Split(file.Name(), "--")
			var endereco string
			if len(parts) >= 3 {
				endereco = "0x" + parts[len(parts)-1]
			} else {
				endereco = file.Name()
			}
			fmt.Printf("  %d. %s\n", i+1, endereco)
		}
		fmt.Print("Escolha uma carteira (número): ")
		var escolha int
		fmt.Scanln(&escolha)
		if escolha < 1 || escolha > len(files) {
			return fmt.Errorf("escolha inválida")
		}
		keyFile = files[escolha-1].Name()
	} else {
		// Usa a única carteira encontrada
		keyFile = files[0].Name()
		// Extrai endereço para mostrar
		parts := strings.Split(keyFile, "--")
		if len(parts) >= 3 {
			endereco := "0x" + parts[len(parts)-1]
			fmt.Printf("Carteira encontrada: %s\n", endereco)
		}
	}

	keyPath := keystorePath + "/" + keyFile
	if strings.Contains(keyPath, "\\") {
		// Windows path
		keyPath = strings.ReplaceAll(keyPath, "/", "\\")
	}

	// Solicita senha
	fmt.Print("Digite a senha da sua carteira: ")
	var senha string
	fmt.Scanln(&senha)
	senha = strings.TrimSpace(senha)

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

	// Só agora que a carteira foi carregada, habilita a blockchain
	blockchainEnabled = true

	fmt.Printf("✓ Carteira carregada: %s\n", contaBlockchain.Hex())
	fmt.Printf("✓ Blockchain totalmente conectada e pronta para uso!\n")
	return nil
}

// comprarPacoteBlockchain compra um pacote de cartas na blockchain
func comprarPacoteBlockchain() error {
	fmt.Printf("[DEBUG] comprarPacoteBlockchain() iniciado\n")
	fmt.Printf("[DEBUG] blockchainEnabled=%v, chavePrivada!=nil=%v\n", blockchainEnabled, chavePrivada != nil)
	fmt.Printf("[DEBUG] contaBlockchain=%s, contractAddress=%s\n", contaBlockchain.Hex(), contractAddress.Hex())
	
	if !blockchainEnabled {
		return fmt.Errorf("blockchain não está habilitada")
	}

	if chavePrivada == nil {
		return fmt.Errorf("carteira não carregada")
	}

	// Prepara a chamada
	fmt.Printf("[DEBUG] Preparando chamada comprarPacote()...\n")
	data, err := contractABI.Pack("comprarPacote")
	if err != nil {
		fmt.Printf("[ERRO] Falha ao preparar chamada: %v\n", err)
		return fmt.Errorf("erro ao preparar chamada: %v", err)
	}
	fmt.Printf("[DEBUG] Chamada preparada, dados: %s...\n", common.Bytes2Hex(data[:min(20, len(data))]))

	// Valor: 1 ETH
	valor := big.NewInt(1000000000000000000) // 1 ETH em wei
	fmt.Printf("[DEBUG] Valor da transação: %s wei (1 ETH)\n", valor.String())

	// Envia transação
	fmt.Printf("[DEBUG] Enviando transação...\n")
	tx, err := enviarTransacaoBlockchain(data, valor)
	if err != nil {
		fmt.Printf("[ERRO] Falha ao enviar transação: %v\n", err)
		return fmt.Errorf("erro ao enviar transação: %v", err)
	}

	fmt.Printf("[DEBUG] Transação criada: Hash=%s, Nonce=%d\n", tx.Hash().Hex(), tx.Nonce())
	fmt.Printf("Transação enviada: %s\n", tx.Hash().Hex())
	fmt.Println("Aguardando confirmação...")

	// Aguarda confirmação
	fmt.Printf("[DEBUG] Aguardando confirmação da transação...\n")
	receipt, err := aguardarConfirmacaoBlockchain(tx.Hash())
	if err != nil {
		fmt.Printf("[ERRO] Falha ao aguardar confirmação: %v\n", err)
		return err
	}

	fmt.Printf("[DEBUG] Receipt recebido: Status=%d, BlockNumber=%d, GasUsed=%d\n", receipt.Status, receipt.BlockNumber.Uint64(), receipt.GasUsed)

	if receipt.Status == 0 {
		fmt.Printf("[ERRO] Transação falhou (Status=0)\n")
		return fmt.Errorf("transação falhou")
	}

	fmt.Println("✓ Pacote comprado com sucesso!")
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
	fmt.Printf("[DEBUG] enviarTransacaoBlockchain() iniciado\n")
	fmt.Printf("[DEBUG] contaBlockchain=%s, contractAddress=%s\n", contaBlockchain.Hex(), contractAddress.Hex())
	
	nonce, err := blockchainClient.PendingNonceAt(context.Background(), contaBlockchain)
	if err != nil {
		fmt.Printf("[ERRO] Falha ao obter nonce: %v\n", err)
		return nil, err
	}
	fmt.Printf("[DEBUG] Nonce obtido: %d\n", nonce)

	gasPrice, err := blockchainClient.SuggestGasPrice(context.Background())
	if err != nil {
		fmt.Printf("[ERRO] Falha ao obter gasPrice: %v\n", err)
		return nil, err
	}
	fmt.Printf("[DEBUG] GasPrice sugerido: %s wei\n", gasPrice.String())

	chainID, err := blockchainClient.NetworkID(context.Background())
	if err != nil {
		fmt.Printf("[ERRO] Falha ao obter chainID: %v\n", err)
		return nil, err
	}
	fmt.Printf("[DEBUG] ChainID: %s\n", chainID.String())
	fmt.Printf("[DEBUG] GasLimit: %d\n", gasLimit)

	tx := types.NewTransaction(nonce, contractAddress, valor, gasLimit, gasPrice, data)
	fmt.Printf("[DEBUG] Transação criada (antes de assinar)\n")
	
	txAssinada, err := types.SignTx(tx, types.NewEIP155Signer(chainID), chavePrivada.PrivateKey)
	if err != nil {
		fmt.Printf("[ERRO] Falha ao assinar transação: %v\n", err)
		return nil, err
	}
	fmt.Printf("[DEBUG] Transação assinada com sucesso\n")

	fmt.Printf("[DEBUG] Enviando transação para a blockchain...\n")
	err = blockchainClient.SendTransaction(context.Background(), txAssinada)
	if err != nil {
		fmt.Printf("[ERRO] Falha ao enviar transação: %v\n", err)
		return nil, err
	}
	fmt.Printf("[DEBUG] Transação enviada com sucesso!\n")

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

