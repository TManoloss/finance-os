import json
from datetime import date, datetime, timedelta
from app.services.base_agent import BaseAgent
import httpx

PERSONAL_INFLATION_PROMPT = """
Você é um analista financeiro especialista em inflação pessoal.
Analise os dados de inflação pessoal abaixo e escreva um resumo em português, tom direto, focando no impacto real em reais.

Dados:
{inflation_data_json}

Foque no impacto em reais, não em percentuais abstratos.
Exemplo de tom: "Seu supermercado habitual ficou 23% mais caro nos últimos 12 meses — você está pagando R$187 a mais por mês pelo mesmo carrinho de compras."

Retorne obrigatoriamente um JSON com:
- personal_inflation_rate: float (taxa total ponderada)
- ipca_comparison: str (comparação curta com IPCA oficial)
- top_categories_inflation: lista de objetos [{category, variation_percent, impact_reais}]
- insights: str (o parágrafo narrativo solicitado)
"""

class PersonalInflationAgent(BaseAgent):
    def get_last_month(self, d: date) -> date:
        first = d.replace(day=1)
        last_month = first - timedelta(days=1)
        return last_month.replace(day=1)

    async def run(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # 1. Definir períodos
            today = date.today()
            current_month_start = today.replace(day=1)
            last_month_start = self.get_last_month(current_month_start)
            
            # Fetch IPCA
            ipca_rate = await self._fetch_ipca()
            
            # 2. Buscar gastos por categoria no mês atual
            current_spending = await conn.fetch("""
                SELECT c.name as category, SUM(t.amount) as total
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                JOIN categories c ON t.category_id = c.id
                WHERE acc.user_id = $1 
                AND t.date >= $2 AND t.date < $3
                AND t.direction = 'debit'
                GROUP BY c.name
            """, user_id, current_month_start, current_month_start + timedelta(days=32)) 
            # Note: timedelta(days=32) + replace(day=1) is a bit hacky but works for month end
            
            # Re-fetch with exact bounds for current month
            next_month_start = (current_month_start + timedelta(days=32)).replace(day=1)
            current_spending = await conn.fetch("""
                SELECT c.name as category, SUM(t.amount) as total
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                JOIN categories c ON t.category_id = c.id
                WHERE acc.user_id = $1 
                AND t.date >= $2 AND t.date < $3
                AND t.direction = 'debit'
                GROUP BY c.name
            """, user_id, current_month_start, next_month_start)

            # 3. Buscar gastos por categoria no mês anterior
            last_spending = await conn.fetch("""
                SELECT c.name as category, SUM(t.amount) as total
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                JOIN categories c ON t.category_id = c.id
                WHERE acc.user_id = $1 
                AND t.date >= $2 AND t.date < $3
                AND t.direction = 'debit'
                GROUP BY c.name
            """, user_id, last_month_start, current_month_start)

            if not current_spending or not last_spending:
                return {
                    "personal_inflation_rate": 0,
                    "ipca_comparison": f"IPCA: {ipca_rate}%",
                    "top_categories_inflation": [],
                    "insights": "Dados insuficientes para calcular a inflação pessoal este mês. Continue registrando suas transações!"
                }

            # 4. Calcular variação
            current_dict = {r['category']: float(r['total']) for r in current_spending}
            last_dict = {r['category']: float(r['total']) for r in last_spending}
            
            comparison_data = []
            total_last = sum(last_dict.values())
            weighted_inflation = 0
            
            for cat, amount in current_dict.items():
                if cat in last_dict:
                    prev_amount = last_dict[cat]
                    if prev_amount > 0:
                        variation = ((amount - prev_amount) / prev_amount) * 100
                        impact = amount - prev_amount
                        weight = prev_amount / total_last if total_last > 0 else 0
                        weighted_inflation += variation * weight
                        
                        comparison_data.append({
                            "category": cat,
                            "current_amount": round(amount, 2),
                            "previous_amount": round(prev_amount, 2),
                            "variation_percent": round(variation, 2),
                            "impact_reais": round(impact, 2)
                        })

            # 5. Gerar insights via LLM
            inflation_context = {
                "user_id": user_id,
                "period": current_month_start.strftime("%Y-%m"),
                "ipca_official": ipca_rate,
                "categories": comparison_data,
                "personal_inflation_weighted": round(weighted_inflation, 2)
            }
            
            prompt = f"Dados de inflação para o usuário {user_id}:\n{json.dumps(inflation_context, default=str)}"
            response_text = await self.llm.completion(prompt, system_prompt=PERSONAL_INFLATION_PROMPT)
            
            # Extract JSON
            try:
                start_idx = response_text.find("{")
                end_idx = response_text.rfind("}")
                if start_idx == -1 or end_idx == -1:
                    raise ValueError("JSON not found")
                report_data = json.loads(response_text[start_idx:end_idx+1])
            except Exception:
                report_data = {
                    "personal_inflation_rate": round(weighted_inflation, 2),
                    "ipca_comparison": f"Sua inflação: {round(weighted_inflation, 2)}% vs IPCA: {ipca_rate}%",
                    "top_categories_inflation": comparison_data[:3],
                    "insights": response_text
                }

            # 6. Salvar snapshot
            await conn.execute("""
                INSERT INTO inflation_snapshots (user_id, period_month, personal_inflation_rate, ipca_rate, category_breakdown)
                VALUES ($1, $2, $3, $4, $5)
                ON CONFLICT (user_id, period_month) DO UPDATE 
                SET personal_inflation_rate = EXCLUDED.personal_inflation_rate,
                    ipca_rate = EXCLUDED.ipca_rate,
                    category_breakdown = EXCLUDED.category_breakdown
            """, user_id, current_month_start, float(report_data.get("personal_inflation_rate", weighted_inflation)), 
                 ipca_rate, json.dumps(comparison_data))

            # 7. Salvar no cache de relatórios
            # Usar 30 dias de validade para inflação mensal
            expires_at = datetime.now() + timedelta(days=30)
            await conn.execute("""
                INSERT INTO report_cache (user_id, report_type, period_key, result_json, expires_at)
                VALUES ($1, $2, $3, $4, $5)
                ON CONFLICT (user_id, report_type, period_key) DO UPDATE
                SET result_json = EXCLUDED.result_json,
                    expires_at = EXCLUDED.expires_at,
                    computed_at = NOW()
            """, user_id, "personal_inflation", current_month_start.strftime("%Y-%m"), 
                 json.dumps(report_data), expires_at)

            return report_data

        finally:
            await conn.close()

    async def _fetch_ipca(self):
        try:
            async with httpx.AsyncClient(timeout=10.0) as client:
                # Pegar último valor do IPCA
                resp = await client.get("https://api.bcb.gov.br/dados/serie/bcdata.sgs.433/dados/ultimos/1?formato=json")
                if resp.status_code == 200:
                    data = resp.json()
                    return float(data[0]['valor'])
                return 4.5
        except Exception:
            return 4.5 # Default fallback
