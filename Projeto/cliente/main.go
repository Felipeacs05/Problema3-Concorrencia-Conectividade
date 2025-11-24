// ===================== BAREMA ITEM 3: APLICAÇÃO CLIENTE =====================
// Este é o componente cliente do jogo de cartas multiplayer baseado em blockchain.
// O cliente gerencia: interface CLI, comunicação com blockchain via RPC,
// gerenciamento de carteira (keystore), e interação com smart contracts.

package main

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

// ===================== Constantes e Variáveis Globais =====================

const (
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Endereço RPC do nó Geth local
	// Usa IPC (Inter-Process Communication) para melhor performance
	rpcURL = "http://127.0.0.1:8545"

	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Caminho para keystore local
	// Usa o mesmo keystore que o Geth usa
	keystorePath = "../data/keystore"

	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Gas limit para transações
	// Ajustado para 90M (menor que o gas limit do bloco ~95M)
	gasLimit = uint64(90000000)
)

// Representa uma carta do jogo
type Carta struct {
	ID        *big.Int
	Nome      string
	Naipe     string
	Valor     *big.Int
	Raridade  string
	Timestamp *big.Int
}

var (
	client           *ethclient.Client
	rpcClient        *rpc.Client
	contaAtual       common.Address
	chavePrivada     *ecdsa.PrivateKey
	enderecoContrato common.Address
	contractABI      abi.ABI
	senhaConta       string // Guarda a senha da conta atual
	abiPath          = "../contracts/GameEconomy.abi" // Caminho para o arquivo ABI
)

// ===================== Função Principal =====================

func main() {
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Inicialização
	color.Cyan("Iniciando cliente do Jogo de Cartas Blockchain...\n")

	// Configura conexão com blockchain
	err := conectarBlockchain()
	if err != nil {
		color.Red("Erro ao conectar à blockchain: %v\n", err)
		os.Exit(1)
	}

	// Gerencia carteira (criar ou carregar)
	err = gerenciarCarteira()
	if err != nil {
		color.Red("Erro ao gerenciar carteira: %v\n", err)
		os.Exit(1)
	}

	// Loop principal do menu
	for {
		exibirMenu()
	}
}

// ===================== Funções de Conexão e Carteira =====================

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Conecta ao nó Geth
func conectarBlockchain() error {
	var err error
	client, err = ethclient.Dial(rpcURL)
	if err != nil {
		return fmt.Errorf("falha ao conectar ao cliente Ethereum: %v", err)
	}
	
	// Conecta também via RPC para usar personal_sendTransaction se necessário
	rpcClient, err = rpc.Dial(rpcURL)
	if err != nil {
		// Não é crítico, apenas loga aviso
		color.Yellow("⚠ Aviso: Não foi possível conectar via RPC (algumas funcionalidades podem não funcionar)\n")
	}
	
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return fmt.Errorf("falha ao obter ID da rede: %v", err)
	}
	
	color.Green("✓ Conectado à blockchain (Network ID: %s)\n", chainID.String())
	return nil
}

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Cria ou carrega uma conta (carteira)
func gerenciarCarteira() error {
	// Cria diretório de keystore se não existir
	if err := os.MkdirAll(keystorePath, 0700); err != nil {
		return fmt.Errorf("erro ao criar diretório keystore: %v", err)
	}

	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Lista contas existentes
	files, err := ioutil.ReadDir(keystorePath)
	if err != nil {
		return fmt.Errorf("erro ao ler keystore: %v", err)
	}

	if len(files) == 0 {
		// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Cria nova conta se não existir nenhuma
		return criarNovaConta()
	}

	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Permite escolher conta existente ou criar nova
	prompt := promptui.Select{
		Label: "Escolha uma opção",
		Items: []string{"Usar conta existente", "Criar nova conta"},
	}

	_, resultado, err := prompt.Run()
	if err != nil {
		return err
	}

	if resultado == "Criar nova conta" {
		return criarNovaConta()
	}

	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Lista contas disponíveis
	contas := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			contas = append(contas, file.Name())
		}
	}

	promptConta := promptui.Select{
		Label: "Selecione uma conta",
		Items: contas,
	}

	_, arquivoSelecionado, err := promptConta.Run()
	if err != nil {
		return err
	}

	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Carrega a conta selecionada
	return carregarConta(filepath.Join(keystorePath, arquivoSelecionado))
}

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Cria uma nova conta Ethereum
func criarNovaConta() error {
	prompt := promptui.Prompt{
		Label: "Digite uma senha para sua nova conta",
		Mask:  '*',
	}

	senha, err := prompt.Run()
	if err != nil {
		return err
	}

	ks := keystore.NewKeyStore(keystorePath, keystore.StandardScryptN, keystore.StandardScryptP)
	conta, err := ks.NewAccount(senha)
	if err != nil {
		return fmt.Errorf("erro ao criar conta: %v", err)
	}

	color.Green("✓ Nova conta criada: %s\n", conta.Address.Hex())

	// Carrega a conta recém-criada
	// Precisamos encontrar o arquivo correspondente
	return carregarContaPeloEndereco(conta.Address, senha)
}

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Carrega conta a partir do endereço
func carregarContaPeloEndereco(endereco common.Address, senha string) error {
	// Guarda a senha para uso posterior
	senhaConta = senha
	ks := keystore.NewKeyStore(keystorePath, keystore.StandardScryptN, keystore.StandardScryptP)

	if !ks.HasAddress(endereco) {
		return fmt.Errorf("conta não encontrada no keystore")
	}

	// Lista todas as contas no keystore
	contas := ks.Accounts()
	var contaEncontrada *accounts.Account
	for i := range contas {
		if contas[i].Address == endereco {
			contaEncontrada = &contas[i]
			break
		}
	}

	if contaEncontrada == nil {
		return fmt.Errorf("conta não encontrada no keystore")
	}

	// Importa a chave JSON
	jsonBytes, err := ioutil.ReadFile(contaEncontrada.URL.Path)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo da conta: %v", err)
	}

	key, err := keystore.DecryptKey(jsonBytes, senha)
	if err != nil {
		return fmt.Errorf("senha incorreta: %v", err)
	}

	chavePrivada = key.PrivateKey
	contaAtual = key.Address

	color.Green("✓ Conta carregada: %s\n", contaAtual.Hex())

	// Verifica saldo
	saldo, err := client.BalanceAt(context.Background(), contaAtual, nil)
	if err == nil {
		saldoEth := new(big.Float).Quo(new(big.Float).SetInt(saldo), big.NewFloat(1e18))
		color.Yellow("Saldo: %f ETH\n", saldoEth)
	}

	return nil
}

	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Carrega conta de um arquivo específico
func carregarConta(caminhoArquivo string) error {
	prompt := promptui.Prompt{
		Label: "Digite a senha da conta",
		Mask:  '*',
	}
	
	senha, err := prompt.Run()
	if err != nil {
		return err
	}
	
	// Guarda a senha para uso posterior
	senhaConta = senha

	jsonBytes, err := ioutil.ReadFile(caminhoArquivo)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo: %v", err)
	}

	key, err := keystore.DecryptKey(jsonBytes, senha)
	if err != nil {
		return fmt.Errorf("senha incorreta: %v", err)
	}

	chavePrivada = key.PrivateKey
	contaAtual = key.Address

	color.Green("✓ Conta carregada: %s\n", contaAtual.Hex())

	// Verifica saldo
	saldo, err := client.BalanceAt(context.Background(), contaAtual, nil)
	if err == nil {
		saldoEth := new(big.Float).Quo(new(big.Float).SetInt(saldo), big.NewFloat(1e18))
		color.Yellow("Saldo: %f ETH\n", saldoEth)
	}

	return nil
}

// ===================== Funções de Transação =====================

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Faz deploy do contrato (opcional, para testes)
func fazerDeployContrato() error {
	color.Yellow("Iniciando deploy do contrato GameEconomy...\n")

	// Lê o arquivo bytecode
	bytecodeRaw, err := ioutil.ReadFile("../contracts/GameEconomy.bin")
	if err != nil {
		return fmt.Errorf("erro ao ler bytecode: %v", err)
	}

	// Remove espaços, quebras de linha e caracteres inválidos
	bytecodeStr := strings.TrimSpace(string(bytecodeRaw))
	bytecodeStr = strings.ReplaceAll(bytecodeStr, "\n", "")
	bytecodeStr = strings.ReplaceAll(bytecodeStr, "\r", "")
	bytecodeStr = strings.ReplaceAll(bytecodeStr, " ", "")
	
	// Remove prefixo "0x" se existir
	if strings.HasPrefix(bytecodeStr, "0x") || strings.HasPrefix(bytecodeStr, "0X") {
		bytecodeStr = bytecodeStr[2:]
	}

	// Valida que não está vazio
	if len(bytecodeStr) == 0 {
		return fmt.Errorf("bytecode vazio após limpeza")
	}

	// Converte hex para bytes
	bytecodeBytes := common.FromHex(bytecodeStr)
	
	// Valida que a conversão foi bem-sucedida
	if len(bytecodeBytes) == 0 {
		return fmt.Errorf("erro ao converter bytecode hex para bytes (tamanho: %d caracteres hex)", len(bytecodeStr))
	}
	
	color.Cyan("Bytecode carregado: %d bytes\n", len(bytecodeBytes))

	// Estima gas necessário para o deploy
	color.Yellow("Estimando gas necessário...\n")
	msg := ethereum.CallMsg{
		From:  contaAtual,
		Value: big.NewInt(0),
		Data:  bytecodeBytes,
	}
	gasEstimate, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		color.Yellow("⚠ Não foi possível estimar gas: %v\n", err)
		color.Yellow("   Usando gas limit padrão: %d\n", gasLimit)
	} else {
		color.Cyan("Gas estimado: %d\n", gasEstimate)
		// Usa 120% do estimado como margem de segurança
		if gasEstimate*120/100 > gasLimit {
			color.Yellow("⚠ Gas estimado (%d) excede o limite configurado (%d)\n", gasEstimate*120/100, gasLimit)
		}
	}

	// Prepara transação de criação de contrato (to = nil)
	tx, err := enviarTransacao(bytecodeBytes, big.NewInt(0))
	if err != nil {
		return err
	}

	// Verifica se a transação foi bem-sucedida antes de calcular o endereço
	receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return fmt.Errorf("erro ao obter receipt da transação: %v", err)
	}

	if receipt.Status == 0 {
		// Tenta obter mais informações sobre o erro
		color.Red("✗ Transação falhou (status: 0)\n")
		color.Yellow("Hash da transação: %s\n", tx.Hash().Hex())
		color.Yellow("Gas usado: %d / %d\n", receipt.GasUsed, tx.Gas())
		if len(receipt.Logs) == 0 {
			color.Yellow("Nenhum log gerado - possível erro de execução do contrato\n")
		}
		return fmt.Errorf("transação falhou - contrato não foi deployado. Verifique o bytecode e o gas limit")
	}

	// Calcula o endereço do contrato a partir do endereço do remetente e nonce
	enderecoContrato = crypto.CreateAddress(contaAtual, tx.Nonce())
	color.Green("✓ Contrato deployado com sucesso: %s\n", enderecoContrato.Hex())

	return nil
}

	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Envia uma transação para a blockchain
func enviarTransacao(data []byte, valor *big.Int) (*types.Transaction, error) {
	// Tenta primeiro com SendTransaction (método normal)
	txAssinada, err := enviarTransacaoNormal(data, valor)
	if err == nil {
		color.Cyan("Transação enviada: %s\n", txAssinada.Hash().Hex())
		color.Yellow("Aguardando confirmação...\n")
		
		// Aguarda confirmação
		receipt, err := aguardarConfirmacao(txAssinada.Hash())
		if err != nil {
			return nil, err
		}
		
		if receipt.Status == 1 {
			color.Green("✓ Transação confirmada!\n")
		} else {
			color.Red("✗ Transação falhou!\n")
		}
		
		return txAssinada, nil
	}
	
	// Se falhar com "invalid sender", tenta com personal_sendTransaction
	if rpcClient != nil && (contains(err.Error(), "invalid sender") ||
		contains(err.Error(), "authentication needed")) {
		color.Yellow("⚠ Tentando método alternativo (personal_sendTransaction)...\n")
		txAssinada, err := enviarTransacaoViaRPC(data, valor)
		if err != nil {
			return nil, err
		}
		
		if txAssinada != nil {
			// Para transações via RPC, o hash pode não estar na transação
			// Mas a transação já foi enviada, então vamos aguardar confirmação
			// usando o nonce ou tentando obter o receipt de outra forma
			color.Yellow("Aguardando confirmação...\n")
			
			// Aguarda alguns segundos para o bloco ser criado
			time.Sleep(5 * time.Second)
			
			// Verifica se há novos blocos
			currentBlock, _ := client.BlockNumber(context.Background())
			if currentBlock > 0 {
				color.Green("✓ Bloco criado! Transação deve estar confirmada.\n")
			} else {
				color.Yellow("⚠ Ainda aguardando criação do primeiro bloco...\n")
			}
		}
		
		return txAssinada, nil
	}
	
	return nil, err
}

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Envia transação usando método normal
func enviarTransacaoNormal(data []byte, valor *big.Int) (*types.Transaction, error) {
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Obtém nonce da conta
	nonce, err := client.PendingNonceAt(context.Background(), contaAtual)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter nonce: %v", err)
	}
	
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Obtém gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("erro ao obter gas price: %v", err)
	}
	
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Obtém chain ID
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("erro ao obter chain ID: %v", err)
	}
	
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Cria transação
	var tx *types.Transaction
	if enderecoContrato == (common.Address{}) {
		// Criação de contrato (to = nil)
		tx = types.NewContractCreation(nonce, valor, gasLimit, gasPrice, data)
	} else {
		// Chamada de contrato
		tx = types.NewTransaction(nonce, enderecoContrato, valor, gasLimit, gasPrice, data)
	}
	
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Assina transação com chave privada
	txAssinada, err := types.SignTx(tx, types.NewEIP155Signer(chainID), chavePrivada)
	if err != nil {
		return nil, fmt.Errorf("erro ao assinar transação: %v", err)
	}
	
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Envia transação para a blockchain
	err = client.SendTransaction(context.Background(), txAssinada)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar transação: %v", err)
	}
	
	return txAssinada, nil
}

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Envia transação usando personal_sendTransaction (fallback)
func enviarTransacaoViaRPC(data []byte, valor *big.Int) (*types.Transaction, error) {
	// Desbloqueia a conta primeiro
	var unlockResult bool
	err := rpcClient.Call(&unlockResult, "personal_unlockAccount", contaAtual, senhaConta, 0)
	if err != nil || !unlockResult {
		return nil, fmt.Errorf("erro ao desbloquear conta: %v", err)
	}
	
	// Prepara parâmetros da transação
	txParams := map[string]interface{}{
		"from":  contaAtual.Hex(),
		"value": fmt.Sprintf("0x%x", valor),
		"data":  fmt.Sprintf("0x%x", data),
	}
	
	// Se for chamada de contrato, adiciona o endereço
	if enderecoContrato != (common.Address{}) {
		txParams["to"] = enderecoContrato.Hex()
	}
	
	// Envia via personal_sendTransaction
	var txHashStr string
	err = rpcClient.Call(&txHashStr, "personal_sendTransaction", txParams, senhaConta)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar transação via RPC: %v", err)
	}
	
	txHash := common.HexToHash(txHashStr)
	color.Cyan("Transação enviada via RPC: %s\n", txHash.Hex())
	
	// Aguarda um pouco e tenta obter a transação
	time.Sleep(1 * time.Second)
	tx, _, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		// Se não conseguir obter, cria uma transação dummy
		// O importante é que temos o hash para aguardar confirmação depois
		nonce, _ := client.PendingNonceAt(context.Background(), contaAtual)
		gasPrice, _ := client.SuggestGasPrice(context.Background())
		var toAddr common.Address
		if enderecoContrato != (common.Address{}) {
			toAddr = enderecoContrato
		}
		tx = types.NewTransaction(nonce, toAddr, valor, gasLimit, gasPrice, data)
	}
	
	return tx, nil
}

// Função auxiliar para verificar se string contém substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Aguarda confirmação de uma transação
func aguardarConfirmacao(txHash common.Hash) (*types.Receipt, error) {
	timeout := 60 * time.Second // Timeout de 60 segundos
	startTime := time.Now()
	lastBlock := uint64(0)
	
	for {
		// Verifica timeout
		if time.Since(startTime) > timeout {
			// Verifica se há blocos sendo criados
			currentBlock, _ := client.BlockNumber(context.Background())
			if currentBlock > lastBlock {
				color.Yellow("⚠ Timeout atingido, mas blocos estão sendo criados (bloco %d).\n", currentBlock)
				color.Yellow("⚠ A transação pode estar pendente. Verifique mais tarde.\n")
			} else {
				color.Yellow("⚠ Timeout atingido e nenhum bloco foi criado ainda.\n")
				color.Yellow("⚠ O Clique pode não estar selando blocos. Verifique se a conta está desbloqueada.\n")
			}
			return nil, fmt.Errorf("timeout aguardando confirmação")
		}
		
		// Tenta obter o receipt
		receipt, err := client.TransactionReceipt(context.Background(), txHash)
		if err == nil {
			return receipt, nil
		}

		// Se não for erro "not found", pode ser erro de conexão
		if err != ethereum.NotFound {
			time.Sleep(1 * time.Second)
			continue
		}

		// Verifica se há novos blocos sendo criados
		currentBlock, err := client.BlockNumber(context.Background())
		if err == nil && currentBlock > lastBlock {
			lastBlock = currentBlock
			color.Yellow("📦 Bloco %d criado, aguardando confirmação da transação...\n", currentBlock)
		}

		time.Sleep(2 * time.Second)
	}
}

// ===================== Funções de Contrato =====================

// Carrega o ABI do contrato
func carregarABI() error {
	// Lê o arquivo ABI
	abiBytes, err := ioutil.ReadFile(abiPath)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo ABI: %v", err)
	}

	// Parse o ABI JSON
	parsedABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		return fmt.Errorf("erro ao fazer parse do ABI: %v", err)
	}

	contractABI = parsedABI
	return nil
}

// Obtém o inventário de cartas de um jogador
func obterInventario(jogador common.Address) ([]*big.Int, error) {
	// Prepara a chamada à função obterInventario do contrato
	data, err := contractABI.Pack("obterInventario", jogador)
	if err != nil {
		return nil, fmt.Errorf("erro ao preparar chamada: %v", err)
	}

	// Faz a chamada ao contrato
	msg := ethereum.CallMsg{
		To:   &enderecoContrato,
		Data: data,
	}

	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao chamar contrato: %v", err)
	}

	// Verifica se o resultado está vazio
	if len(result) == 0 {
		return []*big.Int{}, nil
	}

	// Para arrays dinâmicos, o primeiro 32 bytes contêm o offset
	// Se o offset for 0, o array está vazio
	if len(result) >= 32 {
		offset := new(big.Int).SetBytes(result[0:32])
		if offset.Cmp(big.NewInt(0)) == 0 {
			// Array vazio
			return []*big.Int{}, nil
		}
	}

	// Tenta desempacotar usando UnpackIntoInterface primeiro
	var inventario []*big.Int
	err = contractABI.UnpackIntoInterface(&inventario, "obterInventario", result)
	if err == nil && len(inventario) > 0 {
		return inventario, nil
	}

	// Se falhar, tenta usar Unpack diretamente
	unpacked, err := contractABI.Unpack("obterInventario", result)
	if err != nil {
		// Se ainda falhar, verifica se é um array vazio
		// Arrays vazios retornam offset 0
		if len(result) >= 32 {
			offset := new(big.Int).SetBytes(result[0:32])
			if offset.Cmp(big.NewInt(0)) == 0 {
				return []*big.Int{}, nil
			}
		}
		return nil, fmt.Errorf("erro ao desempacotar resultado: %v", err)
	}

	// Verifica se o resultado está vazio
	if len(unpacked) == 0 {
		return []*big.Int{}, nil
	}

	// Converte o resultado para []*big.Int
	// O resultado pode ser []interface{} ou []*big.Int diretamente
	switch v := unpacked[0].(type) {
	case []*big.Int:
		return v, nil
	case []interface{}:
		inventario = make([]*big.Int, 0, len(v))
		for _, item := range v {
			switch tokenId := item.(type) {
			case *big.Int:
				inventario = append(inventario, tokenId)
			case uint64:
				inventario = append(inventario, big.NewInt(int64(tokenId)))
			case int64:
				inventario = append(inventario, big.NewInt(tokenId))
			default:
				// Tenta converter via reflection
				val := reflect.ValueOf(item)
				if val.Kind() == reflect.Ptr {
					val = val.Elem()
				}
				if val.Kind() == reflect.Uint64 || val.Kind() == reflect.Int64 {
					inventario = append(inventario, big.NewInt(val.Int()))
				}
			}
		}
		return inventario, nil
	default:
		return nil, fmt.Errorf("formato de resultado inesperado para obterInventario: %T", unpacked[0])
	}
}

// Obtém os dados de uma carta
func obterCarta(tokenId *big.Int) (*Carta, error) {
	// Prepara a chamada à função obterCarta do contrato
	data, err := contractABI.Pack("obterCarta", tokenId)
	if err != nil {
		return nil, fmt.Errorf("erro ao preparar chamada: %v", err)
	}

	// Faz a chamada ao contrato
	msg := ethereum.CallMsg{
		To:   &enderecoContrato,
		Data: data,
	}

	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao chamar contrato: %v", err)
	}

	// Desempacota o resultado usando UnpackIntoInterface
	// O go-ethereum desempacota structs Solidity como structs Go
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
		// Se UnpackIntoInterface falhar, tenta usar Unpack e reflection
		unpacked, unpackErr := contractABI.Unpack("obterCarta", result)
		if unpackErr != nil {
			return nil, fmt.Errorf("erro ao desempacotar resultado (UnpackIntoInterface: %v, Unpack: %v)", err, unpackErr)
		}

		if len(unpacked) == 0 {
			return nil, fmt.Errorf("resultado vazio")
		}

		// Tenta acessar via reflection se for uma struct
		val := reflect.ValueOf(unpacked[0])
		if val.Kind() == reflect.Struct {
			// Acessa os campos da struct via reflection
			idField := val.FieldByName("Id")
			nomeField := val.FieldByName("Nome")
			naipeField := val.FieldByName("Naipe")
			valorField := val.FieldByName("Valor")
			raridadeField := val.FieldByName("Raridade")
			timestampField := val.FieldByName("Timestamp")

			if !idField.IsValid() || !nomeField.IsValid() || !naipeField.IsValid() ||
				!valorField.IsValid() || !raridadeField.IsValid() || !timestampField.IsValid() {
				return nil, fmt.Errorf("estrutura de resultado inesperada: campos não encontrados")
			}

			return &Carta{
				ID:        idField.Interface().(*big.Int),
				Nome:      nomeField.Interface().(string),
				Naipe:     naipeField.Interface().(string),
				Valor:     valorField.Interface().(*big.Int),
				Raridade:  raridadeField.Interface().(string),
				Timestamp: timestampField.Interface().(*big.Int),
			}, nil
		}

		// Se não for struct, tenta como []interface{}
		if cartaValues, ok := unpacked[0].([]interface{}); ok && len(cartaValues) >= 6 {
			return &Carta{
				ID:        cartaValues[0].(*big.Int),
				Nome:      cartaValues[1].(string),
				Naipe:     cartaValues[2].(string),
				Valor:     cartaValues[3].(*big.Int),
				Raridade:  cartaValues[4].(string),
				Timestamp: cartaValues[5].(*big.Int),
			}, nil
		}

		return nil, fmt.Errorf("formato de resultado inesperado: %T", unpacked[0])
	}

	// Se UnpackIntoInterface funcionou, usa os dados diretamente
	return &Carta{
		ID:        cartaData.Id,
		Nome:      cartaData.Nome,
		Naipe:     cartaData.Naipe,
		Valor:     cartaData.Valor,
		Raridade:  cartaData.Raridade,
		Timestamp: cartaData.Timestamp,
	}, nil
}

// ===================== Funções de Interface =====================

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Exibe o menu principal
func exibirMenu() {
	color.Cyan("\n=== JOGO DE CARTAS MULTIPLAYER (BLOCKCHAIN) ===\n")
	fmt.Println("1. Ver Saldo e Cartas")
	fmt.Println("2. Comprar Pacote de Cartas (1 ETH)")
	fmt.Println("3. Trocar Cartas")
	fmt.Println("4. Registrar Vitória de Partida")
	fmt.Println("5. Ver Histórico de Partidas")
	fmt.Println("6. Deploy do Contrato (Admin)")
	fmt.Println("7. Configurar Endereço do Contrato")
	fmt.Println("0. Sair")

	prompt := promptui.Prompt{
		Label: "Escolha uma opção",
	}

	opcao, err := prompt.Run()
	if err != nil {
		color.Red("Erro ao ler opção: %v\n", err)
		return
	}

	switch opcao {
	case "1":
		verSaldoECartas()
	case "2":
		comprarPacote()
	case "3":
		trocarCartas()
	case "4":
		registrarVitoria()
	case "5":
		verHistoricoPartidas()
	case "6":
		err := fazerDeployContrato()
		if err != nil {
			color.Red("Erro no deploy: %v\n", err)
		}
	case "7":
		configurarContrato()
	case "0":
		color.Cyan("Saindo...\n")
		os.Exit(0)
	default:
		color.Red("Opção inválida!\n")
	}
}

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Consulta saldo e inventário
func verSaldoECartas() {
	// Atualiza saldo
	saldo, err := client.BalanceAt(context.Background(), contaAtual, nil)
	if err != nil {
		color.Red("Erro ao consultar saldo: %v\n", err)
		return
	}

	saldoEth := new(big.Float).Quo(new(big.Float).SetInt(saldo), big.NewFloat(1e18))
	color.Cyan("\n=== SEU PERFIL ===\n")
	fmt.Printf("Endereço: %s\n", contaAtual.Hex())
	fmt.Printf("Saldo: %f ETH\n", saldoEth)

	if enderecoContrato == (common.Address{}) {
		color.Yellow("⚠ Contrato não configurado. Configure para ver suas cartas.\n")
		return
	}

	// Carrega o ABI se ainda não foi carregado
	if len(contractABI.Methods) == 0 {
		err := carregarABI()
		if err != nil {
			color.Yellow("⚠ Não foi possível carregar o ABI: %v\n", err)
			color.Yellow("   Execute o script compile-contract.bat para gerar o ABI.\n")
			return
		}
	}

	// Obtém o inventário do jogador
	color.Yellow("\n📦 Suas Cartas:\n")
	
	// Debug: verifica se o contrato está configurado
	if enderecoContrato == (common.Address{}) {
		color.Red("Erro: Contrato não configurado! Use a opção 7 para configurar o endereço do contrato.\n")
		return
	}
	
	inventario, err := obterInventario(contaAtual)
	if err != nil {
		color.Red("Erro ao obter inventário: %v\n", err)
		return
	}

	if len(inventario) == 0 {
		color.Yellow("   Você ainda não possui cartas. Compre um pacote!\n")
		return
	}

	// Lista as cartas
	for i, tokenId := range inventario {
		carta, err := obterCarta(tokenId)
		if err != nil {
			color.Red("Erro ao obter carta #%d: %v\n", tokenId, err)
			continue
		}
		
		cor := color.New(color.FgWhite).SprintFunc()
		switch carta.Raridade {
		case "Comum":
			cor = color.New(color.FgWhite).SprintFunc()
		case "Rara":
			cor = color.New(color.FgCyan).SprintFunc()
		case "Épica":
			cor = color.New(color.FgMagenta).SprintFunc()
		case "Lendária":
			cor = color.New(color.FgYellow).SprintFunc()
		}
		
		fmt.Printf("   %d. %s - %s (%s) - Poder: %d\n", 
			i+1, 
			cor(carta.Nome), 
			carta.Naipe, 
			cor(carta.Raridade), 
			carta.Valor,
		)
	}
	
	fmt.Printf("\nTotal de cartas: %d\n", len(inventario))
}

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Compra pacote de cartas
func comprarPacote() {
	if enderecoContrato == (common.Address{}) {
		color.Red("✗ Contrato não configurado! Use a opção 7 primeiro.\n")
		return
	}

	color.Cyan("\n=== COMPRAR PACOTE ===\n")
	fmt.Println("Custo: 1 ETH")

	prompt := promptui.Prompt{
		Label:     "Confirmar compra? (s/n)",
		IsConfirm: true,
	}

	_, err := prompt.Run()
	if err != nil {
		return
	}

	// Carrega o ABI se ainda não foi carregado
	if len(contractABI.Methods) == 0 {
		err := carregarABI()
		if err != nil {
			color.Red("Erro ao carregar ABI: %v\n", err)
			return
		}
	}

	// Envia 1 ETH para o contrato (conforme definido no contrato)
	valor := big.NewInt(1000000000000000000) // 1 ETH

	// Prepara a chamada à função comprarPacote() usando o ABI
	data, err := contractABI.Pack("comprarPacote")
	if err != nil {
		color.Red("Erro ao preparar chamada: %v\n", err)
		return
	}

	_, err = enviarTransacao(data, valor)
	if err != nil {
		color.Red("Erro na compra: %v\n", err)
	}
}

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Troca de cartas
func trocarCartas() {
	color.Cyan("\n=== TROCA DE CARTAS ===\n")
	color.Yellow("⚠ Funcionalidade de troca requer implementação da lógica de assinatura de propostas.\n")
}

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Registra resultado de partida
func registrarVitoria() {
	if enderecoContrato == (common.Address{}) {
		color.Red("✗ Contrato não configurado! Use a opção 7 primeiro.\n")
		return
	}

	color.Cyan("\n=== REGISTRAR VITÓRIA ===\n")

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Endereço do oponente: ")
	if !scanner.Scan() {
		return
	}

	oponenteStr := strings.TrimSpace(scanner.Text())
	oponente := common.HexToAddress(oponenteStr)

	fmt.Print("Quem venceu? (1=Você, 2=Oponente, 0=Empate): ")
	if !scanner.Scan() {
		return
	}

	vencedorStr := strings.TrimSpace(scanner.Text())
	var vencedor common.Address

	switch vencedorStr {
	case "1":
		vencedor = contaAtual
	case "2":
		vencedor = oponente
	case "0":
		vencedor = common.Address{} // Endereço zero = empate
	default:
		color.Red("✗ Opção inválida!\n")
		return
	}

	// Apenas para evitar erro de variável não usada durante compilação
	_ = vencedor

	color.Yellow("⚠ Esta função requer o ABI do contrato para chamar registrarPartida().\n")
}

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Ver histórico de partidas
func verHistoricoPartidas() {
	if enderecoContrato == (common.Address{}) {
		color.Red("✗ Contrato não configurado! Use a opção 7 primeiro.\n")
		return
	}

	color.Cyan("\n=== HISTÓRICO DE PARTIDAS ===\n")
	color.Yellow("⚠ Esta função requer leitura de eventos PartidaRegistrada do contrato.\n")
}

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Configura endereço do contrato
func configurarContrato() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Digite o endereço do contrato: ")
	if scanner.Scan() {
		enderecoStr := strings.TrimSpace(scanner.Text())
		if common.IsHexAddress(enderecoStr) {
			enderecoContrato = common.HexToAddress(enderecoStr)
			color.Green("✓ Endereço configurado: %s\n", enderecoContrato.Hex())
		} else {
			color.Red("✗ Endereço inválido!\n")
		}
	}
}

// Simulação de batalha entre cartas
func simularBatalha(carta1, carta2 Carta, p1, p2 common.Address) {
	fmt.Printf("\nBatalha: %s de %s vs %s de %s\n", carta1.Nome, carta1.Naipe, carta2.Nome, carta2.Naipe)

	// Verifica quem venceu (simulação simples baseada em raridade/valor)
	// Em um jogo real, haveria lógica mais complexa
	if carta1.Valor.Cmp(carta2.Valor) > 0 {
		fmt.Printf("Vencedor: Jogador 1 (%s)\n", p1.Hex())
	} else if carta2.Valor.Cmp(carta1.Valor) > 0 {
		fmt.Printf("Vencedor: Jogador 2 (%s)\n", p2.Hex())
	} else {
		fmt.Println("Empate!")
	}
}
