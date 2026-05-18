import json
import logging
from datetime import date, datetime, timedelta
from app.services.base_agent import BaseAgent

logger = logging.getLogger(__name__)

TIMELINE_NARRATIVE_PROMPT = """
Você é um analista financeiro pessoal que conta histórias baseadas em dados.
Sua tarefa é gerar uma narrativa curta (1-2 frases) para um evento financeiro significativo.
O tom deve ser leve, pessoal e direto, como se você conhecesse bem o usuário.

Evento: {event_type}
Dados do Evento: {event_data}

Exemplos:
- "Depois de 11 meses, a parcela do notebook foi quitada. R$340/mês de volta."
- "Maio foi seu melhor mês dos últimos 12: sobrou R$1.240."
- "Você ficou 18 dias sem delivery — seu recorde histórico."
- "Detectamos uma nova assinatura: Netflix (R$55,90). Mais uma para a lista!"

Gere a narrativa em português.
"""

LIFESTYLE_DRIFT_PROMPT = """
Você é um analista financeiro especialista em comportamento.
Analise a mudança no estilo de vida do usuário baseada nos dados abaixo.
Compare a janela recente (últimos 3 meses) com a janela anterior (3 meses antes).

Dados:
{drift_data}

Gere uma narrativa em português (2-3 frases) com números reais, sem julgamento.
Exemplo: "Nos últimos 5 meses seu custo de vida aumentou 18%, mas a renda ficou estável. O aumento veio de alimentação (+R$290/mês) e lazer (+R$180/mês). Impacto anual: R$5.640."
"""

class TimelineDriftAgent(BaseAgent):
    async def run(self, user_id: str):
        # Dispara ambas as análises
        timeline = await self.build_financial_timeline(user_id)
        drift = await self.detect_lifestyle_drift(user_id)
        return {
            "timeline_events_count": len(timeline),
            "lifestyle_drift": drift
        }

    async def build_financial_timeline(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            events = []
            
            # 1. Detect Installment Ends (parcelas quitadas recentemente)
            installment_ends = await conn.fetch("""
                SELECT merchant_name, total_amount, installments_total, start_date, id
                FROM installments i
                JOIN connected_accounts acc ON i.account_id = acc.id
                WHERE acc.user_id = $1 
                AND i.installment_current = i.installments_total
                AND i.next_due_date > NOW() - INTERVAL '30 days'
            """, user_id)
            
            for inst in installment_ends:
                events.append({
                    "event_type": "installment_end",
                    "event_date": date.today(),
                    "title": f"Parcela quitada: {inst['merchant_name']}",
                    "event_data": {
                        "merchant": inst['merchant_name'],
                        "monthly_value": float(inst['total_amount'] / inst['installments_total']),
                        "total_amount": float(inst['total_amount'])
                    }
                })

            # 2. Detect New Installments
            new_installments = await conn.fetch("""
                SELECT merchant_name, total_amount, installments_total, start_date
                FROM installments i
                JOIN connected_accounts acc ON i.account_id = acc.id
                WHERE acc.user_id = $1 
                AND i.installment_current = 1
                AND i.start_date > NOW() - INTERVAL '30 days'
            """, user_id)

            for inst in new_installments:
                events.append({
                    "event_type": "installment_start",
                    "event_date": inst['start_date'],
                    "title": f"Novo parcelamento: {inst['merchant_name']}",
                    "event_data": {
                        "merchant": inst['merchant_name'],
                        "monthly_value": float(inst['total_amount'] / inst['installments_total']),
                        "total_amount": float(inst['total_amount']),
                        "installments": inst['installments_total']
                    }
                })

            # 3. Debt Free (se antes tinha parcelas e agora não tem nenhuma ativa)
            active_installments = await conn.fetchval("""
                SELECT COUNT(*)
                FROM installments i
                JOIN connected_accounts acc ON i.account_id = acc.id
                WHERE acc.user_id = $1 
                AND i.installment_current < i.installments_total
            """, user_id)
            
            if active_installments == 0 and len(installment_ends) > 0:
                events.append({
                    "event_type": "debt_free",
                    "event_date": date.today(),
                    "title": "Dívidas quitadas!",
                    "event_data": {"message": "Todas as parcelas foram pagas."}
                })

            # 4. Spending Peaks
            peaks = await conn.fetch("""
                WITH daily_spending AS (
                    SELECT date, SUM(amount) as total
                    FROM transactions t
                    JOIN connected_accounts acc ON t.account_id = acc.id
                    WHERE acc.user_id = $1 AND t.direction = 'debit'
                    AND t.date > NOW() - INTERVAL '30 days'
                    GROUP BY date
                ),
                stats AS (
                    SELECT AVG(total) as avg_daily
                    FROM daily_spending
                )
                SELECT date, total, avg_daily
                FROM daily_spending, stats
                WHERE total > avg_daily * 3
                ORDER BY date DESC
                LIMIT 1
            """, user_id)
            
            for p in peaks:
                events.append({
                    "event_type": "spending_peak",
                    "event_date": p['date'],
                    "title": "Pico de gastos detectado",
                    "event_data": {
                        "total": float(p['total']),
                        "avg_daily": float(p['avg_daily']),
                        "ratio": float(p['total'] / p['avg_daily'])
                    }
                })

            # 5. Salary Change
            salaries = await conn.fetch("""
                SELECT amount, date, merchant_name, description
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                WHERE acc.user_id = $1 
                AND t.direction = 'credit'
                AND t.amount > 1000
                AND t.date > NOW() - INTERVAL '180 days'
                ORDER BY date DESC
            """, user_id)
            
            if len(salaries) >= 2:
                latest = salaries[0]
                previous = salaries[1]
                if abs(float(latest['amount']) - float(previous['amount'])) / float(previous['amount']) > 0.05:
                    events.append({
                        "event_type": "salary_change",
                        "event_date": latest['date'],
                        "title": "Mudança na renda detectada",
                        "event_data": {
                            "new_amount": float(latest['amount']),
                            "old_amount": float(previous['amount']),
                            "diff": float(latest['amount'] - previous['amount']),
                            "diff_percent": float((latest['amount'] - previous['amount']) / previous['amount'] * 100)
                        }
                    })

            # 6. Subscription Changes
            sub_events = await self.detect_subscription_changes(conn, user_id)
            events.extend(sub_events)

            # Process narratives and save
            for event in events:
                exists = await conn.fetchval("""
                    SELECT 1 FROM financial_timeline_events 
                    WHERE user_id = $1 AND event_type = $2 AND title = $3
                """, user_id, event['event_type'], event['title'])
                
                if not exists:
                    prompt = TIMELINE_NARRATIVE_PROMPT.format(
                        event_type=event['event_type'],
                        event_data=json.dumps(event['event_data'], ensure_ascii=False)
                    )
                    narrative = await self.llm.completion(prompt)
                    event['narrative'] = narrative.strip()
                    
                    await conn.execute("""
                        INSERT INTO financial_timeline_events (user_id, event_type, event_date, title, narrative, event_data)
                        VALUES ($1, $2, $3, $4, $5, $6)
                    """, user_id, event['event_type'], event['event_date'], event['title'], event['narrative'], json.dumps(event['event_data']))

            return events
        finally:
            await conn.close()

    async def detect_subscription_changes(self, conn, user_id: str):
        # Encontrar assinaturas (mercantes recorrentes mensais)
        rows = await conn.fetch("""
            SELECT merchant_name, amount, date
            FROM transactions t
            JOIN connected_accounts acc ON t.account_id = acc.id
            WHERE acc.user_id = $1 AND t.direction = 'debit'
            AND t.date > NOW() - INTERVAL '180 days'
            AND merchant_name IS NOT NULL
        """, user_id)
        
        from collections import defaultdict
        merchant_dates = defaultdict(list)
        for r in rows:
            merchant_dates[r['merchant_name']].append(r['date'])
            
        subscriptions = []
        for merchant, dates in merchant_dates.items():
            if len(dates) >= 2:
                dates.sort()
                # Verificar se há intervalos de ~30 dias
                intervals = [(dates[i] - dates[i-1]).days for i in range(1, len(dates))]
                monthly_intervals = [iv for iv in intervals if 25 <= iv <= 35]
                if len(monthly_intervals) >= 1: # Pelo menos um intervalo mensal
                    subscriptions.append({
                        "merchant": merchant,
                        "last_date": dates[-1],
                        "first_date": dates[0],
                        "count": len(dates)
                    })
        
        events = []
        today = date.today()
        for sub in subscriptions:
            # New Subscription: first appearance in last 45 days
            if sub['first_date'] > today - timedelta(days=45) and sub['count'] >= 2:
                events.append({
                    "event_type": "new_subscription",
                    "event_date": sub['first_date'],
                    "title": f"Nova assinatura: {sub['merchant']}",
                    "event_data": {"merchant": sub['merchant']}
                })
            
            # Cancelled Subscription: last appearance between 45 and 75 days ago
            if today - timedelta(days=75) < sub['last_date'] < today - timedelta(days=45):
                 events.append({
                    "event_type": "cancel_subscription",
                    "event_date": today,
                    "title": f"Assinatura cancelada: {sub['merchant']}",
                    "event_data": {"merchant": sub['merchant']}
                })
        return events

    async def detect_lifestyle_drift(self, user_id: str, window_months: int = 6):
        conn = await self.get_db_connection()
        try:
            today = date.today()
            recent_start = (today.replace(day=1) - timedelta(days=90)).replace(day=1)
            recent_end = today.replace(day=1)
            previous_start = (recent_start - timedelta(days=90)).replace(day=1)
            previous_end = recent_start
            
            recent_spending = await conn.fetchval("""
                SELECT SUM(amount)
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                WHERE acc.user_id = $1 AND t.direction = 'debit'
                AND t.date >= $2 AND t.date < $3
            """, user_id, recent_start, recent_end) or 0
            
            previous_spending = await conn.fetchval("""
                SELECT SUM(amount)
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                WHERE acc.user_id = $1 AND t.direction = 'debit'
                AND t.date >= $2 AND t.date < $3
            """, user_id, previous_start, previous_end) or 0
            
            recent_avg = float(recent_spending) / 3
            previous_avg = float(previous_spending) / 3
            
            if previous_avg == 0:
                return {"status": "insufficient_data"}
                
            variation = (recent_avg - previous_avg) / previous_avg
            
            if abs(variation) > 0.15:
                category_drift = await conn.fetch("""
                    WITH recent AS (
                        SELECT category_id, SUM(amount) as total
                        FROM transactions t
                        JOIN connected_accounts acc ON t.account_id = acc.id
                        WHERE acc.user_id = $1 AND t.direction = 'debit'
                        AND t.date >= $2 AND t.date < $3
                        GROUP BY category_id
                    ),
                    previous AS (
                        SELECT category_id, SUM(amount) as total
                        FROM transactions t
                        JOIN connected_accounts acc ON t.account_id = acc.id
                        WHERE acc.user_id = $1 AND t.direction = 'debit'
                        AND t.date >= $4 AND t.date < $5
                        GROUP BY category_id
                    )
                    SELECT c.name, 
                           COALESCE(r.total, 0) / 3 as recent_avg, 
                           COALESCE(p.total, 0) / 3 as previous_avg
                    FROM categories c
                    LEFT JOIN recent r ON c.id = r.category_id
                    LEFT JOIN previous p ON c.id = p.category_id
                    WHERE r.total IS NOT NULL OR p.total IS NOT NULL
                    ORDER BY ABS(COALESCE(r.total, 0) - COALESCE(p.total, 0)) DESC
                    LIMIT 3
                """, user_id, recent_start, recent_end, previous_start, previous_end)
                
                categories_data = []
                for cat in category_drift:
                    categories_data.append({
                        "name": cat['name'],
                        "diff": float(cat['recent_avg'] - cat['previous_avg'])
                    })
                
                drift_data = {
                    "variation_percent": variation * 100,
                    "recent_avg": recent_avg,
                    "previous_avg": previous_avg,
                    "top_categories": categories_data,
                    "annual_impact": (recent_avg - previous_avg) * 12
                }
                
                prompt = LIFESTYLE_DRIFT_PROMPT.format(drift_data=json.dumps(drift_data, ensure_ascii=False))
                narrative = await self.llm.completion(prompt)
                
                report = {
                    "type": "UPGRADE" if variation > 0 else "DOWNGRADE",
                    "variation_percent": variation * 100,
                    "narrative": narrative.strip(),
                    "data": drift_data
                }
                
                # Gerar evento na timeline
                await conn.execute("""
                    INSERT INTO financial_timeline_events (user_id, event_type, event_date, title, narrative, event_data)
                    VALUES ($1, $2, $3, $4, $5, $6)
                    ON CONFLICT DO NOTHING
                """, user_id, "lifestyle_drift", date.today(), 
                   f"Mudança no padrão de vida: {report['type']}", 
                   report['narrative'], json.dumps(report['data']))
                
                return report
                
            return {"status": "stable", "variation_percent": variation * 100}
        finally:
            await conn.close()
