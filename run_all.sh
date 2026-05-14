#!/bin/bash

# Cores para o log
GREEN='\033[0;32m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${PURPLE}===========================================${NC}"
echo -e "${PURPLE}   FINANCE OS - INITIALIZATION SCRIPT     ${NC}"
echo -e "${PURPLE}===========================================${NC}"

# 1. Limpeza de processos antigos
echo -e "${BLUE}[1/4] Limpando processos antigos...${NC}"
fuser -k 8080/tcp 2>/dev/null
fuser -k 8000/tcp 2>/dev/null
fuser -k 3000/tcp 2>/dev/null
sleep 1

# 2. Build e Start do Backend Go
echo -e "${BLUE}[2/4] Buildando e iniciando Backend (Go) na porta 8080...${NC}"
cd backend
go build -o server cmd/server/main.go
./server > backend.log 2>&1 &
BACKEND_PID=$!
cd ..

# 3. Start dos Agentes Python
echo -e "${BLUE}[3/4] Iniciando Agentes de IA (Python) na porta 8000...${NC}"
cd agents
source venv/bin/activate
PORT=8000 python main.py > agents.log 2>&1 &
AGENTS_PID=$!
cd ..

# 4. Start do Frontend Next.js
echo -e "${BLUE}[4/4] Iniciando Frontend (Next.js) em 0.0.0.0:3000...${NC}"
cd web
echo -e "${GREEN}SISTEMA ONLINE! Acesse via navegador.${NC}"
npm run dev -- -H 0.0.0.0

# Ao fechar o front (Ctrl+C), mata os processos de fundo
trap "echo -e '\n${RED}Frontend encerrado. Backend e Agentes continuam rodando.'; exit; echo -e '${RED}Serviços encerrados.${NC}'; exit" INT
