package game

import (
	"fmt"
	"jogodistribuido/protocolo"
	"jogodistribuido/servidor/seguranca"
	"jogodistribuido/servidor/tipos"
	"log"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	PACOTE_SIZE = 5
)

// GameManagerInterface defines the interface for game management
type GameManagerInterface interface {
	GetClientes() map[string]*tipos.Cliente
	GetSalas() map[string]*tipos.Sala
	GetFilaDeEspera() []*tipos.Cliente
	GetComandosPartida() map[string]chan protocolo.Comando
	GetMeuEndereco() string
	GetMeuEnderecoHTTP() string
	GetBrokerMQTT() string
	GetMQTTClient() mqtt.Client
	GetClusterManager() GameClusterManagerInterface
	GetStore() GameStoreInterface
	PublicarParaCliente(string, protocolo.Mensagem)
	PublicarEventoPartida(string, protocolo.Mensagem)
	NotificarCompraSucesso(string, []tipos.Carta)
	GetStatusEstoque() (map[string]int, int)
}

// GameClusterManagerInterface defines the interface for cluster management in game context
type GameClusterManagerInterface interface {
	SouLider() bool
	GetLider() string
	GetServidores() map[string]*tipos.InfoServidor
}

// GameStoreInterface defines the interface for store management in game context
type GameStoreInterface interface {
	FormarPacote(int) []tipos.Carta
	GetStatusEstoque() (map[string]int, int)
}

// Manager handles all game-related logic
type Manager struct {
	gameInterface GameManagerInterface
	mutex         sync.RWMutex
}

// NewManager creates a new game manager
func NewManager(gi GameManagerInterface) *Manager {
	return &Manager{
		gameInterface: gi,
	}
}

// EntrarFila adds a client to the matchmaking queue
func (m *Manager) EntrarFila(cliente *tipos.Cliente) {
	fila := m.gameInterface.GetFilaDeEspera()

	// Try to find opponent in local queue first
	if len(fila) > 0 {
		oponente := fila[0]
		// Remove from queue (this would need to be handled by the main server)
		// For now, we'll just create the room
		m.CriarSala(oponente, cliente)
		return
	}

	// If no opponent found, add to local queue
	// This would need to be handled by the main server
	log.Printf("Cliente %s (%s) entrou na fila de espera.", cliente.Nome, cliente.ID)
	m.gameInterface.PublicarParaCliente(cliente.ID, protocolo.Mensagem{
		Comando: "AGUARDANDO_OPONENTE",
		Dados:   seguranca.MustJSON(map[string]string{"mensagem": "Procurando oponente em todos os servidores..."}),
	})
}

// CriarSala creates a new game room with two players
func (m *Manager) CriarSala(j1, j2 *tipos.Cliente) {
	salaID := fmt.Sprintf("%d", time.Now().UnixNano())

	sala := &tipos.Sala{
		ID:             salaID,
		Jogadores:      []*tipos.Cliente{j1, j2},
		Prontos:        make(map[string]bool),
		Estado:         "AGUARDANDO_COMPRAS",
		ServidorHost:   j1.ID,
		ServidorSombra: j2.ID,
		Mutex:          sync.Mutex{},
		CartasJogadas:  make(map[string]tipos.Carta),
		TurnoDe:        "",
	}

	// Add room to server (this would need to be handled by the main server)
	log.Printf("Sala criada: %s com jogadores %s e %s", salaID, j1.Nome, j2.Nome)

	// Store the room (this would need to be handled by the main server)
	_ = sala // Placeholder to avoid unused variable warning

	// Notify both players
	m.gameInterface.PublicarParaCliente(j1.ID, protocolo.Mensagem{
		Comando: "PARTIDA_ENCONTRADA",
		Dados:   seguranca.MustJSON(map[string]string{"sala_id": salaID, "oponente": j2.Nome}),
	})

	m.gameInterface.PublicarParaCliente(j2.ID, protocolo.Mensagem{
		Comando: "PARTIDA_ENCONTRADA",
		Dados:   seguranca.MustJSON(map[string]string{"sala_id": salaID, "oponente": j1.Nome}),
	})
}

// ProcessarCompraPacote handles card package purchase
func (m *Manager) ProcessarCompraPacote(clienteID string, sala *tipos.Sala) {
	souLider := m.gameInterface.GetClusterManager().SouLider()

	log.Printf("[COMPRAR_DEBUG] Processando compra para cliente %s, souLider: %v", clienteID, souLider)

	cartas := make([]tipos.Carta, 0)

	if souLider {
		cartas = m.gameInterface.GetStore().FormarPacote(PACOTE_SIZE)
		log.Printf("[COMPRAR_DEBUG] Líder retirou %d cartas do estoque", len(cartas))
	} else {
		// Make HTTP request to leader
		// This would need to be implemented with HTTP client
		log.Printf("[COMPRAR_DEBUG] Encaminhando compra para líder")
	}

	// Add cards to client inventory
	clientes := m.gameInterface.GetClientes()
	cliente, exists := clientes[clienteID]
	if !exists {
		log.Printf("Cliente %s não encontrado", clienteID)
		return
	}

	cliente.Mutex.Lock()
	cliente.Inventario = append(cliente.Inventario, cartas...)
	cliente.Mutex.Unlock()

	// Notify client
	_, total := m.gameInterface.GetStatusEstoque()
	msg := protocolo.Mensagem{
		Comando: "PACOTE_RESULTADO",
		Dados: seguranca.MustJSON(protocolo.ComprarPacoteResp{
			Cartas:          cartas,
			EstoqueRestante: total,
		}),
	}

	// Notify locally via MQTT or remotely via API
	if m.getClienteLocal(clienteID) != nil {
		m.gameInterface.PublicarParaCliente(clienteID, msg)
	} else {
		sala.Mutex.Lock()
		if sala.ServidorHost == clienteID {
			// Notify shadow server
			m.notificarSombra(sala, msg)
		} else {
			// Notify host server
			m.notificarHost(sala, msg)
		}
		sala.Mutex.Unlock()
	}

	// Mark as ready and check if both players are ready
	sala.Mutex.Lock()
	sala.Prontos[clienteID] = true
	sala.Mutex.Unlock()

	m.VerificarEIniciarPartidaSeProntos(sala)
}

// VerificarEIniciarPartidaSeProntos checks if both players are ready and starts the game
func (m *Manager) VerificarEIniciarPartidaSeProntos(sala *tipos.Sala) {
	sala.Mutex.Lock()
	prontos := len(sala.Prontos)
	totalJogadores := len(sala.Jogadores)
	sala.Mutex.Unlock()

	if prontos == totalJogadores {
		log.Printf("Ambos os jogadores prontos, iniciando partida na sala %s", sala.ID)
		m.IniciarPartida(sala)
	} else {
		log.Printf("Aguardando jogadores ficarem prontos: %d/%d", prontos, totalJogadores)
	}
}

// IniciarPartida starts the game
func (m *Manager) IniciarPartida(sala *tipos.Sala) {
	sala.Mutex.Lock()
	sala.Estado = "EM_JOGO"
	// Define o primeiro jogador como o dono do turno inicial
	if len(sala.Jogadores) > 0 {
		sala.TurnoDe = sala.Jogadores[0].ID
	}
	// Inicializa estruturas de jogo se necessário
	if sala.CartasJogadas == nil {
		sala.CartasJogadas = make(map[string]tipos.Carta)
	}
	sala.Mutex.Unlock()

	// Notify both players
	msg := protocolo.Mensagem{
		Comando: "PARTIDA_INICIADA",
		Dados: seguranca.MustJSON(map[string]string{
			"sala_id":  sala.ID,
			"turno_de": sala.TurnoDe,
		}),
	}

	for _, jogador := range sala.Jogadores {
		m.gameInterface.PublicarParaCliente(jogador.ID, msg)
	}

	// Broadcast to room
	m.gameInterface.PublicarEventoPartida(sala.ID, msg)
}

// BroadcastChat handles chat messages
func (m *Manager) BroadcastChat(sala *tipos.Sala, texto, remetenteNome string) {
	msg := protocolo.Mensagem{
		Comando: "CHAT_RECEBIDO",
		Dados: seguranca.MustJSON(protocolo.DadosReceberChat{
			NomeJogador: remetenteNome,
			Texto:       texto,
		}),
	}

	// Send to all players in the room
	for _, jogador := range sala.Jogadores {
		m.gameInterface.PublicarParaCliente(jogador.ID, msg)
	}

	// Broadcast to room
	m.gameInterface.PublicarEventoPartida(sala.ID, msg)
}

// Helper functions
func (m *Manager) getClienteLocal(clienteID string) *tipos.Cliente {
	clientes := m.gameInterface.GetClientes()
	return clientes[clienteID]
}

func (m *Manager) notificarSombra(sala *tipos.Sala, msg protocolo.Mensagem) {
	// Implementation would send HTTP request to shadow server
	log.Printf("Notificando servidor sombra para sala %s", sala.ID)
}

func (m *Manager) notificarHost(sala *tipos.Sala, msg protocolo.Mensagem) {
	// Implementation would send HTTP request to host server
	log.Printf("Notificando servidor host para sala %s", sala.ID)
}

// ProcessarJogada trata a lógica de um jogador jogar uma carta
func (m *Manager) ProcessarJogada(salaID, clienteID, cartaID string) {
	log.Printf("[GAME_DEBUG] === ProcessarJogada INICIADO ===")
	log.Printf("[GAME_DEBUG] salaID=%s, clienteID=%s, cartaID=%s", salaID, clienteID, cartaID)

	salas := m.gameInterface.GetSalas()
	sala, exists := salas[salaID]
	if !exists {
		log.Printf("[GAME_DEBUG] ERRO: Sala %s não encontrada", salaID)
		return
	}

	sala.Mutex.Lock()
	log.Printf("[GAME_DEBUG] Lock adquirido para sala %s", salaID)
	log.Printf("[GAME_DEBUG] Estado atual da sala: Estado=%s, TurnoDe=%s, CartasJogadas=%d", sala.Estado, sala.TurnoDe, len(sala.CartasJogadas))
	defer func() {
		sala.Mutex.Unlock()
		log.Printf("[GAME_DEBUG] Lock liberado para sala %s", salaID)
	}()

	// 1. Validações básicas
	if sala.Estado != "EM_JOGO" {
		log.Printf("[GAME_DEBUG] ERRO: Tentativa de jogada em sala fora de jogo (%s)", sala.Estado)
		return
	}
	if sala.TurnoDe != clienteID {
		log.Printf("[GAME_DEBUG] ERRO: Jogador %s tentou jogar fora de turno (TurnoDe=%s)", clienteID, sala.TurnoDe)
		return
	}
	if _, jaJogou := sala.CartasJogadas[clienteID]; jaJogou {
		log.Printf("[GAME_DEBUG] ERRO: Jogador %s já jogou neste turno", clienteID)
		return
	}

	// 2. Encontrar e remover a carta do inventário do jogador
	clientes := m.gameInterface.GetClientes()
	cliente := clientes[clienteID]
	var cartaJogada tipos.Carta
	cartaEncontrada := false

	cliente.Mutex.Lock()
	for i, c := range cliente.Inventario {
		if c.ID == cartaID {
			cartaJogada = c
			// Remove do slice
			cliente.Inventario = append(cliente.Inventario[:i], cliente.Inventario[i+1:]...)
			cartaEncontrada = true
			break
		}
	}
	cliente.Mutex.Unlock()

	if !cartaEncontrada {
		log.Printf("[GAME_DEBUG] ERRO: Carta %s não encontrada no inventário de %s", cartaID, clienteID)
		return
	}

	// 3. Atualizar Estado da Sala
	sala.CartasJogadas[clienteID] = cartaJogada
	log.Printf("[GAME_DEBUG] JOGADA: %s jogou %s na sala %s", cliente.Nome, cartaJogada.Nome, sala.ID)
	log.Printf("[GAME_DEBUG] Total de cartas jogadas agora: %d", len(sala.CartasJogadas))

	// 4. Verificar se o turno acabou ou passa a vez
	if len(sala.CartasJogadas) >= 2 {
		log.Printf("[GAME_DEBUG] Ambos jogaram! Chamando ResolverTurno")
		// Ambos jogaram: Resolver o combate
		m.ResolverTurno(sala)
	} else {
		log.Printf("[GAME_DEBUG] Apenas um jogou. Mudando turno...")
		// Passar a vez para o oponente
		turnoAnterior := sala.TurnoDe
		for _, jog := range sala.Jogadores {
			if jog.ID != clienteID {
				sala.TurnoDe = jog.ID
				log.Printf("[GAME_DEBUG] TURNO MUDADO: %s -> %s (jogador: %s)", turnoAnterior, sala.TurnoDe, jog.Nome)
				break
			}
		}
		log.Printf("[GAME_DEBUG] Chamando NotificarAtualizacaoJogo com TurnoDe=%s", sala.TurnoDe)
		// Notificar atualização para ambos verem a carta na mesa
		m.NotificarAtualizacaoJogo(sala)
		log.Printf("[GAME_DEBUG] NotificarAtualizacaoJogo retornou")
	}
	log.Printf("[GAME_DEBUG] === ProcessarJogada FINALIZADO ===")
}

func (m *Manager) ResolverTurno(sala *tipos.Sala) {
	// Lógica simples de combate (exemplo: maior atributo vence)
	// Adapte conforme as regras do seu PBL (Elemento, Força, etc)

	var j1, j2 *tipos.Cliente
	j1 = sala.Jogadores[0]
	j2 = sala.Jogadores[1]

	c1 := sala.CartasJogadas[j1.ID]
	c2 := sala.CartasJogadas[j2.ID]

	vencedorTurno := ""

	// Comparação baseada no Valor da carta
	poder1 := c1.Valor
	poder2 := c2.Valor

	// Mapa de força dos naipes para desempate: Espadas > Copas > Ouros > Paus
	naipes := map[string]int{"♠": 4, "♥": 3, "♦": 2, "♣": 1}

	if poder1 > poder2 {
		vencedorTurno = j1.ID
	} else if poder2 > poder1 {
		vencedorTurno = j2.ID
	} else {
		// Empate no valor: decide pelo naipe
		if naipes[c1.Naipe] > naipes[c2.Naipe] {
			vencedorTurno = j1.ID
		} else {
			vencedorTurno = j2.ID
		}
	}

	// Limpar mesa para o próximo turno
	sala.CartasJogadas = make(map[string]tipos.Carta)

	// Notificar Resultado do Turno
	msg := protocolo.Mensagem{
		Comando: "RESULTADO_TURNO",
		Dados: seguranca.MustJSON(map[string]interface{}{
			"vencedor_turno": vencedorTurno,
			"carta_j1":       c1,
			"carta_j2":       c2,
		}),
	}
	m.gameInterface.PublicarEventoPartida(sala.ID, msg)

	// VERIFICAR FIM DE JOGO
	// Se alguém perdeu toda a vida ou cartas acabaram
	fimDeJogo := false
	vencedorPartida := ""
	// ... Implemente sua lógica de fim de jogo aqui ...

	if fimDeJogo {
		sala.Estado = "FINALIZADO"
		m.FinalizarPartidaBlockchain(sala, vencedorPartida, c1, c2) // Função para chamar a blockchain
	} else {
		// Se não acabou, notifica atualização e segue o jogo
		m.NotificarAtualizacaoJogo(sala)
	}
}

func (m *Manager) NotificarAtualizacaoJogo(sala *tipos.Sala) {
	log.Printf("[NOTIFICACAO_DEBUG] === NotificarAtualizacaoJogo INICIADO ===")
	log.Printf("[NOTIFICACAO_DEBUG] salaID=%s, TurnoDe=%s", sala.ID, sala.TurnoDe)
	log.Printf("[NOTIFICACAO_DEBUG] CartasJogadas: %d cartas", len(sala.CartasJogadas))

	// CORREÇÃO CRÍTICA: Usar a estrutura do protocolo corretamente
	// O protocolo espera "turnoDe" (camelCase), não "turno_de"
	dados := protocolo.DadosAtualizacaoJogo{
		TurnoDe:      sala.TurnoDe,
		UltimaJogada: sala.CartasJogadas,
		NumeroRodada: sala.NumeroRodada,
		// Outros campos podem ser preenchidos conforme necessário
	}

	log.Printf("[NOTIFICACAO_DEBUG] Dados montados usando protocolo.DadosAtualizacaoJogo")
	log.Printf("[NOTIFICACAO_DEBUG]   - TurnoDeeeeee='%s'", dados.TurnoDe)
	log.Printf("[NOTIFICACAO_DEBUG]   - NumeroRodada=%d", dados.NumeroRodada)
	log.Printf("[NOTIFICACAO_DEBUG]   - UltimaJogada=%d cartas", len(dados.UltimaJogada))

	msg := protocolo.Mensagem{
		Comando: "ATUALIZACAO_JOGO",
		Dados:   seguranca.MustJSON(dados),
	}

	log.Printf("[NOTIFICACAO_DEBUG] Mensagem criada. Comando=%s", msg.Comando)
	log.Printf("[NOTIFICACAO_DEBUG] Dados JSON: %s", string(msg.Dados))
	log.Printf("[NOTIFICACAO_DEBUG] Chamando PublicarEventoPartida para sala %s", sala.ID)

	m.gameInterface.PublicarEventoPartida(sala.ID, msg)

	log.Printf("[NOTIFICACAO_DEBUG] PublicarEventoPartida retornou")
	log.Printf("[NOTIFICACAO_DEBUG] === NotificarAtualizacaoJogo FINALIZADO ===")
}

// Apenas um esqueleto para a chamada da blockchain que você pediu
func (m *Manager) FinalizarPartidaBlockchain(sala *tipos.Sala, vencedorID string, ultimasCartas ...tipos.Carta) {
	log.Printf("!!! PARTIDA FINALIZADA - INICIANDO REGISTRO NA BLOCKCHAIN !!!")

	// Aqui você chama seu módulo de blockchain
	// go m.gameInterface.GetBlockchain().RegistrarVitoria(vencedorID, ...)

	msg := protocolo.Mensagem{
		Comando: "FIM_DE_JOGO",
		Dados:   seguranca.MustJSON(map[string]string{"vencedor": vencedorID}),
	}
	m.gameInterface.PublicarEventoPartida(sala.ID, msg)
}
