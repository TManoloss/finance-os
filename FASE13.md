## FASE 13 — Inteligência Emocional e Produto Premium

> Features que transformam o app de ferramenta financeira em experiência.
> Foco em valor emocional, gamificação sutil, narrativa e consciência contextual.
> Cada módulo é independente — podem ser desenvolvidos em paralelo no Antigravity.

---

### 13.1 Modo Sobrevivência Financeira

```
PROMPT:

Você é um engenheiro full-stack. Implemente o "Modo Sobrevivência Financeira"
— um estado especial do app ativado automaticamente quando o sistema detecta
risco financeiro iminente.

**backend/internal/service/survival_mode_service.go** — SurvivalModeService:

EvaluateSurvivalMode(userID string) → SurvivalModeStatus:
Calcular score de risco com base em 5 sinais (cada um 0-100, média ponderada):

1. Saldo projetado (peso 35%):
   - 0 pts: projeção negativa até o próximo salário
   - 50 pts: projeção < 20% do salário mensal
   - 100 pts: projeção > 50% do salário mensal

2. Velocidade de gasto (peso 25%):
   - Comparar gasto dos últimos 7 dias vs média histórica da mesma semana do mês
   - 0 pts: > 150% da média / 100 pts: < 80% da média

3. Uso de crédito (peso 20%):
   - Percentual da fatura em aberto vs limite estimado
   - 0 pts: > 80% do limite / 100 pts: < 30%

4. Proximidade do salário (peso 10%):
   - 0 pts: > 15 dias até próximo salário esperado
   - 100 pts: <= 3 dias

5. Recorrência de saldo baixo (peso 10%):
   - Quantas vezes nos últimos 3 meses o usuário ficou com saldo < R$200
   - 0 pts: > 2 vezes / 100 pts: nunca

Níveis de ativação:
- score > 70: TRANQUILO (app normal)
- score 45-70: ATENCAO (banner amarelo discreto)
- score 20-45: PRESSAO (UI muda sutilmente)
- score < 20: CRITICO (Modo Sobrevivência ativo)

GetSurvivalRecommendations(userID string) → []SurvivalRecommendation:
Quando ativo, gerar lista de ações priorizadas:
- Assinaturas canceláveis (ordenadas por valor, excluindo essenciais)
- Parcelamentos com opção de renegociação
- Categorias onde cortar é mais fácil baseado no histórico
- Projeção diária máxima recomendada

Endpoint: GET /reports/survival-mode

Dashboard em modo crítico — mudanças visuais SUTIS:
- Cards de métricas com borda left vermelha de 2px
- Widgets de menor prioridade ficam com opacity: 0.4
- Destaque para: saldo atual, projeção, gastos do dia, próximas contas

Flutter:
- Se crítico: mudar cor da status bar para vermelho escuro
- Notificação push: aviso de saldo projetado negativo
- Widget de limite diário na home com barra de progresso
```

---

### 13.2 Score de Stress Financeiro

```
PROMPT:

Você é um engenheiro Python sênior. Implemente agents/stress_score_agent.py
— indicador de stress financeiro de curto prazo, distinto do health score.

StressScoreAgent(BaseAgent):

calculate_stress_score(user_id) → StressScore:
Score de 0 (crítico) a 100 (tranquilo), 5 componentes:

1. Velocidade de consumo do saldo (30%): taxa de burn dos últimos 7 dias
2. Fração do salário já consumida (25%): gasto atual vs salário esperado
3. Crédito como % do gasto total (20%): se > 60% = stress elevado
4. Parcelamentos vs renda (15%): > 40% da renda = stress estrutural
5. Volatilidade recente (10%): desvio padrão dos gastos diários últimos 14 dias

Níveis: "tranquilo" | "atenção" | "pressão" | "crítico"
Trend: "melhorando" | "estável" | "piorando" vs 7 dias atrás

get_stress_history(user_id, days=30) → list[StressSnapshot]:
Histórico do score para ver evolução

generate_stress_context(user_id, score) → str:
Claude gera 1 frase contextual sobre o momento financeiro.
Tom: direto, sem drama, sem usar a palavra "stress".
Exemplos:
- "Você consumiu 68% do salário em 12 dias — ritmo acelerado."
- "Combinação de parcelas e cartão consume 71% da renda este mês."

Schema:
CREATE TABLE stress_score_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    score NUMERIC(5,2), level TEXT, components JSONB,
    trend TEXT, computed_at TIMESTAMPTZ DEFAULT NOW()
);

UI: badge no header — ponto colorido + label nível
Popover ao clicar: breakdown dos 5 componentes em barras
Flutter: widget abaixo do saldo na home
```

---

### 13.3 Explicador Inteligente de Gastos

```
PROMPT:

Você é um engenheiro Python sênior. Implemente agents/expense_explainer_agent.py
— responde perguntas sobre gastos em linguagem natural com dados reais e específicos.

ExpenseExplainerAgent(BaseAgent):

explain_period_spending(user_id, question: str, period_months=1) → Explanation:
Fluxo:
1. Classificar a pergunta (período, categoria, comparação)
2. Buscar dados relevantes do PostgreSQL
3. Calcular: comparativo período anterior, breakdown por categoria,
   top merchants, dias/horários relevantes, anomalias
4. Gerar resposta com Claude usando dados concretos

Prompt:
"Você é analista financeiro respondendo a pergunta com dados reais.
Seja específico, use números reais, máximo 4 frases. Não invente dados.

Pergunta: {question}
Dados: {financial_context_json}

Exemplos de resposta:
- '72% do aumento veio de delivery (+R$438), principalmente sextas à noite.
   Também houve aumento em transporte (+R$119) após o dia 12.'
- 'Foram 3 fatores: parcela subiu (R$340 a mais), mais visitas ao restaurante
   (+6 vezes) e compra pontual de R$520 no dia 18.'"

Integração com ChatAgent:
- Detectar automaticamente quando pergunta é sobre explicação de gastos
- Padrões: "por que", "quanto gastei", "o que aconteceu", "explica"
- Delegar para ExpenseExplainerAgent

Endpoint: POST /chat/explain
Body: {question: str, context_months: int}

UI no chat:
- Resposta com valores em destaque (bold + cor)
- Mini breakdown em chips coloridos por categoria abaixo da resposta
- Botão "Ver detalhes" filtra tela de transações pelo período
```

---

### 13.4 Timeline de Vida Financeira

```
PROMPT:

Você é um engenheiro full-stack. Implemente a Timeline de Vida Financeira
— histórico narrativo dos eventos financeiros significativos do usuário.

Tipos de evento detectados automaticamente:
- new_subscription / cancel_subscription
- installment_start / installment_end (com 🎉)
- salary_change
- new_merchant_habit / abandoned_habit
- best_month / worst_month
- spending_peak
- lifestyle_drift
- debt_free (zero parcelas)
- streak_no_delivery
- streak_positive_balance

BuildFinancialTimeline(userID string) → []TimelineEvent:
Varrer histórico completo e detectar todos os eventos acima.
Calcular e cachear semanalmente.

GenerateTimelineNarrative(event) → string:
Claude gera 1-2 frases narrativas por evento. Tom leve e pessoal.
Exemplos:
- "Depois de 11 meses, a parcela do notebook foi quitada. R$340/mês de volta."
- "Maio foi seu melhor mês dos últimos 12: sobrou R$1.240."
- "Você ficou 18 dias sem delivery — seu recorde histórico."

Endpoint: GET /reports/timeline?limit=50

Schema:
CREATE TABLE financial_timeline_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    event_type TEXT NOT NULL, event_date DATE NOT NULL,
    title TEXT, narrative TEXT, event_data JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

UI web: timeline vertical com linha central, ícones coloridos por tipo,
filtros: Todos | Marcos | Assinaturas | Parcelamentos | Records

Flutter: ListView com cards, animação staggered de entrada,
eventos de celebração com haptic feedback
```

---

### 13.5 Detecção de Drift de Estilo de Vida

```
PROMPT:

Você é um engenheiro Python sênior. Implemente agents/lifestyle_drift_agent.py
— detecta quando o padrão de vida mudou estruturalmente, não apenas variação pontual.

LifestyleDriftAgent(BaseAgent):

detect_lifestyle_drift(user_id, window_months=6) → LifestyleDriftReport:
Comparar janela recente (últimos 3 meses) vs janela anterior (3 meses antes).
É drift quando variação > 15% E consistente nos 3 meses (não apenas 1 pico).

calculate_cost_of_living_change(user_id) → CostOfLivingChange:
- custo_vida_anterior vs custo_vida_recente
- variacao_renda_percent no mesmo período
- categorias responsáveis pelo drift (top 3)

classify_drift_type() → DriftType:
- UPGRADE: custo subiu, renda acompanhou
- INFLATION: custo subiu, renda estável (pressão crescente)
- DOWNGRADE: custo caiu (contenção intencional ou forçada)
- TRANSITION: mudança brusca em categorias específicas
- STABLE: variação < 10%

Claude gera narrativa em 2-3 frases com números reais, sem julgamento.
Exemplo: "Nos últimos 5 meses seu custo de vida aumentou 18%, mas a renda
ficou estável. O aumento veio de alimentação (+R$290/mês) e lazer (+R$180/mês).
Impacto anual: R$5.640."

Endpoint: GET /reports/lifestyle-drift
Integrar com Timeline (gerar evento quando drift detectado)
```

---

### 13.6 Feed Gamificado de Conquistas

```
PROMPT:

Você é um engenheiro full-stack. Implemente o sistema de conquistas automáticas
— gamificação leve e adulta integrada ao ActivityFeed.

Conquistas detectadas após cada sync:

Streaks:
- 7 dias sem delivery / 14 dias (recorde)
- Semana inteira no positivo / Mês inteiro no positivo

Marcos financeiros:
- Parcelamento quitado 🎉
- Zero parcelamentos ativos
- Melhor mês dos últimos 12
- Maior sobra mensal histórica

Comportamento:
- Gastou abaixo da média esse mês
- Delivery caiu 30% este mês
- Semana sem compras por impulso

CheckAndAwardAchievements(userID) → []AwardedAchievement:
Verificar após cada sync, retornar apenas conquistas NOVAS.
Cada conquista vira FeedEvent com tipo "achievement".

Exemplos de textos no feed:
- "Você terminou o parcelamento da compra de R$2.400 🎉 — R$340/mês de volta."
- "Seu gasto com iFood caiu 22% esse mês 💪"
- "Você passou 14 dias sem delivery — seu recorde histórico 🏆"
- "Melhor mês dos últimos 12: sobrou R$1.240 🌟"

Schema:
CREATE TABLE achievements_awarded (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    achievement_id TEXT NOT NULL,
    awarded_at TIMESTAMPTZ DEFAULT NOW(),
    context_data JSONB,
    UNIQUE(user_id, achievement_id, date_trunc('month', awarded_at))
);

Flutter: bottom sheet animado + confetti leve + haptic de sucesso para conquistas novas
```

---

### 13.7 Memória Financeira

```
PROMPT:

Você é um engenheiro Python sênior. Implemente agents/financial_memory_agent.py
— usa o histórico anual para contextualizar o momento atual.

FinancialMemoryAgent(BaseAgent):

get_same_period_last_year(user_id, current_month) → YearOverYearContext:
Dados do mesmo mês do ano anterior vs mês atual:
- variacao_total, categorias que mais mudaram, narrativa do Claude

detect_seasonal_patterns(user_id) → list[SeasonalPattern]:
Com 12+ meses de dados: identificar meses sistematicamente acima da média.
Gerar alerta preventivo para o mês seguinte se historicamente pesado.

generate_memory_insights(user_id) → list[MemoryInsight]:
Claude gera 2-3 insights baseados em padrões anuais. Tom de quem conhece bem a pessoa.

Prompt:
"Gere 2-3 insights baseados em padrões anuais, em português, tom pessoal,
como se lembrasse algo relevante sobre o histórico do usuário.

Dados históricos: {historical_data}

Exemplos:
- 'Ano passado você também aumentou gastos em dezembro (+34%). Dezembro está chegando.'
- 'Você costuma gastar mais com viagens entre julho e agosto.'
- 'Em março do ano passado teve seu pior mês. Este março está 28% melhor.'"

Endpoint: GET /reports/financial-memory?month=2025-06
Disparar automaticamente no primeiro acesso de cada novo mês

UI: card "MEMÓRIA_FINANCEIRA" no dashboard com 2-3 insights
```

---

### 13.8 Planejamento Automático de Salário

```
PROMPT:

Você é um engenheiro Python sênior. Implemente agents/salary_planner_agent.py
— divide automaticamente o salário detectado em categorias de uso recomendadas.

SalaryPlannerAgent(BaseAgent):

generate_salary_plan(user_id) → SalaryPlan:
Executar quando salário for detectado (FeedEvent salary_detected):

1. Compromissos fixos: parcelas + assinaturas do mês (dados reais)
2. Histórico variável: média dos últimos 3 meses por categoria
3. Plano de alocação:
   - fixed_commitments: valor real comprometido
   - variable_budget: estimativa histórica ajustada
   - recommended_reserve: 10-20% do disponível
   - safe_daily_limit: (disponível - reserva) / dias_até_salário

Claude gera o briefing do mês em 3 frases. Tom de CFO pessoal.
Exemplo: "Seus compromissos fixos consomem R$1.840. Com R$2.242 disponíveis,
seu limite diário é R$58. Se mantiver esse ritmo, sobrará R$340."

Disparar automaticamente ao detectar salary_detected

Schema:
CREATE TABLE salary_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    salary_detected NUMERIC(10,2), fixed_commitments NUMERIC(10,2),
    safe_daily_limit NUMERIC(8,2), plan_data JSONB,
    valid_until DATE, generated_at TIMESTAMPTZ DEFAULT NOW()
);

UI: modal "PLANO_DO_MÊS" que aparece após salário detectado
Barra horizontal: fixo (vermelho) + variável (amarelo) + reserva (verde)
Destaque central: "Limite diário: R$X"
Flutter: widget permanente na home enquanto plano ativo
```

---

### 13.9 Radar Anti-Impulso

```
PROMPT:

Você é um engenheiro full-stack. Implemente o Radar Anti-Impulso — detecta
compras por impulso em tempo quase real e gera observações contextuais SEM julgamento.

AnalyzeRecentTransactions(userID, since) → []ImpulseAlert:
Rodar após cada sync, verificar transações das últimas 24h acima de R$50.

Sinais de impulso (2+ = gera alerta):
- Merchant novo (primeira vez)
- Fora do horário habitual (±2h do padrão da categoria)
- N-ésima compra acima de R$X nos últimos Y dias
- Horário de madrugada (22h-06h)
- < 30 min após outra transação

Claude gera observação em 1 frase. NÃO julga. Só traz consciência com dados.
Exemplos:
- "Essa é sua terceira compra acima de R$400 em 8 dias."
- "Primeira vez nesse estabelecimento, às 23h14."
- "Sua média de gastos às sextas à noite é 73% maior que nos outros dias."

UI: FeedEvent discreto com ícone de radar. Sem popup bloqueante.
Flutter: snackbar no rodapé, sem vibração, auto-dismiss 5 segundos.
```

---

### 13.10 Previsão de Dias Perigosos

```
PROMPT:

Você é um engenheiro Python sênior. Implemente agents/dangerous_days_agent.py
— aprende padrões temporais e gera alertas preventivos antes dos dias de maior gasto.

DangerousDaysAgent(BaseAgent):

identify_dangerous_patterns(user_id) → DangerousPatterns:
Top 5 combinações dia da semana × faixa horária com maior gasto médio histórico.
Dias do mês com maior probabilidade de gasto acima da média.

generate_preventive_alerts(user_id) → list[PreventiveAlert]:
Alertas para os próximos 7 dias baseados nos padrões históricos:
- "Sexta-feira: historicamente seu dia de maior gasto (+63%)"
- "Amanhã é domingo à noite — alta chance de delivery"
- "Dia 5 do mês: você costuma gastar 40% mais"

Claude gera alertas em 1 frase, tom neutro e informativo.
Não usa "cuidado" ou "atenção" — apenas informa.

Endpoint: GET /reports/dangerous-days
Mostrar como FeedEvents tipo "preventive_alert"
UI: card "PADRÕES_TEMPORAIS" com heatmap hora × dia da semana
```

---

### 13.11 Mapa de Dependência Financeira

```
PROMPT:

Você é um engenheiro full-stack. Implemente o Mapa de Dependência Financeira
— visualização Treemap mostrando quanto o orçamento depende de cada merchant/categoria.

Endpoint: GET /reports/dependency-map
Retornar estrutura hierárquica: categorias → merchants dentro de cada categoria,
com amount, percent_of_total, dependency_level por nó.

dependency_level:
- "crítica": > 30% do gasto total em 1 nó
- "alta": 15-30%
- "normal": < 15%

Alertas automáticos:
- "Dependência crítica de delivery: 34% do orçamento"
- "68% dos gastos em apenas 3 merchants"
- "Crédito responde por 71% dos gastos"

UI web: Treemap com recharts ou d3
- Retângulos proporcionais ao gasto
- Cor por categoria (paleta do sistema)
- Hover: tooltip com nome, valor, %, dependency level
- Drill down: clicar em categoria mostra merchants dentro dela

Flutter: Treemap com CustomPaint
- Tap em categoria: drill down
- Card de alerta no topo quando houver nível "crítica"
```

---

### 13.12 CFO Pessoal — Agente Proativo

```
PROMPT:

Você é um engenheiro Python sênior. Evolua o chat agent para o CFO Pessoal
— agente que toma iniciativa e envia insights diários sem ser perguntado.

CFOAgent(BaseAgent):

generate_proactive_insights(user_id) → list[ProactiveInsight]:
Verificar em ordem de prioridade e gerar 1-3 insights relevantes:
1. Survival mode ativo → plano de sobrevivência
2. Conquista nova → celebrar com contexto
3. Drift detectado → alertar com números
4. Melhor performance histórica → reforçar positivamente
5. Padrão sazonal próximo → lembrar com antecedência
6. Meta em risco → alertar com projeção
7. Subscriptions esquecidas → sugerir revisão

Prompt:
"Você é o CFO pessoal do usuário. Escreva 1 insight proativo em português.
Tom: parceiro financeiro de confiança, direto, com dados concretos. Máximo 3 frases.

Contexto: {full_context}

Exemplos:
- 'Você já economizou R$312 este mês comparado à média. Se mantiver esse ritmo,
   terá o melhor saldo dos últimos 8 meses.'
- 'Dezembro está chegando e historicamente é seu mês mais pesado (+34%).
   Com base no ritmo atual, você tem R$890 de margem para planejar.'
- 'Três assinaturas somam R$187/mês. Você não abre dois dos apps há 30 dias.'"

schedule_daily_cfo_message(user_id):
Rodar todo dia às 08:00, salvar como FeedEvent tipo "cfo_insight"
Flutter: push notification diária com o insight (horário configurável)

UI: FeedEvent com destaque especial — borda left 3px accent, label "CFO_PESSOAL"
Sticky por 24h no feed
```

---

### 13.13 Previsão Comportamental

```
PROMPT:

Você é um engenheiro Python sênior. Implemente agents/behavioral_prediction_agent.py
— previsão de comportamentos financeiros futuros baseada em padrões históricos.

BehavioralPredictionAgent(BaseAgent):

predict_month_end_balance(user_id) → BalancePrediction:
- predicted_balance, confidence_interval (low, high), confidence_percent
- Usar: dias restantes, gasto médio histórico, compromissos fixos conhecidos

predict_budget_overshoot_probability(user_id, category_id) → float:
Probabilidade 0-1 de extrapolar padrão histórico na categoria.
Baseado em velocidade atual, dias restantes, histórico de meses similares.

predict_overdraft_risk(user_id, days=7) → OverdraftRisk:
Monte Carlo com 1000 cenários:
- Gasto diário como variável aleatória (distribuição do histórico)
- Compromissos fixos como eventos determinísticos
- Salário como evento probabilístico (data histórica ± 2 dias)
- Retornar: probability_of_overdraft, expected_overdraft_day, worst_case

predict_impulse_probability_now(user_id) → ImpulseProbability:
Probabilidade de compra por impulso nas próximas 6 horas.
Baseado em: horário atual vs padrão histórico, stress score, tempo desde último gasto grande.

Endpoint: GET /reports/predictions
UI: seção "PREVISÕES" com probabilidades em barras de progresso:
- "Fechar mês positivo: 78% de chance"
- "Gasto alto nas próximas 6h: baixo risco"
```

---

### 13.14 Sistema de Missões Financeiras

```
PROMPT:

Você é um engenheiro full-stack. Implemente o Sistema de Missões Financeiras
— gamificação adulta focada em mudança de comportamento real.

Missões disponíveis (geradas baseadas no perfil do usuário):
- 7 dias sem delivery
- Reduza {categoria} em 20% este mês
- Respeite o limite diário por 5 dias consecutivos
- Poupe R${valor} este mês (feche com mais que a média de sobra)
- Quite um parcelamento (antecipe as últimas parcelas de {merchant})

GenerateMissionsForUser(userID) → []Mission:
Selecionar 3 missões relevantes baseadas no perfil:
- Muito delivery → missão no_delivery
- Categoria crescendo → missão reduce_category
- Stress score alto → missão daily_limit
Personalizar valores com dados reais do usuário.
Nunca repetir missão feita no mês anterior.

TrackMissionProgress(userID, missionID) → MissionProgress:
Calcular progresso após cada sync.
Gerar FeedEvent de conquista quando missão concluída.

Schema:
CREATE TABLE missions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    template_id TEXT NOT NULL, title TEXT, description TEXT,
    target_value NUMERIC(10,2), current_value NUMERIC(10,2) DEFAULT 0,
    started_at TIMESTAMPTZ DEFAULT NOW(), ends_at TIMESTAMPTZ,
    status TEXT DEFAULT 'active', completed_at TIMESTAMPTZ
);

UI: página /missions com 3 cards de missão ativas, progress bars,
dias restantes, botão trocar missão (1x por missão por mês)
Flutter: badge no ícone de navegação quando missão > 80% concluída
```

---

### 13.15 Replay Financeiro do Mês (Spotify Wrapped Financeiro)

```
PROMPT:

Você é um engenheiro full-stack. Implemente o Replay Financeiro — feature premium
gerada no dia 1 de cada mês contando a história financeira do mês anterior.

MonthlyReplayAgent(BaseAgent):

generate_monthly_replay(user_id, month, year) → MonthlyReplay:
Coletar: cashflow diário, top 5 merchants, categoria com maior crescimento,
dia mais caro e mais barato, semana mais pesada, conquistas, comparativo
com mês anterior, score de saúde, parcelamentos quitados.

Claude gera narrativa em 3-4 parágrafos curtos, estilo jornalístico pessoal.
Estrutura:
1. Como o mês começou (pós-salário, primeiros gastos)
2. O meio do mês (padrões, pico de gastos)
3. Como terminou (saldo, contenção ou expansão)
4. Destaque único do mês (recorde, conquista, ou ponto de atenção)

SEM bullets. Parágrafos corridos. Números em destaque.

Schema:
CREATE TABLE monthly_replays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id), period_month DATE NOT NULL,
    narrative TEXT, highlight_stat TEXT, replay_data JSONB,
    generated_at TIMESTAMPTZ DEFAULT NOW(), UNIQUE(user_id, period_month)
);

Disparar automaticamente todo dia 1 para o mês anterior.

UI web — página /reports/replay/[month]:
- Fundo #050508 (mais escuro que o dashboard)
- Título: "MAIO_2025.replay" no estilo terminal
- Stat de destaque centralizado e grande
- Mini area chart do cashflow do mês
- Narrativa em tipografia maior, texto puro com números em destaque (bold + cor)
- Grid de conquistas do mês com ícones grandes
- Botão "Compartilhar" → html2canvas → imagem para WhatsApp/Stories

Flutter:
- Tela imersiva (esconde bottom nav e header)
- Scroll com parallax no gráfico
- Share nativo com share_plus
- Notificação no dia 1: "Seu Replay de {mês} está pronto 📊"
```

---

### 13.16 Heatmap de Gastos (GitHub Style)

```
PROMPT:

Você é um engenheiro frontend sênior. Implemente o Heatmap de Gastos
— visualização estilo GitHub contribution graph.

52 colunas (semanas) × 7 linhas (dias), cada célula = 1 dia.
Intensidade por gasto relativo à média histórica:
- Sem gasto: #1a1a1a
- < 50% da média: #2a2a2a
- 50-100%: cor primária 30% opacity
- 100-200%: cor primária 60% opacity
- > 200%: cor primária 100% opacity

No tema terminal: verde (#4caf50) com opacidades variadas.

Marcadores especiais:
- Dia de salário: outline teal na célula
- Conquista alcançada: ponto branco no canto superior direito
- Parcelamento quitado: ponto dourado

Tooltip ao hover: data, valor gasto, top categoria, vs média (%)

Endpoint: GET /reports/spending-heatmap?year=2025
Retornar: [{date, total_spent, top_category, vs_average_ratio}] para 365 dias

Flutter: CustomPaint
- Scroll horizontal para navegar entre meses
- Tap em célula: bottom sheet com detalhes do dia
```

---

### 13.17 Detector de Pequenos Vazamentos

```
PROMPT:

Você é um engenheiro Python sênior. Implemente agents/micro_spending_agent.py
— detecta o impacto acumulado de pequenos gastos que passam despercebidos.

MicroSpendingAgent(BaseAgent):

analyze_micro_transactions(user_id, threshold=30.0, months=1) → MicroSpendingReport:
Todas as transações < R$30:
- total_micro_spending, count, percent_of_total
- by_category, by_merchant (top merchants de micros)
- daily_average, annualized (projeção anual)

identify_micro_patterns(user_id) → list[MicroPattern]:
- "Café toda manhã": transações R$8-15 mesmo merchant entre 07-09h
- "Gorjetas digitais": taxa de entrega + gorjeta + taxa de serviço acumuladas
- "Assinaturas menores esquecidas": cobranças < R$20 recorrentes

calculate_micro_leakage_impact(user_id) → MicroLeakageImpact:
- micro_spending_monthly e annual
- biggest_micro_habit: maior hábito de micro gasto por impacto acumulado
- invested_equivalent: se investido à SELIC por 5 anos

Claude gera 2 frases revelando o impacto. Neutro, baseado em dados.
Exemplo: "Seus gastos abaixo de R$30 somaram R$1.184 este mês — 47 transações
que passaram despercebidas. O maior hábito: café e lanches (R$412/mês)."

Endpoint: GET /reports/micro-spending?threshold=30
UI: card "PEQUENOS_VAZAMENTOS" com total em destaque + lista de padrões
```

---

### 13.18 Timeline de Parcelamentos (Alívio Financeiro Futuro)

```
PROMPT:

Você é um engenheiro full-stack. Implemente a Timeline de Parcelamentos
— visualização focada no alívio financeiro futuro conforme parcelas terminam.

Endpoint: GET /cards/installments/timeline
Retornar mês a mês pelos próximos 12 meses:
- active_installments: quantas parcelas ativas
- total_amount: total de parcelas no mês
- ending_this_month: lista de parcelas que terminam (merchant, valor, total pago)
- new_freedom: dinheiro liberado vs mês anterior
- cumulative_freedom: total liberado desde hoje

UI web — timeline horizontal dos próximos 12 meses:
- Colunas com altura proporcional ao total de parcelas
- Meses onde parcelas terminam: destaque verde + "🎉 +R$X livres"
- Linha de alívio acumulado sobreposta
- Mensagem de destaque: "Em {mês} você libera R$X/mês"

Exemplos de mensagens:
- "Em agosto você libera R$640/mês de parcelas 🎉"
- "Em outubro mais R$280/mês com o fim do parcelamento da TV"
- "Até dezembro: R$1.120/mês totalmente liberados"

Flutter: cards horizontais deslizáveis, um por mês com parcela terminando
Cor verde com opacity proporcional ao valor liberado
Tap: detalhes do que termina naquele mês
```

---

### Schema adicional para Fase 13

```
PROMPT:

Adicione ao backend/internal/db/schema.sql as tabelas da Fase 13:

-- Survival mode snapshots
CREATE TABLE survival_mode_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    risk_score NUMERIC(5,2), level TEXT, is_active BOOLEAN DEFAULT false,
    projected_shortfall NUMERIC(10,2), days_until_salary INT, top_risks JSONB,
    computed_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX ON survival_mode_snapshots(user_id, computed_at DESC);

-- Stress score snapshots
CREATE TABLE stress_score_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    score NUMERIC(5,2), level TEXT, components JSONB, trend TEXT,
    computed_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX ON stress_score_snapshots(user_id, computed_at DESC);

-- Conquistas
CREATE TABLE achievements_awarded (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    achievement_id TEXT NOT NULL, awarded_at TIMESTAMPTZ DEFAULT NOW(), context_data JSONB,
    UNIQUE(user_id, achievement_id, date_trunc('month', awarded_at))
);

-- Missões
CREATE TABLE missions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    template_id TEXT NOT NULL, title TEXT, description TEXT,
    target_value NUMERIC(10,2), current_value NUMERIC(10,2) DEFAULT 0,
    started_at TIMESTAMPTZ DEFAULT NOW(), ends_at TIMESTAMPTZ,
    status TEXT DEFAULT 'active', completed_at TIMESTAMPTZ
);
CREATE INDEX ON missions(user_id, status);

-- Monthly replay
CREATE TABLE monthly_replays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id), period_month DATE NOT NULL,
    narrative TEXT, highlight_stat TEXT, replay_data JSONB,
    generated_at TIMESTAMPTZ DEFAULT NOW(), UNIQUE(user_id, period_month)
);

-- Salary plans
CREATE TABLE salary_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    salary_detected NUMERIC(10,2), fixed_commitments NUMERIC(10,2),
    safe_daily_limit NUMERIC(8,2), plan_data JSONB,
    valid_until DATE, generated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Financial timeline events
CREATE TABLE financial_timeline_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    event_type TEXT NOT NULL, event_date DATE NOT NULL,
    title TEXT, narrative TEXT, event_data JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX ON financial_timeline_events(user_id, event_date DESC);
```

---

### Ordem de execução da Fase 13 no Antigravity

```
Semana 11 — Fase 13 Core (maior impacto emocional):
├── Agente A: 13.1 Modo Sobrevivência + 13.2 Stress Score
├── Agente B: 13.3 Explicador Inteligente + 13.12 CFO Pessoal
├── Agente C: 13.6 Feed de Conquistas + 13.14 Missões
└── Agente D: 13.8 Planejamento de Salário + 13.18 Timeline Parcelamentos

Semana 12 — Fase 13 Intelligence:
├── Agente A: 13.4 Timeline de Vida + 13.5 Lifestyle Drift
├── Agente B: 13.7 Memória Financeira + 13.10 Dias Perigosos
├── Agente C: 13.9 Radar Anti-Impulso + 13.13 Previsão Comportamental
└── Agente D: 13.17 Pequenos Vazamentos + Schema Fase 13

Semana 13 — Fase 13 Visual e Premium:
├── Agente A: 13.11 Mapa de Dependência
├── Agente B: 13.15 Replay Financeiro (Spotify Wrapped)
├── Agente C: 13.16 Heatmap de Gastos
└── Agente D: integrações entre módulos
```
