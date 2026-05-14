from app.services.base_agent import BaseAgent
from datetime import datetime, timedelta
import json

class GoalsAgent(BaseAgent):
    async def run(self, user_id: str, period: str = ""):
        return await self.suggest_goals(user_id)

    async def suggest_goals(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # 1. Buscar top categorias de gastos
            top_cats = await conn.fetch("""
                SELECT c.name, SUM(amount) as total
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                JOIN categories c ON t.category_id = c.id
                WHERE a.user_id = $1 AND direction = 'debit' AND date > NOW() - INTERVAL '30 days'
                GROUP BY c.name ORDER BY total DESC LIMIT 3
            """, user_id)

            # 2. Verificar se tem parcelamentos altos
            inst_sum = await conn.fetchval("""
                SELECT SUM(total_amount / installments_total) FROM installments i
                JOIN connected_accounts a ON i.account_id = a.id
                WHERE a.user_id = $1
            """, user_id)
            
            suggestions_data = {
                "top_categories": [{"name": c['name'], "total": float(c['total'])} for c in top_cats],
                "total_installments": float(inst_sum or 0)
            }

            prompt = f"""
            Com base nos dados financeiros do usuário:
            {json.dumps(suggestions_data)}
            
            Sugira 3 metas financeiras inteligentes e realistas para ele começar agora.
            Gere a resposta em português como uma lista de objetos JSON com 'name', 'goal_type', 'target_amount' e 'reason'.
            Tipos permitidos: 'savings', 'debt_payoff', 'spending_limit', 'income_target'.
            """
            response = await self.llm.completion(prompt)
            return json.loads(response)
        except:
            return []
        finally:
            await conn.close()
