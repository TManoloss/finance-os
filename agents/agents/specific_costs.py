import json
import collections
from datetime import date, datetime, timedelta
from app.services.base_agent import BaseAgent

MEAL_COST_SYSTEM_PROMPT = """
Você é um analista financeiro sênior especializado em custos de vida e alimentação.
Sua tarefa é analisar os dados de gastos com refeições do usuário e fornecer insights acionáveis em português.
Seja específico com números reais. Não use jargão financeiro exagerado.
Foque no custo por refeição e na comparação entre canais (delivery, mercado, restaurante).
"""

MEAL_COST_INSIGHT_PROMPT = """
Com base nos dados de alimentação abaixo, escreva um insight em 2-3 frases em português.

Dados: {meal_data_json}

Exemplo de tom: "Cada refeição sua custa em média R$42. Nos dias que você pede
delivery, esse custo sobe para R$67 por refeição — 60% mais caro que quando
você cozinha ou vai a um restaurante."
"""

CONVENIENCE_SYSTEM_PROMPT = """
Você é um analista financeiro focado em eficiência de gastos e "prêmio de conveniência".
Sua tarefa é analisar o quanto o usuário paga a mais pela praticidade em diferentes categorias.
Forneça insights em português, tom informativo e sem julgamento. O objetivo é conscientização.
"""

CONVENIENCE_INSIGHT_PROMPT = """
Analise o índice de conveniência do usuário e escreva um insight em português.

Dados: {convenience_data}

Formato esperado: "Você gasta R$X/mês a mais pela conveniência — principalmente em [categorias].
Isso representa Y% da sua renda estimada. Em um ano, são R$Z pagos pela praticidade."
"""

CONVENIENCE_PAIRS = [
    {
        "name": "Delivery vs Cozinhar",
        "convenient_keywords": ["ifood", "rappi", "uber eats", "ze delivery", "delivery"],
        "economic_proxy": "supermercado",
        "premium_estimate_percent": 60,  # delivery custa ~60% a mais por refeição
        "category": "Alimentação"
    },
    {
        "name": "Uber/99 vs Transporte Público",
        "convenient_keywords": ["uber", "99app", "99 taxi", "cabify"],
        "economic_proxy": "bilhete único",  # detectar por valor R$4-6
        "premium_estimate_percent": 400,
        "category": "Transporte"
    }
]

class SpecificCostsAgent(BaseAgent):
    async def run(self, user_id: str):
        meal_cost = await self.calculate_real_meal_cost(user_id)
        convenience = await self.calculate_convenience_spending(user_id)
        return {
            "meal_cost": meal_cost,
            "convenience": convenience
        }

    async def calculate_real_meal_cost(self, user_id: str, months: int = 3):
        conn = await self.get_db_connection()
        try:
            start_date = date.today() - timedelta(days=months * 30)
            
            # 1. Buscar transações de alimentação
            rows = await conn.fetch("""
                SELECT t.amount, t.merchant_name, t.description, c.name as category_name, t.date
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                LEFT JOIN categories c ON t.category_id = c.id
                WHERE acc.user_id = $1 
                AND t.date >= $2 
                AND t.direction = 'debit'
                AND (c.name = 'Alimentação' OR t.merchant_name ILIKE ANY($3))
            """, user_id, start_date, ['%ifood%', '%rappi%', '%uber eats%', '%supermercado%', '%restaurante%', '%pão de açúcar%', '%carrefour%'])

            if not rows:
                return {"error": "Sem transações de alimentação suficientes"}

            # 2. Classificar por canal
            channels = collections.defaultdict(float)
            channel_counts = collections.defaultdict(int)
            
            delivery_keywords = ['ifood', 'rappi', 'uber eats', 'ze delivery', 'delivery']
            market_keywords = ['supermercado', 'pão de açúcar', 'carrefour', 'extra', 'atacadão', 'assaí', 'st marche', 'sams club', 'mercado']
            
            for r in rows:
                merchant = (r['merchant_name'] or r['description']).lower()
                amount = float(r['amount'])
                
                if any(k in merchant for k in delivery_keywords):
                    channel = "delivery"
                elif any(k in merchant for k in market_keywords):
                    channel = "supermercado"
                else:
                    channel = "restaurante" # Fallback para Restaurante se for Alimentação
                
                channels[channel] += amount
                channel_counts[channel] += 1

            total_spent = sum(channels.values())
            days_in_period = (date.today() - start_date).days
            total_meals_est = days_in_period * 3
            
            avg_cost_per_meal = total_spent / total_meals_est if total_meals_est > 0 else 0
            
            report_data = {
                "total_spent": total_spent,
                "total_meals_estimated": total_meals_est,
                "avg_cost_per_meal": avg_cost_per_meal,
                "by_channel": {
                    k: {
                        "total": v,
                        "count": channel_counts[k],
                        "percent": (v / total_spent * 100) if total_spent > 0 else 0
                    } for k, v in channels.items()
                },
                "period_days": days_in_period
            }

            # 3. LLM Insights
            prompt = MEAL_COST_INSIGHT_PROMPT.format(meal_data_json=json.dumps(report_data))
            llm_response = await self.llm.completion(prompt, system_prompt=MEAL_COST_SYSTEM_PROMPT)
            report_data["insight"] = llm_response

            # 4. Cache
            expires_at = datetime.now() + timedelta(days=7)
            await conn.execute("""
                INSERT INTO report_cache (user_id, report_type, period_key, result_json, expires_at)
                VALUES ($1, $2, $3, $4, $5)
                ON CONFLICT (user_id, report_type, period_key) DO UPDATE
                SET result_json = EXCLUDED.result_json,
                    expires_at = EXCLUDED.expires_at,
                    computed_at = NOW()
            """, user_id, "meal_cost_analysis", date.today().strftime("%Y-%m"), 
                 json.dumps(report_data), expires_at)

            return report_data
        finally:
            await conn.close()

    async def calculate_convenience_spending(self, user_id: str, months: int = 3):
        conn = await self.get_db_connection()
        try:
            start_date = date.today() - timedelta(days=months * 30)
            
            # Buscar transações para análise de conveniência
            all_txs = await conn.fetch("""
                SELECT t.amount, t.merchant_name, t.description, t.date
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                WHERE acc.user_id = $1 
                AND t.date >= $2 
                AND t.direction = 'debit'
            """, user_id, start_date)

            results = []
            total_premium = 0
            
            for pair in CONVENIENCE_PAIRS:
                convenient_spent = 0
                count = 0
                for tx in all_txs:
                    merchant = (tx['merchant_name'] or tx['description']).lower()
                    if any(k in merchant for k in pair["convenient_keywords"]):
                        convenient_spent += float(tx['amount'])
                        count += 1
                
                if convenient_spent > 0:
                    # Estimar o prêmio pago
                    # Se premium_estimate_percent é 60%, então valor_conveniencia = valor_economico * 1.6
                    # valor_economico = valor_conveniencia / 1.6
                    # premium = valor_conveniencia - valor_economico
                    divisor = 1 + (pair["premium_estimate_percent"] / 100)
                    economic_est = convenient_spent / divisor
                    premium = convenient_spent - economic_est
                    
                    results.append({
                        "name": pair["name"],
                        "spent": convenient_spent,
                        "count": count,
                        "economic_est": economic_est,
                        "premium": premium,
                        "premium_percent": pair["premium_estimate_percent"]
                    })
                    total_premium += premium

            # Estimar renda para o insight
            estimated_income = await conn.fetchval("""
                SELECT SUM(amount) / $2
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                WHERE acc.user_id = $1 AND direction = 'credit' AND t.date >= $3
            """, user_id, months, start_date) or 1000 # Fallback

            report_data = {
                "pairs": results,
                "total_monthly_premium": total_premium / months,
                "total_annual_premium": (total_premium / months) * 12,
                "percent_of_income": ((total_premium / months) / float(estimated_income) * 100) if estimated_income > 0 else 0
            }

            # LLM Insight
            prompt = CONVENIENCE_INSIGHT_PROMPT.format(
                convenience_data=json.dumps(report_data),
            )
            llm_response = await self.llm.completion(prompt, system_prompt=CONVENIENCE_SYSTEM_PROMPT)
            report_data["insight"] = llm_response

            # Cache
            expires_at = datetime.now() + timedelta(days=7)
            await conn.execute("""
                INSERT INTO report_cache (user_id, report_type, period_key, result_json, expires_at)
                VALUES ($1, $2, $3, $4, $5)
                ON CONFLICT (user_id, report_type, period_key) DO UPDATE
                SET result_json = EXCLUDED.result_json,
                    expires_at = EXCLUDED.expires_at,
                    computed_at = NOW()
            """, user_id, "convenience_analysis", date.today().strftime("%Y-%m"), 
                 json.dumps(report_data), expires_at)

            return report_data
        finally:
            await conn.close()
