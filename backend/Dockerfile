# Stage 1: Builder
FROM node:20-alpine AS frontend-builder
# (Se houvesse assets estáticos para buildar aqui)

FROM golang:1.23-alpine AS builder

WORKDIR /app

# Instalar dependências necessárias
RUN apk add --no-cache git

# Copiar go.mod e go.sum para aproveitar cache de camadas
COPY go.mod go.sum ./
RUN go mod download

# Copiar o restante do código
COPY . .

# Buildar o binário estático
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server cmd/server/main.go

# Stage 2: Final Image
FROM alpine:3.20

WORKDIR /app

# Instalar certificados CA e wget para healthcheck
RUN apk add --no-cache ca-certificates wget

# Copiar o binário do builder
COPY --from=builder /app/server /app/server

# Expor a porta padrão
EXPOSE 8080

# Comando para rodar a aplicação
CMD ["/app/server"]

# Healthcheck
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1
