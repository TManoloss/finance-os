# Plano de Implementação: Dashboard Next.js (Fase 7)

**Objetivo:** Criar o frontend web moderno e intuitivo integrado ao ecossistema Finance OS.

---

### Micro-tarefa 7.1: Scaffolding e Auth Base
- [ ] Criar projeto `web/` com Next.js 15, Tailwind v4 e TypeScript.
- [ ] Configurar `next-auth` (Auth.js v5) com Credentials Provider.
- [ ] Criar API client (Axios) com interceptor de 401 para Refresh Token.
- [ ] Implementar tela de Login (Dark Mode).

### Micro-tarefa 7.2: Layout Híbrido e Navegação
- [ ] Implementar Sidebar colapsável (Estado Expandido/Recolhido).
- [ ] Criar Navbar com perfil de usuário e status do Pierre.
- [ ] Configurar rotas protegidas (Middleware do Next.js).

### Micro-tarefa 7.3: Dashboard Analytics
- [ ] Criar Cards de Resumo (Saldo, Gastos, etc).
- [ ] Implementar Banner de Alertas do Pierre (Opção 1).
- [ ] Integrar Recharts (Donut Chart e Area Chart) com dados reais da API `/summary`.

### Micro-tarefa 7.4: Gestão de Transações e Cartões
- [ ] Página de Transações com filtros e paginação.
- [ ] Modal de edição de categoria.
- [ ] Página de Cartões com faturas projetadas e parcelamentos.

### Micro-tarefa 7.5: Chat com Pierre
- [ ] Interface de chat estilo "bolha" integrada.
- [ ] Histórico de conversa e integração com o agente IA.
