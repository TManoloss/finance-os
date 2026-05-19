## SYNC — Endpoint de Sync Interno + Configuração cron-job.org

```
PROMPT:

Você é um engenheiro Go sênior especialista em Echo framework e deploy no Render free tier.

Implemente o endpoint de sync interno e o sistema de keep-alive para manter
o serviço acordado no Render free tier.

---

**PARTE 1 — Endpoint de sync interno**

Criar `backend/internal/handler/sync.go`:

POST /api/v1/internal/sync
- Header obrigatório: X-Sync-Secret (valor vem de env SYNC_SECRET)
- Se header ausente ou errado: retornar 401 imediatamente
- Buscar todos os usuários com pelo menos 1 connected_account no banco
- Para cada usuário, chamar SyncUserAccounts(userID) em goroutines paralelas
  (usar sync.WaitGroup, máximo 3 goroutines simultâneas com semáforo)
- Retornar JSON com resultado:
  {
    "synced_users": int,
    "total_transactions_imported": int,
    "errors": [{"user_id": string, "error": string}],
    "duration_ms": int,
    "triggered_at": string
  }
- Logar início e fim com duração total

GET /api/v1/internal/sync/status
- Header obrigatório: X-Sync-Secret
- Retornar status do último sync:
  {
    "last_sync_at": string,
    "last_sync_duration_ms": int,
    "last_sync_users": int,
    "last_sync_transactions": int,
    "next_scheduled_sync": string
  }
- Buscar dados da tabela sync_logs

GET /health
- SEM autenticação (necessário para o keep-alive do cron-job.org)
- Retornar: {"status": "ok", "timestamp": string, "version": "1.0.0"}
- Registrar essa rota FORA do grupo /api/v1 e FORA do JWT middleware

Schema adicional:
CREATE TABLE sync_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    triggered_by TEXT NOT NULL, -- "cron", "manual", "keepalive"
    synced_users INT DEFAULT 0,
    transactions_imported INT DEFAULT 0,
    errors_count INT DEFAULT 0,
    errors_detail JSONB,
    duration_ms INT,
    started_at TIMESTAMPTZ DEFAULT NOW(),
    finished_at TIMESTAMPTZ
);
CREATE INDEX ON sync_logs(started_at DESC);

---

**PARTE 2 — Keep-alive (evitar cold start do Render)**

Criar `backend/internal/jobs/keepalive.go`:

Struct KeepAliveJob:
- selfURL: string (vem de env SELF_URL, ex: https://finance-os-api.onrender.com)
- interval: 10 minutos
- httpClient com timeout de 15s

Start(ctx context.Context):
- Rodar ticker a cada 10 minutos
- Fazer GET em {selfURL}/health
- Logar APENAS quando falhar (não logar pings bem-sucedidos — evitar poluir logs)
- Se falhar 3 vezes consecutivas: logar WARNING com detalhes
- Parar gracefully quando ctx cancelado

Inicializar o KeepAliveJob em cmd/server/main.go junto com os outros jobs.
Só inicializar se env SELF_URL estiver definida (não rodar em desenvolvimento local).

---

**PARTE 3 — Rota de sync manual no handler existente**

Adicionar em `backend/internal/handler/accounts.go`:

POST /api/v1/accounts/sync (rota protegida por JWT — usuário pode forçar sync manual)
- Extrair user_id do JWT
- Chamar SyncUserAccounts(userID) para apenas esse usuário
- Retornar resultado: {transactions_imported, duration_ms, synced_at}
- Rate limit: máximo 1 sync manual por usuário a cada 30 minutos
  (verificar último sync em sync_logs — se < 30min, retornar 429 com mensagem)

---

**PARTE 4 — Variáveis de ambiente necessárias**

Adicionar ao .env.example:
# Sync
SYNC_SECRET=gere-uma-string-aleatoria-forte-aqui  # openssl rand -hex 32
SELF_URL=https://finance-os-api.onrender.com       # URL do serviço no Render

---

**PARTE 5 — Instruções de configuração do cron-job.org**

Criar o arquivo `docs/cronjob-setup.md` com as instruções:

# Configuração do cron-job.org

## Por que é necessário
O Render free tier dorme após 15 minutos de inatividade.
O cron-job.org faz requisições externas que mantêm o serviço acordado
e disparam os syncs nos horários configurados.

## Passo a passo

1. Acesse https://cron-job.org e crie uma conta gratuita

2. Crie 4 jobs de SYNC (um para cada horário):

   Job 1 — Sync manhã:
   - URL: https://finance-os-api.onrender.com/api/v1/internal/sync
   - Método: POST
   - Header: X-Sync-Secret: {seu SYNC_SECRET do .env}
   - Horário: 07:00 (America/Sao_Paulo)
   - Dias: todos

   Job 2 — Sync almoço:
   - URL: https://finance-os-api.onrender.com/api/v1/internal/sync
   - Método: POST
   - Header: X-Sync-Secret: {seu SYNC_SECRET do .env}
   - Horário: 13:00 (America/Sao_Paulo)
   - Dias: todos

   Job 3 — Sync fim de tarde:
   - URL: https://finance-os-api.onrender.com/api/v1/internal/sync
   - Método: POST
   - Header: X-Sync-Secret: {seu SYNC_SECRET do .env}
   - Horário: 19:00 (America/Sao_Paulo)
   - Dias: todos

   Job 4 — Sync noturno:
   - URL: https://finance-os-api.onrender.com/api/v1/internal/sync
   - Método: POST
   - Header: X-Sync-Secret: {seu SYNC_SECRET do .env}
   - Horário: 23:30 (America/Sao_Paulo)
   - Dias: todos

3. Crie 1 job de KEEP-ALIVE (para acordar o serviço antes de cada sync):

   Job Keep-alive — pré-sync:
   - URL: https://finance-os-api.onrender.com/health
   - Método: GET
   - Horário: a cada 14 minutos (*/14 * * * *)
   - Importante: esse job acorda o serviço; sem ele o Render pode estar
     dormindo quando o sync chegar

4. Configure notificações de falha:
   - Em Settings > Notifications: adicionar seu email
   - Notificar quando job falhar 2 vezes consecutivas

## Consumo de updates Pluggy

4 syncs/dia × 30 dias = 120 updates/mês
Limite da Pluggy (dev tier): 240 updates/mês por CPF/instituição
Margem de segurança: 50% — suficiente para syncs manuais ocasionais

## Verificar se está funcionando

Após configurar, verificar em:
GET https://finance-os-api.onrender.com/api/v1/internal/sync/status
Header: X-Sync-Secret: {seu SYNC_SECRET}

O campo last_sync_at deve atualizar após cada job do cron-job.org.

---

**Requisitos gerais:**
- O SYNC_SECRET deve ter pelo menos 32 caracteres — gerar com: openssl rand -hex 32
- Nunca expor o SYNC_SECRET em logs ou respostas de API
- O endpoint /health deve responder em < 100ms (não fazer queries no banco)
- O sync paralelo com semáforo de 3 goroutines evita sobrecarregar a Pluggy API
- Salvar sempre em sync_logs independente de sucesso ou falha
```