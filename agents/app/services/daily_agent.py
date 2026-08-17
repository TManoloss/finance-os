import json
from datetime import datetime, timedelta
from app.services.base_agent import BaseAgent

DAILY_PROMPT = """
Você é um assistente financeiro pessoal analisando as transações de hoje do usuário.

Analise os dados fornecidos e retorne um JSON com:
- summary: resumo em português (2-3 frases, tom amigável, primeira pessoa como se fosse um assistente).
- alerts: lista de alertas [{ "type": "warning/info/danger", "message": "...", "amount": 0.0 }]
- total_spent: float
- insights: lista de observações relevantes.

Seja direto e útil. Não use linguagem corporativa.
"""

class DailyAgent(BaseAgent):
    async def run(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # 1. Buscar transações das últimas 24h
            end_date = datetime.now()
            start_date = end_date - timedelta(days=1)
            
            rows = await conn.fetch("""
                SELECT t.description, t.amount, t.direction, c.name as category
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                LEFT JOIN categories c ON t.category_id = c.id
                WHERE acc.user_id = $1 AND t.date >= $2
            """, user_id, start_date.date())

            if not rows:
                return {"message": "Sem transações para o período"}

            transactions = [dict(r) for r in rows]
            
            debits = [t for t in transactions if t["direction"] == "debit"]
            total_spent = sum(float(t["amount"]) for t in debits)
            by_category = {}
            for tx in debits:
                category = tx["category"] or "Outros"
                by_category[category] = by_category.get(category, 0) + float(tx["amount"])
            top_category, top_total = max(by_category.items(), key=lambda item: item[1], default=("", 0))
            insights = ([f"{top_category} representou R$ {top_total:.2f} dos gastos de hoje."] if top_category else [])
            report_data = {
                "summary": f"Hoje foram {len(debits)} gastos, somando R$ {total_spent:.2f}.",
                "alerts": [],
                "total_spent": total_spent,
                "insights": insights,
            }
            await self.save_report(user_id, "daily", start_date.date(), end_date.date(), report_data["summary"], json.dumps(insights))
            return report_data

        finally:
            await conn.close()
