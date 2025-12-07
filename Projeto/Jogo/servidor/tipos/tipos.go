package tipos

import (
	"jogodistribuido/protocolo"
	"sync"
	"time"
)

type Carta = protocolo.Carta

// InfoServidor armazena informações sobre um servidor no cluster distribuído
// Usado para descoberta de servidores e monitoramento de saúde
type InfoServidor struct {
	Endereco   string    `json:"endereco"`   // Endereço HTTP do servidor (ex: "servidor1:8080")
	UltimoPing time.Time `json:"ultimo_ping"` // Última vez que recebeu um heartbeat deste servidor
	Ativo      bool      `json:"ativo"`       // Indica se o servidor está ativo e respondendo
}

// Cliente representa um jogador conectado ao servidor via MQTT
// Cada cliente tem seu próprio inventário e pode estar em uma sala
type Cliente struct {
	ID                 string                // ID único do cliente (gerado pelo servidor)
	Nome               string                // Nome do jogador
	Inventario         []protocolo.Carta     // Cartas que o jogador possui
	Sala               *Sala                 // Sala atual do jogador (nil se não estiver em partida)
	EnderecoBlockchain string                // Endereço da carteira blockchain do jogador (opcional)
	Mutex              sync.Mutex           // Mutex para proteger acesso concorrente ao inventário
}

// Sala representa uma partida entre dois jogadores
// Pode ser uma partida local (ambos no mesmo servidor) ou cross-server (jogadores em servidores diferentes)
type Sala struct {
	ID             string                // ID único da sala
	Jogadores      []*Cliente            // Lista dos jogadores na partida (sempre 2)
	Estado         string                // "AGUARDANDO_COMPRA" | "JOGANDO" | "FINALIZADO"
	CartasNaMesa   map[string]Carta      // Cartas jogadas na mesa nesta rodada (nome -> carta)
	PontosRodada   map[string]int        // Pontos de cada jogador na rodada atual
	PontosPartida  map[string]int          // Rodadas ganhas por cada jogador na partida
	NumeroRodada   int                   // Número da rodada atual
	Prontos        map[string]bool       // Indica quais jogadores já compraram cartas e estão prontos
	ServidorHost   string                // Servidor responsável pela lógica da partida (autoridade)
	ServidorSombra string                // Servidor backup (shadow) que replica o estado
	EventSeq       int64                 // Sequência de eventos para ordenação e sincronização
	EventLog       []GameEvent           // Log append-only de eventos da partida (para auditoria)
	Mutex          sync.Mutex            // Mutex para proteger acesso concorrente à sala
	TurnoDe        string                `json:"turno_de"` // ID do jogador que tem a vez de jogar
	CartasJogadas  map[string]Carta      `json:"cartas_jogadas"` // Cartas jogadas no turno atual
}

// GameEvent representa um evento no log da partida
// Usado para auditoria e sincronização entre servidores Host e Shadow
type GameEvent struct {
	EventSeq  int64       `json:"eventSeq"`  // Número sequencial do evento (para ordenação)
	MatchID   string      `json:"matchId"`   // ID da partida onde o evento ocorreu
	Timestamp time.Time   `json:"timestamp"`  // Quando o evento ocorreu
	EventType string      `json:"eventType"` // Tipo do evento (CARD_PLAYED, ROUND_END, etc.)
	PlayerID  string      `json:"playerId"`  // ID do jogador que gerou o evento
	Data      interface{} `json:"data"`       // Dados específicos do evento (varia conforme o tipo)
	Signature string      `json:"signature"` // Assinatura HMAC do evento (para verificação de integridade)
}

// EstadoPartida representa o estado completo de uma partida
// Usado para replicação entre servidores Host e Shadow em partidas cross-server
type EstadoPartida struct {
	SalaID         string           `json:"sala_id"`         // ID da sala
	Estado         string           `json:"estado"`          // Estado atual da partida
	CartasNaMesa   map[string]Carta `json:"cartas_na_mesa"`  // Cartas jogadas na mesa
	PontosRodada   map[string]int   `json:"pontos_rodada"`   // Pontos na rodada atual
	PontosPartida  map[string]int   `json:"pontos_partida"`  // Rodadas ganhas na partida
	NumeroRodada   int              `json:"numero_rodada"`  // Número da rodada atual
	Prontos        map[string]bool  `json:"prontos"`         // Jogadores prontos
	EventSeq       int64            `json:"eventSeq"`       // Sequência de eventos (para sincronização)
	EventLog       []GameEvent      `json:"eventLog"`       // Log de eventos (para auditoria)
	TurnoDe        string           `json:"turnoDe"`        // ID do jogador que deve jogar
	VencedorJogada string           `json:"vencedor_jogada"` // Vencedor da jogada (se houver)
	Jogadores      []JogadorEstado  `json:"jogadores"`       // Inventários dos jogadores (para sincronização)
}

// JogadorEstado contém o estado de um jogador para sincronização
// Usado quando o servidor Shadow precisa sincronizar inventários
type JogadorEstado struct {
	ID         string  `json:"id"`         // ID do jogador
	Inventario []Carta `json:"inventario"` // Cartas que o jogador possui
}

// GameStartRequest é enviado pelo servidor Shadow para o Host quando uma partida deve ser iniciada
// Usado na comunicação cross-server para coordenar o início de partidas
type GameStartRequest struct {
	MatchID    string   `json:"matchId"`    // ID único da partida
	HostServer string   `json:"hostServer"` // Servidor que será o Host (autoridade)
	Players    []Player `json:"players"`    // Lista de jogadores que participarão
	Token      string   `json:"token"`      // Token JWT para autenticação entre servidores
}

// Player representa um jogador em uma requisição de início de partida
// Contém informações básicas do jogador e qual servidor ele está conectado
type Player struct {
	ID     string `json:"id"`     // ID do jogador
	Nome   string `json:"nome"`   // Nome do jogador
	Server string `json:"server"` // Servidor ao qual o jogador está conectado
}

// GameEventRequest é usado para enviar eventos de jogo do Shadow para o Host
// Permite que o Shadow notifique o Host sobre ações dos jogadores
type GameEventRequest struct {
	MatchID   string      `json:"matchId"`   // ID da partida
	EventSeq  int64       `json:"eventSeq"`  // Número sequencial do evento
	EventType string      `json:"eventType"` // Tipo do evento (CARD_PLAYED, etc.)
	PlayerID  string      `json:"playerId"`  // ID do jogador que gerou o evento
	Data      interface{} `json:"data"`      // Dados específicos do evento
	Token     string      `json:"token"`     // Token JWT para autenticação
	Signature string      `json:"signature"` // Assinatura HMAC para verificação de integridade
}

// GameReplicateRequest é usado para replicar o estado completo de uma partida
// Enviado do Host para o Shadow para manter sincronização
type GameReplicateRequest struct {
	MatchID   string        `json:"matchId"`   // ID da partida
	EventSeq  int64         `json:"eventSeq"`  // Sequência do evento (para ordenação)
	State     EstadoPartida `json:"state"`     // Estado completo da partida
	Token     string        `json:"token"`     // Token JWT para autenticação
	Signature string        `json:"signature"` // Assinatura HMAC para verificação de integridade
}
