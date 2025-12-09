package blockchain

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"strings"
	"time"

	"jogodistribuido/servidor/tipos"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	// Endereço RPC do nó Geth (pode ser configurado via variável de ambiente)
	DefaultRPCURL = "http://geth-node:8545"
	// Gas limit para transações
	DefaultGasLimit = uint64(80000000)
)

// Manager gerencia a interação com a blockchain
type Manager struct {
	client          *ethclient.Client
	rpcClient       *rpc.Client
	contractAddress common.Address
	contractABI     abi.ABI
	serverAccount   common.Address
	serverKey       *keystore.Key
	serverPassword  string
	keystorePath    string
	gasLimit        uint64
}

// NewManager cria um novo gerenciador de blockchain
// TÓPICO 1 - Essa função estabelece a conexão com o nó da blockchain (Geth), integrando a aplicação servidora à arquitetura descentralizada.
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
	var abiPathUsed string
	for _, abiPath := range abiPaths {
		abiBytes, err = ioutil.ReadFile(abiPath)
		if err == nil {
			abiPathUsed = abiPath
			log.Printf("[BLOCKCHAIN] ABI carregado de: %s", abiPath)
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

	// Log de debug: verifica se registrarPartidaAdmin existe no ABI
	if _, ok := parsedABI.Methods["registrarPartidaAdmin"]; ok {
		log.Printf("[BLOCKCHAIN] ✓ Função registrarPartidaAdmin encontrada no ABI (carregado de: %s)", abiPathUsed)
	} else {
		log.Printf("[BLOCKCHAIN] ⚠ Função registrarPartidaAdmin NÃO encontrada no ABI (carregado de: %s)", abiPathUsed)
	}

	// Carrega a conta do servidor (para registrar partidas)
	var serverAccount common.Address
	var serverKey *keystore.Key

	// CORREÇÃO: Detecta automaticamente o owner do contrato (não hardcode!)
	// Lê a variável pública 'owner' do contrato
	log.Printf("[BLOCKCHAIN] Detectando owner do contrato automaticamente (contrato: %s)...", contractAddress.Hex())
	ownerData, err := parsedABI.Pack("owner")
	if err != nil {
		log.Printf("[BLOCKCHAIN] Erro ao preparar chamada para owner: %v", err)
	} else {
		msg := ethereum.CallMsg{
			To:   &contractAddress,
			Data: ownerData,
		}
		result, err := client.CallContract(context.Background(), msg, nil)
		if err != nil {
			log.Printf("[BLOCKCHAIN] Erro ao chamar owner do contrato: %v", err)
			log.Printf("[BLOCKCHAIN] Verifique se o contrato está deployado e o endereço está correto")
		} else {
			var ownerAddr common.Address
			err = parsedABI.UnpackIntoInterface(&ownerAddr, "owner", result)
			if err != nil {
				log.Printf("[BLOCKCHAIN] Erro ao desempacotar owner: %v", err)
			} else {
				if ownerAddr == (common.Address{}) {
					log.Printf("[BLOCKCHAIN] ⚠ Owner retornado é zero - contrato pode não estar deployado ou endereço incorreto")
				} else {
					serverAccount = ownerAddr
					log.Printf("[BLOCKCHAIN] ✓ Owner do contrato detectado automaticamente: %s", serverAccount.Hex())
				}
			}
		}
	}

	// Se não conseguiu detectar o owner, tenta carregar do keystore
	if serverAccount == (common.Address{}) {
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
						log.Printf("[BLOCKCHAIN] Conta do servidor carregada do keystore: %s", serverAccount.Hex())
					}
				}
			}
		}
	}

	// Se ainda não tem conta, loga aviso e tenta usar a primeira conta do Geth como fallback
	if serverAccount == (common.Address{}) {
		log.Printf("[BLOCKCHAIN] ⚠ Aviso: Não foi possível detectar conta do servidor automaticamente")
		log.Printf("[BLOCKCHAIN] Tentando obter primeira conta do Geth como fallback...")

		// Tenta obter a primeira conta do Geth
		if rpcClient != nil {
			var accounts []common.Address
			err := rpcClient.Call(&accounts, "eth_accounts")
			if err == nil && len(accounts) > 0 {
				serverAccount = accounts[0]
				log.Printf("[BLOCKCHAIN] ✓ Usando primeira conta do Geth como fallback: %s", serverAccount.Hex())
			} else {
				log.Printf("[BLOCKCHAIN] ⚠ Não foi possível obter contas do Geth")
				log.Printf("[BLOCKCHAIN] Configure keystorePath e serverPassword ou garanta que o owner do contrato está desbloqueado")
			}
		}
	} else {
		log.Printf("[BLOCKCHAIN] ✓ Conta do servidor configurada: %s", serverAccount.Hex())
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
// TÓPICO 7 - Ao ler os logs de eventos da transação ('CartaCriada'), esta função demonstra a capacidade do sistema de auditar e recuperar informações críticas diretamente do ledger público.
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

	// Lê os eventos CartaCriada para obter os IDs das cartas
	// É mais confiável ler CartaCriada (um evento por carta) do que PacoteComprado (array dinâmico)
	eventSignature := m.contractABI.Events["CartaCriada"]
	if eventSignature.ID == (common.Hash{}) {
		return nil, fmt.Errorf("evento CartaCriada não encontrado no ABI")
	}

	var tokenIds []*big.Int
	compradorEncontrado := false

	for _, vLog := range receipt.Logs {
		if vLog.Address != m.contractAddress {
			continue
		}

		// Verifica se é o evento CartaCriada
		// CartaCriada(uint256 indexed tokenId, address indexed proprietario, string nome, string raridade, uint256 valor)
		if len(vLog.Topics) >= 3 && vLog.Topics[0] == eventSignature.ID {
			// Topic[1] = tokenId (indexed)
			// Topic[2] = proprietario (indexed)
			tokenId := new(big.Int).SetBytes(vLog.Topics[1].Bytes())
			proprietario := common.BytesToAddress(vLog.Topics[2].Bytes())

			// Verifica se o proprietário é o jogador que comprou
			if proprietario == jogadorAddress {
				compradorEncontrado = true
				tokenIds = append(tokenIds, tokenId)
			}
		}
	}

	if len(tokenIds) > 0 {
		log.Printf("[BLOCKCHAIN_DEBUG] ✓ Encontrados %d tokenIds do evento CartaCriada: %v", len(tokenIds), tokenIds)
		return tokenIds, nil
	}

	if !compradorEncontrado {
		log.Printf("[BLOCKCHAIN_DEBUG] ⚠ Nenhum evento CartaCriada encontrado para o comprador %s", jogadorAddress.Hex())
		log.Printf("[BLOCKCHAIN_DEBUG] ⚠ Total de logs no recibo: %d", len(receipt.Logs))
	}

	// Fallback: se não conseguiu ler dos eventos, retorna vazio
	// O sistema vai buscar o inventário completo depois (já implementado)
	log.Printf("[BLOCKCHAIN_DEBUG] Usando fallback: sistema buscará inventário completo da blockchain")
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
// Usa o mapeamento público 'cartas' que retorna campos individuais (mais confiável)
func (m *Manager) ObterCarta(cartaID *big.Int) (tipos.Carta, error) {
	// Usa o mapeamento público 'cartas' em vez de 'obterCarta'
	// O mapeamento retorna campos individuais, não uma struct
	data, err := m.contractABI.Pack("cartas", cartaID)
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

	// O mapeamento público retorna os campos individualmente (não como struct)
	// Saída: id, nome, naipe, valor, raridade, timestamp
	values, err := m.contractABI.Unpack("cartas", result)
	if err != nil {
		return tipos.Carta{}, fmt.Errorf("erro ao desempacotar resultado: %v", err)
	}

	if len(values) < 6 {
		return tipos.Carta{}, fmt.Errorf("resposta incompleta: esperado 6 campos, recebido %d", len(values))
	}

	// Extrai os valores individuais
	id, ok := values[0].(*big.Int)
	if !ok {
		return tipos.Carta{}, fmt.Errorf("tipo inválido para id: %T", values[0])
	}

	nome, ok := values[1].(string)
	if !ok {
		return tipos.Carta{}, fmt.Errorf("tipo inválido para nome: %T", values[1])
	}

	naipe, ok := values[2].(string)
	if !ok {
		return tipos.Carta{}, fmt.Errorf("tipo inválido para naipe: %T", values[2])
	}

	valor, ok := values[3].(*big.Int)
	if !ok {
		return tipos.Carta{}, fmt.Errorf("tipo inválido para valor: %T", values[3])
	}

	raridade, ok := values[4].(string)
	if !ok {
		return tipos.Carta{}, fmt.Errorf("tipo inválido para raridade: %T", values[4])
	}

	// Converte para tipos.Carta
	return tipos.Carta{
		ID:       id.String(),
		Nome:     nome,
		Naipe:    naipe,
		Valor:    int(valor.Int64()),
		Raridade: raridade,
	}, nil
}

// CriarPropostaTroca cria uma proposta de troca de cartas na blockchain
// NOTA: Esta função está DEPRECATED - use RegistrarTrocaAdmin para trocas coordenadas pelo servidor
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

// RegistrarTrocaAdmin registra uma troca de cartas na blockchain usando a conta do servidor (admin)
// Esta função permite ao servidor registrar trocas em nome dos jogadores, passando os endereços corretos
// que serão registrados na blockchain para auditabilidade
// RegistrarTrocaAdmin registra uma troca de cartas na blockchain usando a conta do servidor (admin)
// Esta função permite ao servidor registrar trocas em nome dos jogadores, passando os endereços corretos
// que serão registrados na blockchain para auditabilidade
func (m *Manager) RegistrarTrocaAdmin(jogador1, jogador2 common.Address, carta1, carta2 *big.Int) (*big.Int, error) {
	log.Printf("[BLOCKCHAIN] RegistrarTrocaAdmin: jogador1=%s, jogador2=%s, carta1=%s, carta2=%s",
		jogador1.Hex(), jogador2.Hex(), carta1.String(), carta2.String())

	// Prepara a chamada à função registrarTrocaAdmin
	// Esta função aceita os 4 parâmetros: jogador1, jogador2, cartaJogador1, cartaJogador2
	data, err := m.contractABI.Pack("registrarTrocaAdmin", jogador1, jogador2, carta1, carta2)
	if err != nil {
		return nil, fmt.Errorf("erro ao preparar chamada registrarTrocaAdmin: %v", err)
	}

	// Envia a transação usando a conta do servidor (que é o owner do contrato)
	// Isso é crucial: usamos serverAccount para que msg.sender seja o owner
	tx, err := m.enviarTransacao(m.serverAccount, data, big.NewInt(0))
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar transação: %v", err)
	}

	log.Printf("[BLOCKCHAIN] Transação enviada: %s", tx.Hash().Hex())

	// Aguarda confirmação
	receipt, err := m.aguardarConfirmacao(tx.Hash())
	if err != nil {
		return nil, fmt.Errorf("erro ao aguardar confirmação: %v", err)
	}

	if receipt.Status == 0 {
		return nil, fmt.Errorf("transação falhou (status=0)")
	}

	log.Printf("[BLOCKCHAIN] Transação confirmada! Status=%d, Logs=%d", receipt.Status, len(receipt.Logs))

	// Lê o evento TrocaExecutada para obter o ID da proposta
	for _, vLog := range receipt.Logs {
		if vLog.Address == m.contractAddress && len(vLog.Topics) >= 4 {
			// O ID da proposta é o primeiro argumento indexado (Topic[1])
			propostaID := new(big.Int).SetBytes(vLog.Topics[1].Bytes())
			log.Printf("[BLOCKCHAIN] ✓ Troca registrada com ID: %s", propostaID.String())
			log.Printf("[BLOCKCHAIN] Jogador1 (ofertante): %s", jogador1.Hex())
			log.Printf("[BLOCKCHAIN] Jogador2 (desejado): %s", jogador2.Hex())
			return propostaID, nil
		}
	}

	// Se não encontrou o evento, ainda assim a troca foi bem sucedida
	log.Printf("[BLOCKCHAIN] ✓ Troca registrada (ID não encontrado nos logs)")
	return big.NewInt(0), nil
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
// Tenta usar registrarPartidaAdmin (onlyOwner) primeiro, se não existir usa registrarPartida
// TÓPICO 4 - Esta função interage com o contrato para registrar permanentemente o resultado de uma partida, assegurando que o histórico de vitórias seja mantido de forma descentralizada e imutável.
func (m *Manager) RegistrarPartida(jogador1, jogador2, vencedor common.Address) error {
	// Tenta usar registrarPartidaAdmin primeiro (onlyOwner - funciona igual às trocas)
	data, err := m.contractABI.Pack("registrarPartidaAdmin", jogador1, jogador2, vencedor)
	if err != nil {
		// Se a função não existir no ABI, tenta usar a função antiga registrarPartida
		// (mas isso requer que jogador1 seja msg.sender, então não funciona com servidor)
		// Por enquanto, apenas ignora silenciosamente - o contrato precisa ser atualizado
		log.Printf("[BLOCKCHAIN] registrarPartidaAdmin não encontrado no ABI (contrato precisa ser atualizado)")
		return fmt.Errorf("função registrarPartidaAdmin não encontrada no ABI")
	}

	log.Printf("[BLOCKCHAIN] Registrando partida (admin): jogador1=%s, jogador2=%s, vencedor=%s",
		jogador1.Hex(), jogador2.Hex(), vencedor.Hex())
	log.Printf("[BLOCKCHAIN] Dados da transação (primeiros 20 bytes): %s", common.Bytes2Hex(data[:min(20, len(data))]))

	// Envia a transação usando a conta do servidor (que é o owner)
	tx, err := m.enviarTransacao(m.serverAccount, data, big.NewInt(0))
	if err != nil {
		return fmt.Errorf("erro ao enviar transação: %v", err)
	}

	log.Printf("[BLOCKCHAIN] Transação enviada: %s", tx.Hash().Hex())

	// Aguarda confirmação
	receipt, err := m.aguardarConfirmacao(tx.Hash())
	if err != nil {
		return fmt.Errorf("erro ao aguardar confirmação: %v", err)
	}

	if receipt.Status == 0 {
		return fmt.Errorf("transação falhou (status=0)")
	}

	log.Printf("[BLOCKCHAIN] ✓ Partida registrada com sucesso! Block: %d", receipt.BlockNumber.Uint64())
	log.Printf("[BLOCKCHAIN] DEBUG: Receipt tem %d logs (eventos)", len(receipt.Logs))
	if len(receipt.Logs) == 0 {
		log.Printf("[BLOCKCHAIN] ⚠ AVISO: Transação bem-sucedida mas nenhum evento foi emitido! O contrato pode não ter a função registrarPartidaAdmin")
	} else {
		for i, logEntry := range receipt.Logs {
			log.Printf("[BLOCKCHAIN] DEBUG: Log[%d]: %d tópicos, endereço=%s", i, len(logEntry.Topics), logEntry.Address.Hex())
		}
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

// TÓPICO 2 - A comunicação com a blockchain ocorre aqui via RPC, enviando transações que serão propagadas e validadas pelos nós da rede, garantindo a interação distribuída.
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

	// Se for a conta do servidor, tenta assinar com a chave privada primeiro
	if from == m.serverAccount {
		if m.serverKey != nil {
			// Tem chave privada: assina localmente
			txAssinada, err := types.SignTx(tx, types.NewEIP155Signer(chainID), m.serverKey.PrivateKey)
			if err != nil {
				return nil, fmt.Errorf("erro ao assinar transação: %v", err)
			}

			err = m.client.SendTransaction(context.Background(), txAssinada)
			if err != nil {
				return nil, fmt.Errorf("erro ao enviar transação: %v", err)
			}

			return txAssinada, nil
		} else {
			// Não tem chave privada: tenta usar eth_sendTransaction (conta deve estar desbloqueada)
			log.Printf("[BLOCKCHAIN] Conta do servidor sem chave privada, tentando eth_sendTransaction (conta deve estar desbloqueada no Geth)")
		}
	}

	// Para outras contas ou servidor sem chave privada, usa eth_sendTransaction (requer que a conta esteja desbloqueada no Geth via --unlock)
	if m.rpcClient != nil {
		// Prepara parâmetros para eth_sendTransaction
		txParams := map[string]interface{}{
			"from":     from.Hex(),
			"to":       m.contractAddress.Hex(),
			"value":    fmt.Sprintf("0x%x", valor),
			"data":     fmt.Sprintf("0x%x", data),
			"gas":      fmt.Sprintf("0x%x", gasToUse),
			"gasPrice": fmt.Sprintf("0x%x", gasPrice),
		}

		// Envia via eth_sendTransaction (a conta já deve estar desbloqueada no Geth)
		var txHashStr string
		err = m.rpcClient.Call(&txHashStr, "eth_sendTransaction", txParams)
		if err != nil {
			// Erro mais descritivo para "unknown account"
			if strings.Contains(err.Error(), "unknown account") {
				return nil, fmt.Errorf("conta %s não está desbloqueada no Geth. Configure --unlock no docker-compose ou forneça keystorePath e serverPassword", from.Hex())
			}
			return nil, fmt.Errorf("erro ao enviar transação via eth_sendTransaction: %v", err)
		}

		log.Printf("[BLOCKCHAIN] Transação enviada via eth_sendTransaction: %s", txHashStr)

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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetContractAddress retorna o endereço do contrato
func (m *Manager) GetContractAddress() common.Address {
	return m.contractAddress
}
