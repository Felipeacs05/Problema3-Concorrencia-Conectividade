// ===================== BAREMA ITEM 1: ARQUITETURA =====================
// Utilitário cross-platform para gerenciar a blockchain privada Ethereum
// Funciona em Windows e Linux sem modificações

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
)

// ===================== Estruturas =====================

type GenesisConfig struct {
	Config struct {
		ChainID            int `json:"chainId"`
		HomesteadBlock     int `json:"homesteadBlock"`
		EIP150Block        int `json:"eip150Block"`
		EIP155Block        int `json:"eip155Block"`
		EIP158Block        int `json:"eip158Block"`
		ByzantiumBlock     int `json:"byzantiumBlock"`
		ConstantinopleBlock int `json:"constantinopleBlock"`
		PetersburgBlock    int `json:"petersburgBlock"`
		IstanbulBlock      int `json:"istanbulBlock"`
		BerlinBlock        int `json:"berlinBlock"`
		LondonBlock        int `json:"londonBlock"`
		ShanghaiBlock      int `json:"shanghaiBlock"`
		Clique             struct {
			Period int `json:"period"`
			Epoch  int `json:"epoch"`
		} `json:"clique"`
	} `json:"config"`
	Difficulty string                 `json:"difficulty"`
	GasLimit   string                 `json:"gasLimit"`
	ExtraData  string                 `json:"extraData"`
	Alloc      map[string]interface{} `json:"alloc"`
}

// ===================== Funções Principais =====================

// criarConta cria uma nova conta Ethereum no keystore
func criarConta(keystorePath, password string) (common.Address, error) {
	ks := keystore.NewKeyStore(keystorePath, keystore.StandardScryptN, keystore.StandardScryptP)
	
	account, err := ks.NewAccount(password)
	if err != nil {
		return common.Address{}, fmt.Errorf("erro ao criar conta: %v", err)
	}
	
	fmt.Printf("✓ Conta criada: %s\n", account.Address.Hex())
	fmt.Printf("✓ Senha: %s\n", password)
	fmt.Printf("✓ Keystore: %s\n", keystorePath)
	
	return account.Address, nil
}

// extrairEnderecoDoKeystore extrai o endereço do primeiro arquivo keystore encontrado
func extrairEnderecoDoKeystore(keystorePath string) (common.Address, error) {
	files, err := os.ReadDir(keystorePath)
	if err != nil {
		return common.Address{}, fmt.Errorf("erro ao ler keystore: %v", err)
	}
	
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "UTC--") {
			// Formato: UTC--2025-11-22T21-05-39.067881112Z--83d53c0bf346da6ca032e91d97372cdd9171372d
			// O endereço está após o último --
			parts := strings.Split(file.Name(), "--")
			if len(parts) >= 3 {
				addressHex := parts[len(parts)-1]
				address := common.HexToAddress("0x" + addressHex)
				return address, nil
			}
		}
	}
	
	return common.Address{}, fmt.Errorf("nenhum arquivo keystore encontrado")
}

// gerarExtraData gera o campo extraData para o Clique com o endereço do signer
func gerarExtraData(signerAddress common.Address) string {
	// Formato: 0x + 64 zeros (32 bytes) + endereço (20 bytes = 40 hex) + 130 zeros (65 bytes)
	addressHex := strings.TrimPrefix(signerAddress.Hex(), "0x")
	
	extraData := "0x"
	extraData += strings.Repeat("0", 64) // 32 bytes de zeros
	extraData += addressHex                // 20 bytes do endereço
	extraData += strings.Repeat("0", 130)  // 65 bytes de zeros
	
	return extraData
}

// gerarGenesisJSON gera o arquivo genesis.json com a configuração do Clique
func gerarGenesisJSON(signerAddress common.Address, genesisPath string) error {
	genesis := GenesisConfig{}
	
	// Configuração básica
	genesis.Config.ChainID = 1337
	
	// Configura todos os hard forks no bloco 0 (para suportar todas as features)
	genesis.Config.HomesteadBlock = 0
	genesis.Config.EIP150Block = 0
	genesis.Config.EIP155Block = 0
	genesis.Config.EIP158Block = 0
	genesis.Config.ByzantiumBlock = 0
	genesis.Config.ConstantinopleBlock = 0
	genesis.Config.PetersburgBlock = 0
	genesis.Config.IstanbulBlock = 0
	genesis.Config.BerlinBlock = 0
	genesis.Config.LondonBlock = 0
	genesis.Config.ShanghaiBlock = 0 // Importante para suportar PUSH0 (Solidity 0.8.20+)
	
	// Configuração do Clique (PoA)
	genesis.Config.Clique.Period = 5
	genesis.Config.Clique.Epoch = 30000
	
	genesis.Difficulty = "0x1"
	genesis.GasLimit = "0x8000000"
	genesis.ExtraData = gerarExtraData(signerAddress)
	
	// Aloca saldo inicial para o signer (1 milhão de ETH)
	genesis.Alloc = map[string]interface{}{
		signerAddress.Hex(): map[string]string{
			"balance": "1000000000000000000000000", // 1 milhão de ETH em Wei
		},
	}
	
	// Serializa para JSON
	jsonData, err := json.MarshalIndent(genesis, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar genesis: %v", err)
	}
	
	// Salva no arquivo
	err = os.WriteFile(genesisPath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("erro ao salvar genesis.json: %v", err)
	}
	
	fmt.Printf("✓ Genesis.json criado: %s\n", genesisPath)
	fmt.Printf("✓ Signer configurado: %s\n", signerAddress.Hex())
	
	return nil
}

// criarArquivoSenha cria um arquivo de senha sem quebra de linha (cross-platform)
func criarArquivoSenha(password, filePath string) error {
	return os.WriteFile(filePath, []byte(password), 0644)
}

// atualizarDockerCompose atualiza o docker-compose.yml com o novo endereço do signer
func atualizarDockerCompose(signerAddress common.Address, dockerComposePath string) error {
	// Lê o arquivo docker-compose.yml
	content, err := os.ReadFile(dockerComposePath)
	if err != nil {
		return fmt.Errorf("erro ao ler docker-compose.yml: %v", err)
	}
	
	contentStr := string(content)
	addressHex := signerAddress.Hex()
	
	// Regex para encontrar endereços Ethereum no formato 0x seguido de 40 caracteres hex
	// Procura por padrões como --unlock=0x... ou --miner.etherbase=0x...
	addressPattern := regexp.MustCompile(`0x[0-9a-fA-F]{40}`)
	
	// Substitui todos os endereços encontrados pelo novo endereço
	contentStr = addressPattern.ReplaceAllString(contentStr, addressHex)
	
	// Salva o arquivo atualizado
	err = os.WriteFile(dockerComposePath, []byte(contentStr), 0644)
	if err != nil {
		return fmt.Errorf("erro ao salvar docker-compose.yml: %v", err)
	}
	
	return nil
}

// ===================== Função Main =====================

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: blockchain-utils <comando> [argumentos]")
		fmt.Println("\nComandos disponíveis:")
		fmt.Println("  criar-conta <keystore-path> [senha]")
		fmt.Println("  gerar-genesis <keystore-path> <genesis-path>")
		fmt.Println("  extrair-endereco <keystore-path>")
		fmt.Println("  atualizar-docker-compose <endereco> <docker-compose-path>")
		os.Exit(1)
	}
	
	comando := os.Args[1]
	
	switch comando {
	case "criar-conta":
		if len(os.Args) < 3 {
			fmt.Println("Erro: forneça o caminho do keystore")
			fmt.Println("Uso: blockchain-utils criar-conta <keystore-path> [senha]")
			os.Exit(1)
		}
		
		keystorePath := os.Args[2]
		senha := "123456" // Senha padrão
		if len(os.Args) >= 4 {
			senha = os.Args[3]
		}
		
		// Cria diretório se não existir
		os.MkdirAll(keystorePath, 0700)
		
		address, err := criarConta(keystorePath, senha)
		if err != nil {
			fmt.Printf("ERRO: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("\n✓ Conta criada com sucesso!\n")
		fmt.Printf("  Endereço: %s\n", address.Hex())
		fmt.Printf("  Senha: %s\n", senha)
		
	case "gerar-genesis":
		if len(os.Args) < 4 {
			fmt.Println("Erro: forneça o caminho do keystore e do genesis.json")
			fmt.Println("Uso: blockchain-utils gerar-genesis <keystore-path> <genesis-path>")
			os.Exit(1)
		}
		
		keystorePath := os.Args[2]
		genesisPath := os.Args[3]
		
		// Extrai endereço do keystore
		address, err := extrairEnderecoDoKeystore(keystorePath)
		if err != nil {
			fmt.Printf("ERRO: %v\n", err)
			fmt.Println("\nDica: Crie uma conta primeiro com: blockchain-utils criar-conta <keystore-path>")
			os.Exit(1)
		}
		
		// Gera genesis.json
		err = gerarGenesisJSON(address, genesisPath)
		if err != nil {
			fmt.Printf("ERRO: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("\n✓ Genesis.json gerado com sucesso!\n")
		
	case "extrair-endereco":
		if len(os.Args) < 3 {
			fmt.Println("Erro: forneça o caminho do keystore")
			fmt.Println("Uso: blockchain-utils extrair-endereco <keystore-path>")
			os.Exit(1)
		}
		
		keystorePath := os.Args[2]
		
		address, err := extrairEnderecoDoKeystore(keystorePath)
		if err != nil {
			fmt.Printf("ERRO: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Println(address.Hex())
		
	case "atualizar-docker-compose":
		if len(os.Args) < 4 {
			fmt.Println("Erro: forneça o endereço e o caminho do docker-compose.yml")
			fmt.Println("Uso: blockchain-utils atualizar-docker-compose <endereco> <docker-compose-path>")
			os.Exit(1)
		}
		
		addressStr := os.Args[2]
		dockerComposePath := os.Args[3]
		
		address := common.HexToAddress(addressStr)
		if address == (common.Address{}) {
			fmt.Printf("ERRO: endereço inválido: %s\n", addressStr)
			os.Exit(1)
		}
		
		err := atualizarDockerCompose(address, dockerComposePath)
		if err != nil {
			fmt.Printf("ERRO: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("✓ docker-compose.yml atualizado com endereço: %s\n", address.Hex())
		
	default:
		fmt.Printf("Erro: comando desconhecido: %s\n", comando)
		os.Exit(1)
	}
}

