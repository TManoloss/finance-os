# Design Spec: Dashboard Next.js (Fase 7)

## Visão Geral
Dashboard web para gestão financeira familiar, focado em clareza, proatividade da IA e visualização rica de dados.

## Design System
- **Tema:** Dark Mode exclusivo (#0A0A0F).
- **Cores de Destaque:** Roxo Elétrico (#7C6FFF) para marca/CTAs, Teal (#4ECDC4) para receitas.
- **Tipografia:** Inter (limpa e legível).

## Estrutura de Navegação
- **Sidebar Híbrida:** 
  - Estado Expandido: Ícones + Nomes claros (foco em usuários leigos/família).
  - Estado Recolhido: Apenas ícones (foco em espaço para dados).
  - Inclui mini-card de status do Pierre (IA).

## Componentes Principais (Home/Dashboard)
1. **Cards de Resumo:** Saldo, Gastos do Mês, Receitas, Economia.
2. **Alertas do Pierre (Opção 1):** Banner proativo no topo, estilo carrossel, logo abaixo do saldo total.
3. **Gráficos (Recharts):** 
   - Donut Chart: Gastos por categoria.
   - Area Chart: Evolução diária de gastos vs receitas.
4. **Últimas Transações:** Lista simplificada com ícones de categoria.

## Stack Tecnológica
- Next.js 15 (App Router)
- Tailwind CSS v4
- shadcn/ui (Componentes)
- Next-Auth v5 (Autenticação)
- Recharts (Gráficos)
- Axios (API Client com interceptors para Refresh Token)

## Fluxo de Autenticação
- Login via Credenciais (Email/Senha) integrado ao backend Go.
- Armazenamento de JWT com rotação de Refresh Token.
