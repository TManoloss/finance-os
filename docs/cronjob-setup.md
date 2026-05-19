# Configuração do cron-job.org

## Por que é necessário
O Render free tier dorme após 15 minutos de inatividade.
O cron-job.org faz requisições externas que mantêm o serviço acordado
e disparam os syncs nos horários configurados.

## Passo a passo

1. Acesse https://cron-job.org e crie uma conta gratuita

2. Crie 4 jobs de SYNC (um para cada horário):

   Job 1 — Sync manhã:
   - URL: https://finance-os-api.onrender.com/api/v1/internal/sync
   - Método: POST
   - Header: X-Sync-Secret: {seu SYNC_SECRET do .env}
   - Horário: 07:00 (America/Sao_Paulo)
   - Dias: todos

   Job 2 — Sync almoço:
   - URL: https://finance-os-api.onrender.com/api/v1/internal/sync
   - Método: POST
   - Header: X-Sync-Secret: {seu SYNC_SECRET do .env}
   - Horário: 13:00 (America/Sao_Paulo)
   - Dias: todos

   Job 3 — Sync fim de tarde:
   - URL: https://finance-os-api.onrender.com/api/v1/internal/sync
   - Método: POST
   - Header: X-Sync-Secret: {seu SYNC_SECRET do .env}
   - Horário: 19:00 (America/Sao_Paulo)
   - Dias: todos

   Job 4 — Sync noturno:
   - URL: https://finance-os-api.onrender.com/api/v1/internal/sync
   - Método: POST
   - Header: X-Sync-Secret: {seu SYNC_SECRET do .env}
   - Horário: 23:30 (America/Sao_Paulo)
   - Dias: todos

3. Crie 1 job de KEEP-ALIVE (para acordar o serviço antes de cada sync):

   Job Keep-alive — pré-sync:
   - URL: https://finance-os-api.onrender.com/health
   - Método: GET
   - Horário: a cada 14 minutos (*/14 * * * *)
   - Importante: esse job acorda o serviço; sem ele o Render pode estar
     dormindo quando o sync chegar

4. Configure notificações de falha:
   - Em Settings > Notifications: adicionar seu email
   - Notificar quando job falhar 2 vezes consecutivas

## Consumo de updates Pluggy

4 syncs/dia × 30 dias = 120 updates/mês
Limite da Pluggy (dev tier): 240 updates/mês por CPF/instituição
Margem de segurança: 50% — suficiente para syncs manuais ocasionais

## Verificar se está funcionando

Após configurar, verificar em:
GET https://finance-os-api.onrender.com/api/v1/internal/sync/status
Header: X-Sync-Secret: {seu SYNC_SECRET}

O campo last_sync_at deve atualizar após cada job do cron-job.org.
