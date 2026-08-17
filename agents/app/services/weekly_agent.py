import json
from datetime import datetime, timedelta
from app.services.base_agent import BaseAgent

WEEKLY_PROMPT = """
Você é um assistente financeiro pessoal analisando o resumo semanal do usuário.

Analise os dados da semana atual vs semana anterior e retorne um JSON com:
- summary: resumo amigável em português (3-4 frases).
- alerts: [{ "type": "warning/info/danger", "message": "..." }]
- top_categories: [{ "name": "...", "total": 0.0 }]
- vs_previous_week_percent: float
- insights: lista de observações relevantes.

Seja direto e útil. Use uma linguagem que motive o usuário a economizar.
"""

class WeeklyAgent(BaseAgent):
    async def run(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # 1. Datas
            end_date = datetime.now()
            start_date = end_date - timedelta(days=7)
            prev_start = start_date - timedelta(days=7)
            
            # 2. Buscar transações da semana atual
            rows_current = await conn.fetch("""
                SELECT t.description, t.amount, t.direction, c.name as category, t.date
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                LEFT JOIN categories c ON t.category_id = c.id
                WHERE acc.user_id = $1 AND t.date BETWEEN $2 AND $3
            """, user_id, start_date.date(), end_date.date())

            # 3. Buscar totais da semana anterior (contexto)
            prev_total = await conn.fetchval("""
                SELECT SUM(amount)
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                WHERE acc.user_id = $1 AND t.direction = 'debit' AND t.date BETWEEN $2 AND $3
            """, user_id, prev_start.date(), (start_date - timedelta(days=1)).date())

            if not rows_current:
                return {"message": "Sem transações suficientes para gerar relatório semanal"}

            transactions = [dict(r) for r in rows_current]
            context_data = {
                "current_week_transactions": transactions,
                "previous_week_total_debit": float(prev_total or 0)
            }
            
            debits = [t for t in transactions if t["direction"] == "debit"]
            current_total = sum(float(t["amount"]) for t in debits)
            by_category = {}
            for tx in debits:
                category = tx["category"] or "Outros"
                by_category[category] = by_category.get(category, 0) + float(tx["amount"])
            top_categories = [{"name": name, "total": total} for name, total in sorted(by_category.items(), key=lambda item: item[1], reverse=True)[:3]]
            previous_total = float(prev_total or 0)
            change = ((current_total - previous_total) / previous_total * 100) if previous_total else 0
            direction = "aumentaram" if change > 0 else "diminuíram" if change < 0 else "ficaram estáveis"
            insights = [f"Seus gastos {direction} {abs(change):.1f}% em relação à semana anterior."] if previous_total else []
            report_data = {
                "summary": f"Você gastou R$ {current_total:.2f} nesta semana.",
                "alerts": [],
                "top_categories": top_categories,
                "vs_previous_week_percent": change,
                "insights": insights,
            }
            await self.save_report(user_id, "weekly", start_date.date(), end_date.date(), report_data["summary"], json.dumps(insights))
            return report_data

        finally:
            await conn.close()
