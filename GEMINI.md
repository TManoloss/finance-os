# GEMINI.md — Personal Finance OS
> Guia de desenvolvimento com Antigravity + Gemini CLI
> Stack: Go (Echo) · Python · Flutter · Next.js · PostgreSQL · Pluggy API

---

## Contexto do projeto

Personal Finance OS é um app de gestão financeira pessoal/familiar com:
- Conexão automática via Open Finance (Pluggy API)
- Classificação inteligente de transações por LLM
- Agentes de análise diária, semanal e mensal
- App mobile Flutter + Dashboard web Next.js
- Backend API em Go com Echo framework
- Serviço de agentes em Python (FastAPI)
- Banco de dados PostgreSQL (Supabase free tier)
- Deploy: Render free tier (Go + Python separados), Vercel (Next.js)

---

## Estrutura do repositório

```
finance-os/
├── GEMINI.md               ← este arquivo
├── docker-compose.yml
├── backend/                ← Go + Echo
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── auth/
│   │   ├── pluggy/
│   │   ├── transactions/
│   │   ├── cards/
│   │   ├── accounts/
│   │   └── reports/
│   ├── go.mod
│   └── Dockerfile
├── agents/                 ← Python + FastAPI
│   ├── agents/
│   │   ├── daily.py
│   │   ├── weekly.py
│   │   ├── monthly.py
│   │   └── chat.py
│   ├── classifier/
│   │   └── classifier.py
│   ├── main.py
│   └── Dockerfile
├── web/                    ← Next.js 15 App Router
│   ├── app/
│   ├── components/
│   └── package.json
└── mobile/                 ← Flutter
    ├── lib/
    └── pubspec.yaml
```

---

## Como usar este arquivo com Gemini CLI / Antigravity

**Gemini CLI:** Cada bloco `PROMPT:` abaixo pode ser colado diretamente no terminal:
```bash
gemini "$(cat <<'EOF'
<cole o prompt aqui>
EOF
)"
```

**Antigravity:** Crie um agente no Manager para cada seção, usando o prompt correspondente como instrução inicial do agente. Ative paralelismo entre agentes independentes (ex: backend + frontend simultaneamente).

---

## FASE 1 — Infraestrutura base

### 1.1 Docker Compose + PostgreSQL

```
PROMPT:

Você é um engenheiro DevOps especialista em Go e PostgreSQL.

Crie o arquivo `docker-compose.yml` na raiz do projeto `finance-os/` com:

Serviços:
- postgres: imagem postgres:16-alpine, porta 5432, variáveis POSTGRES_DB=financedb POSTGRES_USER=finance POSTGRES_PASSWORD=finance123, volume persistente em ./data/postgres, healthcheck com pg_isready
- adminer: imagem adminer, porta 8080, depende do postgres (para visualização do banco em dev)

Requisitos:
- Usar um único arquivo `.env` na raiz do projeto para centralizar as configurações do Backend (Go) e Agentes (Python).
- Criar o arquivo .gitignore incluindo .env, data/, binários Go e __pycache__
- Criar README.md com instruções de como iniciar o ambiente local com `docker compose up -d`

Não instale nada. Apenas gere os arquivos.
```

---

### 1.2 Schema PostgreSQL completo

```
PROMPT:

Você é um DBA especialista em PostgreSQL para aplicações financeiras.

Crie o arquivo `backend/internal/db/schema.sql` com o schema completo do Personal Finance OS.

Tabelas necessárias:

1. `users` — usuários do sistema (uso familiar, múltiplos usuários)
   - id UUID PK, name TEXT, email TEXT UNIQUE, password_hash TEXT, created_at TIMESTAMPTZ

2. `connected_accounts` — contas bancárias conectadas via Pluggy
   - id UUID PK, user_id FK, pluggy_item_id TEXT, institution_name TEXT, account_type TEXT (checking/savings/credit), balance NUMERIC(12,2), currency TEXT DEFAULT 'BRL', last_synced_at TIMESTAMPTZ, created_at TIMESTAMPTZ

3. `transactions` — transações importadas do Open Finance
   - id UUID PK, account_id FK, pluggy_transaction_id TEXT UNIQUE, amount NUMERIC(12,2), direction TEXT (debit/credit), description TEXT, merchant_name TEXT, category_id FK NULLABLE, date DATE, is_recurring BOOLEAN DEFAULT false, created_at TIMESTAMPTZ

4. `categories` — categorias de gastos (system + custom)
   - id UUID PK, user_id FK NULLABLE (NULL = categoria do sistema), name TEXT, color TEXT, icon TEXT, parent_id FK NULLABLE (para subcategorias), created_at TIMESTAMPTZ
   - Inserir as categorias padrão do sistema via INSERT: Alimentação, Transporte, Saúde, Lazer, Assinaturas, Moradia, Educação, Investimentos, Renda, Outros

5. `installments` — parcelamentos de cartão
   - id UUID PK, transaction_id FK, account_id FK, total_amount NUMERIC(12,2), installments_total INT, installment_current INT, merchant_name TEXT, start_date DATE, next_due_date DATE

6. `agent_reports` — relatórios gerados pelos agentes Python
   - id UUID PK, user_id FK, agent_type TEXT (daily/weekly/monthly), period_start DATE, period_end DATE, summary_markdown TEXT, insights JSONB, created_at TIMESTAMPTZ

7. `category_rules` — regras de classificação automática por merchant
   - id UUID PK, user_id FK NULLABLE, merchant_pattern TEXT, category_id FK, priority INT DEFAULT 0

8. `refresh_tokens` — tokens JWT de refresh
   - id UUID PK, user_id FK, token TEXT UNIQUE, expires_at TIMESTAMPTZ, created_at TIMESTAMPTZ

Requisitos:
- Usar UUID como PK em todas as tabelas (extensão pgcrypto ou gen_random_uuid())
- Criar índices em: transactions(account_id, date), transactions(category_id), transactions(merchant_name), connected_accounts(user_id)
- Adicionar constraints de CHECK onde aplicável (direction IN ('debit','credit'), etc)
- Criar arquivo separado `backend/internal/db/seed.sql` com categorias padrão e um usuário admin de teste
```

---

### 1.3 Skeleton Go com Echo

```
PROMPT:

Você é um engenheiro Go sênior especialista em APIs REST com Echo framework.

Crie o skeleton completo do backend em `backend/` com a seguinte estrutura:

**go.mod** — módulo: github.com/finance-os/backend
Dependências:
- github.com/labstack/echo/v4 (latest)
- github.com/labstack/echo-jwt/v4
- github.com/golang-jwt/jwt/v5
- github.com/jackc/pgx/v5 (PostgreSQL driver)
- github.com/joho/godotenv
- golang.org/x/crypto (bcrypt)

**cmd/server/main.go** — inicialização do servidor:
- Carregar .env com godotenv
- Conectar ao PostgreSQL via pgx pool
- Registrar middlewares: Logger, Recover, CORS (origens configuráveis via env), RequestID
- Registrar rotas (chamar router.Setup)
- Iniciar servidor na porta definida em env PORT (default 8080)
- Graceful shutdown com context e os.Signal

**internal/config/config.go** — struct Config com todos os campos de env:
DatabaseURL, JWTSecret, JWTRefreshSecret, PluggyClientID, PluggyClientSecret, AgentsServiceURL, Port, CORSOrigins

**internal/router/router.go** — Setup(e *echo.Echo, db *pgxpool.Pool, cfg *config.Config):
- Grupo /api/v1
- Rotas públicas: POST /auth/register, POST /auth/login, POST /auth/refresh
- Rotas protegidas (JWT middleware): GET /me, CRUD /accounts, GET /transactions, POST /transactions/sync, GET /reports, GET /cards/installments, POST /chat

**internal/middleware/auth.go** — middleware JWT com validação de claims customizados (user_id, email)

**internal/handler/** — um arquivo por domínio (auth.go, accounts.go, transactions.go, reports.go, cards.go, chat.go) com structs de Request/Response e handlers vazios que retornam 501 Not Implemented com a mensagem "not implemented yet: <nome do handler>"

**internal/response/response.go** — helpers padronizados:
- Success(c echo.Context, status int, data interface{}) error
- Error(c echo.Context, status int, message string) error
- Structs APIResponse{Success bool, Data interface{}, Error string}

**Dockerfile** — multi-stage build: builder com go:1.23-alpine, imagem final alpine:3.20, binário em /app/server, EXPOSE 8080, CMD ["/app/server"]

Requisitos gerais:
- Zero lógica de negócio ainda — apenas estrutura, tipos e handlers stub
- Cada arquivo deve ter comentários de seção explicando o que será implementado
- O código deve compilar sem erros
```

---

## FASE 2 — Autenticação JWT

```
PROMPT:

Você é um engenheiro Go sênior. Implemente o sistema de autenticação JWT no backend Go existente em `backend/`.

Implemente em `internal/handler/auth.go`:

**POST /auth/register**
- Body: {name, email, password}
- Validar: email formato válido, senha mínimo 8 caracteres, nome não vazio
- Hash da senha com bcrypt (custo 12)
- Inserir em `users` via pgx
- Retornar: {user: {id, name, email, created_at}, access_token, refresh_token}

**POST /auth/login**
- Body: {email, password}
- Buscar usuário por email
- Verificar senha com bcrypt.CompareHashAndPassword
- Gerar access_token (JWT, exp 15min, claims: user_id, email)
- Gerar refresh_token (JWT opaco, exp 7 dias) — salvar hash em refresh_tokens
- Retornar: {access_token, refresh_token, expires_in: 900}

**POST /auth/refresh**
- Body: {refresh_token}
- Validar token na tabela refresh_tokens (não expirado, não revogado)
- Revogar o token atual (deletar da tabela)
- Gerar novo par access_token + refresh_token (rotação)
- Retornar: {access_token, refresh_token, expires_in: 900}

**GET /me** (protegida)
- Extrair user_id dos claims JWT
- Buscar usuário no banco
- Retornar: {id, name, email, created_at}

Implementar em `internal/service/auth_service.go` toda a lógica de negócio separada do handler.
Implementar em `internal/repository/user_repository.go` as queries SQL com pgx.

Requisitos:
- Usar pgx.Pool injetado via handler struct (não global)
- Retornar erros tipados (ErrUserNotFound, ErrInvalidCredentials, ErrEmailTaken)
- Logar erros internos mas nunca expor detalhes ao cliente
- Testes unitários básicos para o service em `internal/service/auth_service_test.go`
```

---

## FASE 3 — Integração Pluggy (Open Finance)

```
PROMPT:

Você é um engenheiro Go sênior especialista em integração com APIs de Open Finance.

Implemente o client da Pluggy API em `backend/internal/pluggy/`.

**pluggy/client.go** — struct Client com:
- BaseURL: https://api.pluggy.ai
- ClientID e ClientSecret do config
- httpClient com timeout 30s
- método Authenticate() que faz POST /auth e armazena o APIKey (válido 2h), com renovação automática

**pluggy/models.go** — structs Go mapeando os responses da Pluggy:
- Item (item_id, connector_id, status, created_at, updated_at)
- Account (id, item_id, name, marketing_name, balance, currency_code, type, subtype)
- Transaction (id, account_id, date, description, amount, currency_code, type, category, payment_data)
- Connector (id, name, primary_color, logo_url, type)

**pluggy/accounts.go**:
- GetAccounts(itemID string) ([]Account, error)
- GetAccount(accountID string) (*Account, error)

**pluggy/transactions.go**:
- GetTransactions(accountID string, from, to time.Time, pageSize int) ([]Transaction, error) — com paginação automática (percorrer todas as páginas)
- GetTransaction(transactionID string) (*Transaction, error)

**pluggy/connect_token.go**:
- CreateConnectToken(itemID *string) (string, error) — gera token para o Pluggy Connect Widget (usado no frontend para autorizar novos bancos)

**internal/handler/accounts.go** — implementar:
- POST /accounts/connect-token — gera token para o widget de conexão de banco no frontend
- GET /accounts — lista contas conectadas do usuário (da tabela connected_accounts)
- POST /accounts/sync — dispara sincronização manual de todas as contas do usuário

**internal/service/sync_service.go** — SyncUserAccounts(userID string):
- Buscar todas as connected_accounts do usuário
- Para cada conta, chamar Pluggy.GetTransactions dos últimos 90 dias
- Inserir transações novas (ignorar duplicatas por pluggy_transaction_id com ON CONFLICT DO NOTHING)
- Atualizar saldo em connected_accounts
- Disparar classificação automática (chamar classifier — stub por enquanto)
- Registrar last_synced_at

**internal/jobs/scheduler.go** — cron job diário (00:30 BRT) que chama SyncUserAccounts para todos os usuários com contas conectadas

Requisitos:
- Tratar rate limits da Pluggy (429) com retry exponencial (máx 3 tentativas)
- Logar cada sincronização com duração e número de transações importadas
- O APIKey da Pluggy deve ser cacheado em memória com expiração, nunca no banco
```

---

## FASE 4 — Transações e Classificação

```
PROMPT:

Você é um engenheiro Go sênior. Implemente o módulo de transações e classificação no backend Go em `backend/`.

**internal/handler/transactions.go** — implementar:

GET /transactions
- Query params: account_id?, category_id?, from_date?, to_date?, direction?, page (default 1), page_size (default 50)
- Retornar: {transactions: [...], total: int, page: int, page_size: int, total_pages: int}
- Cada transaction deve incluir: category (objeto completo), account (nome + instituição)

PATCH /transactions/:id/category
- Body: {category_id: string}
- Atualizar categoria de uma transação manualmente
- Criar/atualizar regra em category_rules para o merchant_name desta transação
- Retornar a transação atualizada

GET /transactions/summary
- Query params: from_date, to_date (default: mês atual)
- Retornar: {
    total_spent: float,
    total_received: float,
    by_category: [{category_id, category_name, color, total, percentage, transaction_count}],
    by_day: [{date, total_spent, total_received}],
    top_merchants: [{merchant_name, total, count}]
  }

**internal/service/classifier_service.go** — ClassifyTransaction(tx *Transaction) (categoryID string):
1. Verificar se existe regra exata em category_rules para tx.merchant_name (case insensitive)
2. Se não, verificar regras com LIKE pattern
3. Se não, chamar o serviço Python de agentes via HTTP POST para /classify
   - Body: {merchant_name, description, amount, direction}
   - Retornar category_id
4. Se o serviço Python falhar, usar categoria "Outros" como fallback

**internal/repository/transaction_repository.go**:
- GetTransactions(filters TransactionFilters) ([]Transaction, int, error)
- GetSummary(userID string, from, to time.Time) (*Summary, error)
- UpdateCategory(txID, categoryID string) error
- InsertBatch(transactions []Transaction) error (com ON CONFLICT DO NOTHING)

Requisitos:
- Todas as queries devem filtrar por user_id implicitamente via JOIN com connected_accounts
- GetSummary deve usar uma única query com CTEs para performance
- Usar pgx batch para InsertBatch
```

---

## FASE 5 — Módulo de Cartões e Parcelamentos

```
PROMPT:

Você é um engenheiro Go sênior especialista em finanças pessoais.

Implemente o módulo de cartões e parcelamentos no backend Go em `backend/internal/`.

**internal/handler/cards.go**:

GET /cards/installments
- Listar todos os parcelamentos ativos do usuário
- Incluir: merchant_name, valor_parcela, parcela_atual/total, próximo_vencimento, valor_restante_total
- Ordenar por próximo_vencimento ASC

GET /cards/invoice/:account_id
- Calcular fatura projetada do cartão de crédito
- Retornar: {
    account: {...},
    due_date: date,
    closed_amount: float (gastos já fechados),
    open_amount: float (gastos do ciclo atual),
    installments_total: float (soma de todas as parcelas futuras),
    subscriptions_total: float (recorrentes detectados),
    projected_total: float (soma de tudo),
    breakdown: [{description, amount, type}]
  }

GET /cards/subscriptions
- Detectar transações recorrentes (mesmo merchant, valor similar, periodicidade mensal)
- Retornar lista de assinaturas detectadas com: merchant, valor_mensal, próximo_cobrança_estimado, status (ativa/irregular)

**internal/service/installments_service.go**:
- DetectAndSaveInstallments(transactions []Transaction) — detectar transações parceladas pelo padrão "(X/Y)" na descrição ou pelo campo payment_data da Pluggy, criar/atualizar registros em installments
- GetProjectedInvoice(accountID string, referenceMonth time.Time) (*Invoice, error)

**internal/service/subscription_service.go**:
- DetectSubscriptions(userID string) ([]Subscription, error) — agrupar transações por merchant_name nos últimos 90 dias, identificar padrão mensal (cobrança no mesmo período ± 5 dias), retornar assinaturas com confiança > 80%

Requisitos:
- Detecção de parcelamentos deve funcionar tanto por regex na descrição quanto pelo campo da Pluggy
- Assinaturas "esquecidas" (sem cobrança nos últimos 45 dias) devem ter status "irregular"
- Todas as respostas devem incluir o breakdown detalhado para transparência
```

---

## FASE 6 — Serviço Python de Agentes

```
PROMPT:

Você é um engenheiro Python sênior especialista em LLMs e agentes de análise financeira.

Crie o serviço de agentes em `agents/` com FastAPI.

**Dependências** (requirements.txt):
fastapi, uvicorn[standard], anthropic, pydantic, httpx, python-dotenv, apscheduler

**main.py** — app FastAPI:
- POST /classify — classificar transação por merchant + descrição
- POST /agents/daily — executar agente diário para um user_id
- POST /agents/weekly — executar agente semanal
- POST /agents/monthly — executar agente mensal
- POST /chat — responder pergunta financeira em linguagem natural
- GET /health — healthcheck

**classifier/classifier.py** — ClassifierService:
- Recebe: {merchant_name, description, amount, direction}
- Prompt para Claude: dado o merchant e descrição, retornar o category_id mais adequado da lista fixa de categorias
- Usar cache simples em dict para merchants já classificados (evitar chamadas repetidas)
- Retornar: {category_id, category_name, confidence}

**agents/base_agent.py** — classe BaseAgent com:
- método build_context(user_id, period) — buscar dados do PostgreSQL para o período
- método run(user_id, period) — abstrato
- método save_report(user_id, agent_type, period, summary, insights) — salvar em agent_reports
- cliente Anthropic injetado no construtor

**agents/daily.py** — DailyAgent(BaseAgent):
Analisar transações das últimas 24h e identificar:
1. Gastos fora do padrão histórico (> 2x a média diária do mesmo dia da semana)
2. Cobranças duplicadas (mesmo merchant, mesmo valor, mesmo dia)
3. Novas assinaturas detectadas
4. Total gasto vs média dos últimos 30 dias

Prompt template para o Claude (incluir no arquivo como constante DAILY_PROMPT):
```
Você é um assistente financeiro pessoal analisando as transações de hoje do usuário.

Dados do dia {date}:
{transactions_json}

Histórico de referência (média dos últimos 30 dias):
{historical_context}

Analise e retorne um JSON com:
- summary: resumo em português (2-3 frases, tom amigável, primeira pessoa como se fosse um assistente)
- alerts: lista de alertas [{type, message, amount, severity: low/medium/high}]
- total_spent: float
- vs_average_percent: float (variação % em relação à média)
- insights: lista de observações relevantes

Seja direto e útil. Não use linguagem corporativa.
```

**agents/weekly.py** — WeeklyAgent(BaseAgent):
Analisar semana atual vs semana anterior e vs média das últimas 4 semanas:
- Variação por categoria (quais categorias cresceram/caíram)
- Top 5 maiores gastos da semana
- Dias da semana com maior gasto
- Progresso em relação ao orçamento mensal (se definido)

**agents/monthly.py** — MonthlyAgent(BaseAgent):
Análise completa do mês:
- Fechamento: total gasto, total recebido, saldo
- Comparativo com mês anterior e média dos últimos 3 meses
- Categorias acima do esperado
- Projeção para o próximo mês baseada em parcelamentos + assinaturas
- Top merchants do mês
- Tendências de gastos

**agents/chat.py** — ChatAgent(BaseAgent):
- Recebe: {user_id, message, conversation_history: [...]}
- Busca contexto relevante baseado na pergunta (últimas transações, resumo do mês, etc)
- Responde em linguagem natural em português
- Exemplos de perguntas suportadas: "quanto gastei com iFood esse mês?", "qual foi meu maior gasto essa semana?", "estou gastando mais que o normal?"
- Manter contexto da conversa (passar histórico para o Claude)

Requisitos:
- Conexão com PostgreSQL via asyncpg (não usar ORM)
- Variáveis de ambiente: ANTHROPIC_API_KEY, DATABASE_URL, GO_BACKEND_URL
- Usar claude-sonnet-4-20250514 como modelo
- Max tokens: 1024 para classify, 2048 para agentes, 1024 para chat
- Logar tempo de execução de cada chamada LLM
```

---

## FASE 7 — Dashboard Next.js

```
PROMPT:

Você é um engenheiro frontend sênior especialista em Next.js 15 e design de interfaces financeiras.

Crie o dashboard web em `web/` usando Next.js 15 App Router.

**Setup inicial:**
- TypeScript, Tailwind CSS v4, shadcn/ui
- next-auth v5 para autenticação (JWT strategy, provider credentials)
- Axios com interceptor para refresh token automático
- Recharts para gráficos
- date-fns para formatação de datas
- numeral.js para formatação de valores monetários (R$ 1.234,56)

**Design system** (aplicar em globals.css e tailwind.config):
Tema escuro exclusivo — sem toggle light/dark. Paleta:
- Background primário: #0A0A0F (quase preto com toque azul)
- Background cards: #111118
- Background elevated: #1A1A24
- Border sutil: #2A2A3A
- Texto primário: #F0F0F5
- Texto secundário: #8888A0
- Accent primário: #7C6FFF (roxo elétrico — cor principal do Pierre)
- Accent secundário: #4ECDC4 (teal para receitas/positivo)
- Danger: #FF6B6B
- Warning: #FFD93D
- Success: #6BCB77
- Fonte: Inter (Google Fonts)

**Páginas e rotas:**

`/` (redirect para /dashboard se logado, /login se não)

`/login` — página de login:
- Logo + nome do app centralizado
- Form com email e senha
- Botão "Entrar" com loading state
- Background com gradiente sutil

`/dashboard` — página principal:
- Header: logo, nome do usuário, avatar inicial
- Sidebar colapsável: Dashboard, Transações, Cartões, Relatórios, Configurações
- Cards de resumo no topo: Saldo total, Gasto no mês, Receita no mês, Economizado
- Gráfico de gastos por categoria (donut chart com Recharts)
- Gráfico de gastos por dia do mês (area chart)
- Lista das últimas 10 transações com ícone de categoria e valor colorido (vermelho=débito, verde=crédito)
- Card "Alertas dos Agentes" mostrando o último relatório diário

`/transactions` — extrato completo:
- Filtros: conta, categoria, período (date picker), direção
- Tabela com paginação
- Coluna de categoria clicável para reclassificar (modal com select de categorias)
- Exportar para CSV (client-side)

`/cards` — cartões e parcelamentos:
- Lista de cartões conectados com saldo devedor
- Fatura projetada por cartão
- Lista de parcelamentos ativos com progress bar (X/Y parcelas)
- Lista de assinaturas detectadas com próxima cobrança

`/reports` — relatórios dos agentes:
- Tabs: Diário, Semanal, Mensal
- Card com o relatório mais recente em markdown renderizado
- Histórico de relatórios anteriores
- Chat box no rodapé para perguntas ao agente LLM

`/settings` — configurações:
- Seção "Contas conectadas": lista + botão "Conectar novo banco" (embed do Pluggy Connect Widget)
- Seção "Perfil": alterar nome e senha

**Componentes reutilizáveis** em `components/`:
- TransactionRow — linha de transação com ícone, merchant, data, valor
- CategoryBadge — badge colorido com ícone da categoria
- AmountDisplay — valor formatado em BRL, colorido por direção
- AgentInsightCard — card do relatório do agente com severity badge
- ConnectBankButton — botão que abre o Pluggy Connect Widget
- LoadingSkeleton — skeleton para loading states

**lib/api.ts** — cliente Axios configurado com baseURL do env, interceptor de refresh token

Requisitos:
- 100% dark mode (sem suporte a light)
- Responsive: funcionar em telas de 375px até 1440px
- Loading skeletons em todos os dados assíncronos
- Error boundaries com mensagens amigáveis
- Usar Server Components onde possível, Client Components apenas para interatividade
```

---

## FASE 8 — App Flutter

```
PROMPT:

Você é um engenheiro Flutter sênior especialista em apps financeiros.

Crie o app mobile em `mobile/` com Flutter.

**pubspec.yaml** — dependências:
- dio: HTTP client com interceptors
- flutter_secure_storage: armazenar JWT de forma segura
- riverpod + hooks_riverpod: state management
- go_router: navegação declarativa
- fl_chart: gráficos
- intl: formatação de datas e moeda (pt_BR)
- cached_network_image: imagens com cache
- shimmer: loading skeletons
- flutter_svg: ícones SVG

**Design system** — aplicar em ThemeData:
Mesmo tema escuro do Next.js:
- scaffoldBackgroundColor: Color(0xFF0A0A0F)
- cardColor: Color(0xFF111118)
- Accent: Color(0xFF7C6FFF)
- Fonte: Inter (via google_fonts)
- Border radius padrão: 16px nos cards, 12px nos botões
- Cards com borda sutil: Border.all(color: Color(0xFF2A2A3A), width: 1)

**Telas:**

`SplashScreen` — logo animado, verificar token salvo, redirecionar

`LoginScreen` — design idêntico ao web, email + senha, botão com loading

`HomeScreen` (BottomNavigationBar com 4 tabs):
1. Dashboard — igual ao web: saldo, gráfico donut, transações recentes, alerta do agente
2. Transações — lista com pull-to-refresh, filtros por swipe (bottom sheet), pesquisa
3. Cartões — fatura projetada, parcelamentos em cards deslizáveis, assinaturas
4. Chat — interface de chat com o agente LLM (estilo WhatsApp, bolhas de mensagem)

**Widgets reutilizáveis:**
- TransactionTile — ListTile customizado com ícone de categoria colorido
- SummaryCard — card de métrica com ícone, label, valor, variação %
- DonutChart — wrapper do fl_chart para gastos por categoria
- AgentMessageBubble — bolha de chat do agente com avatar
- LoadingShimmer — shimmer cards durante carregamento
- CurrencyText — widget que formata sempre em R$ com estilo correto

**Providers (Riverpod):**
- authProvider — estado de autenticação, funções login/logout
- accountsProvider — contas conectadas
- transactionsProvider — transações com filtros
- summaryProvider — resumo financeiro do mês
- chatProvider — histórico de conversa com o agente

**lib/services/api_service.dart** — DioClient:
- baseUrl do .env
- Interceptor para adicionar Authorization header
- Interceptor para refresh automático do token (401 → refresh → retry)
- Métodos tipados para cada endpoint da API Go

Requisitos:
- Suporte a iOS e Android
- Tela de login com keyboard handling correto (scroll quando teclado aparece)
- Pull-to-refresh em todas as listas
- Haptic feedback nos botões principais
- Splash screen nativa (não flutter)
- App icon dark com logo do projeto
```

---

## FASE 9 — UI Design System (Prompt para gerador de interface)

```
PROMPT PARA GERADOR DE UI (v0, Lovable, Builder.io, ou similar):

Design a complete dark-theme personal finance app UI system inspired by Pierre Finance, with the following exact specifications:

**Color palette (non-negotiable):**
- Background primary: #0A0A0F
- Background surface: #111118
- Background elevated: #1A1A24
- Border: #2A2A3A (1px, subtle)
- Text primary: #F0F0F5
- Text secondary: #8888A0
- Accent purple: #7C6FFF (primary brand color, use for CTAs, active states, charts accent)
- Accent teal: #4ECDC4 (positive values, income, success states)
- Red: #FF6B6B (expenses, negative values, alerts)
- Yellow: #FFD93D (warnings, neutral alerts)
- Font: Inter, weights 400/500/600

**Typography scale:**
- Display: 32px/500 (total balance)
- Heading: 20px/600
- Subheading: 16px/500
- Body: 14px/400
- Caption: 12px/400, color: #8888A0
- Mono amounts: 18px/600, font-variant-numeric: tabular-nums

**Component specifications:**

1. DASHBOARD SCREEN (mobile 390px wide):
- Status bar area: dark
- Top section: greeting ("Bom dia, Manoel"), date subtitle
- Balance card: full-width card (#111118), large total balance in white, subtitle "Saldo total" in #8888A0, small +/- monthly variation badge in teal/red
- Quick stats row: 3 cards side-by-side — "Gasto no mês" (red accent), "Recebido" (teal accent), "Economizado" (purple accent). Each: label caption + large value
- Section "Gastos por categoria": donut chart (purple/teal/red/yellow segments) + legend list below
- Section "Últimas transações": list of transaction rows
- Bottom nav: 4 icons (home, list, credit-card, chat), active state = purple with glow

2. TRANSACTION ROW component:
- Left: circular icon (40px) with category color background (10% opacity) + category emoji/icon in full color
- Center: merchant name (body/500), category label (caption), date (caption right-aligned)
- Right: amount (mono/600, red for debit, teal for credit)
- Separator: 1px #2A2A3A

3. AGENT INSIGHT CARD:
- Card with left border 3px colored by severity (yellow=info, red=high, purple=insight)
- Top: icon + "Agente Diário" label + timestamp caption
- Body: insight text in body size
- Subtle background matching severity color at 5% opacity

4. CHAT INTERFACE:
- Dark background #0A0A0F
- User messages: right-aligned, #7C6FFF bubble, white text, border-radius 18px 18px 4px 18px
- Agent messages: left-aligned, #1A1A24 bubble, #F0F0F5 text, border-radius 18px 18px 18px 4px
- Agent avatar: small circle with "F" (Finance OS logo)
- Input bar: #111118 background, #2A2A3A border, send button in #7C6FFF

5. INVOICE/CARDS SCREEN:
- Card visual: physical card representation with gradient (#1A1A24 to #2A2A3A), card number masked (•••• 4521), bank name, card network logo area
- Below: "Fatura projetada" section with breakdown rows
- Installments: progress bar (purple fill, #2A2A3A track), "3/12 parcelas" label

6. WEB DASHBOARD (1440px):
- Left sidebar: 240px, #0D0D14 background, logo top, nav items with active purple highlight + left border accent
- Main area: #0A0A0F background, content max-width 1200px centered
- Top row: 4 metric cards
- Second row: 60/40 split — area chart left, donut chart right
- Third row: full-width transactions table with subtle row hover (#1A1A24)

**Interaction states:**
- Hover: background lighten by 5% (#1A1A24 → #1E1E2C)
- Active/pressed: accent color at 20% opacity background
- Focus: 2px #7C6FFF outline, offset 2px
- Disabled: 40% opacity

**Micro-details that match Pierre's quality:**
- All monetary values use non-breaking space between R$ and number
- Negative amounts always show minus sign (not parentheses)
- Dates: "hoje", "ontem", or "DD/MM" for older
- Category icons use emoji for mobile, outlined icons for web
- Loading states: shimmer animation with #1A1A24 → #222230 → #1A1A24 gradient sweep
- Empty states: centered illustration (simple SVG), heading, subtext, primary CTA button

Generate all screens with pixel-perfect dark mode, no light mode toggle, production-ready quality matching fintech apps like Nubank and Pierre.
```

---

## FASE 10 — Deploy e CI/CD

```
PROMPT:

Você é um engenheiro DevOps especialista em deploy de aplicações Go e Python no Render.

Configure o deploy completo do Personal Finance OS.

**render.yaml** — na raiz do projeto, definindo:

Serviço 1 — Go API:
- type: web
- name: finance-os-api
- runtime: docker
- dockerfilePath: ./backend/Dockerfile
- envVars: DATABASE_URL (fromDatabase), JWT_SECRET, JWT_REFRESH_SECRET, PLUGGY_CLIENT_ID, PLUGGY_CLIENT_SECRET, AGENTS_SERVICE_URL, PORT=8080, CORS_ORIGINS

Serviço 2 — Python Agents:
- type: web
- name: finance-os-agents
- runtime: docker
- dockerfilePath: ./agents/Dockerfile
- envVars: DATABASE_URL, ANTHROPIC_API_KEY, GO_BACKEND_URL, PORT=8000

Database:
- type: pgsql
- name: finance-os-db
- plan: free
- databaseName: financedb

**Self-ping para evitar cold start do Render free tier:**
Em `backend/internal/jobs/keepalive.go`:
- Cron a cada 10 minutos que faz GET para a própria URL (env SELF_URL)
- Logar apenas quando o ping falhar

**backend/Dockerfile** melhorado:
- Cache de dependências Go no layer intermediário
- Build com CGO_DISABLED=1 GOOS=linux para binário estático
- Health check no Dockerfile: HEALTHCHECK CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

**agents/Dockerfile:**
- Base: python:3.12-slim
- Cache de pip no layer de dependências
- Usuário não-root para segurança
- CMD: uvicorn main:app --host 0.0.0.0 --port $PORT

**GitHub Actions** em `.github/workflows/ci.yml`:
- Trigger: push em main e pull_request
- Jobs paralelos:
  - test-go: go test ./... com coverage
  - test-python: pytest agents/tests/
  - lint-go: golangci-lint
  - lint-python: ruff check
- Deploy automático no merge para main (via Render deploy hook)

**.env.example** completo com todos os campos necessários e comentários explicando cada variável

Requisitos:
- Render free tier dorme após 15 min de inatividade — o keepalive resolve isso
- Banco de dados Supabase pode ser usado no lugar do banco Render (apenas mudar DATABASE_URL)
- Documentar no README.md como configurar todas as variáveis no painel do Render
```

---

## Ordem de execução recomendada no Antigravity

```
Agentes paralelos (Antigravity Manager):

Semana 1:
├── Agente A: Fase 1.1 + 1.2 (Docker + Schema)
└── Agente B: Fase 1.3 (Skeleton Go)

Semana 2:
├── Agente A: Fase 2 (Auth JWT)
└── Agente B: Fase 6 (início do serviço Python — classify endpoint)

Semana 3:
├── Agente A: Fase 3 (Pluggy integration)
└── Agente B: Fase 7 (Next.js dashboard — auth + layout)

Semana 4:
├── Agente A: Fase 4 (Transações + classificação)
└── Agente B: Fase 7 cont. (páginas dashboard + transações)

Semana 5:
├── Agente A: Fase 5 (Cartões + parcelamentos)
└── Agente B: Fase 6 cont. (agentes daily/weekly/monthly)

Semana 6:
├── Agente A: Fase 8 (Flutter)
└── Agente B: Fase 10 (Deploy + CI/CD)
```

---

## Notas de desenvolvimento

- **Echo vs Gin**: Echo foi escolhido por ter melhor suporte nativo a middleware, binding de request e grupos de rotas. Usar `echo.Context` em todos os handlers, nunca `http.Request` direto.
- **Pluggy Connect Widget**: O widget de conexão de banco é um JavaScript embed da Pluggy. No Flutter, usar `webview_flutter` para abrir o widget em um modal. No Next.js, usar o script oficial da Pluggy.
- **Custos**: Com uso pessoal/familiar (~4 usuários, sync diário), o custo estimado na API Anthropic é < US$5/mês usando claude-sonnet-4-20250514.
- **Segurança**: Nunca expor o PLUGGY_CLIENT_SECRET no frontend. O connect token é gerado sempre pelo backend. JWTs com rotação de refresh token.
- **Dados**: Todas as queries filtram implicitamente por user_id via JOIN. Nunca retornar dados de outros usuários.

## FASE 11 — Inteligência Financeira Avançada ("Você vê o que normalmente não veria")

> Esta fase transforma o app de um visualizador de dados em um sistema de inteligência financeira real.
> Cada módulo abaixo é um agente ou feature independente que pode ser desenvolvido em paralelo.

---

### 11.1 Cashflow Timeline — Linha do tempo real do dinheiro

```
PROMPT:

Você é um engenheiro Python sênior especialista em análise financeira.

Implemente o módulo `agents/cashflow_timeline.py` e o endpoint Go correspondente.

**O que é:** Um gráfico de linha mostrando o saldo real dia a dia — não agregado por mês, mas cada dia individual. O usuário vê exatamente quando o dinheiro entrou, quando saiu, e como o saldo evoluiu ao longo do tempo.

**agents/cashflow_timeline.py** — CashflowTimelineService:

`build_daily_cashflow(user_id, from_date, to_date)` → lista de DailyBalance:
- Para cada dia no período: saldo_inicio_dia, total_entradas, total_saidas, saldo_fim_dia
- Calcular saldo_inicio usando o saldo atual e trabalhando retroativamente com as transações
- Incluir: maior_gasto_do_dia (merchant + valor), maior_receita_do_dia
- Marcar dias "críticos": saldo abaixo de um threshold configurável (ex: R$500)

`detect_cashflow_patterns(user_id)` → CashflowPatterns:
- salary_day: dia do mês em que normalmente cai a renda (detectar por maior crédito recorrente)
- low_balance_days: lista de dias do mês historicamente com saldo baixo
- pre_salary_stress_period: quantos dias antes do salário o saldo fica abaixo de R$500
- peak_spending_days: dias da semana com maior gasto (ex: "você gasta 60% mais às sextas")
- monthly_cashflow_cycle: descrição em texto do ciclo financeiro pessoal do usuário

**backend/internal/handler/reports.go** — adicionar:
GET /reports/cashflow?from=&to=
- Retornar o cashflow diário com os padrões detectados
- Incluir: projeção dos próximos 7 dias baseada nos padrões históricos

**web/components/CashflowChart.tsx** — área chart com:
- Linha de saldo ao longo do tempo
- Área preenchida: verde quando acima da média, vermelho quando abaixo
- Marcadores nos dias críticos (círculo vermelho com tooltip)
- Marcador no dia do salário (ícone de entrada)
- Tooltip ao hover: saldo, entradas, saídas, maior gasto do dia
- Linha pontilhada para os 7 dias futuros projetados

**Flutter: lib/widgets/cashflow_timeline_chart.dart**
- Mesmo comportamento usando fl_chart LineChart
- Gestos: pinch para zoom, scroll horizontal para navegar no histórico
```

---

### 11.2 Agente de Padrões de Comportamento (Behavioral Intelligence)

```
PROMPT:

Você é um engenheiro Python sênior especialista em análise comportamental financeira.

Implemente `agents/behavioral_agent.py` — o agente que detecta padrões de comportamento
que o usuário nunca perceberia olhando para os números brutos.

**BehavioralAgent(BaseAgent)**:

`analyze_emotional_spending(user_id)` → EmotionalSpendingReport:
Detectar correlação entre dia da semana/hora e valor dos gastos:
- Calcular média de gasto por dia da semana (seg-dom)
- Calcular média de gasto por faixa horária (manhã/tarde/noite/madrugada)
- Identificar os 3 padrões mais fortes (ex: "Você gasta 73% mais às sextas à noite")
- Detectar "revenge spending": sequência de gastos altos após período de contenção
- Retornar insights em texto gerado pelo Claude explicando o padrão em linguagem humana

`detect_lifestyle_inflation(user_id)` → LifestyleInflationReport:
Comparar média de gastos por categoria nos últimos 3 meses vs 3 meses anteriores:
- Identificar categorias com crescimento > 20% sem aumento de renda correspondente
- Calcular "lifestyle inflation rate": percentual de aumento dos gastos discricionários
- Projetar impacto anual se o padrão continuar
- Exemplo de output: "Seus gastos com lazer cresceram R$340/mês nos últimos 3 meses.
  Se mantido, você gastará R$4.080 a mais por ano nessa categoria."

`calculate_opportunity_cost(user_id)` → OpportunityCostReport:
Para cada hábito de gasto recorrente detectado:
- Calcular custo mensal, anual, e em 5 anos
- Calcular o equivalente em investimento (usar taxa SELIC atual como referência: 10.5% a.a.)
- Exemplo: "Seu hábito de delivery (R$847/mês) custará R$10.164/ano.
  Investido à taxa SELIC, seriam R$64.200 em 5 anos."
- Não ser moralista — apresentar como informação, não como julgamento

`analyze_spending_velocity(user_id)` → SpendingVelocityReport:
- Calcular em que dia do mês normalmente 25%, 50%, 75% do orçamento mensal é consumido
- Detectar se há aceleração de gastos no fim do mês ("corrida para gastar")
- Comparar velocidade atual do mês com meses anteriores
- Alert se o usuário vai estourar o padrão histórico antes do fim do mês

Prompt template BEHAVIORAL_INSIGHTS_PROMPT (incluir como constante):
```
Você é um analista financeiro pessoal com acesso ao histórico completo de transações do usuário.

Analise os seguintes padrões detectados algoritmicamente:
{patterns_json}

Gere insights em português, tom direto mas empático, primeira pessoa como se fosse um
conselheiro financeiro de confiança. Não use linguagem corporativa. Seja específico com
números reais. Máximo 3 insights por análise, ordenados por impacto financeiro.

Formato de resposta JSON:
{
  "insights": [
    {
      "title": "título curto do insight",
      "description": "explicação em 2-3 frases com números específicos",
      "impact_monthly": float,
      "impact_annual": float,
      "type": "pattern|warning|opportunity",
      "severity": "info|medium|high"
    }
  ],
  "summary": "resumo geral em 1 frase"
}
```

Novo endpoint em `main.py`:
POST /agents/behavioral — executar análise comportamental completa para um user_id
Retornar: emotional_spending, lifestyle_inflation, opportunity_costs, spending_velocity
```

---

### 11.3 Detector de Gastos Invisíveis

```
PROMPT:

Você é um engenheiro Python sênior. Implemente `agents/invisible_spending_agent.py`
— o agente que encontra dinheiro sendo perdido silenciosamente.

**InvisibleSpendingAgent(BaseAgent)**:

`detect_forgotten_subscriptions(user_id)` → lista de ForgottenSubscription:
- Buscar todas as assinaturas detectadas (cobranças recorrentes mensais)
- Cruzar com frequência de uso: se o merchant é apenas de assinatura (Netflix, Spotify, etc.)
  não tem como detectar uso, então marcar como "não verificável"
- Para merchants que TAMBÉM aparecem como uso direto (ex: academia que cobra mensalidade),
  verificar se há transações de uso além da mensalidade
- Calcular: custo_mensal, custo_anual, meses_ativo, total_gasto_historico
- Ordenar por "custo do esquecimento" (valor × tempo sem atenção do usuário)

`detect_duplicate_charges(user_id, lookback_days=30)` → lista de DuplicateCharge:
- Buscar transações com mesmo merchant_name E mesmo valor nos últimos lookback_days
- Calcular janela de suspeita: mesma cobrança em até 5 dias de diferença
- Incluir: merchant, valor, datas, total_cobrado_duplicado
- Severity: HIGH se valor > R$50, MEDIUM se > R$20, LOW se menor

`detect_price_increases(user_id)` → lista de PriceIncrease:
- Para cada assinatura recorrente, comparar valor atual vs valor 6 meses atrás
- Detectar aumentos silenciosos (sem ação do usuário)
- Calcular impacto mensal e anual do aumento
- Exemplo: "Netflix aumentou R$7/mês (de R$39,90 para R$46,90). Custo anual: +R$84"

`detect_unused_services(user_id)` → lista de UnusedService:
- Detectar merchants com padrão de "cobrança mensal mas zero uso" (academias, apps)
- Identificar cobranças de apps que provavelmente têm versão gratuita (usar lista curada)
- Detectar cobranças em horário incomum (madrugada) que podem indicar serviços esquecidos

`calculate_total_invisible_waste(user_id)` → InvisibleWasteSummary:
- Somar tudo: duplicatas + assinaturas esquecidas + aumentos não percebidos
- Retornar: total_mensal_perdido, total_anual_perdido, número de itens encontrados
- Gerar com Claude um parágrafo de "você está perdendo R$X/mês em gastos invisíveis"

Endpoint: POST /agents/invisible-spending
UI: Card de destaque no dashboard — "💸 R$ X perdidos em gastos invisíveis" que abre modal com breakdown
```

---

### 11.4 Motor de Projeção Financeira

```
PROMPT:

Você é um engenheiro Python sênior especialista em modelagem financeira.

Implemente `agents/projection_engine.py` — o motor que projeta o futuro financeiro
baseado em dados reais, não em orçamentos manuais.

**ProjectionEngine(BaseAgent)**:

`project_end_of_month(user_id)` → EndOfMonthProjection:
Projetar quanto vai sobrar/faltar até o fim do mês:
- Gastos fixos confirmados restantes: parcelas, assinaturas com data conhecida
- Gastos variáveis projetados: média histórica dos dias restantes × dias restantes
- Receitas esperadas: detectar salário esperado pela data histórica de entrada
- Resultado: {saldo_atual, gastos_fixos_restantes, gastos_variaveis_estimados,
              receitas_esperadas, saldo_projetado_fim_mes, confianca_percent}
- Confiança baseada em: quão consistente é o padrão histórico

`project_next_3_months(user_id)` → lista de MonthProjection:
Para cada um dos próximos 3 meses:
- Parcelas abertas que vencem nesse mês (dados reais do módulo de cartões)
- Assinaturas confirmadas
- Sazonalidade: histórico do mesmo mês em anos anteriores (se disponível)
- Gastos variáveis: média móvel ponderada (meses recentes têm mais peso)
- Retornar: total_comprometido (fixo), total_estimado (variável), saldo_projetado

`project_large_expense_impact(user_id, amount, date)` → ImpactProjection:
Simular o impacto de uma compra planejada:
- "Se eu comprar R$3.000 parcelado em 12x agora, como fica meu cashflow?"
- Calcular: impacto_mensal, meses_de_impacto, novo_saldo_projetado_por_mes
- Comparar com padrão histórico: "Você ficaria abaixo do seu saldo mínimo histórico
  em 3 dos próximos 6 meses"

`detect_financial_risks(user_id)` → lista de FinancialRisk:
- RISCO ALTO: saldo projetado fim do mês < 0
- RISCO MÉDIO: saldo projetado < saldo_mínimo_histórico do usuário
- RISCO BAIXO: comprometimento de renda > 70% em parcelamentos
- OPORTUNIDADE: sobra projetada > média histórica (sugerir guardar a diferença)

Endpoint Go: GET /reports/projection?months=3
Componente web: gráfico de barras empilhadas — fixo (roxo escuro) + variável estimado (roxo claro) + receita esperada (teal) por mês, com linha de saldo resultante
```

---

### 11.5 Feed de Eventos Financeiros (Financial Activity Feed)

```
PROMPT:

Você é um engenheiro full-stack. Implemente o Financial Activity Feed —
um timeline estilo feed de notícias mostrando eventos financeiros relevantes,
não apenas transações brutas.

**backend/internal/service/feed_service.go** — FeedEvent:
```go
type FeedEvent struct {
    ID          string
    Type        FeedEventType
    Title       string
    Description string
    Amount      *float64
    Severity    string // info, warning, alert
    RelatedTx   []string
    CreatedAt   time.Time
    ReadAt      *time.Time
}

type FeedEventType string
const (
    EventDuplicateCharge    FeedEventType = "duplicate_charge"
    EventUnusualSpending    FeedEventType = "unusual_spending"
    EventSubscriptionChange FeedEventType = "subscription_change"
    EventNewMerchant        FeedEventType = "new_merchant"
    EventMilestone          FeedEventType = "milestone"
    EventInstallmentAlert   FeedEventType = "installment_alert"
    EventSalaryDetected     FeedEventType = "salary_detected"
    EventLowBalance         FeedEventType = "low_balance"
    EventMonthlyClose       FeedEventType = "monthly_close"
    EventAgentInsight       FeedEventType = "agent_insight"
)
```

`GenerateFeedEvents(userID string, transactions []Transaction)` → []FeedEvent:
Gerar eventos automaticamente a cada sincronização:
- Para cada transação nova: verificar se é merchant novo (primeira vez), duplicata, gasto incomum
- Detectar salário: crédito grande no padrão histórico → EventSalaryDetected
- Saldo abaixo do threshold → EventLowBalance
- Parcela final de um parcelamento → EventMilestone ("Você quitou o parcelamento da TV 🎉")
- Parcelamento com 1 mês restante → EventInstallmentAlert

**Endpoints:**
GET /feed?page=1&unread_only=false — listar feed paginado, mais recentes primeiro
PATCH /feed/:id/read — marcar como lido
PATCH /feed/read-all — marcar todos como lidos
GET /feed/unread-count — contador para badge no app

**web/components/ActivityFeed.tsx:**
- Lista vertical de cards de evento
- Cada card: ícone colorido por tipo, título, descrição, timestamp relativo ("há 2 horas")
- Badge vermelho no ícone de sino no header com contagem de não lidos
- Filtros: Todos | Alertas | Insights | Transações
- Animação de entrada (slide + fade) para novos eventos

**Flutter: lib/screens/feed_screen.dart:**
- Lista com pull-to-refresh
- Swipe para marcar como lido
- Notificação push para eventos de severity "alert" (usar flutter_local_notifications)
```

---

### 11.6 Comparador "Você vs Você Mesmo"

```
PROMPT:

Você é um engenheiro Python sênior. Implemente `agents/comparison_agent.py`
— comparações do usuário com ele mesmo em diferentes períodos.

**ComparisonAgent(BaseAgent)**:

`compare_period(user_id, period_a_start, period_a_end, period_b_start, period_b_end)` → PeriodComparison:
Comparar dois períodos quaisquer (semana vs semana, mês vs mês, etc.):
- Total gasto: diferença absoluta e percentual
- Por categoria: quais cresceram, quais reduziram, quais são novas
- Top merchants: novos no período B que não estavam no A, e vice-versa
- Dias mais caros: comparação dos dias mais pesados de cada período
- Resumo gerado pelo Claude: "Comparando essa semana com a anterior, você gastou
  R$234 a mais, principalmente em alimentação (+R$180) e transporte (+R$89)..."

`detect_spending_anomalies(user_id)` → lista de SpendingAnomaly:
Para cada categoria e merchant, calcular:
- Média histórica (últimos 3 meses) e desvio padrão
- Qualquer transação/período > média + 2σ é uma anomalia
- Anomalia de merchant: "Você gastou R$340 no iFood essa semana, sua média semanal é R$87"
- Anomalia de categoria: "Gastos com saúde esse mês: R$890 vs média de R$120/mês"
- Score de anomalia: quantos desvios padrão acima da média

`build_personal_benchmarks(user_id)` → PersonalBenchmarks:
Construir os benchmarks pessoais do usuário baseados no histórico:
- gasto_diario_medio: média de gasto por dia (excluindo dias sem transações)
- gasto_semanal_tipico: percentil 50 dos gastos semanais dos últimos 3 meses
- gasto_mensal_tipico: percentil 50 dos gastos mensais
- maior_gasto_historico: maior transação única já registrada
- categoria_mais_volatil: categoria com maior coeficiente de variação
- dia_mais_caro_semana: dia com maior média histórica de gastos

`generate_monthly_report_vs_history(user_id, month, year)` → MonthlyComparisonReport:
- Mês atual vs mesmo mês ano anterior (sazonalidade)
- Mês atual vs média dos últimos 3 meses
- Tendência: os gastos estão crescendo, estáveis ou diminuindo?
- Usar regressão linear simples nas últimas 6 observações mensais para detectar tendência
- Claude gera o texto: narrativa do mês em português, tom de coach financeiro

Endpoint: GET /reports/comparison?type=monthly&ref_month=2025-04
```

---

### 11.7 Simulador "E se?" (What-If Scenarios)

```
PROMPT:

Você é um engenheiro full-stack. Implemente o módulo de simulação financeira
"E se?" que permite ao usuário simular cenários sem afetar dados reais.

**backend/internal/handler/simulator.go** — endpoints de simulação (nunca salvam no banco):

POST /simulator/purchase
Body: {amount, installments, monthly_interest_rate?, description}
Retorna: impacto no cashflow mês a mês, total_de_juros se parcelado,
custo_real_total, meses_impactados, alerta se vai comprometer > X% da renda

POST /simulator/cut-subscription
Body: {merchant_name, monthly_amount}
Retorna: economia_mensal, economia_anual, economia_5_anos,
equivalente_investido_selic_5_anos, "o que você poderia fazer com esse dinheiro"

POST /simulator/save-goal
Body: {goal_name, target_amount, target_date?}
Retorna: quanto_poupar_por_mes para atingir a meta,
percentual_da_renda_necessario,
ajuste_sugerido (qual categoria cortar para viabilizar)
data_estimada_se_poupar_media_historica

POST /simulator/extra-income
Body: {amount, type: "one_time"|"recurring"}
Retorna: impacto no cashflow, sugestão de alocação baseada nos padrões do usuário
(ex: "com base no seu histórico, você provavelmente gastaria X em Y — sugerimos
guardar Z% e usar W% para quitar o parcelamento do cartão")

**web/components/WhatIfSimulator.tsx:**
- Interface de calculadora estilo "painel de controle"
- Tabs: Compra Parcelada | Cortar Assinatura | Meta de Economia | Renda Extra
- Cada aba tem inputs simples e resultado em cards visuais
- Gráfico de impacto no cashflow dos próximos 6 meses após a simulação
- Botão "Salvar simulação" que gera um agent_report do tipo "simulation"

**Flutter: lib/screens/simulator_screen.dart:**
- Bottom sheet deslizável com os 4 cenários
- Resultado animado (counter animation nos valores)
```

---

### 11.8 Score de Saúde Financeira Pessoal

```
PROMPT:

Você é um engenheiro Python sênior especialista em modelagem financeira.

Implemente `agents/health_score_agent.py` — um score numérico de saúde
financeira calculado 100% a partir dos dados reais do usuário.

**HealthScoreAgent(BaseAgent)**:

`calculate_health_score(user_id)` → HealthScore:
Score de 0 a 100, composto por 6 dimensões (cada uma de 0 a 100, com peso diferente):

1. **Fluxo de Caixa** (peso 25%):
   - 100 pts: sobra > 30% da renda todo mês consistentemente
   - 0 pts: saldo negativo ou gastos > receitas
   - Fórmula: média dos últimos 3 meses de (receita - gasto) / receita × 100

2. **Controle de Parcelamentos** (peso 20%):
   - 100 pts: zero parcelamentos ativos
   - 0 pts: > 50% da renda comprometida em parcelas
   - Fórmula: 100 - (total_parcelas_mensais / receita_mensal × 100)

3. **Consistência de Gastos** (peso 20%):
   - Mede a volatilidade dos gastos mensais (coeficiente de variação)
   - Alta consistência = gastos previsíveis = score alto
   - Fórmula: 100 - (desvio_padrao_gastos / media_gastos × 100), clampado em [0,100]

4. **Assinaturas vs Renda** (peso 15%):
   - 100 pts: < 5% da renda em assinaturas
   - 0 pts: > 20% da renda em assinaturas
   - Linear entre os dois extremos

5. **Diversificação de Gastos** (peso 10%):
   - Mede concentração: se > 50% dos gastos é em 1 categoria, score baixo
   - Usar índice HHI (Herfindahl-Hirschman) das categorias
   - Score = 100 × (1 - HHI_normalizado)

6. **Tendência de Melhora** (peso 10%):
   - Comparar últimos 2 meses vs 2 meses anteriores
   - Gastos diminuindo ou renda aumentando = score alto
   - Regressão linear simples no cashflow dos últimos 4 meses

`get_score_history(user_id, months=6)` → lista de HealthScoreSnapshot:
Recalcular o score para cada mês anterior e retornar histórico
Permitir ver a evolução do score ao longo do tempo

`get_score_recommendations(user_id, score: HealthScore)` → lista de Recommendation:
Usar Claude para gerar 3 recomendações específicas baseadas nas dimensões mais fracas:
- Ordenar dimensões por score mais baixo
- Para cada uma das 2 piores: gerar recomendação específica com dado concreto
- Ex: "Sua dimensão de Parcelamentos está em 34/100. Você tem R$1.240/mês
  comprometidos (41% da sua renda estimada). Quitar o parcelamento X
  em fevereiro liberaria R$340/mês e subiria seu score para ~52."

Endpoint: GET /reports/health-score
UI: Gauge chart circular (0-100) com cor gradiente (vermelho→amarelo→verde)
Abaixo: 6 barras de progresso para cada dimensão
Abaixo: 3 cards de recomendação
No Flutter: widget prominente na home screen abaixo do saldo
```

---

### 11.9 Detector de Sazonalidade e Gastos Futuros Previsíveis

```
PROMPT:

Você é um engenheiro Python sênior. Implemente `agents/seasonality_agent.py`
— detecção de gastos sazonais e previsão de despesas futuras que o usuário
vai ter mas ainda não está pensando.

**SeasonalityAgent(BaseAgent)**:

`detect_annual_patterns(user_id)` → lista de AnnualPattern:
Se houver dados de mais de 12 meses:
- Comparar cada mês com o mesmo mês do ano anterior por categoria
- Identificar: dezembro tem gastos X% maiores (festas), julho Y% maiores (férias)
- Detectar gastos anuais recorrentes: IPTU, IPVA, renovação de plano anual, etc.
- Retornar: mês, categoria, gasto_historico_medio, gasto_esperado_proximo_ano

`predict_upcoming_large_expenses(user_id, horizon_days=90)` → lista de UpcomingExpense:
Combinar múltiplas fontes de previsão:
- Parcelamentos com vencimento nos próximos horizon_days (dados reais)
- Assinaturas anuais com renovação detectada nos próximos horizon_days
- Padrões sazonais: "em julho você historicamente gasta +R$800 em viagens"
- Contas com padrão anual (IPVA detectado pelo merchant name pattern)
Retornar ordenado por data, com confidence_level (high/medium/low)

`build_annual_expense_calendar(user_id)` → dict[month] → lista de PlannedExpense:
Calendário anual de despesas previstas mês a mês:
- Janeiro: IPTU (detectado), renovações anuais detectadas
- Julho: padrão de viagens
- Dezembro: padrão de gastos elevados
- Cada mês com: total_fixo_previsto, total_sazonal_estimado, total_geral_estimado

Endpoint: GET /reports/upcoming-expenses?days=90
UI web: calendário visual (grid de 12 meses) com altura proporcional ao gasto esperado
Flutter: lista de "próximas despesas" na home com chip de data e valor
```

---

### 11.10 Análise de Merchant Intelligence

```
PROMPT:

Você é um engenheiro Python sênior. Implemente `agents/merchant_intelligence.py`
— análise profunda de relacionamento do usuário com cada merchant/estabelecimento.

**MerchantIntelligenceAgent(BaseAgent)**:

`get_merchant_profile(user_id, merchant_name)` → MerchantProfile:
Perfil completo do relacionamento com um merchant:
- primeira_compra: data e valor
- total_gasto_historico: soma de todas as transações
- frequencia_media: quantas vezes por mês/semana
- ticket_medio: valor médio por transação
- maior_compra: valor e data
- ultima_compra: data e valor
- tendencia: gasto está crescendo, estável ou diminuindo (últimos 3 meses)
- dias_preferidos: em quais dias da semana compra mais
- horarios_preferidos: manhã/tarde/noite (se disponível)
- ranking_na_categoria: "3º maior gasto em Alimentação"

`get_top_merchants(user_id, period_months=3, limit=20)` → lista de MerchantSummary:
- Top merchants por valor total no período
- Incluir: rank, merchant_name, total, percentual_do_total_gasto, ticket_medio, frequencia

`detect_merchant_anomalies(user_id)` → lista de MerchantAnomaly:
- Merchant com gasto > 3x a média histórica no último mês
- Merchant novo com gasto alto (potencial problema: primeiro gasto > R$200)
- Merchant com frequência muito acima do normal (possível vício de consumo)

`generate_merchant_insights(user_id)` → lista de MerchantInsight:
Usar Claude para gerar insights sobre os top 5 merchants:
Prompt: dado o perfil de cada merchant, identificar padrões interessantes.
Ex: "Você é cliente do iFood há 18 meses e já gastou R$4.200 lá.
     Sua frequência aumentou 40% nos últimos 2 meses — isso coincide com
     um aumento de R$287/mês nos seus gastos totais com alimentação."

Endpoint: GET /merchants?limit=20&period_months=3
Endpoint: GET /merchants/:merchant_name/profile
UI web: tabela de merchants com mini sparkline de tendência em cada linha
Ao clicar: drawer lateral com o MerchantProfile completo em gráficos
Flutter: tela de "Seus estabelecimentos" com cards deslizáveis
```

---

### 11.11 Sistema de Metas e Acompanhamento

```
PROMPT:

Você é um engenheiro full-stack. Implemente o sistema de metas financeiras
com acompanhamento automático baseado nos dados reais do usuário.

**Schema adicional** em `backend/internal/db/schema.sql`:
```sql
CREATE TABLE financial_goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    name TEXT NOT NULL,
    goal_type TEXT NOT NULL, -- savings, debt_payoff, spending_limit, income_target
    target_amount NUMERIC(12,2),
    current_amount NUMERIC(12,2) DEFAULT 0,
    start_date DATE NOT NULL,
    target_date DATE,
    category_id UUID REFERENCES categories(id), -- para metas de limite de gasto
    status TEXT DEFAULT 'active', -- active, completed, failed, paused
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**backend/internal/service/goals_service.go**:
- `UpdateGoalProgress(userID string)` — chamado após cada sync:
  - Para metas de spending_limit: somar gastos na categoria no mês atual vs limite
  - Para metas de savings: calcular saldo economizado desde start_date
  - Para metas de debt_payoff: verificar parcelas quitadas
  - Marcar como completed/failed automaticamente quando aplicável
  - Gerar FeedEvent quando meta é atingida ou está em risco

**Endpoints:**
GET /goals — listar metas com progresso atual
POST /goals — criar meta
PATCH /goals/:id — atualizar meta
DELETE /goals/:id — remover meta
GET /goals/:id/history — histórico de progresso da meta

**agents/goals_agent.py** — GoalsAgent:
`suggest_goals(user_id)` → lista de GoalSuggestion:
Baseado nos dados reais do usuário, sugerir 3 metas relevantes:
- Se tem parcelamentos altos: sugerir meta de "quitar parcelamentos até X"
- Se tem assinaturas esquecidas: sugerir "reduzir assinaturas em R$X"
- Se cashflow positivo consistente: sugerir meta de poupança com valor calculável
Usar Claude para formatar as sugestões em linguagem amigável

**UI web/components/GoalsPanel.tsx:**
- Cards de meta com progress bar animada
- Cor: verde se no caminho certo, amarelo se atrasada, vermelho se em risco
- Clicar na meta: modal com histórico de progresso (mini line chart)
- Botão "Sugerir metas" que chama o agente

**Flutter: lib/screens/goals_screen.dart:**
- Circular progress indicators para cada meta
- Notificação quando meta é atingida (animação de confetti com confetti package)
```

---

### 11.12 Relatório Narrativo Mensal (O "Extrato Inteligente")

```
PROMPT:

Você é um engenheiro Python sênior especialista em geração de conteúdo com LLMs.

Implemente `agents/narrative_report_agent.py` — gerador do relatório mensal
completo em formato narrativo, como se um contador pessoal estivesse explicando
o mês para você.

**NarrativeReportAgent(BaseAgent)**:

`generate_monthly_narrative(user_id, month, year)` → NarrativeReport:

Coletar TODOS os dados do mês:
- Resumo financeiro: total_gasto, total_recebido, saldo_gerado
- Comparativo: vs mês anterior, vs mesmo mês ano anterior (se disponível)
- Análise por categoria: top 5 categorias com variações
- Top 5 merchants do mês
- Parcelamentos pagos e pendentes
- Assinaturas ativas
- Anomalias detectadas
- Score de saúde do mês
- Eventos notáveis do feed

Prompt MONTHLY_NARRATIVE_PROMPT:
```
Você é o assistente financeiro pessoal de {user_name}.
Escreva o relatório financeiro do mês de {month_name} de {year} em português.

Dados completos do mês:
{full_month_data_json}

Diretrizes de escrita:
- Tom: direto, honesto, sem julgamentos morais, como um contador de confiança
- Estrutura: introdução (1 parágrafo) → análise de gastos → destaques positivos →
  pontos de atenção → parcelamentos e compromissos → perspectiva para o próximo mês
- Use os números reais. Seja específico.
- Não use bullets ou headers — escreva em parágrafos corridos como uma carta
- Máximo 400 palavras
- Termine com 1 ação concreta recomendada para o próximo mês

Não invente dados. Use apenas o que está nos dados fornecidos.
```

Salvar o relatório em agent_reports com agent_type='monthly_narrative'
Gerar automaticamente no dia 1 de cada mês para o mês anterior

**Endpoint:** GET /reports/narrative?month=2025-01
**UI web:** Página de relatório com o texto em tipografia elegante, sem boxes ou cards —
apenas texto bem formatado com os números em destaque (bold + cor accent)
**Flutter:** Tela de leitura com scroll suave, fonte maior, fundo #0D0D14
```

---

### Schema adicional para a Fase 11

```
PROMPT:

Adicione ao arquivo `backend/internal/db/schema.sql` as tabelas necessárias
para os módulos da Fase 11:

```sql
-- Feed de eventos financeiros
CREATE TABLE feed_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    amount NUMERIC(12,2),
    severity TEXT DEFAULT 'info',
    related_tx_ids UUID[],
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX ON feed_events(user_id, created_at DESC);
CREATE INDEX ON feed_events(user_id, read_at) WHERE read_at IS NULL;

-- Score de saúde financeira (histórico)
CREATE TABLE health_score_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    score NUMERIC(5,2),
    cashflow_score NUMERIC(5,2),
    installments_score NUMERIC(5,2),
    consistency_score NUMERIC(5,2),
    subscriptions_score NUMERIC(5,2),
    diversification_score NUMERIC(5,2),
    trend_score NUMERIC(5,2),
    period_month DATE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, period_month)
);

-- Metas financeiras
CREATE TABLE financial_goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    name TEXT NOT NULL,
    goal_type TEXT NOT NULL,
    target_amount NUMERIC(12,2),
    current_amount NUMERIC(12,2) DEFAULT 0,
    start_date DATE NOT NULL,
    target_date DATE,
    category_id UUID REFERENCES categories(id),
    status TEXT DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Simulações salvas
CREATE TABLE saved_simulations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    simulation_type TEXT NOT NULL,
    input_params JSONB,
    result_json JSONB,
    name TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```
```

---

## Ordem de execução recomendada no Antigravity

```
Agentes paralelos (Antigravity Manager):

Semana 1:
├── Agente A: Fase 1.1 + 1.2 (Docker + Schema)
└── Agente B: Fase 1.3 (Skeleton Go)

Semana 2:
├── Agente A: Fase 2 (Auth JWT)
└── Agente B: Fase 6 (início do serviço Python — classify endpoint)

Semana 3:
├── Agente A: Fase 3 (Pluggy integration)
└── Agente B: Fase 7 (Next.js dashboard — auth + layout)

Semana 4:
├── Agente A: Fase 4 (Transações + classificação)
└── Agente B: Fase 7 cont. (páginas dashboard + transações)

Semana 5:
├── Agente A: Fase 5 (Cartões + parcelamentos)
└── Agente B: Fase 6 cont. (agentes daily/weekly/monthly)

Semana 6:
├── Agente A: Fase 8 (Flutter)
└── Agente B: Fase 10 (Deploy + CI/CD)

Semana 7 — Fase 11 (Intelligence Layer):
├── Agente A: 11.1 Cashflow Timeline + 11.4 Motor de Projeção
├── Agente B: 11.2 Behavioral Agent + 11.6 Comparador Você vs Você
├── Agente C: 11.3 Gastos Invisíveis + 11.10 Merchant Intelligence
└── Agente D: 11.5 Activity Feed + 11.8 Health Score

Semana 8 — Fase 11 cont.:
├── Agente A: 11.7 Simulador What-If + UI web
├── Agente B: 11.9 Sazonalidade + 11.12 Relatório Narrativo
└── Agente C: 11.11 Metas + UI Flutter da Fase 11
```

---
## FASE 12 — Análise de Padrões Comportamentais Profundos

> Análises que o usuário nunca perceberia sozinho olhando para os números.
> Todos os módulos são agentes Python independentes que consomem o PostgreSQL diretamente.

---

### 12.1 Inflação Pessoal — Quanto suas compras ficaram mais caras

```
PROMPT:

Você é um engenheiro Python sênior especialista em análise financeira.

Implemente `agents/personal_inflation_agent.py` — detecta quanto os preços
que o usuário paga estão subindo ao longo do tempo, merchant por merchant
e categoria por categoria.

**PersonalInflationAgent(BaseAgent)**:

`calculate_merchant_inflation(user_id, merchant_name, months=12)` → MerchantInflation:
Para um merchant com histórico suficiente (mínimo 6 transações em períodos distintos):
- Calcular ticket médio por mês para esse merchant
- Aplicar regressão linear nos ticket médios mensais para extrair tendência
- Calcular taxa de inflação mensal e anual implícita
- Separar: inflação de preço real (mesmo produto mais caro) vs inflação de consumo
  (você passou a comprar itens mais caros no mesmo lugar)
- Retornar: ticket_medio_6_meses_atras, ticket_medio_atual, variacao_percent,
  taxa_anualizada, tipo_inflacao (preco/consumo/misto), confianca

`calculate_category_inflation(user_id, category_id, months=12)` → CategoryInflation:
- Mesmo cálculo mas agregado por categoria
- Decompor em: inflação de frequência (você vai mais vezes) vs inflação de ticket
- Identificar qual merchant dentro da categoria mais contribuiu para o aumento
- Comparar com IPCA do período (buscar via API do Banco Central — endpoint público,
  sem auth: https://api.bcb.gov.br/dados/serie/bcdata.sgs.433/dados?formato=json)

`build_personal_inflation_index(user_id)` → PersonalInflationIndex:
Construir o índice de inflação pessoal do usuário (estilo IPCA mas dos seus gastos reais):
- Pegar as top 10 categorias por peso no orçamento
- Calcular inflação de cada uma nos últimos 12 meses
- Ponderar pelo peso de cada categoria no total de gastos
- Resultado: um número percentual — "sua inflação pessoal foi de X% nos últimos 12 meses"
- Comparar com IPCA oficial do mesmo período
- Se inflação pessoal > IPCA: "seus gastos subiram X% acima da inflação oficial"

`detect_silent_price_increases(user_id)` → lista de SilentPriceIncrease:
Merchants onde o ticket médio aumentou mais de 15% nos últimos 6 meses
sem aumento correspondente na frequência (você não está comprando mais,
só pagando mais):
- Ordenar por impacto mensal absoluto (em reais)
- Top 5 com: merchant, aumento_percent, impacto_mensal_em_reais, primeiro_sinal_de_aumento

Prompt INFLATION_NARRATIVE_PROMPT:
```
Analise os dados de inflação pessoal abaixo e escreva um parágrafo em português,
tom direto, com os números mais relevantes.

Dados:
{inflation_data_json}

Foque no impacto em reais, não em percentuais abstratos.
Exemplo de tom: "Seu supermercado habitual ficou 23% mais caro nos últimos 12 meses —
você está pagando R$187 a mais por mês pelo mesmo carrinho de compras."
```

Endpoint: GET /reports/inflation
UI: tabela de merchants com seta de tendência colorida (vermelho=subindo, cinza=estável)
e o índice pessoal em destaque no topo comparado ao IPCA
Flutter: card na home "Sua inflação pessoal: X% (IPCA: Y%)"
```

---

### 12.2 Fim de Semana vs Dia Útil — Sua vida financeira em dois modos

```
PROMPT:

Você é um engenheiro Python sênior. Implemente `agents/weekday_weekend_agent.py`
— análise completa do comportamento financeiro em dias úteis vs fins de semana.

**WeekdayWeekendAgent(BaseAgent)**:

`compare_weekday_vs_weekend(user_id, months=3)` → WeekdayWeekendReport:
Separar todas as transações em: seg-sex (útil) vs sab-dom (fim de semana).
Para cada grupo calcular:
- gasto_medio_por_dia: média de gasto num dia de cada tipo
- top_categorias: quais categorias dominam cada tipo de dia
- top_merchants: merchants mais frequentes em cada contexto
- ticket_medio: valor médio por transação
- horario_pico: faixa horária de maior gasto em cada contexto
- percentual_do_gasto_mensal: quanto % do gasto total ocorre em cada contexto

`calculate_weekend_premium(user_id)` → WeekendPremium:
O "prêmio de fim de semana" — quanto a mais você gasta por dia no final de semana:
- gasto_diario_medio_util: float
- gasto_diario_medio_fds: float
- premium_absoluto: diferença em reais por dia
- premium_percentual: quantas vezes maior
- premium_mensal_estimado: impacto no mês (premium × 8 fins de semana)
- premium_anual: impacto anual
- categoria_mais_responsavel: qual categoria mais explica a diferença

`analyze_monday_effect(user_id)` → MondayEffectReport:
Detectar o "efeito segunda-feira" — compensação após o fim de semana:
- Comparar gastos de segunda com outros dias úteis
- Detectar se há queda de gastos na segunda (contenção pós-fds) ou aumento
  (compras postergadas do fim de semana)
- Analisar: é consistente ou varia dependendo do quanto foi gasto no fim de semana anterior?

`analyze_day_of_week_full(user_id)` → DayOfWeekBreakdown:
Perfil completo dos 7 dias da semana:
- Para cada dia (seg a dom): gasto_medio, top_3_categorias, top_3_merchants,
  horario_pico, adjetivo_gerado_pelo_llm (ex: "segunda é seu dia mais contido",
  "sexta é seu dia mais impulsivo")
- Ranking dos dias por gasto médio
- Detectar: existe um "dia de reset" onde os gastos são mínimos?

`generate_weekend_insights(user_id, report: WeekdayWeekendReport)` → lista de str:
Usar Claude para gerar 3 insights em linguagem natural baseados nos dados:
Prompt: dado o relatório completo, identificar os padrões mais interessantes
e inesperados. Foco em insights que o usuário provavelmente não sabia.
Ex: "Seus fins de semana custam o equivalente a um dia útil e meio.
     Em um ano, isso representa R$4.200 a mais só por causa do comportamento de FDS."

Endpoint: GET /reports/weekday-weekend
UI: gráfico de barras lado a lado (útil vs fds) por categoria
+ heatmap semanal (7 colunas × N semanas) com intensidade de cor por valor gasto
Flutter: widget de "Perfil da semana" com os 7 dias como cards horizontais deslizáveis
```

---

### 12.3 Efeito do Salário — O que acontece depois que o dinheiro cai

```
PROMPT:

Você é um engenheiro Python sênior. Implemente `agents/salary_effect_agent.py`
— análise de como o comportamento de gastos muda nos dias após a entrada de renda.

**SalaryEffectAgent(BaseAgent)**:

`detect_salary_events(user_id)` → lista de SalaryEvent:
Identificar todas as entradas de renda histórica:
- Buscar créditos grandes e recorrentes (mesmo valor ± 10%, mesmo período do mês)
- Para cada ocorrência: data, valor, dia_do_mes
- Calcular: dia_medio_do_salario, variacao_de_data (sempre cai no mesmo dia?)
- Retornar lista de SalaryEvent{date, amount, day_of_month}

`analyze_post_salary_spending(user_id, window_days=14)` → PostSalaryReport:
Para cada evento de salário detectado, analisar os N dias seguintes:
- Gasto acumulado por dia (D+1, D+2, ... D+14)
- Percentual do salário consumido em cada janela (D+3, D+7, D+14)
- Quais categorias explodem primeiro após o salário
- Qual merchant recebe o primeiro gasto significativo após o salário
- Curva média de consumo pós-salário (média de todos os eventos históricos)

`calculate_salary_consumption_curve(user_id)` → ConsumptionCurve:
A "curva de consumo do salário" — como o dinheiro vai embora ao longo do mês:
- percent_consumed_day_3: % médio gasto nos primeiros 3 dias
- percent_consumed_day_7: % médio gasto na primeira semana
- percent_consumed_day_15: % médio gasto na primeira quinzena
- percent_consumed_day_30: % médio gasto no mês todo (deve ser próximo de 100%)
- days_until_half_gone: mediana de dias para consumir 50% do salário
- spending_acceleration: os gastos aceleram ou desaceleram com o tempo?

`detect_post_salary_patterns(user_id)` → lista de PostSalaryPattern:
Padrões específicos detectados:
- FRONT_LOADED: > 40% gasto nos primeiros 5 dias ("você gasta metade do salário
  na primeira semana")
- IMPULSE_FIRST: primeiro gasto significativo (> R$200) ocorre em D+1 ou D+2
- SUBSCRIPTION_CLUSTER: assinaturas se concentram nos dias pós-salário
- END_OF_MONTH_STRESS: gasto cai drasticamente na última semana (saldo baixo)
- BALANCED: distribuição relativamente uniforme ao longo do mês

Prompt SALARY_EFFECT_PROMPT:
```
Analise o padrão de consumo pós-salário do usuário e escreva um insight em português.

Dados:
{salary_effect_data}

Seja específico com os números. Tom: analista financeiro direto, sem julgamentos.
Exemplo: "Você consome 47% do seu salário nos primeiros 7 dias do mês.
Na última semana, seus gastos caem 60% em relação à média — sinal claro de
que o dinheiro está acabando antes do fim do mês."
```

Endpoint: GET /reports/salary-effect
UI: gráfico de área acumulada mostrando % do salário consumido ao longo dos 30 dias
com linha de referência em 50% e 100%
Flutter: card "Seu salário dura X dias" na home screen
```

---

### 12.4 Análise de Impulso — Compras planejadas vs não planejadas

```
PROMPT:

Você é um engenheiro Python sênior especialista em análise comportamental.

Implemente `agents/impulse_agent.py` — detector de compras por impulso
baseado em sinais comportamentais das transações.

**ImpulseAgent(BaseAgent)**:

`calculate_impulse_score(transaction: Transaction, user_history: list) → float`:
Score de 0 a 1 de probabilidade de ser compra por impulso, baseado em:
- Horário incomum: transação fora do horário habitual do usuário nessa categoria (+0.2)
- Merchant novo: primeira vez nesse estabelecimento (+0.2)
- Valor atípico: > 2x o ticket médio do usuário nessa categoria (+0.2)
- Dia incomum: dia da semana fora do padrão para essa categoria (+0.1)
- Sequência rápida: < 30 min após outra transação (compras em sequência) (+0.15)
- Horário de madrugada (22h-06h) (+0.15)

`classify_transactions_by_impulse(user_id, months=3)` → ImpulseClassification:
Classificar todas as transações do período:
- impulse_transactions: lista de transações com score > 0.6
- planned_transactions: score < 0.3
- uncertain: 0.3-0.6
- impulse_total_amount: soma total das compras por impulso
- impulse_percent_of_spending: % do gasto total que é por impulso
- impulse_avg_amount: ticket médio das compras por impulso vs planejadas
- top_impulse_categories: quais categorias têm mais compras por impulso
- top_impulse_hours: horários com mais compras por impulso
- impulse_regret_proxy: merchants de impulso que não aparecem novamente
  (comprou uma vez, nunca voltou — proxy de arrependimento)

`analyze_impulse_trends(user_id)` → ImpulseTrends:
- impulse_rate_by_month: % de gastos por impulso em cada mês dos últimos 6
- tendencia: está aumentando ou diminuindo?
- melhor_mes: mês com menor taxa de impulso
- pior_mes: mês com maior taxa
- correlacao_estresse_financeiro: meses com saldo mais baixo têm mais impulso?

`generate_impulse_report(user_id)` → ImpulseReport:
Usar Claude para gerar análise narrativa:
Incluir: % de gastos por impulso, valor mensal médio, padrão de horário,
categorias mais afetadas, e uma observação sobre o proxy de arrependimento.
Tom: neutro, baseado em dados, sem julgamento moral.

Endpoint: GET /reports/impulse
UI: donut chart (impulso vs planejado vs incerto) + lista de transações
com badge de impulso + heatmap de horário × dia da semana
Flutter: badge discreto em cada transação com score alto na lista de extrato
```

---

### 12.5 Custo Real por Refeição — O que você realmente gasta para se alimentar

```
PROMPT:

Você é um engenheiro Python sênior. Implemente `agents/meal_cost_agent.py`
— calcula o custo real de alimentação do usuário de forma consolidada.

**MealCostAgent(BaseAgent)**:

`calculate_real_meal_cost(user_id, months=3)` → MealCostReport:
Consolidar TODOS os gastos relacionados a alimentação:
- Identificar subcategorias dentro de Alimentação: mercado/supermercado,
  delivery (iFood, Rappi, etc), restaurante/lanchonete, padaria, café
- Total gasto em alimentação no período
- Estimar número de refeições: (dias_no_periodo × 3) como denominador base
  — o usuário come ~3 refeições por dia, então esse é o total teórico
- custo_medio_por_refeicao: total / total_refeicoes_estimadas
- custo_por_subcategoria: quanto cada canal representa por refeição equivalente
- percentual_delivery: % das refeições que provavelmente foram delivery
- percentual_externo: % que foi fora de casa (restaurante + delivery)
- percentual_mercado: % que foi em casa (supermercado como proxy)

`compare_food_channels(user_id)` → FoodChannelComparison:
Comparar custo por refeição equivalente entre canais:
- ticket_medio_mercado: gasto médio por ida ao mercado ÷ refeições estimadas daquela compra
- ticket_medio_delivery: gasto médio por pedido de delivery
- ticket_medio_restaurante: gasto médio por refeição em restaurante
- canal_mais_caro: qual canal tem maior custo por refeição
- economia_potencial: se 30% dos deliveries fossem substituídos por cozinha,
  economia mensal estimada (sem recomendar — apenas informar)

`analyze_food_trends(user_id)` → FoodTrends:
- gasto_alimentacao_por_mes: série histórica dos últimos 6 meses
- tendencia_delivery: delivery está aumentando ou diminuindo como % do total?
- dia_semana_mais_delivery: qual dia tem mais pedidos de delivery
- horario_pico_delivery: almoço ou jantar dominam?
- merchant_favorito_por_canal: top merchant de cada canal

`calculate_annual_food_projection(user_id)` → float:
Com base na média dos últimos 3 meses, projetar gasto anual com alimentação.
Comparar com benchmarks públicos (usar valor fixo de referência:
família brasileira gasta em média R$1.200/mês com alimentação — IBGE POF 2023).

Prompt MEAL_COST_INSIGHT_PROMPT:
```
Com base nos dados de alimentação abaixo, escreva um insight em 2-3 frases em português.
Seja específico com números reais. Não use jargão financeiro.

Dados: {meal_data_json}

Exemplo de tom: "Cada refeição sua custa em média R$42. Nos dias que você pede
delivery, esse custo sobe para R$67 por refeição — 60% mais caro que quando
você cozinha ou vai a um restaurante."
```

Endpoint: GET /reports/meal-cost
UI: breakdown visual de alimentação com ícones por canal
Flutter: widget "Sua refeição média custa R$X" na seção de categorias
```

---

### 12.6 Índice de Conveniência — Quanto você paga pela praticidade

```
PROMPT:

Você é um engenheiro Python sênior. Implemente `agents/convenience_index_agent.py`
— calcula quanto o usuário paga a mais pela conveniência em diferentes aspectos
da vida financeira.

**ConvenienceIndexAgent(BaseAgent)**:

Definir pares de conveniência vs alternativa econômica (como constante CONVENIENCE_PAIRS):
```python
CONVENIENCE_PAIRS = [
    {
        "name": "Delivery vs Cozinhar",
        "convenient_merchants": ["ifood", "rappi", "ubereats", "delivery"],
        "economic_proxy": "supermercado",
        "premium_estimate_percent": 60,  # delivery custa ~60% a mais por refeição
        "category": "alimentação"
    },
    {
        "name": "Uber/99 vs Transporte Público",
        "convenient_merchants": ["uber", "99", "cabify"],
        "economic_proxy": "bilhete único",  # detectar por valor R$4-6
        "premium_estimate_percent": 400,
        "category": "transporte"
    },
    {
        "name": "Mercado de Bairro vs Supermercado",
        "convenient_merchants": [],  # detectar por ticket pequeno + frequência alta
        "economic_merchants": ["carrefour", "extra", "pão de açúcar", "atacadão"],
        "premium_estimate_percent": 25,
        "category": "alimentação"
    },
    {
        "name": "Farmácia de Plantão vs Farmácia Normal",
        "convenient_merchants": [],  # detectar por horário (22h-06h)
        "economic_proxy": "farmácia",
        "premium_estimate_percent": 30,
        "category": "saúde"
    }
]
```

`calculate_convenience_spending(user_id, months=3)` → ConvenienceReport:
Para cada par de conveniência detectado nos dados do usuário:
- valor_mensal_gasto_conveniencia: quanto gasta no canal conveniente
- valor_alternativa_estimado: quanto custaria na alternativa econômica
- premium_mensal: diferença em reais
- premium_percentual: diferença em %
- frequencia_mensal: quantas vezes por mês usa esse canal conveniente

`calculate_total_convenience_cost(user_id)` → ConvenienceSummary:
- total_premium_mensal: soma de todos os premiums de conveniência
- total_premium_anual: × 12
- percentual_da_renda: % da renda mensal estimada que vai para conveniência
- maior_conveniente: qual categoria tem maior custo de conveniência
- equivalente_investido: se esse valor fosse investido à SELIC, quanto seria em 5 anos

Prompt CONVENIENCE_INSIGHT_PROMPT:
```
Analise o índice de conveniência do usuário e escreva um insight em português.
Tom: informativo, sem julgamento. O objetivo é apenas mostrar o número real.

Dados: {convenience_data}

Formato: "Você gasta R$X/mês a mais pela conveniência — principalmente em [categorias].
Isso representa Y% da sua renda estimada. Em um ano, são R$Z pagos pela praticidade."
```

Endpoint: GET /reports/convenience-index
UI: lista de pares de conveniência com valor do premium em destaque
+ total mensal em card de destaque no topo
Flutter: seção "O que você paga pela praticidade" na tela de relatórios
```

---

### 12.7 Categoria que Cresce Silenciosamente

```
PROMPT:

Você é um engenheiro Python sênior. Implemente `agents/silent_growth_agent.py`
— detecta qual categoria está crescendo mais rapidamente de forma despercebida.

**SilentGrowthAgent(BaseAgent)**:

`detect_silent_growth_categories(user_id, months=6)` → lista de SilentGrowthCategory:
Para cada categoria com pelo menos 3 meses de dados:
- Calcular gasto mensal para cada mês dos últimos N meses
- Aplicar regressão linear simples para extrair taxa de crescimento mensal
- Filtrar: apenas categorias que NÃO são as maiores em valor absoluto
  (as maiores o usuário já monitora — o interesse é nas menores que crescem em silêncio)
- Calcular:
  - taxa_crescimento_mensal_percent: coeficiente da regressão em %
  - gasto_ha_6_meses: valor no primeiro mês da janela
  - gasto_atual: valor no último mês
  - variacao_absoluta: diferença em reais
  - projecao_12_meses: se continuar crescendo nesse ritmo, quanto será daqui 12 meses
  - r_squared: qualidade do fit da regressão (confiança da detecção)
- Filtrar apenas: r_squared > 0.5 (tendência consistente, não ruído)
- Ordenar por taxa_crescimento_mensal_percent DESC

`generate_silent_growth_alert(user_id, top_category: SilentGrowthCategory)` → str:
Usar Claude para gerar um alerta em linguagem natural para a categoria
que mais cresce silenciosamente:
```
Dado o crescimento silencioso da categoria abaixo, escreva um alerta em 2 frases.
Tom: direto, sem alarme exagerado. Mostre os números e a projeção.

Categoria: {category_name}
Crescimento: {growth_data}

Exemplo: "Seus gastos com 'Assinaturas' cresceram R$87/mês nos últimos 6 meses —
um aumento de 340% que provavelmente passou despercebido. Se continuar,
serão R$1.044 a mais por ano nessa categoria."
```

`calculate_growth_heatmap(user_id, months=6)` → GrowthHeatmap:
Matriz de crescimento: categorias × meses, valor = % de crescimento mês a mês
Permite visualizar a aceleração ou desaceleração de cada categoria ao longo do tempo

Endpoint: GET /reports/silent-growth
UI: tabela de categorias com sparkline de 6 meses + indicador de tendência
+ card de destaque "Categoria crescendo mais rápido: X (+Y%/mês)"
Flutter: alerta na home quando detectar crescimento > 30%/mês em qualquer categoria
```

---

### 12.8 Análise de Ticket Médio — Você compra mais ou paga mais caro?

```
PROMPT:

Você é um engenheiro Python sênior. Implemente `agents/ticket_analysis_agent.py`
— decompõe o crescimento de gastos entre aumento de frequência vs aumento de preço.

**TicketAnalysisAgent(BaseAgent)**:

`decompose_spending_growth(user_id, category_id, months=6)` → SpendingDecomposition:
Para uma categoria com crescimento detectado, decompor a causa:
- periodo_a: primeiros 3 meses da janela
- periodo_b: últimos 3 meses da janela
- frequencia_a: número de transações/mês no período A
- frequencia_b: número de transações/mês no período B
- ticket_medio_a: valor médio por transação no período A
- ticket_medio_b: valor médio por transação no período B
- variacao_gasto_total: (total_b - total_a) / total_a em %
- contribuicao_frequencia: quanto do aumento veio de comprar mais vezes
- contribuicao_ticket: quanto do aumento veio de pagar mais caro por vez
- tipo_crescimento:
  - FREQUENCY_DRIVEN: você está comprando mais vezes (mesmo preço)
  - PRICE_DRIVEN: você está pagando mais por cada compra (mesma frequência)
  - MIXED: ambos contribuíram
  - TICKET_UP_FREQUENCY_DOWN: você vai menos mas gasta mais por vez (upgrade de qualidade?)

`analyze_all_categories(user_id)` → lista de SpendingDecomposition:
Rodar a decomposição para todas as categorias com dados suficientes
Ordenar por variacao_gasto_total DESC (categorias que mais cresceram primeiro)

`generate_ticket_narrative(user_id, decompositions: list)` → str:
Usar Claude para sintetizar as descobertas mais interessantes:
Prompt: dadas as decomposições, identificar o padrão mais interessante e gerar
insight em 2-3 frases. Focar no padrão menos óbvio (não o maior, mas o mais
revelador sobre o comportamento do usuário).
Ex: "Seus gastos com restaurantes cresceram 45%, mas não porque você saiu mais —
você foi 10% menos vezes, mas o ticket médio subiu de R$67 para R$112.
Seu padrão de jantar fora ficou mais premium."

Endpoint: GET /reports/ticket-analysis?category_id=optional
UI: para cada categoria, barra dividida mostrando contribuição de frequência vs preço
Flutter: breakdown na tela de detalhes de cada categoria
```

---

### 12.9 Perfil Completo dos 7 Dias da Semana

```
PROMPT:

Você é um engenheiro Python sênior. Implemente `agents/weekly_profile_agent.py`
— constrói o perfil financeiro completo de cada dia da semana do usuário.

**WeeklyProfileAgent(BaseAgent)**:

`build_day_profiles(user_id, months=3)` → lista de DayProfile (7 profiles):
Para cada dia da semana (0=segunda, 6=domingo):
- gasto_medio: média de gasto em dias desse tipo
- gasto_mediano: mediana (mais robusta a outliers)
- desvio_padrao: volatilidade dos gastos nesse dia
- top_3_categorias: categorias dominantes com % do gasto do dia
- top_3_merchants: merchants mais frequentes
- horario_pico: faixa horária com mais transações (manhã/almoço/tarde/noite/madrugada)
- num_transacoes_medio: quantas transações em média nesse dia
- probabilidade_gasto_zero: % de dias desse tipo sem nenhuma transação (dia de descanso?)
- adjetivo: gerado pelo Claude — uma palavra que caracteriza o dia financeiramente

`detect_day_anomalies(user_id) → lista de DayAnomaly`:
Comparar o gasto de cada dia específico com a média histórica do mesmo dia da semana:
- Qualquer dia com gasto > média + 2σ é uma anomalia positiva
- Qualquer dia com gasto < média - 2σ é uma anomalia negativa (dia incomum de contenção)
- Retornar os 5 dias mais anômalos dos últimos 30 dias com explicação

`find_reset_day(user_id)` → ResetDayReport:
Detectar se existe um "dia de reset" — um dia da semana consistentemente com
gastos mínimos ou zero. Muitas pessoas têm um dia que naturalmente não gastam.
- dia_reset: dia da semana (se existir, com p > 0.7)
- probabilidade_zero_gasto: % de vezes que esse dia teve gasto zero
- gasto_medio_no_reset_day: mesmo que não seja zero, quanto costuma ser

`generate_week_narrative(user_id, profiles: list) → WeekNarrative`:
Usar Claude para gerar uma narrativa da semana financeira do usuário:
```
Você tem os perfis financeiros dos 7 dias da semana do usuário abaixo.
Escreva um parágrafo em português descrevendo a "semana financeira típica" dele.
Mencione os dias mais e menos ativos, os padrões de horário, e qualquer
particularidade interessante. Tom: analítico mas acessível.

Perfis: {day_profiles_json}
```

Endpoint: GET /reports/weekly-profile
UI: heatmap de 7 colunas (dias) × 24 linhas (horas) com intensidade de cor
por volume de gastos — revela instantaneamente quando o usuário gasta
Flutter: carrossel horizontal com card para cada dia da semana
```

---

### 12.10 Análise de Lealdade e Abandono de Merchants

```
PROMPT:

Você é um engenheiro Python sênior. Implemente `agents/loyalty_agent.py`
— analisa o relacionamento de longo prazo do usuário com cada merchant.

**LoyaltyAgent(BaseAgent)**:

`classify_merchant_relationships(user_id) → MerchantRelationshipMap`:
Classificar cada merchant em categorias de relacionamento:
- LEAL: presente em > 80% dos meses dos últimos 6 meses
- FREQUENTE: presente em 50-80% dos meses
- OCASIONAL: presente em 20-50% dos meses
- EXPERIMENTADO: usado apenas 1-2 vezes, não retornou
- ABANDONADO: foi leal/frequente mas não aparece nos últimos 60 dias
- RESGATADO: foi abandonado mas voltou recentemente

`calculate_abandonment_cost(user_id) → AbandonmentReport`:
Para merchants classificados como EXPERIMENTADO:
- Total gasto nesses merchants (dinheiro "desperdiçado" em experiências não repetidas)
- Ticket médio das compras únicas
- Categorias com mais abandono (onde você mais experimenta e não volta)
- Merchant mais caro que foi abandonado

`detect_loyalty_shifts(user_id) → lista de LoyaltyShift`:
Detectar quando o usuário trocou de merchant dentro da mesma categoria:
- Ex: parou de usar iFood e começou a usar Rappi no mesmo período
- Ex: trocou de supermercado
- Retornar: categoria, merchant_abandonado, merchant_adotado, data_da_troca,
  diferenca_de_ticket (ficou mais barato ou mais caro?)

`calculate_customer_lifetime_value(user_id) → lista de MerchantLTV`:
Para os top 20 merchants por total gasto:
- total_gasto_historico: soma de todas as transações
- primeira_compra: data
- ultima_compra: data
- duracao_relacionamento_dias: última - primeira
- frequencia_media_dias: intervalo médio entre compras
- projecao_12_meses: baseada na frequência e ticket atual

`generate_loyalty_insights(user_id) → lista de str`:
Usar Claude para 3 insights sobre padrões de lealdade:
- Qual merchant tem o relacionamento mais valioso (LTV + duracao)
- Qual categoria tem mais abandono (onde você mais experimenta)
- Se houve alguma troca significativa de merchant recentemente

Endpoint: GET /reports/loyalty
UI: tabela de merchants com badge de classificação colorido (leal=verde, abandonado=cinza)
+ seção "Quanto você gastou experimentando" com total de merchants abandonados
Flutter: tela "Seus estabelecimentos" com filtros por classificação
```

---

### 12.11 Correlação entre Semanas e Comportamento de Compensação

```
PROMPT:

Você é um engenheiro Python sênior especialista em análise comportamental.

Implemente `agents/compensation_agent.py` — detecta se o usuário compensa
períodos de contenção com gastos maiores depois, e vice-versa.

**CompensationAgent(BaseAgent)**:

`detect_compensation_patterns(user_id, months=4) → CompensationReport`:
Analisar autocorrelação dos gastos semanais:
- Para cada semana, calcular: foi semana de "alta" (> mediana) ou "baixa" (< mediana)?
- Analisar a semana seguinte: há tendência de compensação (baixa → alta → baixa)?
- Calcular coeficiente de autocorrelação lag-1 (semana N vs semana N+1)
- Se autocorrelação negativa forte: padrão de compensação presente
- Se positiva: padrão de momentum (semanas ruins tendem a continuar ruins)

`quantify_compensation(user_id) → CompensationQuantification`:
Se padrão de compensação detectado:
- quanto_gasta_apos_semana_contida: média de gasto nas semanas que seguem baixas
- quanto_gasta_apos_semana_pesada: média de gasto nas semanas que seguem altas
- amplitude_compensacao: diferença entre os dois em reais
- ciclo_medio_dias: duração média do ciclo contenção/compensação

`detect_post_stress_spending(user_id) → lista de StressSpendingEvent`:
Detectar eventos de "gasto pós-estresse":
- Semana com saldo muito baixo seguida de semana de gastos altos logo após salário
- Período de contenção forçada (saldo baixo) → explosão de gastos quando dinheiro entra
- Retornar: data_inicio_contencao, data_salario, valor_gasto_48h_pos_salario,
  vs_media_historica_mesmo_periodo

`generate_compensation_narrative(user_id, report) → str`:
Usar Claude:
```
Analise o padrão de compensação financeira do usuário.
Dados: {compensation_data}

Se há padrão claro, descreva em 2-3 frases o ciclo típico com números reais.
Se não há padrão, diga que os gastos são relativamente independentes entre semanas.
Tom: analítico, neutro. Não moralize.
```

Endpoint: GET /reports/compensation-pattern
UI: gráfico de linha dos gastos semanais com coloração alternada
mostrando o padrão de alta/baixa visualmente
```

---

### 12.12 Primeira Semana vs Última Semana do Mês

```
PROMPT:

Você é um engenheiro Python sênior. Implemente `agents/monthly_weeks_agent.py`
— análise comparativa das 4 semanas do mês revelando o ciclo financeiro mensal.

**MonthlyWeeksAgent(BaseAgent)**:

`split_month_into_weeks(year, month) → lista de WeekWindow`:
Dividir o mês em 4 janelas (não necessariamente de 7 dias cada):
- Semana 1: dias 1-7
- Semana 2: dias 8-14
- Semana 3: dias 15-21
- Semana 4: dias 22-fim

`analyze_monthly_weeks(user_id, months=3) → MonthlyWeeksReport`:
Para cada semana do mês, calcular médias históricas (média das semanas 1
de todos os meses analisados, etc.):
- gasto_medio_semana_1, 2, 3, 4: médias por semana
- top_categoria_semana_1, 2, 3, 4: categoria dominante de cada semana
- variacao_s1_vs_s4: quanto a primeira semana difere da última em %
- padrao_detectado:
  - DECLINING: gasto cai progressivamente (controle ao longo do mês)
  - INCREASING: gasto aumenta (impulso acumulado ou estouro)
  - U_SHAPED: alto no início, baixo no meio, alto no fim
  - SPIKE_FIRST: enorme na semana 1, estável depois (pós-salário)
  - UNIFORM: distribuição relativamente uniforme

`calculate_week_behavior(user_id) → WeekBehaviorProfile`:
- semana_mais_cara: qual das 4 semanas é historicamente mais cara
- semana_mais_barata: qual é mais econômica
- diferenca_s1_s4_percent: variação percentual entre primeira e última semana
- ultimo_dia_do_mes_behavior: o último dia do mês costuma ter gastos ou não?
  (detectar padrão de "comprar antes do fim do mês")
- primeira_compra_do_mes: qual categoria recebe o primeiro gasto de cada mês?

`generate_monthly_cycle_insight(user_id, profile) → str`:
Usar Claude para descrever o ciclo financeiro mensal em 2-3 frases com dados reais.
Destacar o padrão mais marcante (ex: U_SHAPED, SPIKE_FIRST) de forma acessível.

Endpoint: GET /reports/monthly-weeks
UI: 4 cards lado a lado representando as semanas, altura proporcional ao gasto médio
com a categoria dominante em cada semana como ícone
Flutter: visualização de "seu mês financeiro típico" na tela de relatórios
```

---

### Schema adicional para a Fase 12

```
PROMPT:

Adicione ao `backend/internal/db/schema.sql` as tabelas de cache para
os relatórios da Fase 12 (evitar recalcular tudo a cada request):

```sql
-- Cache de relatórios computados (evitar recálculo frequente)
CREATE TABLE report_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    report_type TEXT NOT NULL,  -- inflation, weekday_weekend, salary_effect, etc.
    period_key TEXT NOT NULL,   -- ex: "2025-01" ou "2025-Q1" ou "last_90_days"
    result_json JSONB NOT NULL,
    computed_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    UNIQUE(user_id, report_type, period_key)
);
CREATE INDEX ON report_cache(user_id, report_type);
CREATE INDEX ON report_cache(expires_at); -- para limpeza de expirados

-- Snapshots de inflação pessoal (histórico mensal)
CREATE TABLE inflation_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    period_month DATE NOT NULL,
    personal_inflation_rate NUMERIC(6,3),
    ipca_rate NUMERIC(6,3),
    category_breakdown JSONB,
    computed_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, period_month)
);

-- Perfis de comportamento por dia da semana (cache semanal)
CREATE TABLE day_profiles_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    day_of_week SMALLINT NOT NULL, -- 0=seg, 6=dom
    avg_spending NUMERIC(10,2),
    top_categories JSONB,
    top_merchants JSONB,
    peak_hour_range TEXT,
    computed_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, day_of_week)
);
```

Regras de cache:
- Relatórios de inflação: expiram em 7 dias
- Perfis de dia da semana: expiram em 3 dias
- Relatórios de comportamento (impulso, compensação): expiram em 1 dia
- O serviço Python verifica o cache antes de computar
- Após computar, salva no cache automaticamente
```

---acab

### Ordem de execução da Fase 12 no Antigravity

```
Semana 9 — Fase 12 (Behavioral Deep Dive):
├── Agente A: 12.1 Inflação Pessoal + 12.7 Categoria Silenciosa
├── Agente B: 12.2 Fim de Semana vs Útil + 12.9 Perfil dos 7 Dias
├── Agente C: 12.3 Efeito do Salário + 12.12 Primeira vs Última Semana
└── Agente D: 12.4 Análise de Impulso + 12.11 Padrão de Compensação

Semana 10 — Fase 12 cont.:
├── Agente A: 12.5 Custo Real por Refeição + 12.6 Índice de Conveniência
├── Agente B: 12.8 Análise de Ticket Médio + 12.10 Lealdade e Abandono
└── Agente C: Schema adicional + sistema de cache + UIs Flutter/Next.js
```

---
