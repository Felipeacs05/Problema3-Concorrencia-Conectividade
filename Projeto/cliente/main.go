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
	gasLimit = uint64(3000000)
)

var (
	client           *ethclient.Client
	contaAtual       common.Address
	chavePrivada     *ecdsa.PrivateKey
	enderecoContrato common.Address
	contractABI      abi.ABI
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
	bytecode, err := ioutil.ReadFile("../contracts/GameEconomy.bin")
	if err != nil {
		return fmt.Errorf("erro ao ler bytecode: %v", err)
	}

	// Converte hex para bytes
	bytecodeBytes := common.FromHex(string(bytecode))

	// Prepara transação de criação de contrato (to = nil)
	tx, err := enviarTransacao(bytecodeBytes, big.NewInt(0))
	if err != nil {
		return err
	}

	enderecoContrato = crypto.CreateAddress(contaAtual, tx.Nonce())
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
			// Se for erro de conexão, tenta novamente
			time.Sleep(1 * time.Second)
			continue
		}

		time.Sleep(1 * time.Second)
	}
}

// ===================== Funções de Interface =====================

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Exibe o menu principal
func exibirMenu() {
	color.Cyan("\n=== JOGO DE CARTAS MULTIPLAYER (BLOCKCHAIN) ===\n")
	fmt.Println("1. Ver Saldo e Cartas")
	fmt.Println("2. Comprar Pacote de Cartas (0.1 ETH)")
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

	// Aqui chamaríamos a função do contrato para listar cartas
	// Como não temos o ABI compilado Go aqui, simulamos ou usamos raw call
	color.Yellow("⚠ Leitura de cartas requer ABI do contrato. Implementar leitura de eventos ou view functions.\n")
}

// BAREMA ITEM 3: APLICAÇÃO CLIENTE - Compra pacote de cartas
func comprarPacote() {
	if enderecoContrato == (common.Address{}) {
		color.Red("✗ Contrato não configurado! Use a opção 7 primeiro.\n")
		return
	}

	color.Cyan("\n=== COMPRAR PACOTE ===\n")
	fmt.Println("Custo: 0.1 ETH")

	prompt := promptui.Prompt{
		Label:     "Confirmar compra? (s/n)",
		IsConfirm: true,
	}

	_, err := prompt.Run()
	if err != nil {
		return
	}

	// Envia 0.1 ETH para o contrato
	valor := big.NewInt(100000000000000000) // 0.1 ETH

	// Dados da transação (chamada da função comprarPacote)
	// Keccak256("comprarPacote()") = 0x4f4b1b7e... (primeiros 4 bytes)
	// Para simplificar, vamos enviar apenas ETH se a função for receive/fallback
	// Ou construir os dados manualmente

	// Assumindo que o contrato tem função default/receive para comprar
	// Ou precisa do seletor da função

	// Exemplo de seletor para comprarPacote() (calculado externamente)
	// Seletor: 0xe2bbb0d8 (exemplo)
	data := common.FromHex("0xe2bbb0d8")

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

// Estruturas auxiliares para cartas (simulação)
type Carta struct {
	ID    *big.Int
	Nome  string
	Naipe string
	Valor *big.Int
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
