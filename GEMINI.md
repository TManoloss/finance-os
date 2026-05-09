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
- Usar .env para as variáveis sensíveis. Criar também o arquivo .env.example com as chaves sem valores
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