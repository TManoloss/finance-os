import json
from datetime import datetime, timedelta
from app.services.base_agent import BaseAgent

MONTHLY_PROMPT = """
Você é um assistente financeiro pessoal realizando o fechamento mensal do usuário.

Analise o desempenho financeiro do mês atual em comparação com a média histórica e projete o próximo mês. 
Retorne um JSON com:
- summary: fechamento detalhado em português (5-6 frases).
- top_merchants: [{ "name": "...", "total": 0.0 }]
- health_score: int (0-100)
- projections: { "next_month_estimated_spent": float }
- insights: lista de observações estratégicas.

Seja analítico, mas mantenha o tom de apoio. Identifique vazamentos de dinheiro (assinaturas esquecidas, etc).
"""

class MonthlyAgent(BaseAgent):
    async def run(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # 1. Datas
            end_date = datetime.now()
            start_date = end_date.replace(day=1) # Primeiro dia do mês atual
            
            # 2. Buscar transações do mês atual
            rows_month = await conn.fetch("""
                SELECT t.description, t.amount, t.direction, c.name as category, t.date
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                LEFT JOIN categories c ON t.category_id = c.id
                WHERE acc.user_id = $1 AND t.date BETWEEN $2 AND $3
            """, user_id, start_date.date(), end_date.date())

            if not rows_month:
                return {"message": "Sem dados suficientes para o fechamento mensal"}

            transactions = [dict(r) for r in rows_month]
            
            # 3. Buscar parcelas futuras
            installments = await conn.fetch("""
                SELECT merchant_name, total_amount / installments_total as part_amount, next_due_date
                FROM installments
                JOIN connected_accounts acc ON installments.account_id = acc.id
                WHERE acc.user_id = $1 AND installment_current < installments_total
            """, user_id)
            
            context_data = {
                "month_transactions": transactions,
                "future_installments": [dict(r) for r in installments]
            }
            
            debits = [t for t in transactions if t["direction"] == "debit"]
            credits = [t for t in transactions if t["direction"] == "credit"]
            spent = sum(float(t["amount"]) for t in debits)
            income = sum(float(t["amount"]) for t in credits)
            by_merchant = {}
            for tx in debits:
                merchant = tx["description"] or "Sem identificação"
                by_merchant[merchant] = by_merchant.get(merchant, 0) + float(tx["amount"])
            top_merchants = [{"name": name, "total": total} for name, total in sorted(by_merchant.items(), key=lambda item: item[1], reverse=True)[:3]]
            days_elapsed = max(end_date.day, 1)
            projected_spent = spent / days_elapsed * 30
            commitment = sum(float(item["part_amount"] or 0) for item in installments)
            insights = [f"Há R$ {commitment:.2f} em parcelas futuras."] if commitment else []
            report_data = {
                "summary": f"No mês, entraram R$ {income:.2f} e saíram R$ {spent:.2f}.",
                "top_merchants": top_merchants,
                "health_score": round(max(0, min(100, ((income - spent) / income * 100) if income else 0))),
                "projections": {"next_month_estimated_spent": projected_spent},
                "insights": insights,
            }
            await self.save_report(user_id, "monthly", start_date.date(), end_date.date(), report_data["summary"], json.dumps(insights))
            return report_data

        finally:
            await conn.close()
