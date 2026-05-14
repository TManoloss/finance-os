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
            
            # 2. Gerar Insights via LLM
            prompt = f"Dados das últimas 24h para o usuário {user_id}:\n{json.dumps(transactions, default=str)}"
            
            response_text = await self.llm.completion(prompt, system_prompt=DAILY_PROMPT)
            
            if response_text.startswith("ERRO_SISTEMA"):
                return {"error": response_text}

            # 3. Processar e salvar
            try:
                # Tenta encontrar o JSON dentro da resposta
                start_idx = response_text.find("{")
                end_idx = response_text.rfind("}")
                if start_idx == -1 or end_idx == -1:
                    raise ValueError("JSON não encontrado na resposta")
                
                clean_json = response_text[start_idx:end_idx+1]
                report_data = json.loads(clean_json)
                
                await self.save_report(
                    user_id, 
                    "daily", 
                    start_date.date(), 
                    end_date.date(), 
                    report_data["summary"], 
                    json.dumps(report_data.get("insights", []))
                )
                return report_data
            except Exception as e:
                print(f"Erro ao processar JSON do agente: {e}")
                return {"error": "Falha ao gerar relatório"}

        finally:
            await conn.close()
