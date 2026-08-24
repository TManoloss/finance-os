# Plano de Validação e Transformação do FinanceOS

> Fonte oficial de verdade para escopo, sequência, aceite e evidências da transformação do FinanceOS.

**Baseline auditado:** 2026-08-18  
**Experiência principal:** mobile  
**Glossário canônico:** [`CONTEXT.md`](../CONTEXT.md)  
**Moeda principal:** BRL  
**Horário financeiro:** `America/Sao_Paulo`

## 1. Regra de governança

Este documento controla a passagem de uma funcionalidade entre os estados abaixo. A existência de código, rota, tabela ou tela não prova que a funcionalidade funciona de ponta a ponta.

| Estado | Significado |
| --- | --- |
| `real` | Usa dados reais no caminho principal e não inventa resultados, mas ainda não possui toda a evidência de aceite. |
| `parcial` | Parte do fluxo é real; faltam contrato, estados, segurança, integração ou aceite. |
| `mock` | Exibe ou calcula dados demonstrativos, constantes ou resultados fictícios. |
| `oculta` | Não está acessível ao usuário enquanto não cumpre os requisitos de validação. |
| `validada` | Cumpre critérios de aceite e possui evidência reproduzível com dados reais. |

### Regra de bloqueio

Nenhum recurso `mock` pode mudar diretamente para `validada`. A promoção exige, nesta ordem:

1. remoção de constantes e resultados fictícios;
2. teste automatizado do cálculo ou contrato;
3. teste com dados reais de um usuário de validação;
4. registro de responsável, data, evidência e resultado neste documento.

Na dúvida, o estado mais restritivo prevalece. Falha silenciosa, ação sem efeito real ou dependência obrigatória de IA bloqueia a fase.

## 2. Produto e limites

FinanceOS acompanha a vida financeira pessoal, detecta fatos relevantes, explica suas causas e ajuda o usuário a decidir e acompanhar próximos passos. Pluggy/Open Finance é a fonte principal; contas e transações manuais continuam sendo fontes válidas e independentes.

### Proposta de valor

- Reunir saldos, movimentações e compromissos em uma visão cotidiana confiável.
- Transformar histórico financeiro em alertas, análises, metas e simulações verificáveis.
- Explicar resultados em linguagem clara sem transferir a autoridade do cálculo para IA.
- Preservar a identidade visual premium atual no mobile e, após sua validação, na web.

### Limites obrigatórios

- FinanceOS não envia, recebe, transfere nem paga dinheiro.
- FinanceOS não executa investimentos e não é uma plataforma de investimentos.
- Uma simulação nunca é apresentada como operação realizada ou retorno garantido.
- Pierre não substitui cálculo determinístico, aconselhamento profissional ou consentimento do usuário.
- Funcionalidades sem dados reais permanecem ocultas.

## 3. Linguagem e fluxo de domínio

As definições normativas de **Usuário, Conexão, Conta, Transação, Categoria, Regra, Compromisso, Alerta, Meta, Simulação, Análise** e **Pierre** estão no [`CONTEXT.md`](../CONTEXT.md). Código, API e interface devem adotar esses nomes; `portfólio`, `investimento`, `pagamento` e `transferência` não são sinônimos válidos para capacidades existentes.

Fluxo principal:

```text
Conectar → importar → classificar → detectar → explicar → orientar → acompanhar
```

- **Conectar:** autorizar Pluggy ou criar uma conta manual.
- **Importar:** obter contas, saldos e transações sem duplicação.
- **Classificar:** aplicar categoria e indicar confiança/revisão.
- **Detectar:** produzir compromissos, padrões e alertas a partir de fatos.
- **Explicar:** mostrar origem, período, qualidade e confiança do resultado.
- **Orientar:** oferecer decisão ou próximo passo sem executar movimentação financeira.
- **Acompanhar:** medir efeitos, metas e mudanças ao longo do tempo.

## 4. Evidências do baseline

Esta é uma auditoria estática do repositório, não uma validação de produção.

| Camada | Evidência atual | Leitura |
| --- | --- | --- |
| Banco | [`schema.sql`](../backend/internal/db/schema.sql) possui usuários, contas, transações, categorias, regras, parcelas, feed, metas, simulações salvas, snapshots e Replay. | O armazenamento existe, mas parte das tabelas não tem fluxo completo de escrita/leitura. |
| Backend | [`router.go`](../backend/internal/router/router.go) expõe autenticação, contas, transações, categorias, relatórios, feed, metas e simulador sob `/api/v1`. | API ampla, com contratos ainda fragmentados e alguns handlers mockados. |
| Sincronização | [`scheduler.go`](../backend/internal/jobs/scheduler.go) agenda `30 0 * * *` em `America/Sao_Paulo`; [`accounts.go`](../backend/internal/handler/accounts.go) oferece sincronização manual. | Agendamento principal é real; o status interno ainda anuncia horários externos divergentes. |
| Propriedade | Listagem/criação de transações filtra por conta do usuário, mas [`UpdateCategory`](../backend/internal/repository/transaction_repository.go) atualiza apenas por `transaction_id`. | Correção de categoria não isola usuário e não conclui revisão/aprendizado. |
| Agentes | [`main.py`](../agents/main.py) publica classificação, rotinas diária/semanal/mensal e análises especializadas. | Há muitos cálculos, mas contratos, fallback determinístico e testes são incompletos. |
| Saúde | [`health_score_agent.py`](../agents/app/services/health_score_agent.py) fixa `consistency`, `subscriptions` e `trend`. | Score é `parcial/mock`. |
| Mobile | [`app_router.dart`](../mobile/lib/core/router/app_router.dart) e [`main_layout.dart`](../mobile/lib/core/layout/main_layout.dart) mantêm cinco branches com os rótulos antigos. | A shell existe, mas não implementa a arquitetura canônica. |
| Sessão mobile | [`auth_provider.dart`](../mobile/lib/features/auth/presentation/auth_provider.dart) guarda access/refresh tokens em storage seguro e tenta renovar ao iniciar. | Implementação `real`, aceite de restauração ainda pendente. |
| Simulador | [`simulator.go`](../backend/internal/handler/simulator.go) projeta de saldo fixo, meses fixos, comprometimento fixo e retorno de 45%; [`simulator_screen.dart`](../mobile/lib/features/simulator/presentation/simulator_screen.dart) chama `/simulator/cut`, rota inexistente. | `mock`; deve ser ocultado. |
| Replay | Web consome `/reports/monthly-replay`; [`replay_screen.dart`](../mobile/lib/features/reports/presentation/replay_screen.dart) declara dados mockados. | Web `parcial`; mobile `mock`. |
| Web | Páginas atuais cobrem dashboard, transações, cartões, relatórios, chat, simulador e configurações. | Compatibilidade precisa ser preservada enquanto os contratos evoluem. |
| Testes | Existem três testes Go de serviço, nenhum teste Python, e dois testes Flutter. | Cobertura insuficiente para qualquer promoção ampla a `validada`. |

## 5. Inventário funcional

### 5.1 Capacidades do produto

| Capacidade | Estado atual | Evidência | Estado de saída |
| --- | --- | --- | --- |
| Cadastro, login e refresh token | `real` | handlers de auth e testes de serviço | `validada` após contrato e restauração mobile/web |
| Sessão persistida mobile | `real` | `AuthNotifier._checkAuthStatus` | `validada` após fechar/reabrir em dispositivo |
| Credenciais Pluggy por usuário | `real` | validação, criptografia e persistência em `accounts.go` | `validada` após cenário real e erros operacionais |
| Widget e Conexão Pluggy | `real` | onboarding mobile validado no SM-S948B abriu o widget oficial com token emitido pelo backend | comprovante completo de importação com conexão concluída |
| Conta manual | `real` | `POST /accounts` | aceite mobile com independência de Conexões |
| Listar/configurar/excluir Conta | `real` | rotas de accounts com filtro de usuário | aceite incluindo cascata e reconexão |
| Sincronização agendada | `parcial` | cron 00:30 e sync interno com texto de status divergente | um agendamento e status coerente |
| Sincronização manual | `parcial` | `POST /accounts/sync`, assíncrono e rate limit | progresso/comprovante e falha parcial explícitos |
| Importação idempotente | `parcial` | identificadores Pluggy e serviço de sync | teste de repetição sem duplicação |
| Extrato e paginação | `real` | `GET /transactions` | busca, todos os filtros e contrato tipado |
| Lançamento manual | `real` | `POST /transactions` atualiza saldo em transação de banco | aceite mobile e recálculos dependentes |
| Categorias e regras | `parcial` | categorias globais/usuário e classificador | confiança, revisão e regra aprendida na correção |
| Correção de categoria | `parcial` | PATCH sem `user_id` no update | propriedade, revisão concluída e aprendizado |
| Resumo financeiro | `real` | `GET /transactions/summary` | substituído/composto por `/overview` no mobile |
| Feed e leitura | `parcial` | listar, contar, ler um e ler todos | origem navegável e eventos faltantes |
| Alertas adaptativos | `parcial` | serviço usa limites fixos para renda, gasto alto e saldo | regras adaptativas e testes de histórico curto |
| Parcelas e assinaturas | `parcial` | endpoints de cards e detecção em sync | linha temporal em Planejar e origem navegável |
| Metas | `parcial` | listar/criar/sugerir; apenas dois modos recalculados | ciclo completo, quatro modos e ajustes auditáveis |
| Simulação de compra | `oculta` | UI removida/redirecionada e endpoint responde `503`, sem projeção fictícia | projeção real antes de reexpor |
| Simulação de corte | `oculta` | contrato mobile corrigido; UI oculta e endpoint responde `503` | economia mensal/anual/acumulada antes de reexpor |
| Saúde financeira | `oculta` | UI redirecionada/oculta e backend responde `503` enquanto dimensões forem fixas | todas dimensões calculadas e metadados de qualidade |
| Análises especializadas | `parcial` | endpoints Go/Python e telas web/mobile | grupos consolidados, lazy load e fallback |
| Pierre | `parcial` | chat e fallback Groq/Gemini | somente perguntas abertas/explicações, sem bloquear produto |
| Replay web | `parcial` | endpoint e tela com mês fixo na entrada de relatórios | dados do período escolhido e aceite mensal |
| Replay mobile | `mock` | tela explicitamente demonstrativa | dados reais e seis resultados mínimos |
| Navegação mobile canônica | `parcial` | cinco branches com destinos/rótulos antigos | `Hoje / Movimentações / Planejar / Análises / Mais` |
| Paridade web | `parcial` | fluxos principais existem com taxonomia anterior | taxonomia e agrupamentos canônicos após mobile |

### 5.2 Inventário de análises existentes e destino consolidado

Todos os endpoints abaixo já estão registrados no backend. “Existente” não significa “validado”.

| Grupo canônico | Análises/endpoints existentes |
| --- | --- |
| Saúde e risco | `health-score`, `stress-score`, `survival-mode`, `dangerous-days`, `impulse-radar` |
| Futuro financeiro | `cashflow`, `projection`, `upcoming-expenses`, `salary-plan`, `installment-timeline` |
| Comportamento | `behavioral`, `behavioral-prediction`, `impulse`, `compensation-pattern`, `micro-spending`, `weekday-weekend` |
| Evolução do custo de vida | `comparison`, `personal-inflation`, `silent-growth`, `lifestyle-drift` |
| Renda e ciclo mensal | `weekly-profile`, `salary-effect`, `monthly-weeks`, `weekday-weekend` |
| Custos e compromissos | `invisible-spending`, `meal-cost`, `convenience-index`, `ticket-analysis`, `loyalty`, `installment-timeline`, cards/installments/subscriptions |
| Relações e dependências | `dependency-map`, merchants e perfis de estabelecimento |
| História financeira | `timeline`, `financial-memory`, `monthly-replay`, `spending-heatmap` |
| Engajamento | `gamification`, metas e missões |
| Explicação sob demanda | `narrative`, `chat` e gatilho explícito `reports/trigger/:type` |

Sobreposições a unificar sem apagar cálculo:

- impulso + compensação + perfil comportamental;
- projeção + estresse + sobrevivência;
- inflação pessoal + crescimento silencioso + mudança de estilo de vida.

## 6. Arquitetura de informação mobile

```text
FinanceOS
├── Hoje
│   ├── saldo disponível → Contas
│   ├── última sincronização
│   ├── ritmo semanal
│   ├── principal Alerta → Feed/origem
│   ├── próximo Compromisso
│   └── Movimentações recentes
├── Movimentações
│   ├── extrato, busca e filtros
│   ├── detalhe, Categoria e revisão
│   └── + lançamento manual
├── Planejar
│   ├── Metas
│   ├── parcelas e assinaturas
│   ├── próximas despesas e plano salarial
│   └── Simulações
├── Análises
│   ├── nove grupos consolidados
│   ├── Replay
│   └── Pierre
└── Mais
    ├── Contas
    ├── Open Finance e sincronização
    ├── perfil
    ├── integrações
    └── configurações
```

O botão `+` existe somente em **Movimentações** e cria lançamento manual. Rotas antigas permanecem como aliases/redirecionamentos até mobile e web deixarem de consumi-las.

### Ritmos de uso

- **Diário:** abrir Hoje, conferir sincronização, Alerta, Compromisso e Movimentações; corrigir itens prioritários.
- **Semanal:** revisar ritmo, categorias pendentes, assinaturas/parcelas e progresso de Metas.
- **Mensal:** fechar período, revisar saúde/evolução, consumir Replay, ajustar Planejamento e registrar decisões.

## 7. Requisitos e critérios de aceite

### Fase 1 — Verdade e segurança dos dados

| ID | Requisito | Critério de aceite | Dependências |
| --- | --- | --- | --- |
| FOS-101 | Remover números e resultados fictícios da experiência acessível. | Busca estática e testes não encontram projeções/percentuais demonstrativos em rotas acessíveis. | Inventário de telas e endpoints. |
| FOS-102 | Ocultar Replay mobile e Simulador enquanto forem mock. | Não há rota, menu ou CTA navegável para recurso mockado. | FOS-101. |
| FOS-103 | Remover ações Enviar, Receber, Pagar, Transferir e Investir. | Nenhuma ação sugere movimentação bancária inexistente. | Nenhuma. |
| FOS-104 | Isolar toda mutação por usuário. | Tentativa com ID de outro usuário retorna 404/403 e não altera dados. | Matriz de mutações. |
| FOS-105 | Completar correção de Categoria. | PATCH valida propriedade, define `needs_review=false`, grava confiança explícita e aprende Regra do Usuário. | Campos de revisão/confiança. |
| FOS-106 | Unificar sync às 00:30 São Paulo. | Scheduler, status e documentação informam o mesmo próximo horário; manual continua explícito. | Timezone disponível. |
| FOS-107 | Preservar endpoints consumidos pela web. | Teste de regressão web passa durante a migração. | Lista de consumidores. |

### Fase 2 — Entrada de dados e primeiro acesso

| ID | Requisito | Critério de aceite | Dependências |
| --- | --- | --- | --- |
| FOS-201 | Onboarding oferece Pluggy em destaque e Conta manual. | Novo Usuário conclui qualquer opção sem beco sem saída. | FOS-104, auth. |
| FOS-202 | Validar credenciais e abrir widget Pluggy. | Credencial inválida, indisponibilidade e conexão incompleta têm mensagens acionáveis. | Pluggy sandbox/produção. |
| FOS-203 | Emitir comprovante de importação. | Mostra contas, quantidade de transações, período, revisões e horário. | Status de sync por Usuário/Conexão. |
| FOS-204 | Gerenciar Conexões. | Reconectar, sincronizar e excluir não afetam Contas manuais independentes. | Modelo explícito de Conexão. |
| FOS-205 | Tratar sincronização parcial e ausência de dados. | Sucesso parcial não aparece como sucesso completo; zero dados é estado válido. | Resultado estruturado de sync. |
| FOS-206 | Restaurar sessão mobile. | Fechar e reabrir mantém sessão válida ou solicita login após refresh inválido. | Secure storage e refresh. |

### Fase 3 — Núcleo diário e navegação

| ID | Requisito | Critério de aceite | Dependências |
| --- | --- | --- | --- |
| FOS-301 | Implementar cinco destinos canônicos. | Estado e back stack de cada aba são estáveis; aliases antigos funcionam. | FOS-102, desenho de rotas. |
| FOS-302 | Criar `GET /overview`. | Uma resposta traz saldos, ritmo, sync, revisões, Alerta, Compromisso e recentes, todos do Usuário. | FOS-104/106. |
| FOS-303 | Montar Hoje somente com dados reais. | Cada card possui loading, vazio, erro e origem navegável quando aplicável. | FOS-302. |
| FOS-304 | Restringir `+` a lançamento manual. | Botão só aparece em Movimentações e abre formulário funcional. | FOS-301. |
| FOS-305 | Ligar saldo e Alerta às origens. | Saldo abre Contas; Alerta/notificação abre Feed ou entidade causadora. | FOS-302. |

### Fase 4 — Movimentações, Categorias e Alertas

| ID | Requisito | Critério de aceite | Dependências |
| --- | --- | --- | --- |
| FOS-401 | Tipar contrato mobile de Transação. | Modelo cobre Conta, Categoria, recorrência, confiança e revisão; campo incompatível falha em teste. | Contrato Go documentado. |
| FOS-402 | Ampliar `GET /transactions`. | Busca e filtros de período, Conta, Categoria, tipo e `needs_review` combinam com paginação. | Índices medidos se necessário. |
| FOS-403 | Revisar Categoria no detalhe. | Confirmar/corrigir atualiza a linha sem reload integral e produz Regra quando cabível. | FOS-105/401. |
| FOS-404 | Tornar Alertas adaptativos. | Gasto alto, salário, saldo crítico e duplicidade obedecem às fórmulas do plano. | 90 dias de histórico ou fallback declarado. |
| FOS-405 | Produzir eventos faltantes. | Mudança de assinatura, fechamento mensal e insight relevante aparecem uma vez no Feed. | Detectores e idempotência. |
| FOS-406 | Navegar da notificação à origem. | Todo Alerta pode ser lido e, quando houver origem, abre Transação/Compromisso relacionado. | Referências de origem no evento. |

Fórmulas obrigatórias do FOS-404:

- gasto elevado: `max(BRL 1.000, 3 × mediana dos débitos dos últimos 90 dias)`;
- salário: descrição compatível **ou** crédito recorrente em 2 dos últimos 3 meses, variação máxima de 20%;
- saldo crítico: saldo disponível menor que Compromissos até a próxima renda; BRL 500 apenas sem histórico;
- duplicidade: mesmo estabelecimento e valor em até 2 dias.

### Fase 5 — Planejamento real

| ID | Requisito | Critério de aceite | Dependências |
| --- | --- | --- | --- |
| FOS-501 | Separar Contas de Planejar. | Saldos ficam em Contas; parcelas/assinaturas em Planejar; nenhuma linguagem de portfólio. | FOS-301. |
| FOS-502 | Implementar quatro modos de Meta. | Limite, renda, economia e dívida usam as fontes definidas e período explícito. | Conta/Compromisso vinculável e saldo-base. |
| FOS-503 | Registrar ajustes manuais. | Ajuste soma ao automático, possui data/nota e nunca sobrescreve histórico. | Tabela incremental de ajustes. |
| FOS-504 | Completar ciclo de Meta. | Criar, editar, pausar, concluir, excluir e sugerir respeitam propriedade. | FOS-104/502. |
| FOS-505 | Recalcular progresso nos eventos corretos. | Sync, lançamento, correção relevante e ajuste disparam cálculo idempotente. | FOS-502/503. |
| FOS-506 | Unificar linha temporal de Planejamento. | Parcelas, assinaturas, fim de Compromissos, despesas e plano salarial ordenam-se por data. | Contratos consolidados. |

Modos do FOS-502:

- limite de gasto = débitos da Categoria no período;
- meta de renda = créditos no período;
- economia = evolução da Conta vinculada desde o saldo-base;
- quitação de dívida = redução do Compromisso vinculado desde o valor-base.

### Fase 6 — Simulador baseado em dados reais

| ID | Requisito | Critério de aceite | Dependências |
| --- | --- | --- | --- |
| FOS-601 | Projetar compra com histórico real. | Entrada inclui valor, parcelas, descrição e primeiro vencimento; sem constantes. | Saldos, renda, variáveis e Compromissos. |
| FOS-602 | Explicar saída da compra. | Retorna impacto, fluxo/saldo mensal, limite diário, renda comprometida, Alertas e confiança. | FOS-601 e qualidade de dados. |
| FOS-603 | Simular corte sem investimento fictício. | Retorna economia mensal, anual e acumulada, sem rentabilidade suposta. | Nenhuma. |
| FOS-604 | Persistir Simulações. | Salvar/nomear/listar/excluir valida propriedade e reproduz premissas. | Migração incremental. |
| FOS-605 | Declarar histórico insuficiente. | Dados insuficientes não são preenchidos com fallback numérico oculto. | Métrica de qualidade. |

### Fase 7 — Saúde e inteligência

| ID | Requisito | Critério de aceite | Dependências |
| --- | --- | --- | --- |
| FOS-701 | Calcular saúde sem dimensão fixa. | Fluxo, reserva, Compromissos, assinaturas, estabilidade e concentração vêm de dados. | Histórico e FOS-501. |
| FOS-702 | Expor metadados dos scores. | Todo score informa período, qualidade, confiança e dimensões usadas. | Contrato consolidado. |
| FOS-703 | Consolidar nove grupos. | Índice lista grupos/resumos/severidade; detalhes carregam sob demanda. | Mapeamento do inventário. |
| FOS-704 | Manter IA opcional. | Sem Groq/Gemini, cálculo, gráfico, Alerta e resumo determinístico continuam disponíveis. | Separação cálculo/narrativa. |
| FOS-705 | Restringir Pierre. | Pierre consulta resultados calculados e responde apenas perguntas abertas/explicações. | FOS-704. |
| FOS-706 | Implementar Replay mobile real. | Mostra gasto, maior compra, Categoria crescente, estabelecimento, melhora/piora e orientação. | Fechamento mensal e dados suficientes. |

### Fase 8 — Paridade web e limpeza

| ID | Requisito | Critério de aceite | Dependências |
| --- | --- | --- | --- |
| FOS-801 | Aplicar taxonomia e grupos na web. | Fluxos web usam a mesma linguagem após aceite mobile. | Fases 1–7 validadas no mobile. |
| FOS-802 | Migrar endpoints sem quebra. | Endpoint antigo só vira legado após nenhum consumidor mobile/web. | Telemetria/busca de consumidores. |
| FOS-803 | Remover código demonstrativo e duplicado. | Providers, telas, rotas e componentes sem consumidor são apagados com testes verdes. | FOS-802. |
| FOS-804 | Bloquear domínio de investimentos. | Não existe UI/domínio de ativos sem posição, preço médio, rentabilidade e patrimônio históricos reais. | Decisão futura explícita. |

## 8. Interfaces e evolução de dados

### API alvo

- `GET /overview`: saldos, ritmo semanal, sincronização, quantidade em revisão, principal Alerta, próximo Compromisso e Movimentações recentes.
- `GET /transactions`: acrescentar `search` e `needs_review`; retornar Conta, Categoria, recorrência, confiança e revisão.
- `PATCH /transactions/:id/category`: validar propriedade, concluir revisão e aprender Regra.
- Metas: acrescentar Conta/Compromisso vinculado, saldo/valor-base, modo de acompanhamento e histórico de ajustes.
- `PATCH /goals/:id`, `DELETE /goals/:id`, `POST /goals/:id/adjustments`.
- Simulador: contrato real de projeção, `GET` de Simulações salvas e exclusão por ID.
- Inteligência: índice consolidado de grupos, resumo, severidade, período e qualidade; detalhes continuam sob demanda.

### Regras de migração

- Migrações são incrementais e reversíveis; registros existentes são preservados.
- Colunas novas começam nullable/default seguro até backfill verificado.
- Endpoint antigo permanece funcional durante a troca de consumidores.
- Mudança de JSON exige teste de contrato Go ↔ Flutter antes do merge.
- Nenhuma mutação aceita somente o ID da entidade: sempre combina ID e Usuário no acesso ao banco.

## 9. Plano de execução e checklist

Uma fase só termina quando todos os requisitos da fase têm evidência registrada e os gates globais passam.

### Fase 0 — Governança

- [x] Criar glossário canônico sem detalhes técnicos.
- [x] Registrar baseline, inventário, requisitos, aceite, testes, riscos e DoD.
- [ ] Nomear responsável por produto, backend, agentes, mobile, web e QA.
- [ ] Criar usuário/conjunto de dados de validação sem dados pessoais de produção.

### Fase 1 — Verdade e segurança

- [x] FOS-101: números fictícios removidos da UI e endpoints incompletos bloqueados com `503`.
- [x] FOS-102: Simulador, Saúde incompleta e Replay mobile ocultos.
- [x] FOS-103: ações bancárias e módulo de investimentos inexistentes removidos da UI.
- [ ] FOS-104/FOS-105: implementados; falta teste de integração com dois Usuários e dados reais.
- [x] FOS-106: agenda/status unificados em 00:30 `America/Sao_Paulo` e cobertos por teste unitário.
- [ ] FOS-107: build web passou; regressão funcional e lint legado ainda pendentes.
- [x] Busca estática por mocks e ações falsas revisada.
- [ ] Matriz de propriedade de mutações aprovada.
- [x] Simulador e Replay mobile ocultos.

### Fase 2 — Entrada e primeiro acesso

- [ ] FOS-201: implementado; opções Pluggy/manual e formulário manual renderizados no dispositivo, falta aceite com novo Usuário.
- [ ] FOS-202: validação de credenciais e widget real implementados; widget abriu no dispositivo, faltam cenários inválido/indisponível/incompleto.
- [ ] FOS-203: endpoint e tela de comprovante implementados; falta concluir uma nova Conexão e validar os números importados.
- [ ] FOS-204: gerenciamento por Conexão implementado com reconexão, sync, nome PF/PJ e revogação na Pluggy; falta aceite destrutivo controlado.
- [x] FOS-205: execução manual retorna `running/completed/partial/failed`, quantidade nova real e motivo operacional; parcial validada no dispositivo.
- [x] FOS-206: restauração básica de sessão validada no dispositivo.
- [ ] Onboarding Pluggy e manual aceitos de ponta a ponta em dispositivo.
- [x] Comprovante e falhas parciais validados no SM-S948B; atualização da fonte permanece bloqueada pelo Item MeuPluggy atual.

### Fase 3 — Núcleo diário

- [x] FOS-301: cinco destinos canônicos e aliases antigos implementados.
- [x] FOS-302: `GET /overview` consolidado, publicado e validado com sessão real.
- [x] FOS-303: Hoje usa o overview real e explicita Alerta/Compromisso ausentes.
- [x] FOS-304: `+` restrito a Movimentações; Meta usa ação textual própria.
- [x] FOS-305: saldo e Alerta abrem a origem; alertas com IDs filtram as Movimentações relacionadas.
- [x] Cinco abas e aliases cobertos por teste; cinco abas validadas no dispositivo.
- [x] `/overview` validado por contrato Go ↔ Flutter e no SM-S948B.

### Fase 4 — Movimentações e Alertas

- [x] FOS-401: contrato Flutter tipado e validado com resposta real no SM-S948B.
- [x] FOS-402: busca, filtros combináveis e paginação publicados no `GET /transactions`.
- [x] FOS-403: detalhe mobile confirma/corrige Categoria, conclui revisão e aprende Regra do Usuário.
- [ ] FOS-404: fórmulas adaptativas implementadas e testadas; validação com nova Transação real bloqueada pelo Item MeuPluggy.
- [x] FOS-405 e FOS-406 implementados.
- [ ] Fórmulas adaptativas testadas com e sem histórico.
- [x] Origem de todo Alerta navegável quando existente.

### Fase 5 — Planejamento

- [ ] FOS-501 a FOS-506 implementados.
- [ ] Quatro modos de Meta e ajustes auditáveis validados.
- [ ] Linha temporal única aceita.

### Fase 6 — Simulador

- [x] FOS-601 a FOS-605 implementados.
- [x] Nenhuma constante financeira fictícia restante.
- [x] Histórico insuficiente exibido explicitamente.
- [x] Projeção de compra, corte de gastos, persistência e exclusão validados no SM-S948B.

### Fase 7 — Inteligência

- [x] FOS-701 a FOS-704 implementados.
- [ ] FOS-705: Pierre limitado a resultados calculados, com fallback determinístico; aceite no Render/dispositivo pendente.
- [ ] FOS-706: seis resultados do Replay cobertos por contrato; aceite no Render/dispositivo pendente.
- [x] Todos os cálculos funcionam sem provedor de IA (100% determinísticos no Go/Postgres).
- [ ] Replay mobile aceito com categoria crescente, melhora/piora e orientação no SM-S948B.

### Fase 8 — Web e limpeza

- [ ] FOS-801 a FOS-804 implementados.
- [ ] Regressão web completa verde.
- [ ] Consumidores antigos zerados antes de remoção.

## 10. Matriz de testes

| Camada | Casos mínimos | Comando/evidência |
| --- | --- | --- |
| Backend | auth; propriedade de cada mutação; filtros; Regra aprendida; sync idempotente; Alertas; Metas; Simulações; scores | `go test ./...` em `backend/` |
| Agentes | cada cálculo sem LLM; falha/quota Groq; falha/quota Gemini; histórico insuficiente; contratos | suíte Python definida pelo projeto |
| Mobile | cinco abas; sessão; onboarding; Categoria; lançamento; Metas; Alertas; Planejar; Replay; loading/vazio/erro | `flutter analyze`, `flutter test` |
| Contrato | respostas Go parseadas pelos modelos Flutter, incluindo campos ausentes/inválidos | teste versionado no backend ou mobile |
| Web | login; dashboard; transações; Contas; Simulador; relatórios | lint/build/test do `web/` |
| Build | APK debug instalável | `flutter build apk --debug` |
| Dispositivo | fluxos críticos e restauração no S26 Ultra | registro manual com build, data e resultado |

### Execução do baseline mobile — 2026-08-18

**Dispositivo:** Samsung SM-S948B, Android 16/API 36  
**APK:** debug, SHA-256 `059fb824e63004ad33ca4e2e47fe6a71777affdbc6147533167c5bb6c1ae460d`

| Verificação | Resultado | Evidência observada |
| --- | --- | --- |
| `flutter test` | passou | 3 testes executados, 3 aprovados |
| `flutter analyze` | falhou o gate | 29 avisos de lint; nenhum erro de compilação reportado |
| `flutter build apk --debug` | passou | APK gerado e instalado via ADB |
| Inicialização limpa | passou | app abriu após `force-stop`; nenhum crash Flutter/Android encontrado no log recente |
| Restauração de sessão | passou no cenário básico | app reinstalado com `-r`, encerrado e reaberto diretamente na sessão autenticada |
| Cinco branches da shell | passou estruturalmente | todas as abas abriram e mantiveram conteúdo sem crash |
| Navegação canônica | falhou | rótulos exibidos: `Home / Portfolio / Goal / History / Settings` |
| Verdade das ações | falhou | Home expõe `Receber`, `Enviar`, `Pagar`; Portfolio expõe `Enviar`, `Receber`, `Transferir`, `Investir` |
| Verdade dos números | falhou | Portfolio exibe `76% investido` fixo |
| Ocultação de mocks | falhou | Simulador está acessível em Mais e abre normalmente |
| Posição do botão `+` | falhou | botão aparece em Metas e Movimentações |
| Identidade visual | aprovada para preservação | tema escuro, cards, tipografia e destaque coral renderizam corretamente no aparelho |

### Execução da Fase 1 — 2026-08-18

**Backend:** commit `3aa7299` enviado a `origin/main`; Render respondeu `200` em `/health` após o push.  
**Dispositivo:** Samsung SM-S948B, Android 16/API 36; APK debug reinstalado via ADB.

| Verificação | Resultado | Evidência observada |
| --- | --- | --- |
| Testes backend | passou no escopo executável | `go test ./internal/... ./cmd/...`; novos testes de propriedade e agenda passaram |
| Build backend | passou | `go build ./cmd/server` |
| Testes mobile | passou | 3 testes executados, 3 aprovados |
| Analyze mobile | gate legado pendente | 27 avisos informativos preexistentes; nenhum erro de compilação |
| Build e instalação mobile | passou | APK debug gerado e instalado no SM-S948B |
| Home | passou FOS-101/103 no escopo visual | sem Receber/Enviar/Pagar, cartão fictício ou final `4688` |
| Contas | passou FOS-101/103 no escopo visual | sem `76% investido` ou ações Enviar/Receber/Transferir/Investir; Conta manual preservada |
| Mais | passou FOS-102 | Simulador e Saúde incompleta não estão navegáveis |
| Movimentações | passou FOS-103 no escopo visual | seletor usa Contas/Movimentações e não Investimentos/Despesas |
| Logs do dispositivo | passou | nenhuma `FATAL EXCEPTION`, `E/flutter` ou `Unhandled Exception` na rodada |
| Web build | passou | Next 16 compilou, tipou e gerou as páginas; Simulador/Saúde redirecionam |
| Web lint | falhou por baseline | 169 ocorrências preexistentes fora do escopo principal desta fase |
| `go test ./...` | bloqueado por baseline | utilitários `package main` duplicados na raiz de `backend/`; pacotes de produto passam isoladamente |

### Execução parcial da Fase 2 — 2026-08-18

**Backend:** commits `43f529b`, `c9f885a` e `7f7a3e4` enviados a `origin/main`; Render respondeu `200` em `/health`.
**Dispositivo:** Samsung SM-S948B, Android 16/API 36.
**APK:** debug mais recente, SHA-256 `cae1f7544d4f738c57918c5e53b0ac9511654607e36f39056311ada91ab0ae8f`.

| Verificação | Resultado | Evidência observada |
| --- | --- | --- |
| Testes backend | passou | `GOCACHE=/tmp/financeos-go-cache go test ./internal/... ./cmd/...`; teste do estado do comprovante incluído |
| Testes mobile | passou | 4 testes aprovados, incluindo presença das opções Pluggy/manual |
| Analyze mobile | gate legado pendente | 25 avisos informativos preexistentes; nenhum erro de compilação ou aviso novo do onboarding |
| Build e instalação | passou | APK debug gerado e reinstalado via ADB no SM-S948B |
| Sessão existente | passou | após reinstalação e `force-stop`, aplicativo restaurou diretamente a sessão autenticada |
| Entrada pela Pluggy | passou no escopo executado | opção recomendada abriu o widget oficial da Pluggy com credenciais já configuradas |
| Entrada manual | passou visualmente | formulário de nome, saldo e tipo renderizou integralmente; criação não foi enviada para evitar alterar dados do Usuário |
| Comprovante | implementação aprovada, aceite pendente | `GET /accounts/import-summary` combina Usuário + `item_id` e retorna contas, Transações, período, revisões e atualização reais |
| Conexões PF/PJ | passou no dispositivo | duas conexões Nubank agrupadas por `pluggy_item_id`; após sincronização exibiram finais distintos `14-8` e `17-6`; rótulo editável oferece `Nubank PF`/`Nubank PJ` |
| Revogação de Conexão | implementação aprovada, execução destrutiva não realizada | confirmação explícita, propriedade por Usuário e `DELETE /items/{id}` na Pluggy antes da remoção local |
| Falha silenciosa em Contas | corrigida no APK | erro de `/accounts` deixou de virar lista vazia; a UI agora exibe falha de carregamento |
| Migração no Render | passou após correção | imagem Docker não copiava `schema.sql`; commit `7f7a3e4` foi confirmado como live e `/accounts` voltou a responder com os campos novos |
| Sincronização por Conexão | passou no cenário executado | ações das duas conexões foram disparadas separadamente e preencheram seus identificadores sem duplicar os cartões na interface |
| Ambiente Pluggy | atenção | widget aberto identifica a aplicação Pluggy como demo; produção exige credenciais/ambiente de produção |
| Logs do dispositivo | passou | nenhuma exceção Flutter ou crash do FinanceOS; apenas mensagens do sistema/Google Play Services |

### Execução da FOS-205 e diagnóstico Pluggy — 2026-08-18

**Backend:** commits `ab793b8`, `1233cf8`, `6c3ffd5`, `f33ba61` e `e9035ff` enviados a `origin/main`; Render respondeu `200` em `/health`.

**Dispositivo:** Samsung SM-S948B, Android 16/API 36.

**APK:** debug final, SHA-256 `f3c23d9feaa4996845e41d8e726ea38f3e15199b4c0369555cd255314a2d74cd`.

| Verificação | Resultado | Evidência observada |
| --- | --- | --- |
| Resultado estruturado | passou | `POST /accounts/sync` devolveu `run_id`; comprovante consultou `GET /accounts/sync/:run_id` e exibiu andamento e parcial sem sucesso falso |
| Zero novas Transações | passou semanticamente | as duas execuções retornaram zero inserts novos e preservaram os totais existentes de 151 e 92 Transações |
| Motivo da parcial | diagnosticado | Pluggy respondeu `400: MeuPluggy item can't be updated`; o app exibiu o motivo real em vez de erro genérico |
| Horário de atualização | corrigido | `updated_at` passou a usar `lastUpdatedAt` da Pluggy; leitura de cache não grava mais `NOW()` como se a instituição tivesse atualizado |
| Reconexão | passou até o consentimento | mobile passou `updateItem` e abriu a tela MeuPluggy da conta final `17-6`; o aceite humano em “Continuar” ficou pendente por conter autorização pessoal |
| Distinção PF/PJ | passou estruturalmente | Conexões mostram finais `14-8` e `17-6`; extrato recebe rótulo da Conexão ou fallback curto `Conta • final ...` |
| Dados recentes | presentes no snapshot | extrato exibiu compras datadas de 18/08/2026; nova coleta da instituição depende de atualizar/recriar o Item Pluggy atual |
| Testes | passou | pacotes Go do produto e 4 testes Flutter aprovados; analyze permaneceu com 25 avisos informativos preexistentes |

### Revalidação das Conexões PF/PJ — 2026-08-20

**Dispositivo:** Samsung SM-S948B, Android 16/API 36, com a sessão preservada.

| Verificação | Resultado | Evidência observada |
| --- | --- | --- |
| Identidade das Conexões | passou | final `14-8` aparece como `Nubank PF` e final `17-6` como `Nubank PJ` em Open Finance |
| Origem no extrato | passou | cada Movimentação mostra o rótulo da Conexão; lançamentos recentes exibiram `Nubank PJ` |
| Atualização dos dados | passou | saldo consolidado mudou para R$ 472,88 e o extrato passou a conter Movimentações da PJ em 19/08/2026 |
| Evidência de Movimentações novas | passou | `A ESTUDANTIL LIVRARIA` (R$ 44,50) e `MORANGO SUL LTDA` (R$ 25,00), ambas identificadas como `Nubank PJ` |
| Logs do dispositivo | passou | nenhum `FATAL EXCEPTION`, `E/flutter` ou `Unhandled Exception` encontrado após a navegação |

### Execução parcial da Fase 3 — 2026-08-20

**Dispositivo:** Samsung SM-S948B, Android 16/API 36, sessão preservada.

**APK:** debug final, SHA-256 `36d25f2a723a2cb8b11ebd0be9e0496273cfe6b048d5f17bbd936f0c4b32a72e`.

| Verificação | Resultado | Evidência observada |
| --- | --- | --- |
| Navegação canônica | passou | `Hoje`, `Movimentações`, `Planejar`, `Análises` e `Mais` abriram no aparelho; aliases antigos permanecem no roteador |
| Ação `+` | passou | botão flutuante aparece somente em Movimentações; Planejar oferece `Nova meta` na barra superior |
| Origem do saldo | passou | toque no saldo da Hoje abriu Contas com saldo real e duas contas Nubank |
| Distinção PF/PJ | passou | Movimentações recentes exibiram `Nubank PJ`; Open Finance mantém `Nubank PF` e `Nubank PJ` por Conexão |
| Análises | passou após correção | cast do contrato `{ timeline: [...] }` foi reproduzido por teste, corrigido e validado nas abas Pierre e Cronograma |
| Agente de parcelas | corrigido e publicado | `datetime` ausente impedia gravar o cache; commit `2ed4a2f` enviado à `main` para deploy no Render |
| Overview autenticado | passou | uma chamada trouxe saldo disponível de R$ 472,88, atualização 20/08 00:07, ritmo semanal, contagem de revisão, Alerta e Movimentações recentes reais |
| Origem do Alerta | parcial | `Saldo em nível crítico` abriu Movimentações; filtro direto pelas Transações relacionadas permanece pendente |
| Verdade dos Compromissos | passou | vencimento 10/08 foi bloqueado; compra recorrente em sorveteria deixou de ser promovida a Compromisso sem confirmação; commits `77f3e37` e `d9a45e1` live no Render |
| Estados vazios | passou | Hoje exibiu `Nenhum compromisso confirmado`; ausência de Alerta usa estado explícito equivalente |
| Origem exata do Alerta | passou | `low_balance` abriu Contas no SM-S948B; alertas com `related_tx_ids` usam filtro autenticado de Movimentações coberto por testes Go/Flutter; commit `108118b` |
| Testes e build | passou | 7 testes Flutter, suíte dos pacotes Go, compilação Python, APK debug e instalação `adb -r`; analyze sem erro e com 22 avisos informativos preexistentes |
| Logs do dispositivo | passou | nenhum `FATAL EXCEPTION`, `E/flutter` ou `Unhandled Exception` após navegar pelas cinco abas |
| Contrato de Movimentações | passou | Conta, Categoria, recorrência, confiança e revisão tipadas; parser Flutter testado e dados reais carregados no SM-S948B |
| Busca, filtros e paginação | passou | commit `ed7074c` live no Render; testes Go cobrem busca, período, Conta, Categoria, tipo, revisão e metadados de página |
| Distinção PF/PJ no extrato | passou | a resposta real exibiu lançamentos de `Nubank PJ` e `Nubank PF` no mesmo extrato sem perder a origem |
| Revisão de Categoria | passou | `FARMACIA E DROGARIA NI`, da `Nubank PJ`, foi confirmada em Saúde no SM-S948B; após recarregar, revisão permaneceu concluída e confiança retornou 100% |
| Logs após revisão | passou | Luna Max removeu avisos do sistema e não encontrou `FATAL EXCEPTION`, `E/flutter`, `Unhandled Exception` ou `AndroidRuntime` |
| Categorias sem duplicação | passou | detector de acessibilidade que antes encontrou oito nomes repetidos retornou vazio no SM-S948B após o commit `7abfbee` ficar live |
| Alertas adaptativos | parcial | commit `4ba39b0` live e testes Go passaram; PF e PJ retornaram zero Transações novas porque o MeuPluggy recusou atualizar ambos os Itens com HTTP 400 |

### Cenários de aceite ponta a ponta

1. Primeiro acesso com Pluggy.
2. Primeiro acesso com Conta manual.
3. Sincronização sem novas Transações.
4. Sincronização parcial ou Pluggy indisponível.
5. Uso cotidiano iniciado por um Alerta.
6. Revisão semanal de Movimentações, Categorias e Planejamento.
7. Fechamento mensal e Replay.
8. IA sem cota ou indisponível.
9. Histórico insuficiente.
10. Usuário sem cartões, parcelas, assinaturas ou Metas.

## 11. Riscos e controles

| Risco | Impacto | Controle |
| --- | --- | --- |
| Vazamento entre Usuários por mutação baseada só em ID | Crítico | filtro `id + user_id`, testes negativos e revisão de todas as mutações |
| Resultado fictício parecer aconselhamento real | Alto | ocultar mock, rotular Simulação e remover linguagem de investimento/operação |
| Mudança de contrato quebrar web/mobile silenciosamente | Alto | compatibilidade temporária e testes Go ↔ Flutter/web |
| Sync duplicar Transações ou apresentar sucesso parcial como total | Alto | chave idempotente, resultado por Conexão e testes repetidos |
| IA indisponível bloquear produto | Alto | cálculo determinístico, fallback narrativo opcional e estados explícitos |
| Score preciso com pouco histórico | Médio | período, qualidade e confiança obrigatórios |
| Exclusão de Conexão apagar Conta manual | Alto | origem explícita e testes de independência |
| Muitas análises degradarem navegação/performance | Médio | nove grupos e carregamento sob demanda |
| Divergência de timezone | Médio | `America/Sao_Paulo` no cálculo, agenda, API e teste |
| Alterações locais ou repositórios separados se misturarem | Médio | commits independentes e escopo por repositório |

## 12. Definition of Done

Uma funcionalidade recebe `validada` somente quando:

- usa dados reais e não contém números, datas, percentuais ou narrativas demonstrativas;
- não oferece ação bancária ou de investimento inexistente;
- autenticação e propriedade são testadas em todas as leituras/mutações sensíveis;
- loading, vazio, erro, sucesso parcial e histórico insuficiente são explícitos;
- cálculo é determinístico e continua funcionando sem IA;
- contrato é tipado e coberto entre backend e consumidor;
- critérios de aceite e cenário ponta a ponta correspondente passam;
- testes relevantes, analyze, build e regressão web passam;
- evidência reproduzível, responsável, data e resultado constam no registro abaixo;
- não há falha silenciosa conhecida no caminho principal.

## 13. Registro de validação

Adicionar uma linha por funcionalidade e por nova tentativa. Não sobrescrever falhas anteriores.

| Requisito | Funcionalidade | Responsável | Evidência | Data | Resultado |
| --- | --- | --- | --- | --- | --- |
| Fase 0 | Glossário e plano de validação | Codex | `CONTEXT.md` e este documento, confrontados com o baseline estático | 2026-08-18 | concluído; não valida funcionalidades de produto |
| FOS-101 | Verdade dos dados | não atribuído | auditoria estática e execução no SM-S948B: ações falsas e `76% investido` visíveis | 2026-08-18 | bloqueado |
| FOS-102 | Ocultação de mocks | não atribuído | Simulador acessível por Mais no SM-S948B | 2026-08-18 | bloqueado |
| FOS-103 | Limites do produto | não atribuído | ações Enviar/Receber/Pagar/Transferir/Investir visíveis no SM-S948B | 2026-08-18 | bloqueado |
| FOS-104 | Isolamento de mutações | não atribuído | `UpdateCategory` atualiza por ID sem Usuário | 2026-08-18 | bloqueado |
| FOS-106 | Sincronização 00:30 | não atribuído | scheduler usa 00:30 São Paulo; status anuncia horários divergentes | 2026-08-18 | parcial |
| FOS-206 | Sessão mobile | não atribuído | reinstalação `-r`, force-stop e reabertura autenticada no SM-S948B | 2026-08-18 | passou cenário básico; expiração/refresh ainda pendente |
| FOS-301 | Navegação canônica | não atribuído | cinco branches abrem, mas mantêm rótulos antigos no SM-S948B | 2026-08-18 | bloqueado |
| FOS-304 | `+` somente em Movimentações | não atribuído | `+` também visível em Metas | 2026-08-18 | bloqueado |
| FOS-101 | Verdade dos dados — nova rodada | Codex | APK da Fase 1 no SM-S948B; busca estática mobile/web | 2026-08-18 | passou na UI; mocks de backend permanecem ocultos |
| FOS-102 | Ocultação de mocks — nova rodada | Codex | Mais sem Simulador/Saúde; rotas mobile ausentes; web redireciona | 2026-08-18 | passou |
| FOS-103 | Limites do produto — nova rodada | Codex | Home, Contas e Movimentações inspecionadas no SM-S948B | 2026-08-18 | passou |
| FOS-104 | Isolamento de mutações — implementação | Codex | commit `3aa7299`; propriedade em Categoria/Meta e chaves Pluggy por Usuário/Conta | 2026-08-18 | testes unitários passam; integração com dois Usuários pendente |
| FOS-105 | Correção de Categoria — implementação | Codex | PATCH usa Usuário; SQL conclui revisão, define confiança e grava Regra atomicamente | 2026-08-18 | teste real pendente |
| FOS-106 | Sincronização 00:30 — implementação | Codex | scheduler/status em São Paulo e `TestNextDailySync` | 2026-08-18 | passou em teste unitário; status autenticado no Render pendente |
| FOS-107 | Compatibilidade web — nova rodada | Codex | build Next 16 passou; endpoints backend preservados | 2026-08-18 | parcial: regressão funcional/lint pendentes |
| FOS-101 | Bloqueio de respostas fictícias | Codex | endpoints de Simulador e Saúde preservados, mas respondem `503` sem números mockados | 2026-08-18 | passou |
| FOS-301 | Navegação canônica — nova rodada | Codex | teste Flutter e execução das cinco abas no SM-S948B; aliases preservados no roteador | 2026-08-20 | passou |
| FOS-304 | `+` somente em Movimentações — nova rodada | Codex | busca estática, teste Flutter e inspeção no SM-S948B | 2026-08-20 | passou |
| FOS-305 | Origem do saldo e Alerta | Codex | saldo da Hoje abriu Contas no SM-S948B; navegação por Alerta ainda ausente | 2026-08-20 | parcial |
| FOS-302 | Overview consolidado | Codex | commit `3d099b7`, contrato Go/Flutter, resposta autenticada do Render e tela Hoje no SM-S948B | 2026-08-20 | passou |
| FOS-303 | Hoje com dados reais | Codex | saldo, sync, ritmo, Alerta e recentes vindos do overview; Compromisso heurístico rejeitado | 2026-08-20 | parcial: vazios explícitos pendentes |
| FOS-305 | Origem do saldo e Alerta — nova rodada | Codex | saldo abriu Contas e Alerta abriu Movimentações no SM-S948B | 2026-08-20 | parcial: origem exata do Alerta pendente |
| FOS-303 | Hoje com estados completos — rodada final | Codex | APK `36d25f2...` no SM-S948B exibiu dados reais e vazio explícito de Compromisso | 2026-08-20 | passou |
| FOS-305 | Origem do saldo e Alerta — rodada final | Codex | `low_balance` abriu Contas; IDs relacionados cobertos no backend e roteador Flutter | 2026-08-20 | passou |
| FOS-401 | Contrato tipado de Movimentação | Codex | testes Flutter, APK `a06ab210...` e extrato real carregado no SM-S948B | 2026-08-21 | passou |
| FOS-402 | Busca, filtros e paginação | Codex | suíte Go, commit `ed7074c` live no Render e compatibilidade mobile validada | 2026-08-21 | passou |
| FOS-403 | Revisão e aprendizado de Categoria | Codex | 10 testes Flutter, APK `d5691de...` e confirmação real no SM-S948B com confiança 100% após reload | 2026-08-21 | passou |
| FOS-404 | Alertas adaptativos | Codex | testes Go e commit `4ba39b0` live; sincronizações PF/PJ exibiram falha parcial explícita do MeuPluggy | 2026-08-21 | parcial: dados reais novos bloqueados pelo provedor |
| FOS-405 | Eventos faltantes e idempotência | Codex | testes Go de formatação/regras e commit `5667fbb` live; detectores cobrem assinatura, fechamento e categoria | 2026-08-21 | passou |
| FOS-406 | Navegação da notificação à origem | Codex | APK `d5691de...` no SM-S948B; alerta de saldo crítico abriu Contas e rota por IDs relacionada abre extrato filtrado | 2026-08-21 | passou |
| FOS-501 | Separar Contas de Planejar | Codex / Antigravity | Aba Planejar isolada de saldos bancários, contendo Metas e Linha do Tempo; sem linguagem de portfólio | 2026-08-21 | passou |
| FOS-502 | Quatro modos de Meta | Codex / Antigravity | Implementados modos `savings`, `debt_payoff`, `spending_limit`, `income_target` com saldo-base e fontes reais | 2026-08-21 | passou |
| FOS-503 | Ajustes manuais incrementais | Codex / Antigravity | Tabela `goal_adjustments` com nota, data e preservação do histórico de auditoria; modal mobile testado | 2026-08-21 | passou |
| FOS-504 | Ciclo de vida completo da Meta | Codex / Antigravity | CRUD e validação de propriedade dos vínculos foram corrigidos; sugestões sem dados reais retornam indisponível | 2026-08-21 | parcial |
| FOS-505 | Recálculo determinístico de progresso | Codex / Antigravity | Recálculo agora ocorre após sincronização, lançamento manual, correção de categoria e ajustes; falta validação integrada com dados de produção | 2026-08-21 | parcial |
| FOS-506 | Linha temporal unificada de Planejamento | Codex / Antigravity | Parcelas usam `next_due_date`/parcela atual e erros não são silenciosos; assinaturas e próximas despesas ainda requerem integração | 2026-08-21 | parcial |
| FOS-601 | Projetar compra com histórico real | Codex / Antigravity | Projeção valida datas e falhas de consulta; assinaturas e compromissos não parcelados ainda requerem integração | 2026-08-21 | parcial |
| FOS-602 | Explicar saída da compra | Codex / Antigravity | Projeção exibiu comprometimento de renda (2.9%), limite diário (R$ 10,12), alertas determinísticos e detalhamento mês a mês no SM-S948B | 2026-08-21 | passou |
| FOS-603 | Simular corte sem investimento fictício | Codex / Antigravity | Economia continua determinística; confiança agora depende do histórico real e falhas de consulta não viram sucesso | 2026-08-21 | parcial |
| FOS-604 | Persistir simulações | Codex / Antigravity | Persistência valida tipo, parâmetros e correspondência básica entre entrada e resultado; recomputação integral no servidor ainda pendente | 2026-08-21 | parcial |
| FOS-605 | Declarar histórico insuficiente | Codex / Antigravity | Compra e corte informam histórico insuficiente sem confiança fabricada | 2026-08-21 | parcial |
| FOS-701 | Calcular saúde sem dimensão fixa | Codex / Antigravity | Concentração e estabilidade deixaram de usar pontuações constantes; validação com séries reais ainda pendente | 2026-08-21 | parcial |
| FOS-702 | Expor metadados dos scores | Codex / Antigravity | Metadados expostos (`period_start`, `period_end`, `quality: HIGH`, `confidence: 95%`, `dimensions_used`) validados na tela Saúde Financeira no SM-S948B (score 48/100) | 2026-08-21 | passou |
| FOS-703 | Consolidar nove grupos | Codex / Antigravity | Grupos canônicos e estados derivados de saúde/dados foram alinhados; detalhes sob demanda e cobertura de cada grupo ainda pendentes | 2026-08-21 | parcial |
| FOS-704 | Manter IA opcional | Codex / Antigravity | Cálculos, pontuações, diagnósticos, gráficos e resumos funcionam 100% no Go/Postgres sem dependência externa de LLM | 2026-08-21 | passou |
| FOS-705 | Restringir Pierre | Codex | Backend envia somente Inteligência/Replay calculados; `ChatAgent` não consulta banco; falha de IA/agents retorna explicação determinística; testes Go e Python verdes | 2026-08-24 | parcial: publicação e aceite no dispositivo pendentes |
| FOS-706 | Implementar Replay mobile real | Codex | Contrato expõe gasto, maior compra, Categoria crescente, estabelecimento, melhora/piora e orientação; ausência de dados não vira saldo positivo; testes Go/Flutter verdes | 2026-08-24 | parcial: publicação e aceite no dispositivo pendentes |


Formato de evidência aceito: link para teste automatizado e sua execução, captura/log sanitizado de cenário real, ou checklist manual reproduzível de dispositivo. “Funciona na minha máquina”, screenshot sem contexto e mera referência a código não promovem estado para `validada`.

## 14. Assunções operacionais

- Mobile é a experiência principal; backend e agentes são ajustados para sustentá-lo.
- Web permanece compatível e recebe paridade depois da validação mobile.
- Pluggy/Open Finance é a fonte principal; entrada manual permanece disponível.
- A identidade visual premium atual é preservada.
- Backend/monorepo e `mobile/` podem exigir históricos Git e commits independentes.
- Alterações locais existentes devem ser preservadas.
- Nenhuma dependência visual nova é adicionada sem necessidade comprovada.
- Não se cria domínio de investimentos sem ativos, posições, preço médio, rentabilidade e histórico patrimonial reais.
