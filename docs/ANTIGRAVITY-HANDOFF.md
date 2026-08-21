# FinanceOS — handoff para o Antigravity CLI

Atualizado em 21/08/2026, horário de São Paulo. Este documento complementa, mas não substitui, [`FINANCEOS-VALIDATION-PLAN.md`](FINANCEOS-VALIDATION-PLAN.md), que é a fonte oficial de requisitos e evidências, e [`../CONTEXT.md`](../CONTEXT.md), que é o glossário canônico.

## 1. O que este produto é para o dono

FinanceOS é o aplicativo financeiro pessoal do próprio usuário. Ele deve acompanhar fatos financeiros reais, explicar o que aconteceu, orientar decisões e acompanhar próximos passos. Não é banco, carteira de pagamentos nem plataforma de investimentos.

Decisões do usuário que não devem ser rediscutidas:

- Mobile é a experiência principal. Backend e agentes devem sustentá-lo primeiro.
- Web deve continuar compatível e receber paridade depois da validação mobile.
- Pluggy/Open Finance é a fonte principal; Conta e Movimentação manuais continuam válidas.
- A identidade visual premium escura atual deve ser preservada.
- Funcionalidade mockada, fictícia ou incompleta deve ficar oculta, nunca parecer real.
- Cálculos financeiros são determinísticos. IA é opcional para explicações e sempre tem fallback.
- Navbar canônica: `Hoje / Movimentações / Planejar / Análises / Mais`.
- FinanceOS não envia, recebe, paga ou investe dinheiro. Não usar ações `Enviar`, `Receber`, `Pagar` ou `Investir` como capacidades do produto.
- Não criar domínio de investimentos sem ativos, posições, preço médio, rentabilidade e histórico patrimonial reais.
- Horário financeiro: `America/Sao_Paulo`. Moeda principal: BRL.

## 2. Contexto real do usuário e do aparelho

- O usuário tem duas Conexões Nubank reais: uma PF e outra PJ.
- Elas precisam permanecer distintas em toda interface. Os rótulos atuais são `Nubank PF` e `Nubank PJ`.
- A Conta PJ tem movimentação frequente e foi a principal evidência de importações recentes.
- Aparelho de validação: Samsung SM-S948B (S26 Ultra), Android 16/API 36.
- Package Android: `com.example.finance_os`.
- A depuração USB da Samsung é desativada por segurança depois de aproximadamente dez minutos. Se `adb` perder o aparelho, peça ao usuário para reconectar/reativar; não trate isso como falha do app.
- A sessão permaneceu autenticada após `adb install -r`, force-stop e reabertura.
- Nunca imprimir `.env`, tokens, credenciais Pluggy, chaves de IA ou conteúdo do armazenamento seguro.

## 3. Repositórios, deploy e regra operacional crítica

Workspace:

```text
/home/manoelfelip/Documentos/projetos/Dinheiro
```

Há dois históricos Git que precisam ser tratados separadamente:

- Repositório raiz: backend, agentes, web, documentação e também uma árvore Flutter histórica.
- Repositório `mobile/`: aplicativo Flutter que está sendo instalado e validado no Samsung.

Não use `git reset --hard`, `git checkout --` nem restaure arquivos em massa. Há muitas mudanças mobile locais legítimas e ainda não commitadas. Não presuma que uma árvore suja é descartável.

Regra explícita do usuário: qualquer modificação de backend que precise chegar ao app deve ser commitada e enviada para `main`; o Render faz o build/deploy a partir dela. Não considere o backend validado no aparelho antes de o usuário ou o Render confirmar `Deploy live for <commit>`.

Backend público usado pelo app:

```text
https://finance-os-d3nm.onrender.com/api/v1
```

Health:

```bash
curl -fsS https://finance-os-d3nm.onrender.com/health
```

## 4. Regras de trabalho já persistidas

Leia [`../AGENTS.md`](../AGENTS.md) antes de trabalhar. Resumo:

- Agente principal decide arquitetura, contratos e lógica complexa.
- Boilerplate e testes repetitivos vão para Luna Max quando disponível.
- Logs de dispositivo devem ser sanitizados e enviados a Luna Max com a instrução exata registrada no arquivo.
- Depois de duas tentativas automáticas malsucedidas para a mesma causa, o principal assume a depuração profunda.
- Preserve alterações do usuário e a separação dos repositórios.

O fluxo que funcionou foi: definir contrato pequeno, delegar teste/UI mecânica, revisar diff, rodar testes, publicar backend isoladamente, esperar Render live, instalar APK, validar no aparelho, triar logs e só então atualizar o documento de validação.

## 5. Estado atual das fases

### Fases 1 a 3

Concluídas no escopo documentado:

- Remoção/ocultação de números e ações falsas na experiência mobile.
- Isolamento de mutações financeiras por Usuário.
- Correção de Categoria conclui revisão, define confiança e aprende Regra do Usuário.
- Sincronização oficial 00:30 em São Paulo e sincronização manual explícita.
- Onboarding Pluggy/Conta manual e estados de erro/parcial.
- Cinco destinos canônicos e aliases de rotas antigas.
- `GET /overview` autenticado e Home real.
- Saldo abre Contas; `low_balance` abre Contas; alertas com IDs abrem Movimentações filtradas.
- Compromissos vencidos ou meramente inferidos não aparecem como fatos futuros.

### Fase 4

- FOS-401: passou. Contrato Flutter tipa Conta, Categoria, recorrência, confiança e revisão.
- FOS-402: passou. Backend aceita busca, período, Conta, Categoria, direção, `needs_review` e paginação.
- FOS-403: passou no Samsung com dado real.
- FOS-404: implementado e coberto por testes Go, commit live; validação com nova Transação real está bloqueada pelo MeuPluggy.
- FOS-405: passou em testes e build. Detectores de mudança de assinatura, fechamento mensal e insight por categoria implementados com idempotência (commit `5667fbb`).
- FOS-406: passou no Samsung com APK `d5691de...`. Alerta de saldo crítico abriu `/accounts`; alertas com IDs vinculados filtram `/movements?ids=...`.

Próximo ponto lógico: Fase 5 (Planejamento Real — FOS-501 a FOS-506).

## 6. Últimos commits importantes na `main`

Em ordem do mais recente:

```text
a5f56be feat(goals): implement 4 goal modes, manual adjustments, and planning timeline (FOS-502..506)
5667fbb feat(feed): implement missing events for subscriptions, monthly close and category insights (FOS-405)
d3f5781 docs: record adaptive alert validation limits
4ba39b0 feat(feed): make financial alerts adaptive
7abfbee fix(categories): deduplicate names in listings
b9f2bac docs: validate transaction category review
2203f44 docs: validate transaction contract and filters
ed7074c feat(transactions): add searchable review filters
bbf8f3d docs: complete phase 3 validation workflow
108118b feat(transactions): filter alert source movements
10e44a8 docs: validate authenticated overview
d9a45e1 fix(overview): require known future commitments
77f3e37 fix(overview): never return expired commitments
3d099b7 feat(api): add authenticated financial overview
2ed4a2f fix(agents): generate installment timeline cache
e9035ff fix(backend): expose Pluggy reconnect errors
f33ba61 fix(backend): shorten transaction account labels
6c3ffd5 fix(backend): import completed widget updates
```

O usuário confirmou o Render live para `a5f56be` em 21/08/2026.

## 7. Implementação atual relevante

### Overview/Home

`GET /overview` entrega em uma chamada:

- saldos;
- ritmo semanal;
- status/horário de sincronização;
- quantidade para revisão;
- Alerta principal;
- próximo Compromisso conhecido;
- Movimentações recentes.

Home foi validada com dados reais. Valores mudam; não transforme evidência histórica em constante. Em uma rodada mostrou R$ 472,88; depois R$ 288,69. Isso é esperado porque são dados reais.

### Movimentações

Backend:

- `GET /transactions` preserva compatibilidade e retorna `account_name` legado e `account {id,name}` novo.
- Retorna `category`, `merchant_name`, `is_recurring`, `confidence_score`, `needs_review`.
- Filtros atuais: `q`/`search`, `ids`, `account_id`, `category_id`, `direction`, `needs_review`, `from_date`, `to_date`, `page`, `page_size`.
- `page_size` máximo: 100; `ids` máximo: 50.
- Entradas inválidas de direção, booleano e datas retornam 400.

Mobile:

- `TransactionModel` é estrito para campos essenciais e falha com `FormatException` em incompatibilidade.
- Linha de Movimentação mostra origem PF/PJ.
- Linha é tocável e abre detalhe com descrição, Conta, valor, data, Categoria, revisão e confiança.
- Badge `Revisar` aparece quando `needsReview=true`.
- Dropdown usa `GET /categories`.
- Confirmação usa `PATCH /transactions/:id/category` e invalida Movimentações, relacionadas, resumo e overview.
- Confirmar a mesma Categoria também conclui revisão.

Validação real do FOS-403:

- Movimentação `FARMACIA E DROGARIA NI`, Conta `Nubank PJ`, R$ 52,75.
- Categoria Saúde foi confirmada sem mudar seu significado.
- Após recarregar: revisão `Concluída`, confiança `100%`.
- APK dessa rodada: SHA-256 `d5691deadb82e0874ec30decb86f2eff611b1abed305f22b0c37456a8d9a105c`.

### Categorias duplicadas

Bug reportado pelo usuário: dropdown mostrava `Lazer Lazer`, `Moradia Moradia` etc.

Reprodução automática encontrou duplicatas em Assinaturas, Educação, Emergências, Investimentos, Lazer, Moradia, Outros e Pet. A API retornava Categorias globais/personalizadas de mesmo nome sem precedência.

Correção `7abfbee`:

- `GET /categories` usa uma Categoria por nome case-insensitive;
- prefere Categoria personalizada do Usuário;
- entre equivalentes, prefere a mais recente;
- não apagou registros nem quebrou referências históricas.

Após o deploy, o mesmo detector de acessibilidade retornou vazio no SM-S948B. Bug validado como corrigido.

### Alertas adaptativos

Implementação em `backend/internal/service/feed_service.go`, commit `4ba39b0`:

- Gasto elevado: maior entre R$ 1.000 e três vezes a mediana de débitos dos últimos 90 dias.
- Salário: descrição compatível (`salário`, folha, pró-labore) ou crédito recorrente em pelo menos dois dos últimos três meses, com variação máxima de 20%.
- Duplicidade: débito do mesmo estabelecimento e valor dentro de dois dias.
- Saldo crítico: saldo disponível abaixo de parcelas e cobranças recorrentes previstas até a próxima renda. R$ 500 é fallback apenas sem histórico de próxima renda.
- Não depende de IA.

Testes puros cobrem piso/mediana e descrições salariais. A suíte `go test ./internal/service` passou.

Limite atual: `GenerateEvents` só roda quando uma sincronização salva Transações novas. Nas validações de 21/08, PF e PJ retornaram zero novas, então as consultas adaptativas não foram exercitadas contra dado novo de produção. Não promover FOS-404 para validado ainda.

## 8. Estado real do Pluggy/MeuPluggy

Há um bloqueio externo consistente nas duas Conexões:

```text
Pluggy recusou a atualização: status 400 (MeuPluggy item cant be updated)
```

O app trata isso corretamente como `Sincronização parcial` e exibe o motivo, sem falha silenciosa.

Última validação manual:

- Nubank PF: 1 Conta, 0 novas Transações, 151 disponíveis, período 10/03/2026 a 09/08/2026, atualização exibida 21/08/2026 00:08.
- Nubank PJ: 1 Conta, 0 novas Transações, 98 disponíveis, período 19/05/2026 a 19/08/2026, atualização exibida 20/08/2026 16:33.
- Ambas terminaram parciais pelo mesmo HTTP 400.

A reconexão já abriu corretamente o widget MeuPluggy com `updateItem`; o aceite final em `Continuar` foi deixado para o usuário porque é autorização bancária pessoal. Não clique consentimentos bancários em nome dele.

Apesar do bloqueio de update do Item, dados PJ novos já haviam aparecido antes no extrato, incluindo movimentações de 20/08. Não conclua que toda importação está quebrada; diferencie dados já disponíveis, atualização do Item e novas Transações efetivamente salvas.

## 9. Alterações locais mobile — cuidado máximo

O repositório `mobile/` está deliberadamente sujo. Não há commit mobile consolidando o trabalho recente. Os arquivos alterados incluem navegação, onboarding, autenticação, dashboard, Metas, Mais, Análises, Configurações, Simulador e Movimentações; há também testes novos não rastreados.

Arquivos diretamente relevantes ao trabalho mais recente:

```text
mobile/lib/features/transactions/data/transaction_model.dart
mobile/lib/features/transactions/presentation/transactions_provider.dart
mobile/lib/features/transactions/presentation/transactions_screen.dart
mobile/test/transaction_model_test.dart
mobile/test/overview_data_test.dart
```

Também existe uma árvore Flutter na raiz (`lib/`, `test/`, `android/`) com mudanças. O APK validado foi sempre gerado dentro de `mobile/`. Não sincronize mecanicamente as duas árvores e não sobrescreva uma com a outra sem entender a história.

## 10. Testes e comandos que funcionaram

Backend, evitando o baseline quebrado de utilitários `package main` duplicados:

```bash
cd backend
go test ./cmd/... ./internal/...
go test ./internal/service
go test ./internal/handler
```

Mobile:

```bash
cd mobile
flutter analyze
flutter test
flutter build apk --debug
adb install -r build/app/outputs/flutter-apk/app-debug.apk
adb shell am force-stop com.example.finance_os
adb shell am start -n com.example.finance_os/.MainActivity
```

Última suíte Flutter completa: 10 testes passaram. `flutter analyze` encontrou 23 avisos informativos preexistentes e nenhum erro; os arquivos de Movimentações analisados isoladamente não tinham issues.

ADB:

```bash
adb devices -l
adb shell uiautomator dump /sdcard/financeos.xml
adb shell cat /sdcard/financeos.xml
adb logcat -d -t 1500 | rg -i "FATAL EXCEPTION|E/flutter|Unhandled Exception|AndroidRuntime|com\.example\.finance_os"
```

Os logs das últimas rodadas não tiveram `FATAL EXCEPTION`, `E/flutter`, `Unhandled Exception` ou `AndroidRuntime` do FinanceOS. Avisos `PackageConfigPersister`, Samsung GameManager e SurfaceFlinger são ruído do sistema.

## 11. Navegação e coordenadas observadas no SM-S948B

Tela lógica reportada pelo ADB: 720 × 1560.

- Tabs inferiores aproximadas: Hoje x110, Movimentações x235, Planejar x360, Análises x485, Mais x610; y1475.
- Mais → `Open Finance e configurações`: aproximadamente x350/y790 na posição inicial.
- Em Configurações, depois de rolar, os dois cartões apareceram na ordem PJ e PF na última rodada.
- Nunca use coordenada de `Excluir conexão` sem inspeção atual da árvore de acessibilidade.
- Prefira sempre `uiautomator dump`, confirme `content-desc` e `bounds`, só então toque.

## 12. Próximas tarefas concretas

1. Confirmar que o workspace e os dois Git continuam no estado esperado; não limpar mudanças.
2. Ler integralmente `AGENTS.md`, `CONTEXT.md` e `docs/FINANCEOS-VALIDATION-PLAN.md`.
3. Iniciar a Fase 6 (Simulador Baseado em Dados Reais):
   - FOS-601: Projetar compra com histórico real (entrada: valor, parcelas, descrição e primeiro vencimento; sem constantes).
   - FOS-602: Explicar saída da compra (impacto, fluxo/saldo mensal, limite diário, renda comprometida, alertas e confiança).
   - FOS-603: Simular corte sem investimento fictício (economia mensal, anual e acumulada, sem rentabilidade suposta).
   - FOS-604: Persistir simulações (salvar, nomear, listar, excluir por ID com validação de propriedade).
   - FOS-605: Declarar histórico insuficiente (dados insuficientes não são preenchidos com fallback numérico oculto).
4. Fase 5 (Planejamento Real — FOS-501..506) foi concluída e validada no SM-S948B (APK SHA-256 `5153fb76667f76a32e6e08b0e050d096976a9325dbd8c064b743bdfe8e185ea5`).
5. Quando backend mudar: testar, commit isolado, push `main`, esperar Render live.
6. Quando mobile mudar: testar, build debug, instalar com `adb -r`, preservar sessão, validar no SM-S948B e triar logs.
7. Atualizar o documento oficial somente com evidência. Mock/test unitário não equivale a validação com dados reais.

## 13. Armadilhas conhecidas

- `go test ./...` no diretório backend pode falhar por utilitários `package main` duplicados que já existiam; use os pacotes de produto indicados acima e não “corrija” arquivos alheios sem escopo.
- Render saudável não prova que o commit novo está live; confirme o hash do deploy.
- `adb` ausente depois de dez minutos normalmente é a política de segurança Samsung.
- Sincronização com zero Transações novas não executa os detectores do Feed.
- Não transformar descrições de Movimentação em Compromissos apenas por recorrência heurística.
- Não mostrar Compromisso vencido como próximo.
- Não ocultar erro Pluggy com mensagem genérica.
- Não misturar Conta com Conexão; uma Conexão pode expor várias Contas.
- Não voltar a chamar a tela de Movimentações de “Investimentos” ou oferecer investimento inexistente.
- Não commitar `.env`, logs com dados pessoais, dumps XML do aparelho ou tokens.

## 14. Critério de handoff concluído

O próximo agente deve conseguir responder, antes de editar:

- qual requisito FOS está sendo executado;
- qual dado será real, calculado, inferido ou opcionalmente explicado por IA;
- qual repositório e quais arquivos serão alterados;
- como o backend chegará ao Render;
- como o resultado será testado e validado no Samsung;
- qual evidência permite atualizar o status oficial.

Se alguma dessas respostas estiver incerta, reler o plano oficial e inspecionar o fluxo existente antes de escrever código.
