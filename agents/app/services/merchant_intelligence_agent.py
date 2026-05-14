from app.services.base_agent import BaseAgent
from datetime import datetime, timedelta
import json

class MerchantIntelligenceAgent(BaseAgent):
	async def run(self, user_id: str, period: str = ""):
		# Default run for compatibility with BaseAgent
		return await self.get_top_merchants(user_id)

	async def get_top_merchants(self, user_id: str, period_months: int = 3):
		conn = await self.get_db_connection()
		try:
			rows = await conn.fetch("""
				SELECT 
					merchant_name,
					COUNT(*) as count,
					SUM(amount) as total,
					AVG(amount) as avg_amount,
					MIN(date) as first_purchase,
					MAX(date) as last_purchase
				FROM transactions t
				JOIN connected_accounts a ON t.account_id = a.id
				WHERE a.user_id = $1 AND direction = 'debit'
				  AND merchant_name IS NOT NULL
				GROUP BY merchant_name
				ORDER BY total DESC
				LIMIT 20
			""", user_id)
			
			return [{
				"merchant": r['merchant_name'],
				"count": r['count'],
				"total": float(r['total']),
				"avg_amount": float(r['avg_amount']),
				"first_purchase": r['first_purchase'].isoformat(),
				"last_purchase": r['last_purchase'].isoformat(),
			} for r in rows]
		finally:
			await conn.close()

	async def get_merchant_profile(self, user_id: str, merchant_name: str):
		conn = await self.get_db_connection()
		try:
			# Detalhes específicos de um merchant
			history = await conn.fetch("""
				SELECT date, amount
				FROM transactions t
				JOIN connected_accounts a ON t.account_id = a.id
				WHERE a.user_id = $1 AND merchant_name = $2
				ORDER BY date ASC
			""", user_id, merchant_name)
			
			if not history:
				return {"error": "Merchant not found"}

			total = sum(float(h['amount']) for h in history)
			count = len(history)
			
			# Insights da IA sobre este merchant
			prompt = f"""
			Analise o histórico de gastos do usuário no estabelecimento '{merchant_name}':
			Total gasto: R$ {total:.2f}
			Frequência: {count} vezes
			Ticket médio: R$ {(total/count):.2f}
			
			Gere um insight curto em português sobre este relacionamento financeiro.
			"""
			insight = await self.llm.completion(prompt)

			return {
				"merchant": merchant_name,
				"total_gasto": total,
				"frequencia": count,
				"ticket_medio": total/count,
				"primeira_compra": history[0]['date'].isoformat(),
				"ultima_compra": history[-1]['date'].isoformat(),
				"insight": insight,
				"history": [{"date": h['date'].isoformat(), "amount": float(h['amount'])} for h in history]
			}
		finally:
			await conn.close()
