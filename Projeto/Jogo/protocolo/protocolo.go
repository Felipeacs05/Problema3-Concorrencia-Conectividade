package protocolo

import "encoding/json"

// Mensagem é o envelope base para todas as mensagens trocadas entre cliente e servidor
// Usa um padrão de comando + dados genéricos para flexibilidade
// Mensagem é o envelope base para todas as mensagens trocadas entre cliente e servidor
type Mensagem struct {
	Comando string          `json:"comando"` // Tipo da operação (LOGIN, JOGAR_CARTA, etc.)
	Dados   json.RawMessage `json:"dados"`   // Payload específico de cada comando (JSON raw para permitir diferentes estruturas)
}

/* ===================== Cartas / Inventário ===================== */

// Carta representa uma carta do jogo com seus atributos
// Cada carta tem um ID único que a identifica no sistema
// TÓPICO 3 - A estrutura Carta é a representação em memória do ativo digital, espelhando os dados imutáveis do ledger (ID, raridade, valor) para uso na aplicação cliente/servidor.
type Carta struct {
	ID       string `json:"id"`                 // Identificador único da carta no estoque global
	Nome     string `json:"nome"`               // Nome da carta para exibição (ex: "Dragão", "Guerreiro")
	Naipe    string `json:"naipe"`              // "♠", "♥", "♦", "♣" - usado para desempate quando valores são iguais
	Valor    int    `json:"valor"`              // Poder da carta (1..13, onde Ás=1, Rei=13)
	Raridade string `json:"raridade,omitempty"` // C=Comum, U=Incomum, R=Rara, L=Lendária
}

// ComprarPacoteReq é enviado pelo cliente quando deseja comprar um pacote de cartas
type ComprarPacoteReq struct {
	Quantidade int `json:"quantidade"` // Quantidade de pacotes desejados (padrão: 1)
}

// ComprarPacoteResp é a resposta do servidor após processar a compra
// Contém as cartas recebidas e o estoque restante para informação do jogador
type ComprarPacoteResp struct {
	Cartas          []Carta `json:"cartas"`          // Cartas recebidas no pacote
	EstoqueRestante int     `json:"estoqueRestante"` // Quantidade de cartas restantes no estoque global
}

// TrocarCartasReq contém os dados necessários para uma proposta de troca de cartas
// O jogador ofertante propõe trocar uma de suas cartas por uma carta do jogador desejado
type TrocarCartasReq struct {
	IDJogadorOferta     string `json:"id_jogador_oferta"`
	NomeJogadorOferta   string `json:"nome_jogador_oferta"`
	IDJogadorDesejado   string `json:"id_jogador_desejado"`
	NomeJogadorDesejado string `json:"nome_jogador_desejado"`
	IDCartaOferecida    string `json:"id_carta_oferecida"`
	IDCartaDesejada     string `json:"id_carta_desejada"`
}

// TrocarCartasResp é a resposta após processar uma troca de cartas
// Inclui o inventário atualizado para sincronização do cliente
type TrocarCartasResp struct {
	Sucesso              bool    `json:"sucesso"`
	Mensagem             string  `json:"mensagem"`
	InventarioAtualizado []Carta `json:"inventario_atualizado,omitempty"`
}

/* ===================== Login / Match / Chat ===================== */

// DadosLogin contém as informações necessárias para autenticar um jogador
// O nome é usado como identificador do jogador na sessão
type DadosLogin struct {
	Nome string `json:"nome"` // Nome único do jogador no sistema
}

// DadosPartidaEncontrada é enviado quando o matchmaking encontra um oponente
// Permite que ambos os jogadores saibam que a partida foi criada e quem é o oponente
type DadosPartidaEncontrada struct {
	SalaID       string `json:"salaID"`       // ID único da sala de jogo criada
	OponenteID   string `json:"oponenteID"`   // ID do oponente para referência
	OponenteNome string `json:"oponenteNome"` // Nome do oponente encontrado
}

// DadosEnviarChat contém a mensagem de chat que o jogador deseja enviar
type DadosEnviarChat struct {
	ClienteID string `json:"cliente_id"`
	Texto     string `json:"texto"` // Conteúdo da mensagem de chat
}

// DadosJogarCarta identifica qual carta o jogador deseja jogar na mesa
type DadosJogarCarta struct {
	CartaID string `json:"cartaID"` // ID da carta a ser jogada
}

// DadosReceberChat contém uma mensagem de chat recebida de outro jogador
type DadosReceberChat struct {
	NomeJogador string `json:"nomeJogador"` // Nome do jogador que enviou a mensagem
	Texto       string `json:"texto"`       // Conteúdo da mensagem
}

/* ===================== Atualizações de jogo ===================== */

// DadosAtualizacaoJogo contém todas as informações sobre o estado atual do jogo
// É enviado periodicamente para manter os clientes sincronizados
type DadosAtualizacaoJogo struct {
	MensagemDoTurno string           `json:"mensagem_do_turno"` // Mensagem descritiva do que aconteceu no turno
	ContagemCartas  map[string]int   `json:"contagem_cartas"`   // nome -> cartas restantes no inventário de cada jogador
	UltimaJogada    map[string]Carta `json:"ultima_jogada"`     // nome -> carta recém jogada na mesa por cada jogador
	VencedorJogada  string           `json:"vencedor_jogada"`   // nome do vencedor da jogada atual / "EMPATE" / ""
	VencedorRodada  string           `json:"vencedor_rodada"`   // nome do vencedor da rodada / "EMPATE" / ""
	NumeroRodada    int              `json:"numero_rodada"`     // Número da rodada atual (1, 2, 3...)
	PontosRodada    map[string]int   `json:"pontos_rodada"`     // nome -> pontos na rodada atual
	PontosPartida   map[string]int   `json:"pontos_partida"`    // nome -> rodadas ganhas na partida
	SalaID          string           `json:"sala_id"`           // ID da sala para roteamento na sombra
	TurnoDe         string           `json:"turnoDe"`           // ID do jogador que deve jogar no próximo turno
}

// DadosFimDeJogo é enviado quando a partida termina
// Informa qual jogador venceu ou se houve empate
type DadosFimDeJogo struct {
	VencedorNome string `json:"vencedorNome"` // Nome do vencedor final / "EMPATE" em caso de empate
	SalaID       string `json:"sala_id"`      // ID da sala para roteamento na sombra
}

// Comando representa uma ação de um jogador em uma partida
// Usado internamente pelo servidor para processar ações dos jogadores
type Comando struct {
	ClienteID string
	Tipo      string // Ex: "JOGAR_CARTA"
	Payload   json.RawMessage
}

/* ===================== Erro ===================== */

// DadosErro contém informações sobre um erro ocorrido durante o processamento
type DadosErro struct {
	Mensagem string `json:"mensagem"` // Descrição do erro ocorrido
}

/* ===================== Ping ===================== */

// DadosPing é usado para medir a latência entre cliente e servidor
type DadosPing struct {
	Timestamp int64 `json:"timestamp"` // Timestamp em milissegundos para cálculo de latência
}

// DadosPong é a resposta do servidor ao ping, ecoando o timestamp original
type DadosPong struct {
	Timestamp int64 `json:"timestamp"` // Timestamp ecoado do ping original
}
