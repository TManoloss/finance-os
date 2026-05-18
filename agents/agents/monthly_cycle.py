import json
from datetime import date, datetime, timedelta
from app.services.base_agent import BaseAgent
import logging

logger = logging.getLogger(__name__)

SALARY_EFFECT_PROMPT = """
Você é um analista financeiro especialista em comportamento do consumidor.
Analise os dados abaixo sobre o "Efeito Salário" (como o usuário gasta o dinheiro logo após recebê-lo).

Dados:
{salary_data_json}

Identifique se existe um pico de gastos logo após o salário, quais categorias são mais afetadas e se há uma queda drástica de padrão no final do ciclo.
Retorne obrigatoriamente um JSON com:
- salary_date_estimated: int (dia do mês aproximado)
- average_salary_amount: float
- consumption_speed: str (rápida, moderada, lenta)
- insights: str (análise narrativa em português, tom amigável e direto)
"""

MONTHLY_WEEKS_PROMPT = """
Você é um analista financeiro. Analise a distribuição de gastos entre as 4 semanas do mês.

Dados:
{weeks_data_json}

Compare a primeira semana com a última. Existe o fenômeno de "viver de luz" na última semana? Ou o usuário mantém consistência?
Retorne obrigatoriamente um JSON com:
- week_1_vs_week_4_ratio: float (gasto semana 1 / gasto semana 4)
- most_expensive_week: int (1, 2, 3 ou 4)
- consistency_score: int (0 a 100)
- insights: str (análise narrativa em português)
"""

class MonthlyCycleAgent(BaseAgent):
    async def run(self, user_id: str):
        # Este agente executa as duas análises e salva no cache
        salary_result = await self.run_salary_effect(user_id)
        weeks_result = await self.run_monthly_weeks(user_id)
        
        return {
            "salary_effect": salary_result,
            "monthly_weeks": weeks_result
        }

    async def detect_salary_events(self, conn, user_id: str):
        # Identificar créditos recorrentes de valor alto
        rows = await conn.fetch("""
            SELECT 
                merchant_name, 
                description, 
                AVG(amount) as avg_amount,
                COUNT(*) as occurrences,
                EXTRACT(DAY FROM date) as avg_day
            FROM transactions t
            JOIN connected_accounts acc ON t.account_id = acc.id
            WHERE acc.user_id = $1 
            AND t.direction = 'credit'
            AND t.amount > 1000
            AND t.date > NOW() - INTERVAL '6 months'
            GROUP BY merchant_name, description
            HAVING COUNT(*) >= 2
            ORDER BY avg_amount DESC
            LIMIT 1
        """, user_id)
        
        if not rows:
            return None
        
        return {
            "merchant_name": rows[0]['merchant_name'],
            "description": rows[0]['description'],
            "avg_amount": float(rows[0]['avg_amount']),
            "avg_day": int(rows[0]['avg_day'])
        }

    async def run_salary_effect(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            salary_info = await self.detect_salary_events(conn, user_id)
            if not salary_info:
                return {"error": "Padrão de salário não detectado nos últimos 6 meses."}

            # Buscar todas as datas de recebimento de salário
            salary_dates = await conn.fetch("""
                SELECT date, amount
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                WHERE acc.user_id = $1 
                AND t.direction = 'credit'
                AND (t.merchant_name = $2 OR t.description = $3)
                AND t.amount BETWEEN $4 * 0.8 AND $4 * 1.2
                ORDER BY date DESC
            """, user_id, salary_info['merchant_name'], salary_info['description'], salary_info['avg_amount'])

            post_salary_spending = []
            for s_date_row in salary_dates[:3]: # Analisar os últimos 3 ciclos
                s_date = s_date_row['date']
                end_date = s_date + timedelta(days=30)
                
                daily_spending = await conn.fetch("""
                    SELECT date, SUM(amount) as total
                    FROM transactions t
                    JOIN connected_accounts acc ON t.account_id = acc.id
                    WHERE acc.user_id = $1
                    AND t.direction = 'debit'
                    AND t.date >= $2 AND t.date < $3
                    GROUP BY date
                    ORDER BY date ASC
                """, user_id, s_date, end_date)
                
                cycle_data = {
                    "salary_date": s_date.isoformat(),
                    "salary_amount": float(s_date_row['amount']),
                    "days": [{"day": (r['date'] - s_date).days, "amount": float(r['total'])} for r in daily_spending]
                }
                post_salary_spending.append(cycle_data)

            # Preparar contexto para LLM
            context = {
                "salary_info": salary_info,
                "cycles": post_salary_spending
            }

            prompt = f"Dados do efeito salário para o usuário {user_id}:\n{json.dumps(context, default=str)}"
            response_text = await self.llm.completion(prompt, system_prompt=SALARY_EFFECT_PROMPT)
            
            try:
                start_idx = response_text.find("{")
                end_idx = response_text.rfind("}")
                report_data = json.loads(response_text[start_idx:end_idx+1])
            except:
                report_data = {"insights": response_text, "error": "Failed to parse JSON"}

            # Cachear resultado
            expires_at = datetime.now() + timedelta(days=7)
            await conn.execute("""
                INSERT INTO report_cache (user_id, report_type, period_key, result_json, expires_at)
                VALUES ($1, $2, $3, $4, $5)
                ON CONFLICT (user_id, report_type, period_key) DO UPDATE
                SET result_json = EXCLUDED.result_json,
                    expires_at = EXCLUDED.expires_at,
                    computed_at = NOW()
            """, user_id, "salary_effect", date.today().strftime("%Y-%m"), json.dumps(report_data), expires_at)

            return report_data
        finally:
            await conn.close()

    async def run_monthly_weeks(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # Analisar os últimos 3 meses fechados
            today = date.today()
            months_data = []
            
            for i in range(1, 4):
                target_month = (today.replace(day=1) - timedelta(days=1)).replace(day=1)
                if i > 1:
                    for _ in range(i-1):
                        target_month = (target_month - timedelta(days=1)).replace(day=1)
                
                month_start = target_month
                month_end = (month_start + timedelta(days=32)).replace(day=1)
                
                weeks = [
                    (month_start, month_start + timedelta(days=7)),
                    (month_start + timedelta(days=7), month_start + timedelta(days=14)),
                    (month_start + timedelta(days=14), month_start + timedelta(days=21)),
                    (month_start + timedelta(days=21), month_end)
                ]
                
                weeks_summary = []
                for idx, (w_start, w_end) in enumerate(weeks):
                    total = await conn.fetchval("""
                        SELECT SUM(amount)
                        FROM transactions t
                        JOIN connected_accounts acc ON t.account_id = acc.id
                        WHERE acc.user_id = $1
                        AND t.direction = 'debit'
                        AND t.date >= $2 AND t.date < $3
                    """, user_id, w_start, w_end)
                    weeks_summary.append({
                        "week": idx + 1,
                        "total": float(total or 0)
                    })
                
                months_data.append({
                    "month": month_start.strftime("%Y-%m"),
                    "weeks": weeks_summary
                })

            # LLM Insights
            prompt = f"Dados das semanas mensais para o usuário {user_id}:\n{json.dumps(months_data, default=str)}"
            response_text = await self.llm.completion(prompt, system_prompt=MONTHLY_WEEKS_PROMPT)
            
            try:
                start_idx = response_text.find("{")
                end_idx = response_text.rfind("}")
                report_data = json.loads(response_text[start_idx:end_idx+1])
            except:
                report_data = {"insights": response_text, "error": "Failed to parse JSON"}

            # Cachear
            expires_at = datetime.now() + timedelta(days=7)
            await conn.execute("""
                INSERT INTO report_cache (user_id, report_type, period_key, result_json, expires_at)
                VALUES ($1, $2, $3, $4, $5)
                ON CONFLICT (user_id, report_type, period_key) DO UPDATE
                SET result_json = EXCLUDED.result_json,
                    expires_at = EXCLUDED.expires_at,
                    computed_at = NOW()
            """, user_id, "monthly_weeks", date.today().strftime("%Y-%m"), json.dumps(report_data), expires_at)

            return report_data
        finally:
            await conn.close()
