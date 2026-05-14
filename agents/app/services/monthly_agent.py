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
            
            # 4. Gerar Insights via LLM
            prompt = f"Fechamento mensal para o usuário {user_id}:\n{json.dumps(context_data, default=str)}"
            response_text = await self.llm.completion(prompt, system_prompt=MONTHLY_PROMPT)
            
            if response_text.startswith("ERRO_SISTEMA"):
                return {"error": response_text}

            # 5. Processar e salvar
            try:
                start_idx = response_text.find("{")
                end_idx = response_text.rfind("}")
                clean_json = response_text[start_idx:end_idx+1]
                report_data = json.loads(clean_json)
                
                await self.save_report(
                    user_id, 
                    "monthly", 
                    start_date.date(), 
                    end_date.date(), 
                    report_data["summary"], 
                    json.dumps(report_data.get("insights", []))
                )
                return report_data
            except Exception as e:
                print(f"Erro ao processar JSON do agente mensal: {e}")
                return {"error": "Falha ao gerar fechamento mensal"}

        finally:
            await conn.close()
