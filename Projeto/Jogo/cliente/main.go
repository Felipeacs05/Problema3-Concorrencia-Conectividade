package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"jogodistribuido/protocolo"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

var (
	meuNome       string
	meuID         string
	mqttClient    mqtt.Client
	salaAtual     string
	oponenteID    string
	oponenteNome  string
	meuInventario []protocolo.Carta
	turnoDeQuem   string // NOVO: Armazena o ID de quem tem o turno
)

func main() {
	fmt.Println("=== Jogo de Cartas Multiplayer Distribuído ===")
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Digite seu nome: ")
	scanner.Scan()
	meuNome = strings.TrimSpace(scanner.Text())
	if meuNome == "" {
		meuNome = "Jogador"
	}

	// --- LÓGICA DE ESCOLHA CORRIGIDA ---
	// Detecta se está rodando dentro do Docker ou fora
	// Se a variável MQTT_BROKER_HOST estiver definida, usa ela
	// Senão, verifica se consegue resolver "broker1" (dentro do Docker)
	// Se não conseguir, assume que está fora e usa localhost com portas mapeadas
	mqttHost := os.Getenv("MQTT_BROKER_HOST")
	isDocker := false

	if mqttHost == "" {
		// Tenta detectar automaticamente se está dentro do Docker
		// Verifica se consegue resolver "broker1"
		if _, err := net.LookupHost("broker1"); err == nil {
			mqttHost = "broker1"
			isDocker = true
		} else {
			mqttHost = "localhost"
			isDocker = false
		}
	} else {
		// Se MQTT_BROKER_HOST foi definido, verifica se é um nome Docker
		isDocker = (mqttHost == "broker1" || mqttHost == "broker2" || mqttHost == "broker3")
	}

	// Mapeia servidores com portas corretas
	var serverMap map[int]string
	if isDocker {
		// Dentro do Docker: usa os nomes dos containers e porta padrão 1883
		serverMap = map[int]string{
			1: "tcp://broker1:1883",
			2: "tcp://broker2:1883",
			3: "tcp://broker3:1883",
		}
	} else {
		// Fora do Docker: usa localhost com as portas mapeadas
		serverMap = map[int]string{
			1: "tcp://localhost:1886", // Porta mapeada do broker1
			2: "tcp://localhost:1884", // Porta mapeada do broker2
			3: "tcp://localhost:1885", // Porta mapeada do broker3
		}
	}

	fmt.Println("\nEscolha o servidor para conectar:")
	fmt.Println("1. Servidor 1")
	fmt.Println("2. Servidor 2")
	fmt.Println("3. Servidor 3")
	fmt.Print("Opção: ")
	scanner.Scan()
	opcaoStr := scanner.Text()
	opcao, err := strconv.Atoi(opcaoStr)
	if err != nil || serverMap[opcao] == "" {
		log.Fatalf("Opção inválida.")
	}
	brokerAddr := serverMap[opcao]
	// --- FIM DA CORREÇÃO ---

	fmt.Printf("\nConectando ao broker MQTT: %s\n", brokerAddr)

	if err := conectarMQTT(brokerAddr); err != nil {
		log.Fatalf("Erro ao conectar ao MQTT: %v", err)
	}

	// --- LÓGICA DE LOGIN CORRIGIDA ---
	if err := fazerLogin(); err != nil {
		log.Fatalf("Erro no processo de login: %v", err)
	}
	// --- FIM DA CORREÇÃO ---

	fmt.Printf("\nBem-vindo, %s! (Seu ID: %s)\n", meuNome, meuID)

	// Tenta inicializar blockchain (opcional)
	fmt.Println("\n=== Configuração Blockchain (Opcional) ===")
	fmt.Println("Deseja conectar sua carteira blockchain?")
	fmt.Println("(Necessário para comprar/trocar cartas)")
	fmt.Print("(s/n): ")
	scanner.Scan()
	resposta := strings.ToLower(strings.TrimSpace(scanner.Text()))

	if resposta == "s" || resposta == "sim" || resposta == "y" || resposta == "yes" {
		// Inicializa blockchain (conexão com Geth e carrega contrato)
		if err := inicializarBlockchain(); err != nil {
			fmt.Printf("[AVISO] Blockchain não disponível: %v\n", err)
			fmt.Println("Você ainda pode jogar, mas não poderá comprar/trocar cartas.")
			blockchainEnabled = false // Garante que está desabilitado
		} else {
			// Carrega carteira (descriptografa keystore)
			if err := carregarCarteira(); err != nil {
				fmt.Printf("[AVISO] Erro ao carregar carteira: %v\n", err)
				fmt.Println("Você ainda pode jogar, mas não poderá comprar/trocar cartas.")
				blockchainEnabled = false // Desabilita se a carteira não foi carregada
				// Limpa variáveis para garantir estado limpo
				chavePrivada = nil
			}
			// Se chegou aqui e blockchainEnabled == true, significa que tudo foi carregado com sucesso
		}
	} else {
		fmt.Println("[INFO] Blockchain não conectada. Você pode jogar, mas não poderá comprar/trocar cartas.")
		blockchainEnabled = false // Garante que está desabilitado
	}

	fmt.Println("\nEntrando na fila de matchmaking...")
	entrarNaFila()

	mostrarAjuda()

	for scanner.Scan() {
		entrada := strings.TrimSpace(scanner.Text())
		processarComando(entrada)

		// Aguarda um pouco para receber mensagens
		time.Sleep(100 * time.Millisecond)
		fmt.Print("> ")
	}
}

func conectarMQTT(broker string) error {
	opts := mqtt.NewClientOptions()
	// --- CORREÇÃO APLICADA AQUI ---
	// Adiciona apenas o broker que o utilizador escolheu.
	opts.AddBroker(broker)
	// --- FIM DA CORREÇÃO ---
	opts.SetClientID("cliente_" + time.Now().Format("20060102150405"))
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetMaxReconnectInterval(10 * time.Second)
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		fmt.Printf("\n[AVISO] Conexão MQTT perdida: %v. Tentando reconectar...\n", err)
	})
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		fmt.Println("\n[INFO] Conectado ao broker MQTT.")
		if meuID != "" {
			// Reinscreve nos tópicos importantes após reconexão
			topicoEventos := fmt.Sprintf("clientes/%s/eventos", meuID)
			client.Subscribe(topicoEventos, 0, handleMensagemServidor)
			if salaAtual != "" {
				topicoPartida := fmt.Sprintf("partidas/%s/eventos", salaAtual)
				client.Subscribe(topicoPartida, 0, handleEventoPartida)
			}
		}
	})

	mqttClient = mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func fazerLogin() error {
	// Cria um canal para esperar a resposta do login de forma segura
	loginResponseChan := make(chan protocolo.Mensagem)

	// Gera um ID temporário único para esta sessão de login
	tempID := uuid.New().String()
	responseTopic := fmt.Sprintf("clientes/%s/eventos", tempID)

	// Inscreve-se no tópico de resposta ANTES de enviar o pedido
	if token := mqttClient.Subscribe(responseTopic, 1, func(c mqtt.Client, m mqtt.Message) {
		var msg protocolo.Mensagem
		if err := json.Unmarshal(m.Payload(), &msg); err == nil {
			loginResponseChan <- msg // Envia a resposta recebida para o canal
		}
	}); token.Wait() && token.Error() != nil {
		return fmt.Errorf("falha ao se inscrever no tópico de resposta: %v", token.Error())
	}

	// Publica a mensagem de login num tópico que o servidor ouve
	dadosLogin := protocolo.DadosLogin{Nome: meuNome}
	msgLogin := protocolo.Mensagem{Comando: "LOGIN", Dados: mustJSON(dadosLogin)}
	payloadLogin, _ := json.Marshal(msgLogin)

	// O tópico de login agora inclui o ID temporário
	loginTopic := fmt.Sprintf("clientes/%s/login", tempID)
	mqttClient.Publish(loginTopic, 1, false, payloadLogin)

	// Aguarda a resposta por um tempo limitado (sem time.Sleep!)
	select {
	case resp := <-loginResponseChan:
		if resp.Comando == "LOGIN_OK" {
			var dados map[string]string
			json.Unmarshal(resp.Dados, &dados)
			meuID = dados["cliente_id"] // Guarda o ID permanente recebido do servidor
			servidor := dados["servidor"]
			fmt.Printf("\n[LOGIN] Conectado ao servidor %s (ID: %s)\n", servidor, meuID)

			// Limpa a inscrição temporária e inscreve-se na permanente
			mqttClient.Unsubscribe(responseTopic)
			permanentTopic := fmt.Sprintf("clientes/%s/eventos", meuID)
			if token := mqttClient.Subscribe(permanentTopic, 1, handleMensagemServidor); token.Wait() && token.Error() != nil {
				return fmt.Errorf("falha ao se inscrever no tópico permanente: %v", token.Error())
			}
			return nil
		}
		return fmt.Errorf("resposta de login inesperada: %s", resp.Comando)
	case <-time.After(5 * time.Second): // Espera por 5 segundos
		return fmt.Errorf("não foi possível obter ID do servidor (timeout)")
	}
}

func entrarNaFila() {
	fmt.Printf("[DEBUG] entrarNaFila() chamado - meuID=%s\n", meuID)

	dados := map[string]string{"cliente_id": meuID}
	payload, _ := json.Marshal(dados)

	fmt.Printf("[DEBUG] Payload: %s\n", string(payload))

	topico := fmt.Sprintf("clientes/%s/entrar_fila", meuID)
	fmt.Printf("[DEBUG] Publicando no tópico: %s\n", topico)

	token := mqttClient.Publish(topico, 0, false, payload)
	if token.Wait() && token.Error() != nil {
		fmt.Printf("[ERRO] Falha ao publicar entrada na fila: %v\n", token.Error())
	} else {
		fmt.Printf("[DEBUG] Mensagem publicada com sucesso no tópico: %s\n", topico)
	}
}

var messageChan = make(chan protocolo.Mensagem, 10)

func handleMensagemServidor(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("[DEBUG] Mensagem recebida no tópico: %s\n", msg.Topic())
	fmt.Printf("[DEBUG] Payload: %s\n", string(msg.Payload()))

	var mensagem protocolo.Mensagem
	if err := json.Unmarshal(msg.Payload(), &mensagem); err != nil {
		log.Printf("[ERRO] Erro ao decodificar mensagem: %v", err)
		return
	}

	fmt.Printf("[DEBUG] Comando recebido: %s\n", mensagem.Comando)

	// Processa a mensagem
	processarMensagemServidor(mensagem)
}

func processarMensagemServidor(msg protocolo.Mensagem) {
	switch msg.Comando {
	case "LOGIN_OK":
		var dados map[string]string
		json.Unmarshal(msg.Dados, &dados)
		meuID = dados["cliente_id"]
		servidor := dados["servidor"]
		fmt.Printf("\n[LOGIN] Conectado ao servidor %s (ID: %s)\n", servidor, meuID)

		// Agora subscreve ao tópico correto com o ID
		topico := fmt.Sprintf("clientes/%s/eventos", meuID)
		token := mqttClient.Subscribe(topico, 0, handleMensagemServidor)
		token.Wait()

	case "AGUARDANDO_OPONENTE":
		fmt.Printf("\n[MATCHMAKING] Aguardando oponente...\n")
		fmt.Printf("[DEBUG] meuID=%s, salaAtual=%s\n", meuID, salaAtual)
		fmt.Print("> ")

	case "PARTIDA_ENCONTRADA":
		fmt.Printf("[DEBUG] PARTIDA_ENCONTRADA recebida!\n")
		fmt.Printf("[DEBUG] Dados brutos: %s\n", string(msg.Dados))

		var dados protocolo.DadosPartidaEncontrada
		if err := json.Unmarshal(msg.Dados, &dados); err != nil {
			fmt.Printf("[ERRO] Falha ao decodificar PARTIDA_ENCONTRADA: %v\n", err)
			return
		}

		fmt.Printf("[DEBUG] SalaID=%s, OponenteID=%s, OponenteNome=%s\n", dados.SalaID, dados.OponenteID, dados.OponenteNome)

		salaAtual = dados.SalaID
		oponenteID = dados.OponenteID
		oponenteNome = dados.OponenteNome

		fmt.Printf("\n[PARTIDA] Partida encontrada contra '%s'! (Sala: %s)\n", oponenteNome, salaAtual)
		fmt.Printf("[DEBUG] Estado atual: meuID=%s, oponenteID=%s, salaAtual=%s\n", meuID, oponenteID, salaAtual)
		fmt.Println("Use /comprar para adquirir seu pacote inicial de cartas.")

		// Subscreve aos eventos da partida
		topicoPartida := fmt.Sprintf("partidas/%s/eventos", salaAtual)
		fmt.Printf("[DEBUG] Inscrevendo-se no tópico da partida: %s\n", topicoPartida)
		if token := mqttClient.Subscribe(topicoPartida, 0, handleEventoPartida); token.Wait() && token.Error() != nil {
			log.Printf("[ERRO] Erro ao se inscrever no tópico da partida: %v", token.Error())
		} else {
			fmt.Printf("[DEBUG] Inscrito com sucesso no tópico da partida: %s\n", topicoPartida)
		}

		// CRÍTICO: Carrega cartas da blockchain ANTES de sincronizar
		log.Printf("[SYNC] Entrei na partida. Carregando inventário da blockchain...")
		if blockchainEnabled && chavePrivada != nil {
			cartas, err := obterInventarioBlockchain()
			if err == nil && len(cartas) > 0 {
				meuInventario = cartas
				log.Printf("[SYNC] Carregadas %d cartas da blockchain", len(cartas))
			} else {
				log.Printf("[SYNC] Erro ao carregar da blockchain ou sem cartas: %v", err)
			}
		}

		// Sincroniza com o servidor AGORA (mesmo que vazio, para garantir que o servidor saiba)
		if len(meuInventario) > 0 {
			log.Printf("[SYNC] Sincronizando %d cartas com o servidor...", len(meuInventario))
			sincronizarCartasComServidor()
		} else {
			log.Printf("[SYNC] AVISO: Inventário vazio! O jogador precisa comprar cartas primeiro.")
		}

	case "TROCA_CONCLUIDA":
		var resp protocolo.TrocarCartasResp
		json.Unmarshal(msg.Dados, &resp)
		fmt.Printf("\n[TROCA] %s\n", resp.Mensagem)
		// Atualiza o inventário se fornecido
		if len(resp.InventarioAtualizado) > 0 {
			meuInventario = resp.InventarioAtualizado
			log.Printf("[TROCA] Inventário atualizado com %d cartas", len(meuInventario))
		}
		mostrarCartas() // Mostra o inventário atualizado
		fmt.Print("> ")

	case "PACOTE_RESULTADO":
		var dados protocolo.ComprarPacoteResp
		json.Unmarshal(msg.Dados, &dados)

		// Se blockchain está habilitada, usa dados da blockchain (fonte da verdade)
		// Não sobrescreve com dados do servidor que podem estar desatualizados
		if blockchainEnabled && chavePrivada != nil {
			// Busca inventário atualizado da blockchain
			cartasBlockchain, err := obterInventarioBlockchain()
			if err == nil && len(cartasBlockchain) > 0 {
				meuInventario = cartasBlockchain
				fmt.Printf("\n╔═══════════════════════════════════════╗\n")
				fmt.Printf("║   PACOTE RECEBIDO! (Blockchain)       ║\n")
				fmt.Printf("║   Você recebeu %d cartas              ║\n", len(cartasBlockchain))
				fmt.Printf("╚═══════════════════════════════════════╝\n")
				fmt.Println("\nSuas cartas (da blockchain):")
				for i, carta := range cartasBlockchain {
					fmt.Printf("  %d. %s %s - Poder: %d (Raridade: %s) [ID: %s]\n",
						i+1, carta.Nome, carta.Naipe, carta.Valor, carta.Raridade, carta.ID)
				}
				// CRÍTICO: Sincroniza com o servidor após atualizar inventário
				log.Printf("[SYNC] Inventário atualizado após compra. Sincronizando %d cartas...", len(cartasBlockchain))
				sincronizarCartasComServidor()
			} else {
				// Fallback: usa dados do servidor se blockchain falhar
				meuInventario = dados.Cartas
				fmt.Printf("\n╔═══════════════════════════════════════╗\n")
				fmt.Printf("║   PACOTE RECEBIDO!                    ║\n")
				fmt.Printf("║   Você recebeu %d cartas              ║\n", len(dados.Cartas))
				fmt.Printf("╚═══════════════════════════════════════╝\n")
				fmt.Println("\nSuas cartas:")
				for i, carta := range dados.Cartas {
					fmt.Printf("  %d. %s %s - Poder: %d (Raridade: %s) [ID: %s]\n",
						i+1, carta.Nome, carta.Naipe, carta.Valor, carta.Raridade, carta.ID)
				}
				// CRÍTICO: Sincroniza com o servidor após atualizar inventário (fallback)
				log.Printf("[SYNC] Inventário atualizado após compra (fallback). Sincronizando %d cartas...", len(dados.Cartas))
				sincronizarCartasComServidor()
			}
		} else {
			// Sem blockchain, usa dados do servidor
			meuInventario = dados.Cartas
			fmt.Printf("\n╔═══════════════════════════════════════╗\n")
			fmt.Printf("║   PACOTE RECEBIDO!                    ║\n")
			fmt.Printf("║   Você recebeu %d cartas              ║\n", len(dados.Cartas))
			fmt.Printf("╚═══════════════════════════════════════╝\n")
			fmt.Println("\nSuas cartas:")
			for i, carta := range dados.Cartas {
				fmt.Printf("  %d. %s %s - Poder: %d (Raridade: %s) [ID: %s]\n",
					i+1, carta.Nome, carta.Naipe, carta.Valor, carta.Raridade, carta.ID)
			}
		}
		fmt.Print("> ")

	case "SISTEMA":
		var dados protocolo.DadosErro
		json.Unmarshal(msg.Dados, &dados)
		fmt.Printf("\n[SISTEMA] %s\n> ", dados.Mensagem)

	case "ERRO", "ERRO_JOGADA":
		var dados protocolo.DadosErro
		json.Unmarshal(msg.Dados, &dados)
		fmt.Printf("\n[ERRO] %s\n> ", dados.Mensagem)

	case "CHAT_RECEBIDO":
		var dados protocolo.DadosReceberChat
		if err := json.Unmarshal(msg.Dados, &dados); err == nil {
			prefixo := dados.NomeJogador
			if dados.NomeJogador == meuNome {
				prefixo = "[VOCÊ]"
			}
			// Usa \r para potencialmente limpar a linha atual antes de imprimir
			fmt.Printf("\r💬 %s: %s\n> ", prefixo, dados.Texto)
		} else {
			log.Printf("Erro ao decodificar dados do chat: %v", err)
		}

	case "ATUALIZACAO_JOGO":
		var dados protocolo.DadosAtualizacaoJogo
		json.Unmarshal(msg.Dados, &dados)

		// ATUALIZA O ESTADO DO TURNO
		log.Printf("[CLIENTE_DEBUG] Recebido ATUALIZACAO_JOGO. TurnoDe recebido: '%s', meuID: '%s'", dados.TurnoDe, meuID)
		if dados.TurnoDe != "" {
			antigoTurno := turnoDeQuem
			turnoDeQuem = dados.TurnoDe
			log.Printf("[CLIENTE_DEBUG] Turno atualizado: '%s' -> '%s'", antigoTurno, turnoDeQuem)
		} else {
			log.Printf("[CLIENTE_DEBUG] AVISO: TurnoDe está vazio na mensagem!")
		}

		// --- INÍCIO DA CORREÇÃO ---
		// Verifica se eu joguei uma carta nesta atualização
		if cartaJogada, euJoguei := dados.UltimaJogada[meuNome]; euJoguei {
			// Se sim, remove a carta do inventário local
			removerCartaDoInventario(cartaJogada.ID)
		}
		// --- FIM DA CORREÇÃO ---

		fmt.Printf("\n--- RODADA %d ---\n", dados.NumeroRodada)
		fmt.Println(dados.MensagemDoTurno)

		if len(dados.UltimaJogada) > 0 {
			fmt.Println("\nCartas na mesa:")
			for nome, carta := range dados.UltimaJogada {
				fmt.Printf("  %s: %s %s (Poder: %d)\n", nome, carta.Nome, carta.Naipe, carta.Valor)
			}
		}

		if dados.VencedorJogada != "" && dados.VencedorJogada != "EMPATE" {
			fmt.Printf("\n🏆 Vencedor da jogada: %s\n", dados.VencedorJogada)
		}

		if dados.VencedorRodada != "" && dados.VencedorRodada != "EMPATE" {
			fmt.Printf("🎯 Vencedor da rodada: %s\n", dados.VencedorRodada)
		}

		if len(dados.ContagemCartas) > 0 {
			fmt.Println("\nCartas restantes:")
			for nome, qtd := range dados.ContagemCartas {
				fmt.Printf("  %s: %d cartas\n", nome, qtd)
			}
		}

		// Mostra de quem é a vez
		quemJoga := oponenteNome
		if turnoDeQuem == meuID {
			quemJoga = "Você"
		}
		fmt.Printf("(Aguardando jogada de %s)\n-------------------\n> ", quemJoga)

	default:
		// Comando não reconhecido, ignora
	}
}

func handleEventoPartida(client mqtt.Client, msg mqtt.Message) {
	log.Printf("[MQTT_RX] Mensagem recebida no tópico: %s", msg.Topic())

	var mensagem protocolo.Mensagem
	if err := json.Unmarshal(msg.Payload(), &mensagem); err != nil {
		log.Printf("[MQTT_RX_ERRO] Erro ao decodificar mensagem: %v", err)
		return
	}

	log.Printf("[MQTT_RX] Comando recebido: %s", mensagem.Comando)

	switch mensagem.Comando {
	case "ATUALIZACAO_JOGO":
		log.Printf("[CLIENTE_DEBUG] === ATUALIZACAO_JOGO RECEBIDA ===")
		log.Printf("[CLIENTE_DEBUG] Payload bruto: %s", string(mensagem.Dados))

		var dados protocolo.DadosAtualizacaoJogo
		if err := json.Unmarshal(mensagem.Dados, &dados); err != nil {
			log.Printf("[CLIENTE_DEBUG] ERRO ao decodificar DadosAtualizacaoJogo: %v", err)
			return
		}

		log.Printf("[CLIENTE_DEBUG] Dados decodificados:")
		log.Printf("[CLIENTE_DEBUG]   - TurnoDe: '%s'", dados.TurnoDe)
		log.Printf("[CLIENTE_DEBUG]   - NumeroRodada: %d", dados.NumeroRodada)
		log.Printf("[CLIENTE_DEBUG]   - UltimaJogada: %d cartas", len(dados.UltimaJogada))
		log.Printf("[CLIENTE_DEBUG]   - meuID: '%s'", meuID)
		log.Printf("[CLIENTE_DEBUG]   - turnoDeQuem ANTES: '%s'", turnoDeQuem)

		// ATUALIZA O ESTADO DO TURNO - CRÍTICO!
		if dados.TurnoDe != "" {
			antigoTurno := turnoDeQuem
			turnoDeQuem = dados.TurnoDe
			log.Printf("[CLIENTE_DEBUG] ✅ Turno atualizado: '%s' -> '%s'", antigoTurno, turnoDeQuem)
			log.Printf("[CLIENTE_DEBUG] ✅ Verificação: turnoDeQuem agora é '%s', meuID é '%s'", turnoDeQuem, meuID)
			if turnoDeQuem == meuID {
				log.Printf("[CLIENTE_DEBUG] ✅✅✅ É A MINHA VEZ AGORA IRMAO! ✅✅✅")
			} else {
				log.Printf("[CLIENTE_DEBUG] ⏳ Não é minha vez, é a vez de: '%s'", turnoDeQuem)
			}
		} else {
			log.Printf("[CLIENTE_DEBUG] ⚠️ AVISO: TurnoDe está vazio na mensagem!")
		}

		// --- INÍCIO DA CORREÇÃO ---
		// Verifica se eu joguei uma carta nesta atualização
		if cartaJogada, euJoguei := dados.UltimaJogada[meuNome]; euJoguei {
			// Se sim, remove a carta do inventário local
			removerCartaDoInventario(cartaJogada.ID)
		}
		// --- FIM DA CORREÇÃO ---

		fmt.Printf("\n--- RODADA %d ---\n", dados.NumeroRodada)
		fmt.Println(dados.MensagemDoTurno)

		if len(dados.UltimaJogada) > 0 {
			fmt.Println("\nCartas na mesa:")
			for nome, carta := range dados.UltimaJogada {
				fmt.Printf("  %s: %s %s (Poder: %d)\n", nome, carta.Nome, carta.Naipe, carta.Valor)
			}
		}

		if dados.VencedorJogada != "" && dados.VencedorJogada != "EMPATE" {
			fmt.Printf("\n🏆 Vencedor da jogada: %s\n", dados.VencedorJogada)
		}

		if dados.VencedorRodada != "" && dados.VencedorRodada != "EMPATE" {
			fmt.Printf("🎯 Vencedor da rodada: %s\n", dados.VencedorRodada)
		}

		if len(dados.ContagemCartas) > 0 {
			fmt.Println("\nCartas restantes:")
			for nome, qtd := range dados.ContagemCartas {
				fmt.Printf("  %s: %d cartas\n", nome, qtd)
			}
		}

		// Mostra de quem é a vez
		if turnoDeQuem == meuID {
			fmt.Println("\n>>> É A SUA VEZ DE JOGAR! <<<")
		} else {
			fmt.Printf("\n(Aguardando jogada de %s)\n", oponenteNome)
		}

		fmt.Println("-------------------")
		fmt.Print("> ")

	case "FIM_DE_JOGO":
		var dados protocolo.DadosFimDeJogo
		json.Unmarshal(mensagem.Dados, &dados)

		fmt.Printf("\n╔═══════════════════════════════════════╗\n")
		if dados.VencedorNome == "EMPATE" {
			fmt.Printf("║   FIM DE JOGO - EMPATE!               ║\n")
		} else {
			fmt.Printf("║   FIM DE JOGO!                        ║\n")
			fmt.Printf("║   Vencedor: %-25s ║\n", dados.VencedorNome)
		}
		fmt.Printf("╚═══════════════════════════════════════╝\n")
		fmt.Print("> ")

	case "ERRO_JOGADA":
		var dados protocolo.DadosErro
		json.Unmarshal(mensagem.Dados, &dados)
		fmt.Printf("\n[JOGADA_INVALIDA] %s\n> ", dados.Mensagem)

	case "RECEBER_CHAT":
		var dados protocolo.DadosReceberChat
		json.Unmarshal(mensagem.Dados, &dados)

		prefixo := dados.NomeJogador
		if dados.NomeJogador == meuNome {
			prefixo = "[VOCÊ]"
		}
		fmt.Printf("\n💬 %s: %s\n> ", prefixo, dados.Texto)

	case "CHAT_RECEBIDO":
		var dados struct {
			NomeJogador string `json:"nomeJogador"`
			Texto       string `json:"texto"`
		}
		if err := json.Unmarshal(mensagem.Dados, &dados); err == nil {
			// Não exibe a própria mensagem de chat que o jogador enviou
			if dados.NomeJogador != meuNome {
				fmt.Printf("\r[%s]: %s\n> ", dados.NomeJogador, dados.Texto)
			}
		}
	default: // Adiciona um caso default para debugging
		log.Printf("[DEBUG] Comando de partida não reconhecido: %s", mensagem.Comando)
		fmt.Print("> ") // Garante que o prompt aparece
	}
}

func processarComando(entrada string) {
	partes := strings.Fields(entrada)
	if len(partes) == 0 {
		return
	}

	comando := partes[0]

	switch comando {
	case "/comprar":
		comprarPacote()

	case "/jogar":
		if len(partes) < 2 {
			fmt.Println("[ERRO] Uso: /jogar <ID_da_carta>")
			return
		}
		cartaID := partes[1]
		jogarCarta(cartaID)

	case "/cartas", "/inventario":
		mostrarCartas()

	case "/ajuda", "/help":
		mostrarAjuda()

	case "/sair":
		fmt.Println("Saindo...")
		os.Exit(0)
	case "/trocar":
		iniciarProcessoDeTroca()

	case "/conectar-carteira", "/conectar":
		reconectarCarteira()
	case "/aceitar":
		if len(partes) < 2 {
			fmt.Println("[ERRO] Uso: /aceitar <ID_DA_PROPOSTA>")
			return
		}
		propostaID := partes[1]
		if err := AceitarPropostaTrocaBlockchain(propostaID); err != nil {
			fmt.Printf("[ERRO] Falha ao aceitar troca: %v\n", err)
		} else {
			fmt.Println("[SUCESSO] Troca realizada com sucesso na Blockchain!")
			// Atualiza inventário local
			mostrarCartas()
		}
	default:
		// Se não for um comando, envia como chat
		if salaAtual != "" {
			enviarChat(entrada)
		} else {
			fmt.Println("[ERRO] Comando não reconhecido. Use /ajuda para ver os comandos disponíveis.")
		}
	}
}

func reconectarCarteira() {
	fmt.Println("\n=== Reconectar Carteira Blockchain ===")

	// Verifica se já está conectada
	if blockchainEnabled && chavePrivada != nil {
		fmt.Printf("✓ Carteira já conectada: %s\n", contaBlockchain.Hex())
		fmt.Println("Use /cartas para ver seu inventário na blockchain.")
		return
	}

	// Tenta inicializar blockchain se ainda não foi feito
	if blockchainClient == nil {
		fmt.Println("Inicializando conexão com blockchain...")
		if err := inicializarBlockchain(); err != nil {
			fmt.Printf("[ERRO] Falha ao conectar à blockchain: %v\n", err)
			fmt.Println("Certifique-se de que o Geth está rodando (docker ps)")
			return
		}
	}

	// Tenta carregar a carteira
	fmt.Println("Carregando carteira...")
	if err := carregarCarteira(); err != nil {
		fmt.Printf("[ERRO] Falha ao carregar carteira: %v\n", err)
		fmt.Println("\nDicas:")
		fmt.Println("- Verifique se a senha está correta")
		fmt.Println("- Execute criar-conta-jogador.bat para criar uma nova conta")
		fmt.Println("- Use fundar-conta.bat para adicionar ETH à sua conta")
		return
	}

	fmt.Println("✓ Carteira conectada com sucesso!")
	fmt.Printf("Endereço: %s\n", contaBlockchain.Hex())
	fmt.Println("Agora você pode usar /comprar para comprar cartas na blockchain!")
}

func comprarPacote() {
	fmt.Printf("[DEBUG] comprarPacote() chamado\n")
	fmt.Printf("[DEBUG] blockchainEnabled=%v, chavePrivada!=nil=%v\n", blockchainEnabled, chavePrivada != nil)
	fmt.Printf("[DEBUG] salaAtual=%s, meuID=%s\n", salaAtual, meuID)

	// Se blockchain está habilitada, usa blockchain
	if blockchainEnabled && chavePrivada != nil {
		fmt.Println("[BLOCKCHAIN] Comprando pacote na blockchain...")
		fmt.Printf("[DEBUG] Chamando comprarPacoteBlockchain()...\n")
		if err := comprarPacoteBlockchain(); err != nil {
			fmt.Printf("[ERRO] Falha ao comprar na blockchain: %v\n", err)
			return
		}
		fmt.Printf("[DEBUG] comprarPacoteBlockchain() concluído com sucesso\n")

		// Atualiza inventário da blockchain
		cartas, err := obterInventarioBlockchain()
		if err == nil {
			meuInventario = cartas
			fmt.Printf("[OK] Você agora possui %d cartas!\n", len(cartas))
		}

		// Notifica o servidor sobre a compra (opcional, para sincronização)
		if salaAtual != "" {
			dados := map[string]string{
				"cliente_id": meuID,
				"endereco":   contaBlockchain.Hex(),
			}
			mensagem := protocolo.Mensagem{
				Comando: "COMPRAR_PACOTE",
				Dados:   mustJSON(dados),
			}
			payload, _ := json.Marshal(mensagem)
			topico := fmt.Sprintf("partidas/%s/comandos", salaAtual)
			mqttClient.Publish(topico, 0, false, payload)
		}
		return
	}

	// Fallback: usa o método antigo via servidor (se não tiver blockchain)
	if salaAtual == "" {
		fmt.Println("[ERRO] Você não está em uma partida.")
		fmt.Println("[INFO] Para comprar cartas na blockchain, use /conectar-carteira primeiro.")
		return
	}

	fmt.Println("[AVISO] Blockchain não conectada.")
	fmt.Println("[INFO] Use /conectar-carteira para conectar sua carteira e comprar na blockchain.")
	fmt.Println("[INFO] Ou usando método via servidor...")
	dados := map[string]string{"cliente_id": meuID}
	mensagem := protocolo.Mensagem{
		Comando: "COMPRAR_PACOTE",
		Dados:   mustJSON(dados),
	}

	payload, _ := json.Marshal(mensagem)
	topico := fmt.Sprintf("partidas/%s/comandos", salaAtual)

	token := mqttClient.Publish(topico, 0, false, payload)
	token.Wait()
}

func jogarCarta(cartaID string) {
	if salaAtual == "" {
		fmt.Println("[ERRO] Você não está em uma partida.")
		return
	}

	// VERIFICAÇÃO DE TURNO NO CLIENTE
	log.Printf("[CLIENTE_DEBUG] Tentando jogar carta. turnoDeQuem='%s', meuID='%s'", turnoDeQuem, meuID)
	if turnoDeQuem != meuID {
		fmt.Printf("[ERRO] Não é a sua vez de jogar. Aguarde o oponente. (Turno atual: '%s', Seu ID: '%s')\n", turnoDeQuem, meuID)
		return
	}

	// Verifica se a carta existe no inventário
	cartaEncontrada := false
	var cartaNome string
	for _, carta := range meuInventario {
		if carta.ID == cartaID {
			cartaEncontrada = true
			cartaNome = carta.Nome
			break
		}
	}

	if !cartaEncontrada {
		fmt.Println("[ERRO] Carta não encontrada no seu inventário.")
		return
	}

	// Envia com cliente_id e carta_id
	dados := map[string]string{
		"cliente_id": meuID,
		"carta_id":   cartaID,
	}
	mensagem := protocolo.Mensagem{
		Comando: "JOGAR_CARTA",
		Dados:   mustJSON(dados),
	}

	payload, _ := json.Marshal(mensagem)
	topico := fmt.Sprintf("partidas/%s/comandos", salaAtual)
	token := mqttClient.Publish(topico, 0, false, payload)
	token.Wait()

	fmt.Printf("[INFO] Jogando carta: %s\n", cartaNome)
}

// --- INÍCIO DA NOVA FUNÇÃO ---
// removerCartaDoInventario remove uma carta do slice meuInventario pelo ID.
func removerCartaDoInventario(cartaID string) {
	novoInventario := []protocolo.Carta{}
	removida := false
	for _, c := range meuInventario {
		if c.ID == cartaID && !removida {
			removida = true // Garante que apenas uma carta seja removida se houver duplicatas (embora IDs devam ser únicos)
		} else {
			novoInventario = append(novoInventario, c)
		}
	}
	meuInventario = novoInventario
}

// --- FIM DA NOVA FUNÇÃO ---

func enviarChat(texto string) {
	if salaAtual == "" {
		return // Não faz sentido enviar chat se não estiver em sala
	}
	dados := protocolo.DadosEnviarChat{ClienteID: meuID, Texto: texto}
	msg := protocolo.Mensagem{
		Comando: "CHAT",
		Dados:   mustJSON(dados),
	}

	payload, _ := json.Marshal(msg)
	topico := fmt.Sprintf("partidas/%s/comandos", salaAtual)
	token := mqttClient.Publish(topico, 0, false, payload)
	token.Wait()
}

func mostrarCartas() {
	// Se blockchain está habilitada, busca da blockchain primeiro
	if blockchainEnabled && chavePrivada != nil {
		fmt.Println("[BLOCKCHAIN] Buscando inventário da blockchain...")
		cartas, err := obterInventarioBlockchain()
		if err == nil {
			meuInventario = cartas
			fmt.Printf("[OK] Inventário atualizado da blockchain\n")
		} else {
			fmt.Printf("[AVISO] Erro ao buscar da blockchain: %v\n", err)
			fmt.Println("Mostrando inventário local...")
		}
	}

	if len(meuInventario) == 0 {
		if blockchainEnabled && chavePrivada != nil {
			fmt.Println("[INFO] Você não possui cartas na blockchain. Use /comprar para adquirir um pacote.")
		} else {
			fmt.Println("[INFO] Você não possui cartas. Use /comprar para adquirir um pacote.")
			fmt.Println("[INFO] Para comprar cartas na blockchain, conecte sua carteira ao iniciar o jogo.")
		}
		return
	}

	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	if blockchainEnabled && chavePrivada != nil {
		fmt.Printf("║          SUAS CARTAS (Blockchain: %s)          ║\n", contaBlockchain.Hex()[:10]+"...")
	} else {
		fmt.Println("║                    SUAS CARTAS                            ║")
	}
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")

	for i, carta := range meuInventario {
		fmt.Printf("%2d. %-15s %s - Poder: %3d (Raridade: %s)\n",
			i+1, carta.Nome, carta.Naipe, carta.Valor, carta.Raridade)
		fmt.Printf("    ID: %s\n", carta.ID)
	}

	fmt.Printf("\nTotal: %d cartas\n", len(meuInventario))

	// Sincroniza com o servidor (se conectado)
	if mqttClient != nil && mqttClient.IsConnected() {
		sincronizarCartasComServidor()
	}
}

func sincronizarCartasComServidor() {
	if len(meuInventario) == 0 {
		log.Printf("[SYNC] AVISO: Tentando sincronizar mas inventário está vazio!")
		return
	}

	log.Printf("[SYNC] === INICIANDO SINCRONIZAÇÃO ===")
	log.Printf("[SYNC] Enviando %d cartas para sincronização com o servidor...", len(meuInventario))
	log.Printf("[SYNC] Sala atual: '%s'", salaAtual)
	log.Printf("[SYNC] Meu ID: '%s'", meuID)

	dados := struct {
		ClienteID string            `json:"cliente_id"`
		Cartas    []protocolo.Carta `json:"cartas"`
	}{
		ClienteID: meuID,
		Cartas:    meuInventario,
	}

	msg := protocolo.Mensagem{
		Comando: "SINCRONIZAR_CARTAS",
		Dados:   mustJSON(dados),
	}

	payload, _ := json.Marshal(msg)
	topico := fmt.Sprintf("partidas/%s/comandos", salaAtual)
	// Se não tiver sala, manda para tópico genérico ou aguarda entrar em sala
	if salaAtual == "" {
		// Tenta mandar para o tópico de comandos do servidor se possível,
		// mas como o tópico depende da sala, talvez seja melhor esperar entrar na sala.
		// Porém, o servidor precisa das cartas ANTES de validar a jogada.
		// O tópico "partidas/+/comandos" é wildcard no subscribe, mas para publicar precisamos de um ID.
		// Se não estamos em sala, não podemos sincronizar ainda.
		log.Printf("[SYNC] Ainda não estou em uma sala. Sincronização adiada.")
		return
	}

	log.Printf("[SYNC] Publicando no tópico: %s", topico)
	log.Printf("[SYNC] Payload: %s", string(payload))

	token := mqttClient.Publish(topico, 0, false, payload)
	if token.Wait() && token.Error() != nil {
		log.Printf("[SYNC] ERRO ao publicar: %v", token.Error())
	} else {
		log.Printf("[SYNC] ✅✅✅ Cartas enviadas para o servidor com sucesso! ✅✅✅")
	}
	log.Printf("[SYNC] === FIM SINCRONIZAÇÃO ===")
}

func mostrarAjuda() {
	fmt.Println("\nComandos disponíveis:")
	fmt.Println("  /cartas, /inventario   - Mostra suas cartas (busca da blockchain se conectado)")
	fmt.Println("  /comprar               - Compra um novo pacote de cartas (blockchain)")
	fmt.Println("  /conectar-carteira    - Conecta/reconecta sua carteira blockchain")
	if blockchainEnabled && chavePrivada != nil {
		fmt.Printf("  [BLOCKCHAIN] Carteira conectada: %s\n", contaBlockchain.Hex())
	} else {
		fmt.Println("  [INFO] Use /conectar-carteira para conectar sua carteira blockchain")
	}
	fmt.Println("  /jogar <ID_da_carta>   - Joga uma carta da sua mão")
	fmt.Println("  /trocar                - Propõe uma troca de cartas com o oponente")
	fmt.Println("  /ajuda                 - Mostra esta lista de comandos")
	fmt.Println("  /sair                  - Sai do jogo")
	fmt.Println("  Qualquer outro texto será enviado como chat.")
}

func iniciarProcessoDeTroca() {
	if !blockchainEnabled || chavePrivada == nil {
		fmt.Println("[ERRO] Você precisa conectar sua carteira (/conectar) para realizar trocas.")
		return
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("\n--- Propor Troca de Cartas (Blockchain) ---")
	mostrarCartas()

	// 1. Obter dados da troca
	fmt.Print("Digite o ID da carta que você quer teuaasas OFERECER: ")
	scanner.Scan()
	cartaOferecidaID := strings.TrimSpace(scanner.Text())

	fmt.Print("Digite o ID da carta que você quer RECEBER: ")
	scanner.Scan()
	cartaDesejadaID := strings.TrimSpace(scanner.Text())

	// COMO O JOGO NÃO FORNECE O ENDEREÇO AUTOMATICAMENTE, PEDIMOS AQUI
	// (Em uma versão final, isso viria nos dados da partida)
	fmt.Print("Digite o Endereço Ethereum do Oponente (0x...): ")
	scanner.Scan()
	enderecoOponente := strings.TrimSpace(scanner.Text())

	if cartaOferecidaID == "" || cartaDesejadaID == "" || enderecoOponente == "" {
		fmt.Println("Dados inválidos. Abortando troca.")
		return
	}

	// 2. Executar na Blockchain LOCALMENTE
	idProposta, err := CriarPropostaTrocaBlockchain(enderecoOponente, cartaOferecidaID, cartaDesejadaID)
	if err != nil {
		fmt.Printf("[ERRO] Falha ao criar proposta na blockchain: %v\n", err)
		return
	}

	fmt.Printf("\n[SUCESSO] Proposta criada na Blockchain! ID da Proposta: %s\n", idProposta)
	fmt.Println("Avise seu oponente para aceitar a proposta usando o comando:")
	fmt.Printf("/aceitar %s\n", idProposta)

	// Opcional: Enviar chat avisando
	enviarChat(fmt.Sprintf("Criei uma proposta de troca na blockchain! ID: %s", idProposta))
}

func mustJSON(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}
