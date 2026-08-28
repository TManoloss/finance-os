# FinanceOS — contexto operacional canônico

Atualizado em 28/08/2026, no fuso `America/Sao_Paulo`.

Este documento permite que uma nova sessão continue o trabalho sem reconstruir o histórico. Ele descreve arquitetura, fluxos, operação e estado corrente. Não substitui:

- [`../CONTEXT.md`](../CONTEXT.md): glossário canônico do domínio, sem implementação;
- [`FINANCEOS-VALIDATION-PLAN.md`](FINANCEOS-VALIDATION-PLAN.md): requisitos FOS, critérios, evidências e status oficiais;
- [`../AGENTS.md`](../AGENTS.md): divisão obrigatória de trabalho entre agente principal e Luna Max.

Quando houver divergência, o código e os testes definem o comportamento atual; o plano define o comportamento desejado. Registre a divergência antes de alterá-la.

## 1. Produto e decisões fixas

FinanceOS é o aplicativo financeiro pessoal do próprio usuário. Ele acompanha fatos financeiros, explica o que aconteceu, apoia decisões e acompanha próximos passos.

- Mobile é a experiência principal; backend e agentes o sustentam primeiro.
- Web permanece compatível e recebe paridade depois da validação mobile.
- Pluggy/Open Finance é a fonte principal. Conta e Movimentação manuais continuam válidas.
- A identidade premium escura existente deve ser preservada.
- Funcionalidade fictícia, mockada ou incompleta fica oculta até usar dados reais.
- Cálculos financeiros são determinísticos. IA é opcional e apenas explica resultados calculados.
- Navbar canônica: `Hoje / Movimentações / Planejar / Análises / Mais`.
- FinanceOS não envia, recebe ou paga dinheiro e não executa investimentos.
- Não criar domínio de investimentos sem ativos, posições, preço médio, rentabilidade e histórico patrimonial reais.
- Horário financeiro oficial: `America/Sao_Paulo`. Moeda principal: BRL.

O usuário possui duas Conexões Nubank reais, uma PF e outra PJ. Elas devem permanecer distinguíveis em toda interface. Uma Conexão pode expor mais de uma Conta; não trate Conexão e Conta como sinônimos.

## 2. Topologia do sistema

```text
Flutter mobile ──HTTPS/JWT──> API Go/Echo ──SQL──> PostgreSQL
                                  │
                                  ├──HTTPS──> Pluggy/Open Finance
                                  │
                                  └──HTTP interno──> FastAPI agents ──opcional──> Groq/Gemini

Next.js web ─────HTTPS/JWT───────> API Go/Echo
```

### Componentes

| Área | Local | Responsabilidade |
| --- | --- | --- |
| API | `backend/` | autenticação, propriedade dos dados, sincronização, consultas, cálculos determinísticos e contratos públicos |
| Agentes | `agents/` | classificação assistida, relatórios legados e explicações opcionais do Pierre |
| Mobile | `mobile/` | experiência principal Flutter/Riverpod/GoRouter instalada no Samsung |
| Web | `web/` | cliente Next.js mantido compatível; paridade completa ainda pendente |
| Banco | `backend/internal/db/schema.sql` | esquema incremental executado na inicialização do backend |
| Deploy | `Dockerfile` | imagem única: FastAPI interno em `127.0.0.1:8000` e Go exposto na porta do Render |
| Ambiente local | `docker-compose.yml` | PostgreSQL e Adminer; serviços de aplicação rodam separadamente |

O backend Go é a autoridade para cálculos financeiros e isolamento por Usuário. O serviço Python nunca deve ser requisito para saldo, score, gráfico, alerta, resumo, Replay ou simulação.

## 3. Repositórios e estado Git

Workspace:

```text
/home/manoelfelip/Documentos/projetos/Dinheiro
```

Existem dois históricos Git:

1. O repositório raiz contém backend, agentes, web, documentação e também rastreia uma cópia da árvore `mobile/`.
2. `mobile/` possui seu próprio `.git` e é o repositório Flutter efetivamente compilado e instalado.

Estado salvo em 28/08/2026:

- raiz: `main` e `origin/main` em `d462f86`;
- mobile: `main` e `origin/main` em `29fa748`;
- o repositório mobile está limpo;
- o repositório raiz mostra os arquivos de `mobile/` modificados porque compara a árvore do repositório mobile novo com o snapshot histórico do repositório pai. Isso é esperado; não restaure nem commite esses arquivos no repositório raiz por acidente.

Nunca use limpeza destrutiva para “resolver” essa diferença. Antes de qualquer commit, execute os dois comandos:

```bash
git status --short
git -C mobile status --short
```

Commits recentes importantes no repositório raiz:

```text
d462f86 fix(locale): format financial messages in pt-BR
ed96640 fix(chat): hide provider failures behind fallback
88c0522 feat(intelligence): complete Pierre and Replay contracts
0168cfa feat(categories): complete hierarchical category taxonomy
392c168 fix(classifier): preserve bakery subcategory precedence
8e8027d fix(db): migrate category parent links safely
9cd5be3 feat(categories): add subcategories and channel-aware convenience index
4e661e0 feat(intelligence): calculate deterministic report insights
f632bef fix(finance): enforce real data validation and goal ownership
a1c700f feat(intelligence): implement real-data financial health pillars, consolidated 9 groups, and monthly replay
c089e92 feat(simulator): implement real-data purchase and cut simulation with persistence
a5f56be feat(goals): implement 4 goal modes, manual adjustments, and planning timeline
5667fbb feat(feed): implement missing events for subscriptions, monthly close and category insights
4ba39b0 feat(feed): make financial alerts adaptive
ed7074c feat(transactions): add searchable review filters
3d099b7 feat(api): add authenticated financial overview
```

Commits recentes no mobile:

```text
29fa748 feat(mobile): complete real-data core experience
44c8380 fix(auth): refresh expired mobile sessions
bf8eac3 fix(transactions): separate category and subcategory selectors
b4be11e test(categories): cover parent category display
9e48ea4 feat(categories): show parent category in transactions
b19ee16 feat(categories): display transaction subcategories
```

## 4. Deploy e publicação

Regra explícita do usuário: toda alteração de backend/agentes que precise chegar ao aplicativo deve ser commitada e enviada para a `main` do repositório raiz. O Render constrói a imagem a partir desse push.

```text
API pública: https://finance-os-d3nm.onrender.com/api/v1
Health:      https://finance-os-d3nm.onrender.com/health
```

Fluxo de publicação:

1. Rode testes proporcionais ao risco.
2. Faça commit apenas dos arquivos do repositório correto.
3. Envie a `main` do repositório raiz com `git push origin main`.
4. Confirme no Render `Deploy live for <hash>`; health verde sozinho não identifica o commit.
5. Repita o cenário autenticado no app.
6. Atualize a evidência no plano somente depois do aceite real.

Última publicação registrada: `d462f86` foi enviada à `main`. O Pierre seguro foi aceito no aparelho após `ed96640`; a formatação BRL de `d462f86` passou na suíte Go, mas a conferência visual final foi interrompida quando a depuração Samsung desconectou.

## 5. Fluxos principais

### Autenticação e sessão

- `POST /auth/register`, `/auth/login` e `/auth/refresh` emitem sessão.
- Tokens ficam no `FlutterSecureStorage`.
- O `ApiClient` adiciona Bearer token e serializa renovações concorrentes.
- Em `401`, tenta refresh uma vez e repete a requisição original.
- Se refresh falhar, remove os tokens e notifica expiração uma única vez; a UI mostra mensagem amigável, nunca `DioException` bruto.
- Ao reabrir, o mobile prioriza o refresh token e restaura a sessão.

Arquivos centrais: `backend/internal/handler/auth.go`, `mobile/lib/core/api/api_client.dart`, `mobile/lib/features/auth/presentation/auth_provider.dart`.

### Primeiro acesso e entrada de dados

- Onboarding oferece Pluggy como opção principal e Conta manual como alternativa.
- Pluggy: credenciais do Usuário → connect token → widget → item/conexão → sincronização → comprovante de importação.
- O comprovante mostra Contas, Transações importadas, período, itens para revisão e atualização.
- Conexões podem ser renomeadas, reconectadas, sincronizadas e excluídas; Contas manuais são independentes.
- Consentimentos bancários e exclusões são sempre confirmados pelo usuário humano.

Arquivos centrais: `backend/internal/handler/accounts.go`, `backend/internal/service/sync_service.go`, `mobile/lib/features/onboarding/`, `mobile/lib/features/cards/presentation/pluggy_connect_screen.dart`.

### Sincronização

- Job oficial: 00:30 no fuso `America/Sao_Paulo`.
- Sincronização manual é assíncrona e limitada; retorna `run_id` e status `running/completed/partial/failed`.
- O serviço importa Contas e Transações de forma idempotente, classifica, produz eventos e recalcula Metas.
- Falha parcial da Pluggy é mostrada com motivo operacional; não converter em sucesso ou erro genérico.
- O Item MeuPluggy das conexões existentes já respondeu `400 (item cant be updated)`. Isso é bloqueio externo conhecido, não prova perda dos dados já importados.

Arquivos centrais: `backend/internal/jobs/scheduler.go`, `backend/internal/handler/accounts.go`, `backend/internal/service/sync_service.go`.

### Hoje e navegação

`GET /overview` agrega saldos, ritmo semanal, sincronização, revisão, Alerta principal, próximo Compromisso conhecido e Movimentações recentes.

- Hoje usa somente esse contrato real.
- Saldo abre Contas.
- Alerta abre Feed ou sua origem.
- Compromisso expirado ou apenas inferido não aparece como fato futuro.
- Rotas canônicas: `/today`, `/movements`, `/planning`, `/analytics`, `/more`.
- Aliases antigos continuam redirecionando durante a migração.

Arquivos centrais: `backend/internal/handler/overview.go`, `mobile/lib/features/dashboard/`, `mobile/lib/core/router/app_router.dart`.

### Movimentações, Categoria e Subcategoria

- `GET /transactions` suporta texto, IDs, período, Conta, Categoria, direção, revisão e paginação.
- O contrato inclui Conta, Categoria/Subcategoria, recorrência, confiança e `needs_review`.
- `PATCH /transactions/:id/category` valida propriedade, conclui revisão, define confiança e aprende Regra do Usuário.
- A hierarquia usa uma Categoria pai e uma Subcategoria filha. Exemplo: `Alimentação > Delivery`; a Subcategoria refina a análise sem virar Categoria de primeiro nível.
- A UI mostra Categoria e, abaixo, Subcategoria; os seletores são separados.
- Listagens deduplicam nomes globais/pessoais e preferem a Categoria do Usuário.
- O classificador preserva regras específicas antes das genéricas; índices como conveniência usam os canais/subcategorias, não apenas a Categoria ampla.

Arquivos centrais: `backend/internal/handler/transactions.go`, `backend/internal/handler/categories.go`, `backend/internal/service/classifier_service.go`, `agents/app/services/classifier.py`, `mobile/lib/features/transactions/`.

### Feed e Alertas

- Gasto elevado: maior entre R$ 1.000 e três vezes a mediana de débitos dos últimos 90 dias.
- Salário: descrição compatível ou crédito recorrente em dois dos últimos três meses, com variação máxima de 20%.
- Saldo crítico: disponível abaixo dos Compromissos até a próxima renda; R$ 500 só é fallback sem histórico.
- Duplicidade: mesmo estabelecimento e valor dentro de dois dias.
- Também existem mudança de assinatura, fechamento mensal e insight relevante por Categoria.
- Eventos são idempotentes, podem ser lidos e abrem a origem quando há IDs relacionados.
- Valores textuais são formatados em pt-BR pelo helper compartilhado `formatAmount`.

Arquivos centrais: `backend/internal/service/feed_service.go`, `backend/internal/handler/feed.go`.

### Planejamento e Metas

- Modos: economia, quitação de dívida, limite de gasto e meta de renda.
- Progresso vem de Conta/Compromisso/Movimentações conforme o modo.
- Ajustes manuais são incrementais e auditáveis; não substituem o valor observado.
- CRUD, pausa, conclusão, exclusão, sugestões e histórico de ajustes existem.
- Recálculo é acionado por sync, lançamento manual, correção de Categoria e ajuste.
- A timeline reúne parcelas, assinaturas, renda prevista, próximas despesas e prazos de Metas.

Arquivos centrais: `backend/internal/service/goals_service.go`, `backend/internal/handler/goals.go`, `mobile/lib/features/goals/presentation/goals_screen.dart`.

### Simulador

- Compra usa saldo, renda, gastos, parcelas e Compromissos disponíveis no histórico.
- Saída: impacto mensal, fluxo por mês, saldo final, limite diário, comprometimento, alertas e confiança.
- Corte mostra economia mensal, anual e acumulada; nunca rentabilidade de investimento.
- Simulações podem ser salvas, listadas e excluídas.
- Histórico insuficiente é explícito; o servidor rejeita resultados arbitrários incompatíveis.

Arquivos centrais: `backend/internal/service/simulator_service.go`, `backend/internal/handler/simulator.go`, `mobile/lib/features/simulator/`.

### Saúde, Inteligência, Replay e Pierre

- Saúde calcula fluxo de caixa, reserva, compromissos/dívidas, assinaturas, estabilidade e concentração com dados reais.
- Respostas informam período, qualidade, confiança e dimensões utilizadas.
- A interface consolidada organiza nove grupos canônicos; cobertura sob demanda de todos os detalhes ainda é pendência de FOS-703.
- Replay mensal real contém seis telas: total gasto, maior compra, Categoria em crescimento, principal estabelecimento, melhora/piora e orientação determinística.
- Pierre recebe do Go apenas Inteligência e Replay já calculados. Ele explica; não consulta o banco nem cria métricas financeiras.
- Falha, cota ou modelo indisponível em Groq/Gemini gera fallback determinístico HTTP 200. Strings `ERRO_*` nunca são exibidas ao usuário.

Arquivos centrais: `backend/internal/service/financial_health_service.go`, `backend/internal/service/deterministic_insights_service.go`, `backend/internal/service/visual_reports_service.go`, `backend/internal/handler/chat.go`, `agents/app/services/chat_agent.py`, `mobile/lib/features/health/`, `mobile/lib/features/reports/`, `mobile/lib/features/chat/`.

## 6. Grupos de endpoints públicos

Todos abaixo de `/api/v1`; exceto autenticação, usam JWT.

| Grupo | Operações principais |
| --- | --- |
| `/auth` | register, login, refresh |
| `/overview`, `/me` | núcleo da sessão e Hoje |
| `/accounts` | listar/criar, Pluggy token, sync/status/comprovante, chaves, labels e exclusões |
| `/transactions` | listar/filtrar, manual, corrigir Categoria, resumo |
| `/categories` | taxonomia hierárquica disponível ao Usuário |
| `/feed` | listar, não lidos, ler um/todos |
| `/goals`, `/planning/timeline` | CRUD, ajustes, recálculo, sugestão e timeline |
| `/simulator` | compra, corte, salvar, listar e excluir |
| `/reports`, `/intelligence` | saúde, grupos analíticos, Replay e detalhes legados |
| `/chat` | Pierre com contexto calculado e fallback |
| `/cards` | parcelas, fatura e assinaturas |
| `/merchants` | estabelecimentos e perfil |

A lista exata e seus aliases está em `backend/internal/router/router.go`; consulte o roteador em vez de duplicar todas as rotas aqui.

## 7. Estado de validação

O registro por requisito no plano oficial é a autoridade. Resumo operacional em 28/08/2026:

| Área | Estado real |
| --- | --- |
| Governança | glossário/plano prontos; responsáveis formais e dataset sem dados pessoais pendentes |
| Verdade e segurança | base implementada; teste integrado de propriedade com dois Usuários e regressão web ainda pendentes |
| Onboarding/Open Finance | implementação presente; primeiro acesso completo e cenários Pluggy adversos ainda pendentes |
| Núcleo diário | cinco abas, overview, origens e sessão validados no Samsung |
| Movimentações | contrato, filtros, revisão e hierarquia validados; Alertas adaptativos aguardam nova Transação real por bloqueio MeuPluggy |
| Planejamento | quatro modos e fluxos implementados; integração completa da timeline e progresso com produção ainda parcial |
| Simulador | fluxo principal aceito no aparelho; integrações de Compromissos/confiança e recomputação integral de salvamento ainda parciais no registro FOS |
| Inteligência | cálculos determinísticos, Pierre e Replay validados; FOS-701 e FOS-703 ainda têm cobertura real/detalhes sob demanda pendentes |
| Web e limpeza | FOS-801 a FOS-804 e regressão funcional ainda pendentes |

Não promova `parcial` para `validado` apenas porque existe código ou teste unitário. A evidência aceita é teste automatizado no contrato correto, cenário real sanitizado ou checklist manual reproduzível.

## 8. Testes e comandos

Ambiente local, depois de configurar `.env` a partir de `.env.example` sem versionar segredos:

```bash
docker compose up -d
cd agents && venv/bin/uvicorn main:app --host 127.0.0.1 --port 8000
cd backend && go run cmd/server/main.go
cd web && npm run dev
cd mobile && flutter run
```

`run_all.sh` é um atalho legado que encerra processos nas portas 8080, 8000 e 3000 antes de iniciar. Inspecione-o e prefira os comandos separados quando houver outros serviços locais.

Backend:

```bash
cd backend
GOCACHE=/tmp/financeos-go-cache go test ./internal/...
```

Agentes:

```bash
cd agents
venv/bin/python -m unittest discover -s tests -v
```

Mobile:

```bash
cd mobile
flutter analyze
flutter test
flutter build apk --debug
sha256sum build/app/outputs/flutter-apk/app-debug.apk
```

Web:

```bash
cd web
npm run build
```

Últimos gates registrados antes deste handoff:

- backend `go test ./internal/...`: passou;
- agentes: 4 testes passaram; existe `FutureWarning` da biblioteca legada `google.generativeai`;
- mobile: 14 testes passaram; `flutter analyze` sem erro/warning e com 22 lints informativos preexistentes; APK debug compilou;
- mobile `29fa748` foi instalado com sucesso no SM-S948B.

## 9. Validação no Samsung

Aparelho: Samsung SM-S948B, Android 16/API 36. Package: `com.example.finance_os`.

A Samsung pode desativar a depuração USB depois de aproximadamente dez minutos. `adb devices -l` vazio normalmente significa que o usuário precisa reconectar/reativar a depuração; não é evidência de crash.

```bash
adb devices -l
adb install -r mobile/build/app/outputs/flutter-apk/app-debug.apk
adb shell am force-stop com.example.finance_os
adb shell am start -n com.example.finance_os/.MainActivity
```

Para investigar erro, capture somente o necessário, remova tokens, credenciais e dados pessoais e siga a instrução de Luna registrada em `AGENTS.md`. Ruídos comuns de Samsung, SurfaceFlinger, GameManager e sistema não são stack trace do FinanceOS.

Nunca automatize consentimento bancário, reconexão final ou exclusão sem confirmação atual do usuário. Antes de tocar por coordenada, inspecione a tela ou a árvore de acessibilidade.

## 10. Última validação real

No SM-S948B, em 24/08/2026:

- sessão restaurou e Hoje carregou dados reais;
- Replay exibiu as seis telas, incluindo crescimento de Categoria, evolução e orientação;
- Pierre inicialmente expôs erro técnico de Groq/Gemini;
- a reprodução foi coberta em `agents/tests/test_chat_agent.py`;
- `ed96640` passou a tratar qualquer resposta `ERRO_*` como indisponibilidade e usar fallback;
- após o deploy, a mesma pergunta retornou somente saldo, evolução, Categoria principal e orientação calculada;
- `d462f86` corrigiu separadores BRL em Replay, Pierre e Feed; testes passaram, mas faltou a última inspeção visual porque o ADB desconectou.

Nenhum valor pessoal observado no aparelho deve ser copiado para documentação, logs versionados ou testes.

## 11. Riscos e pendências prioritárias

1. FOS-703: completar detalhes sob demanda e cobertura real dos nove grupos consolidados.
2. FOS-504 a FOS-506: fechar validação integrada de propriedade, recálculo e timeline de Planejamento.
3. FOS-601/FOS-603 a FOS-605: integrar todos os Compromissos e recomputar no servidor qualquer simulação salva.
4. FOS-201 a FOS-204: validar primeiro acesso Pluggy/manual e estados adversos com dataset seguro.
5. FOS-404: validar Alertas adaptativos quando a Pluggy entregar uma Transação nova.
6. FOS-104/FOS-105: teste de isolamento com dois Usuários.
7. Fase 8: paridade web, regressão e remoção de telas/providers demonstrativos.
8. Migrar `google.generativeai` apenas quando a manutenção dos agentes exigir; hoje é warning conhecido, não bloqueio funcional.

## 12. Protocolo para a próxima sessão

Antes de editar:

1. Leia `AGENTS.md`, `CONTEXT.md`, este handoff e o requisito FOS correspondente no plano.
2. Confira os dois estados Git e preserve mudanças fora do escopo.
3. Declare o fato usado: real, calculado, inferido ou explicado opcionalmente por IA.
4. Identifique o contrato Go ↔ Flutter e o teste que falhará se ele quebrar.
5. Delegue boilerplate/testes mecânicos ao Luna Max; mantenha arquitetura e lógica no agente principal.
6. Publique backend/agentes somente pela `main` raiz e aguarde o hash live no Render.
7. Valide no Samsung quando o requisito for mobile e registre evidência sanitizada.
8. Atualize o plano; este handoff só muda quando arquitetura, operação, estado Git ou próximo trabalho mudarem.

O handoff está completo quando a nova sessão consegue dizer qual requisito está executando, qual dado é real, qual repositório mudará, como será publicado, qual teste prova o contrato e qual evidência permite promover o status.
