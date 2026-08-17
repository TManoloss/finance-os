# FinanceOS Mobile — Handover

Atualizado em 2026-08-17.

## Repositórios

- Mobile independente: `git@github.com:TManoloss/MobileFinanceOS.git` (`main`, commit publicado `826ca04`).
- Monorepo original: `Dinheiro`; backend Go é publicado pelo remoto `TManoloss/finance-os`, com commits `9f79103` e `9fb50c4` já enviados para `main`.
- `mobile/` agora possui seu próprio `.git`. Não use o Git do monorepo para publicar alterações futuras do app mobile.

## Produto e direção visual

- App Flutter dark premium inspirado nas duas imagens de referência fornecidas pelo usuário.
- Paleta e tema: `lib/core/theme/blueprint_theme.dart`.
- Navegação flutuante: `lib/core/layout/main_layout.dart`.
  - Abas: Home, Portfolio, Goal, History e Settings/Mais.
  - Goal e History não exibem o marcador/label laranja; foi uma solicitação explícita.
- Fundação compartilhada: `lib/shared/widgets/premium_page.dart` (`PremiumPage`, `PremiumCard`, `PremiumTitle`). Reutilize estes widgets ao portar telas restantes.

## Fluxo manual, sem Pluggy

- O usuário não quer Open Finance/Pluggy.
- `lib/core/router/app_router.dart` não bloqueia mais login em setup Pluggy.
- A rota e o arquivo `pluggy_setup_screen.dart` foram removidos do mobile.
- API do mobile aponta sempre para `https://finance-os-d3nm.onrender.com/api/v1` em `lib/core/api/api_client.dart`.
- Backend local recebeu:
  - `POST /accounts`: cria conta corrente, poupança ou cartão manual.
  - `POST /transactions`: cria entrada/saída manual e atualiza o saldo da conta.
- Para uma fatura manual: criar cartão em Patrimônio e registrar despesas como saída selecionando esse cartão. Não existe ainda campo explícito de meio de pagamento (PIX/cartão); ele é inferido pela conta escolhida.

## Telas já alteradas

- Home: `features/dashboard/presentation/dashboard_screen.dart`.
- Metas: `features/goals/presentation/goals_screen.dart`.
- Mais: `features/more/presentation/more_screen.dart`.
- Patrimônio/Cartões: `features/cards/presentation/cards_screen.dart`.
- Histórico/Extrato + inserção manual: `features/transactions/presentation/transactions_screen.dart`.
- Pierre: `features/chat/presentation/chat_screen.dart`.
- Saúde financeira: `features/health/presentation/health_screen.dart`.
- Estabelecimentos: `features/merchants/presentation/merchants_screen.dart`.
- Relatórios: `features/reports/presentation/reports_screen.dart` (portabilidade parcial).
- Login e cadastro: `features/auth/presentation/login_screen.dart`, `register_screen.dart`.
- Configurações: `features/settings/presentation/settings_screen.dart` foi simplificada para modo manual.

## Pendências importantes

1. Rodar `flutter analyze` antes de qualquer novo APK. As últimas alterações de Relatórios, Configurações, Login/Cadastro e remoção Pluggy aconteceram após o último APK e após o último push do repositório mobile.
2. Portar por completo Relatórios, Replay, Simulador e componentes em `features/dashboard/presentation/widgets/`; ainda há elementos neo-brutalistas/monospace antigos.
3. Revisar Configurações: preservar edição de dia de fechamento/vencimento do cartão e criar uma tela real para credenciais Groq/Gemini. A tela atual é visualmente nova, porém deliberadamente reduzida.
4. Adicionar seletor explícito de meio de pagamento para saídas, se confirmado pelo usuário; hoje a conta define PIX/dinheiro/cartão.
5. Confirmar no Render que os commits do backend já concluíram deploy antes de testar `POST /accounts` e `POST /transactions` pelo APK.
6. O mobile independente foi publicado antes das últimas alterações de portabilidade; depois de validar, fazer novo commit e `git push` dentro de `mobile/`.

## Validação e APK

- `flutter test` passou antes das últimas mudanças de portabilidade.
- `flutter analyze` foi executado durante a portabilidade; corrija e rode novamente após as mudanças finais.
- APK debug existente: `build/app/outputs/flutter-apk/app-debug.apk`.
- Limite obrigatório para build: Gradle já contém `org.gradle.jvmargs=-Xmx4G ...` em `android/gradle.properties`.
- Para gerar: `flutter build apk --debug` dentro de `mobile/`. Não aumentar o limite de 4 GB.

## Primeiro roteiro da próxima sessão

```bash
cd /home/manoelfelip/Documentos/projetos/Dinheiro/mobile
git status
flutter analyze
flutter test
```

Corrija erros antes de gerar APK. Depois:

```bash
git add -A
git commit -m "feat: finish premium mobile migration"
git push
flutter build apk --debug
```
