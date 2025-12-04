package blockchain

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
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
	"jogodistribuido/servidor/tipos"
)

const (
	// Endereço RPC do nó Geth (pode ser configurado via variável de ambiente)
	DefaultRPCURL = "http://geth-node:8545"
	// Gas limit para transações
	DefaultGasLimit = uint64(80000000)
)

// Manager gerencia a interação com a blockchain
type Manager struct {
	client           *ethclient.Client
	rpcClient        *rpc.Client
	contractAddress  common.Address
	contractABI      abi.ABI
	serverAccount    common.Address
	serverKey        *keystore.Key
	serverPassword   string
	keystorePath     string
	gasLimit         uint64
}

// NewManager cria um novo gerenciador de blockchain
func NewManager(rpcURL, contractAddressHex, keystorePath, serverPassword string) (*Manager, error) {
	// Conecta ao cliente Ethereum
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao cliente Ethereum: %v", err)
	}

	// Conecta também via RPC para métodos personalizados
	rpcClient, err := rpc.Dial(rpcURL)
	if err != nil {
		// Não é crítico, apenas loga aviso
		fmt.Printf("⚠ Aviso: Não foi possível conectar via RPC\n")
	}

	// Parse do endereço do contrato
	contractAddress := common.HexToAddress(contractAddressHex)

	// Carrega o ABI do contrato
	// Tenta vários caminhos possíveis
	abiPaths := []string{
		"../Blockchain/contracts/GameEconomy.abi",
		"../../Blockchain/contracts/GameEconomy.abi",
		"../../../Blockchain/contracts/GameEconomy.abi",
		"/app/Blockchain/contracts/GameEconomy.abi", // Docker
	}
	
	var abiBytes []byte
	for _, abiPath := range abiPaths {
		abiBytes, err = ioutil.ReadFile(abiPath)
		if err == nil {
			break
		}
	}
	
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo ABI (tentou: %v): %v", abiPaths, err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer parse do ABI: %v", err)
	}

	// Carrega a conta do servidor (para registrar partidas)
	var serverAccount common.Address
	var serverKey *keystore.Key

	if keystorePath != "" && serverPassword != "" {
		// Lista arquivos do keystore
		files, err := ioutil.ReadDir(keystorePath)
		if err == nil && len(files) > 0 {
			// Usa o primeiro arquivo encontrado
			keyFile := files[0].Name()
			keyPath := keystorePath + "/" + keyFile

			jsonBytes, err := ioutil.ReadFile(keyPath)
			if err == nil {
				key, err := keystore.DecryptKey(jsonBytes, serverPassword)
				if err == nil {
					serverKey = key
					serverAccount = key.Address
				}
			}
		}
	}

	return &Manager{
		client:          client,
		rpcClient:       rpcClient,
		contractAddress: contractAddress,
		contractABI:     parsedABI,
		serverAccount:   serverAccount,
		serverKey:       serverKey,
		serverPassword:  serverPassword,
		keystorePath:    keystorePath,
		gasLimit:        DefaultGasLimit,
	}, nil
}

// ComprarPacote processa a compra de um pacote de cartas na blockchain
// Retorna os IDs das cartas criadas
func (m *Manager) ComprarPacote(jogadorAddress common.Address, valor *big.Int) ([]*big.Int, error) {
	// Prepara a chamada à função comprarPacote
	data, err := m.contractABI.Pack("comprarPacote")
	if err != nil {
		return nil, fmt.Errorf("erro ao preparar chamada: %v", err)
	}

	// Envia a transação
	tx, err := m.enviarTransacao(jogadorAddress, data, valor)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar transação: %v", err)
	}

	// Aguarda confirmação
	receipt, err := m.aguardarConfirmacao(tx.Hash())
	if err != nil {
		return nil, fmt.Errorf("erro ao aguardar confirmação: %v", err)
	}

	if receipt.Status == 0 {
		return nil, fmt.Errorf("transação falhou")
	}

	// Lê o evento PacoteComprado para obter os IDs das cartas
	// Por enquanto, retorna vazio - o servidor pode consultar o inventário depois
	return []*big.Int{}, nil
}

// ObterInventario retorna o inventário de cartas de um jogador
func (m *Manager) ObterInventario(jogadorAddress common.Address) ([]tipos.Carta, error) {
	// Prepara a chamada à função obterInventario
	data, err := m.contractABI.Pack("obterInventario", jogadorAddress)
	if err != nil {
		return nil, fmt.Errorf("erro ao preparar chamada: %v", err)
	}

	// Faz a chamada ao contrato
	msg := ethereum.CallMsg{
		To:   &m.contractAddress,
		Data: data,
	}

	result, err := m.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao chamar contrato: %v", err)
	}

	// Desempacota o resultado (array de uint256)
	var ids []*big.Int
	err = m.contractABI.UnpackIntoInterface(&ids, "obterInventario", result)
	if err != nil {
		return nil, fmt.Errorf("erro ao desempacotar resultado: %v", err)
	}

	log.Printf("[BLOCKCHAIN_DEBUG] ObterInventario(%s) retornou %d IDs: %v", jogadorAddress.Hex(), len(ids), ids)

	// Para cada ID, obtém os dados da carta
	cartas := make([]tipos.Carta, 0, len(ids))
	for _, id := range ids {
		carta, err := m.ObterCarta(id)
		if err == nil {
			cartas = append(cartas, carta)
		} else {
			log.Printf("[BLOCKCHAIN_DEBUG] Erro ao obter carta %s: %v", id, err)
		}
	}

	return cartas, nil
}

// ObterCarta retorna os dados de uma carta específica
func (m *Manager) ObterCarta(cartaID *big.Int) (tipos.Carta, error) {
	// Prepara a chamada à função obterCarta
	data, err := m.contractABI.Pack("obterCarta", cartaID)
	if err != nil {
		return tipos.Carta{}, fmt.Errorf("erro ao preparar chamada: %v", err)
	}

	// Faz a chamada ao contrato
	msg := ethereum.CallMsg{
		To:   &m.contractAddress,
		Data: data,
	}

	result, err := m.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return tipos.Carta{}, fmt.Errorf("erro ao chamar contrato: %v", err)
	}

	// Desempacota o resultado (struct Carta)
	var cartaData struct {
		Id        *big.Int
		Nome      string
		Naipe     string
		Valor     *big.Int
		Raridade  string
		Timestamp *big.Int
	}

	err = m.contractABI.UnpackIntoInterface(&cartaData, "obterCarta", result)
	if err != nil {
		return tipos.Carta{}, fmt.Errorf("erro ao desempacotar resultado: %v", err)
	}

	// Converte para tipos.Carta
	return tipos.Carta{
		ID:       cartaData.Id.String(),
		Nome:     cartaData.Nome,
		Naipe:    cartaData.Naipe,
		Valor:    int(cartaData.Valor.Int64()),
		Raridade: cartaData.Raridade,
	}, nil
}

// CriarPropostaTroca cria uma proposta de troca de cartas na blockchain
func (m *Manager) CriarPropostaTroca(jogador1, jogador2 common.Address, carta1, carta2 *big.Int) (*big.Int, error) {
	// Prepara a chamada à função criarPropostaTroca
	data, err := m.contractABI.Pack("criarPropostaTroca", jogador2, carta1, carta2)
	if err != nil {
		return nil, fmt.Errorf("erro ao preparar chamada: %v", err)
	}

	// Envia a transação (jogador1 é quem cria a proposta)
	tx, err := m.enviarTransacao(jogador1, data, big.NewInt(0))
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar transação: %v", err)
	}

	// Aguarda confirmação
	receipt, err := m.aguardarConfirmacao(tx.Hash())
	if err != nil {
		return nil, fmt.Errorf("erro ao aguardar confirmação: %v", err)
	}

	if receipt.Status == 0 {
		return nil, fmt.Errorf("transação falhou")
	}

	// Lê o evento PropostaTrocaCriada para obter o ID
	for _, vLog := range receipt.Logs {
		if vLog.Address == m.contractAddress && len(vLog.Topics) >= 4 {
			// O ID da proposta é o primeiro argumento indexado (Topic[1])
			propostaID := new(big.Int).SetBytes(vLog.Topics[1].Bytes())
			log.Printf("[BLOCKCHAIN_DEBUG] Proposta criada detectada no log: %s", propostaID.String())
			return propostaID, nil
		}
	}

	return nil, fmt.Errorf("id da proposta não encontrado nos logs")
}

// AceitarPropostaTroca aceita uma proposta de troca
func (m *Manager) AceitarPropostaTroca(jogador2 common.Address, propostaID *big.Int) error {
	// Prepara a chamada à função aceitarPropostaTroca
	data, err := m.contractABI.Pack("aceitarPropostaTroca", propostaID)
	if err != nil {
		return fmt.Errorf("erro ao preparar chamada: %v", err)
	}

	// Envia a transação (jogador2 é quem aceita)
	tx, err := m.enviarTransacao(jogador2, data, big.NewInt(0))
	if err != nil {
		return fmt.Errorf("erro ao enviar transação: %v", err)
	}

	// Aguarda confirmação
	receipt, err := m.aguardarConfirmacao(tx.Hash())
	if err != nil {
		return fmt.Errorf("erro ao aguardar confirmação: %v", err)
	}

	if receipt.Status == 0 {
		return fmt.Errorf("transação falhou")
	}

	return nil
}

// RegistrarPartida registra o resultado de uma partida na blockchain
func (m *Manager) RegistrarPartida(jogador1, jogador2, vencedor common.Address) error {
	// Prepara a chamada à função registrarPartida
	data, err := m.contractABI.Pack("registrarPartida", jogador1, jogador2, vencedor)
	if err != nil {
		return fmt.Errorf("erro ao preparar chamada: %v", err)
	}

	// Envia a transação usando a conta do servidor
	tx, err := m.enviarTransacao(m.serverAccount, data, big.NewInt(0))
	if err != nil {
		return fmt.Errorf("erro ao enviar transação: %v", err)
	}

	// Aguarda confirmação
	receipt, err := m.aguardarConfirmacao(tx.Hash())
	if err != nil {
		return fmt.Errorf("erro ao aguardar confirmação: %v", err)
	}

	if receipt.Status == 0 {
		return fmt.Errorf("transação falhou")
	}

	return nil
}

// VerificarPropriedadeCarta verifica se um jogador possui uma carta específica
func (m *Manager) VerificarPropriedadeCarta(jogador common.Address, cartaID *big.Int) (bool, error) {
	// Prepara a chamada à função proprietario
	data, err := m.contractABI.Pack("proprietario", cartaID)
	if err != nil {
		return false, fmt.Errorf("erro ao preparar chamada: %v", err)
	}

	// Faz a chamada ao contrato
	msg := ethereum.CallMsg{
		To:   &m.contractAddress,
		Data: data,
	}

	result, err := m.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return false, fmt.Errorf("erro ao chamar contrato: %v", err)
	}

	// Desempacota o resultado (address)
	var proprietario common.Address
	err = m.contractABI.UnpackIntoInterface(&proprietario, "proprietario", result)
	if err != nil {
		return false, fmt.Errorf("erro ao desempacotar resultado: %v", err)
	}

	return proprietario == jogador, nil
}

// enviarTransacao envia uma transação para a blockchain
func (m *Manager) enviarTransacao(from common.Address, data []byte, valor *big.Int) (*types.Transaction, error) {
	// Obtém nonce
	nonce, err := m.client.PendingNonceAt(context.Background(), from)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter nonce: %v", err)
	}

	// Obtém gas price
	gasPrice, err := m.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("erro ao obter gas price: %v", err)
	}

	// Estima gas
	var gasToUse uint64 = m.gasLimit
	msg := ethereum.CallMsg{
		From:  from,
		To:    &m.contractAddress,
		Value: valor,
		Data:  data,
	}
	gasEstimate, err := m.client.EstimateGas(context.Background(), msg)
	if err == nil {
		gasToUse = gasEstimate * 120 / 100
		if gasToUse > m.gasLimit {
			gasToUse = m.gasLimit
		}
		if gasToUse < 21000 {
			gasToUse = 21000
		}
	}

	// Obtém chain ID
	chainID, err := m.client.NetworkID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("erro ao obter chain ID: %v", err)
	}

	// Cria transação
	tx := types.NewTransaction(nonce, m.contractAddress, valor, gasToUse, gasPrice, data)

	// Se for a conta do servidor, assina com a chave privada
	if from == m.serverAccount && m.serverKey != nil {
		txAssinada, err := types.SignTx(tx, types.NewEIP155Signer(chainID), m.serverKey.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("erro ao assinar transação: %v", err)
		}

		err = m.client.SendTransaction(context.Background(), txAssinada)
		if err != nil {
			return nil, fmt.Errorf("erro ao enviar transação: %v", err)
		}

		return txAssinada, nil
	}

	// Para outras contas, usa personal_sendTransaction (requer que a conta esteja desbloqueada)
	if m.rpcClient != nil {
		// Desbloqueia a conta
		var unlockResult bool
		err := m.rpcClient.Call(&unlockResult, "personal_unlockAccount", from, m.serverPassword, 0)
		if err != nil || !unlockResult {
			return nil, fmt.Errorf("erro ao desbloquear conta: %v", err)
		}

		// Prepara parâmetros
		txParams := map[string]interface{}{
			"from":     from.Hex(),
			"to":       m.contractAddress.Hex(),
			"value":    fmt.Sprintf("0x%x", valor),
			"data":     fmt.Sprintf("0x%x", data),
			"gas":      fmt.Sprintf("0x%x", gasToUse),
			"gasPrice": fmt.Sprintf("0x%x", gasPrice),
		}

		// Envia via personal_sendTransaction
		var txHashStr string
		err = m.rpcClient.Call(&txHashStr, "personal_sendTransaction", txParams, m.serverPassword)
		if err != nil {
			return nil, fmt.Errorf("erro ao enviar transação via RPC: %v", err)
		}

		txHash := common.HexToHash(txHashStr)
		// Aguarda um pouco e tenta obter a transação
		time.Sleep(1 * time.Second)
		tx, _, err := m.client.TransactionByHash(context.Background(), txHash)
		if err != nil {
			// Cria uma transação dummy com o hash correto
			tx = types.NewTransaction(nonce, m.contractAddress, valor, gasToUse, gasPrice, data)
		}

		return tx, nil
	}

	return nil, fmt.Errorf("não foi possível enviar transação: conta não é do servidor e RPC não disponível")
}

// aguardarConfirmacao aguarda a confirmação de uma transação
func (m *Manager) aguardarConfirmacao(txHash common.Hash) (*types.Receipt, error) {
	timeout := 60 * time.Second
	startTime := time.Now()

	for {
		if time.Since(startTime) > timeout {
			return nil, fmt.Errorf("timeout aguardando confirmação")
		}

		receipt, err := m.client.TransactionReceipt(context.Background(), txHash)
		if err == nil {
			return receipt, nil
		}

		if err != ethereum.NotFound {
			time.Sleep(1 * time.Second)
			continue
		}

		time.Sleep(2 * time.Second)
	}
}

// GetContractAddress retorna o endereço do contrato
func (m *Manager) GetContractAddress() common.Address {
	return m.contractAddress
}

