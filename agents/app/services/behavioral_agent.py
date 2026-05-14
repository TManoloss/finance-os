from app.services.base_agent import BaseAgent
from datetime import datetime, timedelta
import json
import numpy as np

BEHAVIORAL_INSIGHTS_PROMPT = """
Você é um analista financeiro pessoal com acesso ao histórico completo de transações do usuário.

Analise os seguintes padrões detectados algoritmicamente:
{patterns_json}

Gere insights em português, tom direto mas empático, primeira pessoa como se fosse um
conselheiro financeiro de confiança. Não use linguagem corporativa. Seja específico com
números reais. Máximo 3 insights por análise, ordenados por impacto financeiro.

Formato de resposta JSON:
{{
  "insights": [
    {{
      "title": "título curto do insight",
      "description": "explicação em 2-3 frases com números específicos",
      "impact_monthly": float,
      "impact_annual": float,
      "type": "pattern|warning|opportunity",
      "severity": "info|medium|high"
    }}
  ],
  "summary": "resumo geral em 1 frase"
}}
"""

class BehavioralAgent(BaseAgent):
	async def run(self, user_id: str, period: str = ""):
		emotional = await self.analyze_emotional_spending(user_id)
		lifestyle = await self.detect_lifestyle_inflation(user_id)
		
		patterns = {
			"emotional_spending": emotional,
			"lifestyle_inflation": lifestyle
		}
		
		prompt = BEHAVIORAL_INSIGHTS_PROMPT.format(patterns_json=json.dumps(patterns))
		response = await self.llm.completion(prompt)
		
		try:
			return json.loads(response)
		except:
			return {"error": "Falha ao processar insights comportamentais", "raw": response}

	async def analyze_emotional_spending(self, user_id: str):
		conn = await self.get_db_connection()
		try:
			# Gastos por dia da semana e faixa horária
			rows = await conn.fetch("""
				SELECT 
					TO_CHAR(date, 'Day') as weekday,
					EXTRACT(HOUR FROM created_at) as hour,
					amount
				FROM transactions t
				JOIN connected_accounts a ON t.account_id = a.id
				WHERE a.user_id = $1 AND direction = 'debit'
				  AND date > NOW() - INTERVAL '90 days'
			""", user_id)
			
			if not rows:
				return {}

			# Agrupar por dia
			by_day = {}
			for r in rows:
				day = r['weekday'].strip()
				by_day[day] = by_day.get(day, 0) + float(r['amount'])
			
			avg_day = sum(by_day.values()) / 7
			top_days = sorted(by_day.items(), key=lambda x: x[1], reverse=True)
			
			# Detectar "Revenge Spending" (sequência de gastos altos após economia)
			# Simplificado: busca picos súbitos
			
			return {
				"peak_days": [{"day": d, "total": t, "vs_avg": (t/avg_day - 1) * 100} for d, t in top_days[:2]],
				"summary": f"Seus gastos são {((top_days[0][1]/avg_day - 1) * 100):.1f}% maiores às {top_days[0][0]}"
			}
		finally:
			await conn.close()

	async def detect_lifestyle_inflation(self, user_id: str):
		conn = await self.get_db_connection()
		try:
			# Comparar gastos discricionários (Lazer, Alimentação Fora, etc) 
			# últimos 3 meses vs 3 meses anteriores
			rows = await conn.fetch("""
				SELECT 
					c.name as category,
					CASE WHEN date > NOW() - INTERVAL '90 days' THEN 'recent' ELSE 'previous' END as period,
					SUM(amount) as total
				FROM transactions t
				JOIN connected_accounts a ON t.account_id = a.id
				JOIN categories c ON t.category_id = c.id
				WHERE a.user_id = $1 AND direction = 'debit'
				  AND date > NOW() - INTERVAL '180 days'
				  AND c.name IN ('Lazer', 'Alimentação', 'Assinaturas', 'Outros')
				GROUP BY category, period
			""", user_id)
			
			categories = {}
			for r in rows:
				cat = r['category']
				if cat not in categories: categories[cat] = {"recent": 0, "previous": 0}
				categories[cat][r['period']] = float(r['total'])
			
			inflation_items = []
			for cat, data in categories.items():
				if data['previous'] > 0:
					increase = (data['recent'] / data['previous']) - 1
					if increase > 0.2: # Aumento de 20%
						inflation_items.append({
							"category": cat,
							"increase_pct": increase * 100,
							"monthly_diff": (data['recent'] - data['previous']) / 3
						})
			
			return {
				"inflation_detected": len(inflation_items) > 0,
				"items": inflation_items
			}
		finally:
			await conn.close()
