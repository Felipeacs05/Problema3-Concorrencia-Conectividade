package store

import (
	"jogodistribuido/servidor/tipos"
	"log"
	"math/rand"
	"sync"
)

// StoreInterface define as operações que o Store de cartas expõe
// Permite abstração e testes com mocks
type StoreInterface interface {
	FormarPacote(tamanho int) []tipos.Carta // Retira cartas do estoque e forma um pacote
	GetStatusEstoque() (map[string]int, int) // Retorna o status do estoque por raridade e total
}

// Store gerencia o estoque global de cartas do servidor
// Mantém cartas separadas por raridade e garante acesso thread-safe
type Store struct {
	mutex   sync.RWMutex              // Mutex para proteger acesso concorrente ao estoque
	Estoque map[string][]tipos.Carta  // Mapa de raridade -> lista de cartas (C, U, R, L)
}

// NewStore cria e inicializa um novo Store com o estoque completo de cartas
func NewStore() *Store {
	s := &Store{
		Estoque: make(map[string][]tipos.Carta),
	}
	s.inicializarEstoque()
	return s
}

// randomString gera uma string aleatória de tamanho fixo
// Usado para criar IDs únicos para as cartas
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// inicializarEstoque popula o estoque com cartas de diferentes raridades
// Cria uma distribuição balanceada: mais cartas comuns, menos lendárias
func (s *Store) inicializarEstoque() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Inicializa os mapas por raridade
	s.Estoque = map[string][]tipos.Carta{
		"C": make([]tipos.Carta, 0), // Comuns
		"U": make([]tipos.Carta, 0), // Incomuns
		"R": make([]tipos.Carta, 0), // Raras
		"L": make([]tipos.Carta, 0), // Lendárias
	}

	// Tipos de cartas disponíveis no jogo
	tiposCartas := []string{
		"Dragão", "Guerreiro", "Mago", "Anjo", "Demônio", "Fênix", "Titan", "Sereia",
		"Lobo", "Águia", "Leão", "Tigre", "Cavaleiro", "Arqueiro", "Bárbaro", "Paladino",
	}
	naipes := []string{"♠", "♥", "♦", "♣"}

	// Para cada tipo de carta, cria múltiplas cópias com diferentes raridades
	for _, nome := range tiposCartas {
		// Comuns: 100 cópias, valor 1-50
		for i := 0; i < 100; i++ {
			s.Estoque["C"] = append(s.Estoque["C"], tipos.Carta{ID: randomString(5), Nome: nome, Naipe: naipes[rand.Intn(len(naipes))], Valor: 1 + rand.Intn(50), Raridade: "C"})
		}
		// Incomuns: 50 cópias, valor 51-80
		for i := 0; i < 50; i++ {
			s.Estoque["U"] = append(s.Estoque["U"], tipos.Carta{ID: randomString(5), Nome: nome, Naipe: naipes[rand.Intn(len(naipes))], Valor: 51 + rand.Intn(30), Raridade: "U"})
		}
		// Raras: 20 cópias, valor 81-100
		for i := 0; i < 20; i++ {
			s.Estoque["R"] = append(s.Estoque["R"], tipos.Carta{ID: randomString(5), Nome: nome, Naipe: naipes[rand.Intn(len(naipes))], Valor: 81 + rand.Intn(20), Raridade: "R"})
		}
		// Lendárias: 5 cópias, valor 101-120
		for i := 0; i < 5; i++ {
			s.Estoque["L"] = append(s.Estoque["L"], tipos.Carta{ID: randomString(5), Nome: nome, Naipe: naipes[rand.Intn(len(naipes))], Valor: 101 + rand.Intn(20), Raridade: "L"})
		}
	}

	log.Printf("Estoque inicializado: C=%d, U=%d, R=%d, L=%d",
		len(s.Estoque["C"]), len(s.Estoque["U"]), len(s.Estoque["R"]), len(s.Estoque["L"]))
}

// sampleRaridade retorna uma raridade aleatória baseada em probabilidades
// 70% Comum, 20% Incomum, 9% Rara, 1% Lendária
func sampleRaridade() string {
	x := rand.Intn(100)
	if x < 70 {
		return "C"
	}
	if x < 90 {
		return "U"
	}
	if x < 99 {
		return "R"
	}
	return "L"
}

// FormarPacote retira cartas do estoque e forma um pacote do tamanho especificado
// Tenta respeitar a raridade sorteada, mas se não houver cartas da raridade desejada,
// busca em raridades inferiores (fallback)
func (s *Store) FormarPacote(tamanho int) []tipos.Carta {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	cartas := make([]tipos.Carta, 0, tamanho)
	for i := 0; i < tamanho; i++ {
		// Sorteia uma raridade baseada em probabilidades
		raridade := sampleRaridade()
		// Ordem de fallback: se não tiver da raridade sorteada, tenta as inferiores
		ordem := []string{"L", "R", "U", "C"}
		var start int
		switch raridade {
		case "L":
			start = 0
		case "R":
			start = 1
		case "U":
			start = 2
		default:
			start = 3
		}

		// Tenta encontrar uma carta da raridade desejada ou inferior
		var carta tipos.Carta
		encontrou := false
		for j := start; j < len(ordem); j++ {
			r := ordem[j]
			if len(s.Estoque[r]) > 0 {
				// Retira a última carta do estoque (mais eficiente)
				idx := len(s.Estoque[r]) - 1
				carta = s.Estoque[r][idx]
				s.Estoque[r] = s.Estoque[r][:idx]
				encontrou = true
				break
			}
		}
		// Se não encontrou nenhuma carta no estoque, gera uma comum como fallback
		if !encontrou {
			carta = gerarCartaComum()
		}
		cartas = append(cartas, carta)
	}
	return cartas
}

// GetStatusEstoque retorna o status atual do estoque
// Retorna um mapa com a quantidade por raridade e o total de cartas
func (s *Store) GetStatusEstoque() (map[string]int, int) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	status := make(map[string]int)
	total := 0
	for raridade, cartas := range s.Estoque {
		status[raridade] = len(cartas)
		total += len(cartas)
	}
	return status, total
}

// gerarCartaComum gera uma carta comum quando o estoque está vazio
// Usado como fallback para garantir que sempre há cartas disponíveis
func gerarCartaComum() tipos.Carta {
	nomes := []string{"Guerreiro", "Arqueiro", "Mago", "Cavaleiro", "Ladrão"}
	naipes := []string{"♠", "♥", "♦", "♣"}
	return tipos.Carta{
		ID:       randomString(5),
		Nome:     nomes[rand.Intn(len(nomes))],
		Naipe:    naipes[rand.Intn(len(naipes))],
		Valor:    1 + rand.Intn(50),
		Raridade: "C",
	}
}
