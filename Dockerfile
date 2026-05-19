# --- Build Go Backend ---
FROM golang:1.26-alpine AS go-builder

WORKDIR /app

# Instala dependências de build
RUN apk add --no-cache git

# Copia arquivos de módulo e baixa dependências
COPY backend/go.mod backend/go.sum ./backend/
RUN cd backend && go mod download

# Copia o código fonte e compila
COPY backend/ ./backend/
RUN cd backend && CGO_ENABLED=0 GOOS=linux go build -o server cmd/server/main.go


# --- Build Final Image ---
FROM python:3.12-slim

WORKDIR /app

# Instala dependências do sistema
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copia o binário do Go
COPY --from=go-builder /app/backend/server /app/backend/server

# Copia o serviço de Agentes Python e instala dependências
COPY agents/ /app/agents/
RUN pip install --no-cache-dir -r /app/agents/requirements.txt

# Cria o script de inicialização unificado
RUN echo '#!/bin/bash\n\
echo "Iniciando Serviço de Agentes Python na porta ${AGENTS_PORT:-8000}..."\n\
cd /app/agents && uvicorn main:app --host 127.0.0.1 --port ${AGENTS_PORT:-8000} &\n\
\n\
echo "Iniciando Backend Go na porta $PORT..."\n\
cd /app/backend && ./server' > /app/start.sh

RUN chmod +x /app/start.sh

# Variáveis de ambiente padrão para o Render
ENV PORT=8080
ENV AGENTS_SERVICE_URL=http://127.0.0.1:8000

# O Render expõe apenas a porta definida na variável PORT para o mundo externo
EXPOSE 8080

CMD ["/app/start.sh"]
