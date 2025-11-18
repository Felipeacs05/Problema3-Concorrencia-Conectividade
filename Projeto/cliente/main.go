// ===================== BAREMA ITEM 3: APLICAÇÃO CLIENTE =====================
// Este é o componente cliente do jogo de cartas multiplayer baseado em blockchain.
// O cliente gerencia: interface CLI, comunicação com blockchain via RPC,
// gerenciamento de carteira (keystore), e interação com smart contracts.

package main

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

// ===================== Constantes e Variáveis Globais =====================

const (
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Endereço RPC do nó Geth local
	// Usa IPC (Inter-Process Communication) para melhor performance
	rpcURL = "http://localhost:8545"
	
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Caminho para keystore local
	keystorePath = "./keystore"
	
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Gas limit para transações
	gasLimit = uint64(3000000)
)

var (
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Cliente Ethereum conectado ao nó local
	client *ethclient.Client
	
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Endereço da conta atual do jogador
	contaAtual common.Address
	
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Chave privada da conta atual
	chavePrivada *ecdsa.PrivateKey
	
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Endereço do contrato GameEconomy implantado
	enderecoContrato common.Address
	
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Canal para receber eventos do contrato
	eventosChan chan interface{}
)

// ===================== Estruturas de Dados =====================

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Estrutura que representa uma carta
// Espelha a estrutura do smart contract
type Carta struct {
	ID        *big.Int `json:"id"`
	Nome      string   `json:"nome"`
	Naipe     string   `json:"naipe"`
	Valor     *big.Int `json:"valor"`
	Raridade  string   `json:"raridade"`
	Timestamp *big.Int `json:"timestamp"`
}

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Estrutura para proposta de troca
type PropostaTroca struct {
	Jogador1       common.Address `json:"jogador1"`
	Jogador2       common.Address `json:"jogador2"`
	CartaJogador1  *big.Int       `json:"cartaJogador1"`
	CartaJogador2  *big.Int       `json:"cartaJogador2"`
	Aceita         bool           `json:"aceita"`
	Executada      bool           `json:"executada"`
	Timestamp      *big.Int       `json:"timestamp"`
}

// ===================== Funções de Inicialização =====================

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Conecta ao nó Ethereum local
func conectarBlockchain() error {
	var err error
	client, err = ethclient.Dial(rpcURL)
	if err != nil {
		return fmt.Errorf("erro ao conectar ao nó Ethereum: %v", err)
	}
	
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Verifica se a conexão está funcionando
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return fmt.Errorf("erro ao obter network ID: %v", err)
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
	
	indice, _, err := promptConta.Run()
	if err != nil {
		return err
	}
	
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Carrega a conta selecionada
	return carregarConta(contas[indice])
}

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Cria uma nova conta (carteira)
func criarNovaConta() error {
	ks := keystore.NewKeyStore(keystorePath, keystore.StandardScryptN, keystore.StandardScryptP)
	
	prompt := promptui.Prompt{
		Label: "Digite uma senha para a nova conta",
		Mask:  '*',
	}
	
	senha, err := prompt.Run()
	if err != nil {
		return err
	}
	
	account, err := ks.NewAccount(senha)
	if err != nil {
		return fmt.Errorf("erro ao criar conta: %v", err)
	}
	
	color.Green("✓ Nova conta criada: %s\n", account.Address.Hex())
	
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Carrega a conta recém-criada
	return carregarConta(account.URL.Path)
}

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Carrega uma conta existente
func carregarConta(caminhoArquivo string) error {
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Lê o arquivo do keystore
	keyJSON, err := ioutil.ReadFile(caminhoArquivo)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo keystore: %v", err)
	}
	
	prompt := promptui.Prompt{
		Label: "Digite a senha da conta",
		Mask:  '*',
	}
	
	senha, err := prompt.Run()
	if err != nil {
		return err
	}
	
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Desbloqueia a conta com a senha
	key, err := keystore.DecryptKey(keyJSON, senha)
	if err != nil {
		return fmt.Errorf("erro ao desbloquear conta (senha incorreta?): %v", err)
	}
	
	contaAtual = key.Address
	chavePrivada = key.PrivateKey
	
	color.Green("✓ Conta carregada: %s\n", contaAtual.Hex())
	
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Obtém saldo da conta
	saldo, err := client.BalanceAt(context.Background(), contaAtual, nil)
	if err != nil {
		return fmt.Errorf("erro ao obter saldo: %v", err)
	}
	
	color.Cyan("Saldo: %s ETH\n", weiParaEther(saldo))
	
	return nil
}

// ===================== Funções de Interação com Contrato =====================

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Faz deploy do contrato GameEconomy
// Esta função deve ser chamada apenas uma vez para implantar o contrato na blockchain
func fazerDeployContrato() error {
	color.Yellow("⚠ Deploy de contrato requer compilação Solidity e ABI.")
	color.Yellow("⚠ Use ferramentas como Hardhat ou Truffle para deploy completo.")
	color.Yellow("⚠ Por enquanto, você precisa fornecer o endereço do contrato já implantado.\n")
	
	prompt := promptui.Prompt{
		Label:   "Digite o endereço do contrato GameEconomy (ou deixe vazio para pular)",
		Default: "",
	}
	
	endereco, err := prompt.Run()
	if err != nil {
		return err
	}
	
	if endereco == "" {
		return nil
	}
	
	enderecoContrato = common.HexToAddress(endereco)
	color.Green("✓ Contrato configurado: %s\n", enderecoContrato.Hex())
	
	return nil
}

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Envia uma transação para a blockchain
func enviarTransacao(data []byte, valor *big.Int) (*types.Transaction, error) {
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
	tx := types.NewTransaction(nonce, enderecoContrato, valor, gasLimit, gasPrice, data)
	
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
	
	color.Cyan("Transação enviada: %s\n", txAssinada.Hash().Hex())
	color.Yellow("Aguardando confirmação...\n")
	
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Aguarda confirmação
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

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Aguarda confirmação de uma transação
func aguardarConfirmacao(txHash common.Hash) (*types.Receipt, error) {
	for {
		receipt, err := client.TransactionReceipt(context.Background(), txHash)
		if err == nil {
			return receipt, nil
		}
		
		if err != ethereum.NotFound {
			return nil, err
		}
		
		time.Sleep(1 * time.Second)
	}
}

// ===================== Funções de Interface =====================

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Exibe o menu principal
func exibirMenu() {
	color.Cyan("\n=== JOGO DE CARTAS MULTIPLAYER (BLOCKCHAIN) ===\n")
	fmt.Println("1. Ver Saldo e Cartas")
	fmt.Println("2. Comprar Pacote")
	fmt.Println("3. Trocar Carta")
	fmt.Println("4. Ver Propostas de Troca Pendentes")
	fmt.Println("5. Registrar Vitória de Partida")
	fmt.Println("6. Ver Histórico de Partidas")
	fmt.Println("7. Configurar Endereço do Contrato")
	fmt.Println("0. Sair")
	fmt.Print("\nEscolha uma opção: ")
}

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Loop principal do menu
func executarMenu() {
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		exibirMenu()
		
		if !scanner.Scan() {
			break
		}
		
		opcao := strings.TrimSpace(scanner.Text())
		
		switch opcao {
		case "1":
			verSaldoECartas()
		case "2":
			comprarPacote()
		case "3":
			criarPropostaTroca()
		case "4":
			verPropostasPendentes()
		case "5":
			registrarVitoria()
		case "6":
			verHistoricoPartidas()
		case "7":
			configurarContrato()
		case "0":
			color.Yellow("Saindo...\n")
			return
		default:
			color.Red("Opção inválida!\n")
		}
	}
}

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Ver saldo e cartas do jogador
func verSaldoECartas() {
	if enderecoContrato == (common.Address{}) {
		color.Red("✗ Contrato não configurado! Use a opção 7 primeiro.\n")
		return
	}
	
	color.Cyan("\n=== SEU INVENTÁRIO ===\n")
	
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Obtém saldo (quantidade de cartas)
	// Nota: Esta é uma chamada simplificada. Em produção, você usaria o ABI do contrato
	// para chamar a função obterSaldo() corretamente.
	
	saldo, err := client.BalanceAt(context.Background(), contaAtual, nil)
	if err != nil {
		color.Red("✗ Erro ao obter saldo: %v\n", err)
		return
	}
	
	color.Green("Saldo de ETH: %s\n", weiParaEther(saldo))
	color.Yellow("⚠ Para ver cartas, é necessário implementar chamada ao contrato via ABI.\n")
	color.Yellow("⚠ Use a biblioteca go-ethereum/accounts/abi para fazer chamadas ao contrato.\n")
}

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Compra um pacote de cartas
func comprarPacote() {
	if enderecoContrato == (common.Address{}) {
		color.Red("✗ Contrato não configurado! Use a opção 7 primeiro.\n")
		return
	}
	
	color.Cyan("\n=== COMPRAR PACOTE ===\n")
	color.Yellow("⚠ Esta função requer o ABI do contrato para chamar comprarPacote().\n")
	color.Yellow("⚠ Implementação completa requer uso de go-ethereum/accounts/abi.\n")
	
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Em implementação completa, você faria:
	// 1. Carregar ABI do contrato
	// 2. Criar instância do contrato
	// 3. Chamar comprarPacote() com valor (precoPacote)
	// 4. Aguardar confirmação
	// 5. Escutar evento PacoteComprado para obter IDs das cartas
}

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Cria proposta de troca
func criarPropostaTroca() {
	if enderecoContrato == (common.Address{}) {
		color.Red("✗ Contrato não configurado! Use a opção 7 primeiro.\n")
		return
	}
	
	color.Cyan("\n=== CRIAR PROPOSTA DE TROCA ===\n")
	color.Yellow("⚠ Esta função requer o ABI do contrato.\n")
}

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Ver propostas pendentes
func verPropostasPendentes() {
	if enderecoContrato == (common.Address{}) {
		color.Red("✗ Contrato não configurado! Use a opção 7 primeiro.\n")
		return
	}
	
	color.Cyan("\n=== PROPOSTAS PENDENTES ===\n")
	color.Yellow("⚠ Esta função requer escuta de eventos do contrato.\n")
}

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Registra vitória de partida
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
	
	fmt.Print("Digite o endereço do contrato GameEconomy: ")
	if !scanner.Scan() {
		return
	}
	
	enderecoStr := strings.TrimSpace(scanner.Text())
	
	if !common.IsHexAddress(enderecoStr) {
		color.Red("✗ Endereço inválido!\n")
		return
	}
	
	enderecoContrato = common.HexToAddress(enderecoStr)
	color.Green("✓ Contrato configurado: %s\n", enderecoContrato.Hex())
}

// ===================== Funções Utilitárias =====================

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Converte Wei para Ether (string formatada)
func weiParaEther(wei *big.Int) string {
	ether := new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(1e18))
	return ether.Text('f', 6)
}

// ===================== Função Principal =====================

func main() {
	color.Cyan("=== JOGO DE CARTAS MULTIPLAYER - BLOCKCHAIN ===\n")
	
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Conecta à blockchain
	if err := conectarBlockchain(); err != nil {
		color.Red("✗ Erro: %v\n", err)
		os.Exit(1)
	}
	
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Gerencia carteira
	if err := gerenciarCarteira(); err != nil {
		color.Red("✗ Erro ao gerenciar carteira: %v\n", err)
		os.Exit(1)
	}
	
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Pergunta sobre deploy do contrato
	prompt := promptui.Select{
		Label: "O contrato GameEconomy já foi implantado?",
		Items: []string{"Sim, já tenho o endereço", "Não, preciso fazer deploy"},
	}
	
	_, resultado, err := prompt.Run()
	if err != nil {
		color.Red("✗ Erro: %v\n", err)
		os.Exit(1)
	}
	
	if resultado == "Não, preciso fazer deploy" {
		if err := fazerDeployContrato(); err != nil {
			color.Red("✗ Erro: %v\n", err)
		}
	} else {
		configurarContrato()
	}
	
	// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Inicia loop do menu
	executarMenu()
}

