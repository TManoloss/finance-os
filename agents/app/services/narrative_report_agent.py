from app.services.base_agent import BaseAgent
from datetime import datetime, timedelta
import json

MONTHLY_NARRATIVE_PROMPT = """
Você é o assistente financeiro pessoal de {user_name}.
Escreva o relatório financeiro do mês de {month_name} de {year} em português.

Dados completos do mês:
{full_month_data_json}

Diretrizes de escrita:
- Tom: direto, honesto, sem julgamentos morais, como um contador de confiança.
- Estrutura: introdução (1 parágrafo) -> análise de gastos -> destaques positivos -> pontos de atenção -> parcelamentos e compromissos -> perspectiva para o próximo mês.
- Use os números reais. Seja específico.
- Não use bullets ou headers — escreva em parágrafos corridos como uma carta.
- Máximo 400 palavras.
- Termine com 1 ação concreta recomendada para o próximo mês.

Não invente dados. Use apenas o que está nos dados fornecidos.
"""

class NarrativeReportAgent(BaseAgent):
    async def run(self, user_id: str, month: int = None, year: int = None):
        # Default to current month if not provided
        if month is None:
            month = datetime.now().month
        if year is None:
            year = datetime.now().year

        conn = await self.get_db_connection()
        try:
            # 1. Gather all data for the month
            user_name = await conn.fetchval("SELECT name FROM users WHERE id = $1", user_id)
            
            # Transactions Summary
            summary = await conn.fetchrow("""
                SELECT 
                    COALESCE(SUM(CASE WHEN direction = 'credit' THEN amount ELSE 0 END), 0) as income,
                    COALESCE(SUM(CASE WHEN direction = 'debit' THEN amount ELSE 0 END), 0) as spent
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 
                  AND EXTRACT(MONTH FROM date) = $2
                  AND EXTRACT(YEAR FROM date) = $3
            """, user_id, month, year)

            # Top Categories
            categories = await conn.fetch("""
                SELECT c.name, SUM(amount) as total
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                JOIN categories c ON t.category_id = c.id
                WHERE a.user_id = $1 
                  AND EXTRACT(MONTH FROM date) = $2
                  AND EXTRACT(YEAR FROM date) = $3
                GROUP BY c.name
                ORDER BY total DESC LIMIT 5
            """, user_id, month, year)

            # Health Score
            health = await conn.fetchval("""
                SELECT score FROM health_score_snapshots 
                WHERE user_id = $1 AND EXTRACT(MONTH FROM period_month) = $2
                ORDER BY created_at DESC LIMIT 1
            """, user_id, month)

            full_data = {
                "income": float(summary['income'] or 0),
                "spent": float(summary['spent'] or 0),
                "balance": float(summary['income'] or 0) - float(summary['spent'] or 0),
                "top_categories": [{"name": c['name'], "total": float(c['total'])} for c in categories],
                "health_score": float(health or 0)
            }

            # If no transactions and no income, maybe we shouldn't generate a report or at least warn
            if full_data["income"] == 0 and full_data["spent"] == 0:
                return "Ainda não tenho dados suficientes sobre suas transações este mês para escrever um relatório detalhado. Continue usando o app e conecte suas contas para que eu possa analisar seu comportamento financeiro."

            month_names = {
                1: "Janeiro", 2: "Fevereiro", 3: "Março", 4: "Abril",
                5: "Maio", 6: "Junho", 7: "Julho", 8: "Agosto",
                9: "Setembro", 10: "Outubro", 11: "Novembro", 12: "Dezembro"
            }
            month_name = month_names.get(month, "este mês")
            
            prompt = MONTHLY_NARRATIVE_PROMPT.format(
                user_name=user_name,
                month_name=month_name,
                year=year,
                full_month_data_json=json.dumps(full_data, ensure_ascii=False)
            )

            narrative = await self.llm.completion(prompt)

            # Save report
            await self.save_report(
                user_id, 
                "monthly_narrative", 
                datetime(year, month, 1).date(), 
                datetime(year, month, 28).date(), # Simplified end of month
                narrative, 
                json.dumps(full_data)
            )

            return narrative
        finally:
            await conn.close()
            
    async def save_report(self, user_id: str, agent_type: str, period_start, period_end, summary: str, insights: str):
        conn = await self.get_db_connection()
        try:
            await conn.execute("""
                INSERT INTO agent_reports (user_id, agent_type, period_start, period_end, summary_markdown, insights)
                VALUES ($1, $2, $3, $4, $5, $6)
            """, user_id, agent_type, period_start, period_end, summary, insights)
        finally:
            await conn.close()
