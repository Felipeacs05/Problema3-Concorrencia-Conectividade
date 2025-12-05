// ===================== BAREMA ITEM 2: SMART CONTRACTS =====================
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

/**
 * @title GameEconomy
 * @dev Contrato principal que gerencia a economia do jogo de cartas multiplayer.
 * Implementa NFTs para cartas, sistema de pacotes, trocas e registro de partidas.
 * 
 * Este contrato garante:
 * - Propriedade única e verificável de cartas (NFTs)
 * - Distribuição justa de pacotes (prevenção de duplo gasto)
 * - Trocas atômicas entre jogadores
 * - Transparência total através de eventos na blockchain
 */

// ===================== Interface ERC-721 Simplificada =====================
/**
 * @dev Interface básica para tokens não-fungíveis (NFTs)
 * Representa cartas únicas que não podem ser duplicadas
 */
interface IERC721 {
    function ownerOf(uint256 tokenId) external view returns (address);
    function transferFrom(address from, address to, uint256 tokenId) external;
    function balanceOf(address owner) external view returns (uint256);
}

// ===================== Estruturas de Dados =====================

/**
 * @dev Estrutura que representa uma carta no jogo
 * Cada carta é um NFT único com propriedades específicas
 */
struct Carta {
    uint256 id;        // ID único do token (NFT)
    string nome;       // Nome da carta (ex: "Dragão", "Guerreiro")
    string naipe;      // Naipe: "Espadas", "Copas", "Ouros", "Paus"
    uint256 valor;     // Poder da carta (1-120)
    string raridade;   // "C" (Comum), "U" (Incomum), "R" (Rara), "L" (Lendária)
    uint256 timestamp; // Quando a carta foi criada
}

/**
 * @dev Estrutura para propostas de troca de cartas
 * Permite trocas atômicas entre dois jogadores
 */
struct PropostaTroca {
    address jogador1;        // Quem iniciou a proposta
    address jogador2;        // Com quem quer trocar
    uint256 cartaJogador1;   // Carta que jogador1 oferece
    uint256 cartaJogador2;   // Carta que jogador1 quer receber
    bool aceita;             // Se jogador2 aceitou
    bool executada;          // Se a troca já foi executada
    uint256 timestamp;       // Quando a proposta foi criada
}

/**
 * @dev Estrutura para registro de partidas
 * Armazena informações sobre resultados de partidas para auditabilidade
 */
struct Partida {
    address jogador1;
    address jogador2;
    address vencedor;        // address(0) se empate
    uint256 timestamp;
}

// ===================== Contrato Principal =====================
contract GameEconomy {
    // ===================== Variáveis de Estado =====================
    
    // BAREMA ITEM 8: PACOTES - Contador global de cartas criadas (para IDs únicos)
    uint256 private _tokenCounter;
    
    // BAREMA ITEM 8: PACOTES - Mapeamento de token ID para dados da carta
    mapping(uint256 => Carta) public cartas;
    
    // BAREMA ITEM 8: PACOTES - Mapeamento de endereço para lista de tokens possuídos
    mapping(address => uint256[]) public inventario;
    
    // BAREMA ITEM 8: PACOTES - Mapeamento de token ID para proprietário
    mapping(uint256 => address) public proprietario;
    
    // BAREMA ITEM 8: PACOTES - Mapeamento de proprietário para quantidade de tokens
    mapping(address => uint256) public saldo;
    
    // BAREMA ITEM 8: TROCAS - Mapeamento de ID de proposta para dados da proposta
    mapping(uint256 => PropostaTroca) public propostasTroca;
    uint256 private _propostaCounter;
    
    // BAREMA ITEM 7: PARTIDAS - Lista de partidas registradas
    Partida[] public partidas;
    
    // BAREMA ITEM 8: PACOTES - Preço de um pacote (em wei, moeda fictícia)
    uint256 public precoPacote = 1000000000000000000; // 1 ETH (fictício)
    
    // BAREMA ITEM 8: PACOTES - Quantidade de cartas por pacote
    uint256 public constant CARTAS_POR_PACOTE = 5;
    
    // BAREMA ITEM 7: PARTIDAS - Endereço do dono do contrato (para funções administrativas)
    address public owner;
    
    // ===================== Eventos =====================
    
    /**
     * @dev Emitido quando uma nova carta é criada (mintada)
     * Permite que clientes escutem e atualizem suas interfaces
     */
    event CartaCriada(
        uint256 indexed tokenId,
        address indexed proprietario,
        string nome,
        string raridade,
        uint256 valor
    );
    
    /**
     * @dev Emitido quando uma carta é transferida
     * Rastreia todas as mudanças de propriedade
     */
    event CartaTransferida(
        uint256 indexed tokenId,
        address indexed de,
        address indexed para
    );
    
    /**
     * @dev Emitido quando um pacote é comprado
     * Registra a transação de compra para transparência
     */
    event PacoteComprado(
        address indexed comprador,
        uint256[] tokenIds,
        uint256 timestamp
    );
    
    /**
     * @dev Emitido quando uma proposta de troca é criada
     * Notifica o jogador2 sobre a proposta pendente
     */
    event PropostaTrocaCriada(
        uint256 indexed propostaId,
        address indexed jogador1,
        address indexed jogador2,
        uint256 cartaJogador1,
        uint256 cartaJogador2
    );
    
    /**
     * @dev Emitido quando uma troca é executada
     * Confirma a transferência atômica de ambas as cartas
     */
    event TrocaExecutada(
        uint256 indexed propostaId,
        address indexed jogador1,
        address indexed jogador2,
        uint256 cartaJogador1,
        uint256 cartaJogador2
    );
    
    /**
     * @dev Emitido quando uma partida é registrada
     * Permite auditabilidade completa do histórico de partidas
     */
    event PartidaRegistrada(
        address indexed jogador1,
        address indexed jogador2,
        address indexed vencedor,
        uint256 timestamp
    );
    
    // ===================== Modificadores =====================
    
    /**
     * @dev Garante que apenas o dono do contrato pode executar certas funções
     */
    modifier onlyOwner() {
        require(msg.sender == owner, "Apenas o dono pode executar esta funcao");
        _;
    }
    
    /**
     * @dev Garante que o remetente possui o token especificado
     */
    modifier possuiCarta(uint256 tokenId) {
        require(proprietario[tokenId] == msg.sender, "Voce nao possui esta carta");
        _;
    }
    
    // ===================== Construtor =====================
    
    /**
     * @dev Inicializa o contrato
     * Define o dono e inicializa contadores
     */
    constructor() {
        owner = msg.sender;
        _tokenCounter = 0;
        _propostaCounter = 0;
    }
    
    // ===================== Funções de Criação de Cartas (Minting) =====================
    
    /**
     * @dev Cria uma nova carta (mint NFT)
     * BAREMA ITEM 8: PACOTES - Esta função é chamada internamente ao comprar pacotes
     * @param _proprietario Endereço que receberá a carta
     * @param _nome Nome da carta
     * @param _naipe Naipe da carta
     * @param _valor Poder da carta
     * @param _raridade Raridade da carta
     * @return tokenId ID único da carta criada
     */
    function _criarCarta(
        address _proprietario,
        string memory _nome,
        string memory _naipe,
        uint256 _valor,
        string memory _raridade
    ) internal returns (uint256) {
        uint256 tokenId = _tokenCounter;
        _tokenCounter++;
        
        cartas[tokenId] = Carta({
            id: tokenId,
            nome: _nome,
            naipe: _naipe,
            valor: _valor,
            raridade: _raridade,
            timestamp: block.timestamp
        });
        
        proprietario[tokenId] = _proprietario;
        inventario[_proprietario].push(tokenId);
        saldo[_proprietario]++;
        
        emit CartaCriada(tokenId, _proprietario, _nome, _raridade, _valor);
        
        return tokenId;
    }
    
    /**
     * @dev Gera uma carta aleatória baseada em block hash
     * BAREMA ITEM 8: PACOTES - Usa blockhash como fonte de entropia para sorteio
     * @param _seed Seed adicional para aumentar aleatoriedade
     * @return nome Nome da carta gerada
     * @return naipe Naipe da carta gerada
     * @return valor Valor/poder da carta gerada
     * @return raridade Raridade da carta gerada
     */
    function _gerarCartaAleatoria(uint256 _seed) internal view returns (
        string memory nome,
        string memory naipe,
        uint256 valor,
        string memory raridade
    ) {
        // BAREMA ITEM 8: PACOTES - Combina blockhash com seed para aleatoriedade
        bytes32 hash = keccak256(abi.encodePacked(blockhash(block.number - 1), _seed, block.timestamp));
        
        // Lista de nomes de cartas disponíveis
        string[20] memory nomes = [
            "Dragao", "Guerreiro", "Mago", "Anjo", "Demonio",
            "Fenix", "Titan", "Sereia", "Lobo", "Aguia",
            "Cavaleiro", "Arqueiro", "Barbaro", "Paladino",
            "Ranger", "Bruxo", "Druida", "Monge", "Assassino", "Bardo"
        ];
        
        // Lista de naipes (usando nomes em vez de símbolos Unicode)
        string[4] memory naipes = ["Espadas", "Copas", "Ouros", "Paus"];
        
        // BAREMA ITEM 8: PACOTES - Sorteia raridade (C=70%, U=20%, R=9%, L=1%)
        uint256 raridadeRoll = uint256(hash) % 100;
        if (raridadeRoll < 70) {
            raridade = "C";
            valor = 1 + (uint256(hash) % 50); // 1-50
        } else if (raridadeRoll < 90) {
            raridade = "U";
            valor = 51 + (uint256(hash) % 30); // 51-80
        } else if (raridadeRoll < 99) {
            raridade = "R";
            valor = 81 + (uint256(hash) % 20); // 81-100
        } else {
            raridade = "L";
            valor = 101 + (uint256(hash) % 20); // 101-120
        }
        
        // Sorteia nome e naipe
        nome = nomes[uint256(hash) % 20];
        naipe = naipes[uint256(hash) % 4];
    }
    
    // ===================== Funções Públicas de Compra =====================
    
    /**
     * @dev Compra um pacote de cartas
     * BAREMA ITEM 8: PACOTES - Previne duplo gasto através de atomicidade da transação
     * Cada chamada cria exatamente 5 cartas novas e as atribui ao comprador
     * @return tokenIds Array com os IDs das cartas recebidas
     */
    function comprarPacote() public payable returns (uint256[] memory) {
        // BAREMA ITEM 8: PACOTES - Valida que o pagamento é suficiente
        require(msg.value >= precoPacote, "Valor insuficiente para comprar pacote");
        
        uint256[] memory tokenIds = new uint256[](CARTAS_POR_PACOTE);
        
        // BAREMA ITEM 8: PACOTES - Gera 5 cartas aleatórias
        // Cada carta usa um seed diferente baseado no tokenCounter para garantir unicidade
        for (uint256 i = 0; i < CARTAS_POR_PACOTE; i++) {
            (string memory nome, string memory naipe, uint256 valor, string memory raridade) = 
                _gerarCartaAleatoria(_tokenCounter + i);
            
            tokenIds[i] = _criarCarta(msg.sender, nome, naipe, valor, raridade);
        }
        
        emit PacoteComprado(msg.sender, tokenIds, block.timestamp);
        
        return tokenIds;
    }
    
    // ===================== Funções de Consulta =====================
    
    /**
     * @dev Retorna informações de uma carta específica
     * @param tokenId ID da carta
     * @return Carta Estrutura com todos os dados da carta
     */
    function obterCarta(uint256 tokenId) public view returns (Carta memory) {
        require(proprietario[tokenId] != address(0), "Carta nao existe");
        return cartas[tokenId];
    }
    
    /**
     * @dev Retorna todas as cartas de um jogador
     * @param _jogador Endereço do jogador
     * @return tokenIds Array com IDs de todas as cartas do jogador
     */
    function obterInventario(address _jogador) public view returns (uint256[] memory) {
        return inventario[_jogador];
    }
    
    /**
     * @dev Retorna o saldo (quantidade de cartas) de um jogador
     * @param _jogador Endereço do jogador
     * @return Quantidade de cartas possuídas
     */
    function obterSaldo(address _jogador) public view returns (uint256) {
        return saldo[_jogador];
    }
    
    // ===================== Funções de Transferência =====================
    
    /**
     * @dev Transfere uma carta para outro endereço
     * BAREMA ITEM 8: TROCAS - Função base para transferências
     * @param _para Endereço que receberá a carta
     * @param tokenId ID da carta a ser transferida
     */
    function transferirCarta(address _para, uint256 tokenId) public possuiCarta(tokenId) {
        require(_para != address(0), "Nao pode transferir para endereco zero");
        require(_para != msg.sender, "Nao pode transferir para si mesmo");
        
        address _de = proprietario[tokenId];
        proprietario[tokenId] = _para;
        
        // Remove do inventário do remetente
        uint256[] storage invDe = inventario[_de];
        for (uint256 i = 0; i < invDe.length; i++) {
            if (invDe[i] == tokenId) {
                invDe[i] = invDe[invDe.length - 1];
                invDe.pop();
                break;
            }
        }
        
        // Adiciona ao inventário do destinatário
        inventario[_para].push(tokenId);
        
        saldo[_de]--;
        saldo[_para]++;
        
        emit CartaTransferida(tokenId, _de, _para);
    }
    
    // ===================== Funções de Troca =====================
    
    /**
     * @dev Cria uma proposta de troca de cartas
     * BAREMA ITEM 8: TROCAS - Permite que jogador1 ofereça sua carta pela carta de jogador2
     * @param _jogador2 Endereço do jogador com quem quer trocar
     * @param _minhaCarta ID da carta que está oferecendo
     * @param _cartaDesejada ID da carta que deseja receber
     * @return propostaId ID da proposta criada
     */
    function criarPropostaTroca(
        address _jogador2,
        uint256 _minhaCarta,
        uint256 _cartaDesejada
    ) public possuiCarta(_minhaCarta) returns (uint256) {
        require(_jogador2 != address(0), "Jogador invalido");
        require(_jogador2 != msg.sender, "Nao pode trocar consigo mesmo");
        require(proprietario[_cartaDesejada] == _jogador2, "Jogador2 nao possui esta carta");
        
        uint256 propostaId = _propostaCounter;
        _propostaCounter++;
        
        propostasTroca[propostaId] = PropostaTroca({
            jogador1: msg.sender,
            jogador2: _jogador2,
            cartaJogador1: _minhaCarta,
            cartaJogador2: _cartaDesejada,
            aceita: false,
            executada: false,
            timestamp: block.timestamp
        });
        
        emit PropostaTrocaCriada(propostaId, msg.sender, _jogador2, _minhaCarta, _cartaDesejada);
        
        return propostaId;
    }
    
    /**
     * @dev Aceita uma proposta de troca
     * BAREMA ITEM 8: TROCAS - Executa a troca atômica se ambas as partes aceitarem
     * @param propostaId ID da proposta a ser aceita
     */
    function aceitarPropostaTroca(uint256 propostaId) public {
        PropostaTroca storage proposta = propostasTroca[propostaId];
        
        // CORREÇÃO: Verifica se a proposta existe (jogador1 não pode ser address(0) e timestamp deve ser > 0)
        require(proposta.jogador1 != address(0), "Proposta nao existe");
        require(proposta.timestamp > 0, "Proposta invalida");
        require(proposta.jogador2 == msg.sender, "Voce nao e o destinatario desta proposta");
        require(!proposta.executada, "Proposta ja foi executada");
        require(proprietario[proposta.cartaJogador1] == proposta.jogador1, "Jogador1 nao possui mais esta carta");
        require(proprietario[proposta.cartaJogador2] == proposta.jogador2, "Voce nao possui mais esta carta");
        
        proposta.aceita = true;
        proposta.executada = true;
        
        // BAREMA ITEM 8: TROCAS - Executa troca atômica: ambas as cartas são transferidas simultaneamente
        // CORREÇÃO: Usa função interna que não verifica se msg.sender possui a carta
        // porque na troca, o jogador2 aceita mas precisa transferir a carta do jogador1
        _transferirCartaInterno(proposta.jogador1, proposta.jogador2, proposta.cartaJogador1);
        _transferirCartaInterno(proposta.jogador2, proposta.jogador1, proposta.cartaJogador2);
        
        emit TrocaExecutada(
            propostaId,
            proposta.jogador1,
            proposta.jogador2,
            proposta.cartaJogador1,
            proposta.cartaJogador2
        );
    }
    
    /**
     * @dev Retorna informações de uma proposta de troca
     * @param propostaId ID da proposta
     * @return PropostaTroca Estrutura com dados da proposta
     */
    function obterPropostaTroca(uint256 propostaId) public view returns (PropostaTroca memory) {
        return propostasTroca[propostaId];
    }
    
    /**
     * @dev Registra uma troca de cartas executada pelo servidor (admin)
     * BAREMA ITEM 8: TROCAS - Permite que o servidor registre trocas em nome dos jogadores
     * Esta função é necessária porque o servidor não tem acesso às chaves privadas dos jogadores
     * @param _jogador1 Endereço do jogador que ofereceu a carta
     * @param _jogador2 Endereço do jogador que recebeu a oferta
     * @param _cartaJogador1 ID da carta oferecida pelo jogador1
     * @param _cartaJogador2 ID da carta oferecida pelo jogador2
     */
    function registrarTrocaAdmin(
        address _jogador1,
        address _jogador2,
        uint256 _cartaJogador1,
        uint256 _cartaJogador2
    ) public onlyOwner {
        require(_jogador1 != address(0), "Jogador1 invalido");
        require(_jogador2 != address(0), "Jogador2 invalido");
        require(_jogador1 != _jogador2, "Nao pode trocar consigo mesmo");
        require(proprietario[_cartaJogador1] == _jogador1, "Jogador1 nao possui esta carta");
        require(proprietario[_cartaJogador2] == _jogador2, "Jogador2 nao possui esta carta");
        
        // Cria registro da proposta para auditabilidade
        uint256 propostaId = _propostaCounter;
        _propostaCounter++;
        
        propostasTroca[propostaId] = PropostaTroca({
            jogador1: _jogador1,
            jogador2: _jogador2,
            cartaJogador1: _cartaJogador1,
            cartaJogador2: _cartaJogador2,
            aceita: true,
            executada: true,
            timestamp: block.timestamp
        });
        
        // Executa a transferência atômica das cartas
        _transferirCartaInterno(_jogador1, _jogador2, _cartaJogador1);
        _transferirCartaInterno(_jogador2, _jogador1, _cartaJogador2);
        
        emit PropostaTrocaCriada(propostaId, _jogador1, _jogador2, _cartaJogador1, _cartaJogador2);
        emit TrocaExecutada(propostaId, _jogador1, _jogador2, _cartaJogador1, _cartaJogador2);
    }
    
    /**
     * @dev Função interna para transferir carta entre jogadores (usada pelo admin)
     * @param _de Endereço de origem
     * @param _para Endereço de destino
     * @param tokenId ID da carta
     */
    function _transferirCartaInterno(address _de, address _para, uint256 tokenId) internal {
        require(proprietario[tokenId] == _de, "Remetente nao possui a carta");
        
        proprietario[tokenId] = _para;
        
        // Remove do inventário do remetente
        uint256[] storage invDe = inventario[_de];
        for (uint256 i = 0; i < invDe.length; i++) {
            if (invDe[i] == tokenId) {
                invDe[i] = invDe[invDe.length - 1];
                invDe.pop();
                break;
            }
        }
        
        // Adiciona ao inventário do destinatário
        inventario[_para].push(tokenId);
        
        saldo[_de]--;
        saldo[_para]++;
        
        emit CartaTransferida(tokenId, _de, _para);
    }
    
    // ===================== Funções de Partidas =====================
    
    /**
     * @dev Registra o resultado de uma partida
     * BAREMA ITEM 7: PARTIDAS - Cria registro permanente e auditável na blockchain
     * @param _jogador2 Endereço do oponente
     * @param _vencedor Endereço do vencedor (address(0) se empate)
     */
    function registrarPartida(address _jogador2, address _vencedor) public {
        require(_jogador2 != address(0), "Jogador2 invalido");
        require(_jogador2 != msg.sender, "Nao pode jogar consigo mesmo");
        require(_vencedor == msg.sender || _vencedor == _jogador2 || _vencedor == address(0), 
                "Vencedor deve ser um dos jogadores ou empate");
        
        Partida memory novaPartida = Partida({
            jogador1: msg.sender,
            jogador2: _jogador2,
            vencedor: _vencedor,
            timestamp: block.timestamp
        });
        
        partidas.push(novaPartida);
        
        emit PartidaRegistrada(msg.sender, _jogador2, _vencedor, block.timestamp);
    }
    
    /**
     * @dev Retorna o número total de partidas registradas
     * @return Quantidade de partidas
     */
    function obterTotalPartidas() public view returns (uint256) {
        return partidas.length;
    }
    
    /**
     * @dev Retorna informações de uma partida específica
     * @param indice Índice da partida no array
     * @return Partida Estrutura com dados da partida
     */
    function obterPartida(uint256 indice) public view returns (Partida memory) {
        require(indice < partidas.length, "Indice invalido");
        return partidas[indice];
    }
    
    // ===================== Funções Administrativas =====================
    
    /**
     * @dev Permite ao dono alterar o preço dos pacotes
     * @param _novoPreco Novo preço em wei
     */
    function definirPrecoPacote(uint256 _novoPreco) public onlyOwner {
        precoPacote = _novoPreco;
    }
    
    /**
     * @dev Permite ao dono retirar fundos acumulados
     */
    function retirarFundos() public onlyOwner {
        payable(owner).transfer(address(this).balance);
    }
}

