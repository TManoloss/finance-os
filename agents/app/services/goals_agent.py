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
            
            suggestions = []
            for category in top_cats[:2]:
                total = float(category['total'])
                suggestions.append({
                    "name": f"Reduzir gastos com {category['name']}",
                    "goal_type": "spending_limit",
                    "target_amount": round(total * 0.9, 2),
                    "reason": f"Um limite 10% menor que os R$ {total:.2f} dos últimos 30 dias.",
                })
            installments_total = float(inst_sum or 0)
            if installments_total:
                suggestions.append({
                    "name": "Reservar para parcelas", "goal_type": "savings",
                    "target_amount": round(installments_total, 2),
                    "reason": "Cobrir os compromissos mensais já assumidos.",
                })
            return suggestions
        except:
            return []
        finally:
            await conn.close()
