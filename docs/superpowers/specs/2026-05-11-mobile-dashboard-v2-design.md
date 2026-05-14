# Spec: Mobile Dashboard V2 - Telemetria Avançada

**Data:** 2026-05-11
**Status:** Aprovado
**Contexto:** Upgrade do Dashboard mobile (Flutter) para exibir dados financeiros mais densos e úteis, conforme solicitado pelo usuário.

## 1. Visão Geral
O novo Dashboard substituirá a versão simplificada por uma interface inspirada em design industrial/blueprint, trazendo métricas de liquidez, evolução temporal de gastos e inteligência sobre mercantes.

## 2. Requisitos de Interface (UI)

### 2.1. Cabeçalho de Estado Líquido
- **Widget:** Card centralizado no topo.
- **Métrica:** `Patrimônio Líquido = (Soma de Contas Correntes) - (Soma de Faturas de Cartão)`.
- **Label:** `ESTADO_LIQUIDO_ATUAL`.

### 2.2. Ativos vs Passivos (Visão Híbrida)
- **Cards de Valor:** Dois cards lado a lado.
  - `EM_CONTA`: Saldo disponível em contas tipo 'checking' ou 'savings'.
  - `EM_CARTOES`: Soma das faturas atuais (dívida) de contas tipo 'credit'.
- **Barra de Comprometimento:** Abaixo dos cards, uma barra de progresso (0-100%).
  - Representa a porcentagem da liquidez que está comprometida com dívidas de cartão.
  - Indicador de "LIMITE_SEGURANÇA_EXCEDIDO" se o comprometimento for > 70%.

### 2.3. Telemetria Diária (Gráfico de Evolução)
- **Widget:** Gráfico de barras ou linha.
- **Dados:** Eixo X (Dias do mês), Eixo Y (Valor gasto no débito).
- **Objetivo:** Visualizar picos de consumo durante o período.

### 2.4. Ranking de Mercantes
- **Widget:** Lista simplificada (Top 5).
- **Dados:** Nome do mercante e valor total gasto.
- **Label:** `RANKING_MERCANTES`.

### 2.5. Ação de Atualização (Manual Refresh)
- **Widget:** Botão de "Refresh" na AppBar ou um Floating Action Button flutuante.
- **Ação:** Disparar `ref.refresh(summaryProvider)` no Riverpod.

## 3. Requisitos Técnicos

### 3.1. Backend (Go)
- **Repository:** Atualizar `GetSummary` para retornar também o detalhamento de saldos por tipo de conta.
- **Model:** Incluir `CheckingBalance` e `CreditBalance` no `TransactionSummary`.

### 3.2. Mobile (Flutter)
- **Model:** Atualizar `FinancialSummary` para refletir os novos campos do JSON (by_day, top_merchants, checking_balance, credit_balance).
- **Widgets:** 
  - Criar `DailySpendingChart` (usando `fl_chart`).
  - Criar `MerchantRankingWidget`.
  - Criar `CommitmentBar`.

## 4. Fluxo de Dados
1. Usuário abre o app ou clica em "Atualizar".
2. `summaryProvider` chama `GET /api/v1/transactions/summary`.
3. Backend calcula saldos, agrega gastos por dia e por mercante.
4. Frontend renderiza os novos widgets com os dados agregados.

## 5. Critérios de Aceite
- O Saldo Líquido deve bater com a diferença entre contas e cartões.
- O gráfico de evolução diária deve exibir corretamente os dias do mês atual.
- O botão de refresh deve recarregar os dados visualmente (shimmer/loading).
