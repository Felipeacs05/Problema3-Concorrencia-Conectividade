# ===================== BAREMA ITEM 1: ARQUITETURA =====================
# Dockerfile para o cliente Go do jogo de cartas
# Compila a aplicação e prepara para execução

FROM golang:1.22-alpine AS builder

# BAREMA ITEM 1: ARQUITETURA - Instala dependências de compilação
RUN apk add --no-cache git

# BAREMA ITEM 1: ARQUITETURA - Define diretório de trabalho
WORKDIR /app

# BAREMA ITEM 1: ARQUITETURA - Copia arquivos de dependências
COPY go.mod go.sum ./
RUN go mod download

# BAREMA ITEM 1: ARQUITETURA - Copia código fonte
COPY cliente/ ./cliente/

# BAREMA ITEM 1: ARQUITETURA - Compila a aplicação
WORKDIR /app/cliente
RUN CGO_ENABLED=0 GOOS=linux go build -o jogo-cartas main.go

# BAREMA ITEM 1: ARQUITETURA - Imagem final minimalista
FROM alpine:latest

# BAREMA ITEM 1: ARQUITETURA - Instala dependências de runtime
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# BAREMA ITEM 1: ARQUITETURA - Copia binário compilado
COPY --from=builder /app/cliente/jogo-cartas .

# BAREMA ITEM 1: ARQUITETURA - Cria diretório para keystore
RUN mkdir -p keystore

# BAREMA ITEM 1: ARQUITETURA - Define ponto de entrada
ENTRYPOINT ["./jogo-cartas"]

