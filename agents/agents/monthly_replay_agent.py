from app.services.base_agent import BaseAgent
from datetime import datetime
import json
import logging

logger = logging.getLogger(__name__)

class MonthlyReplayAgent(BaseAgent):
    async def run(self, user_id: str, month: str):
        # month format: "YYYY-MM"
        conn = await self.get_db_connection()
        try:
            # Pegar o nome do usuário
            user_row = await conn.fetchrow("SELECT name FROM users WHERE id = $1", user_id)
            user_name = user_row['name'] if user_row else "Usuário"

            dt = datetime.strptime(month, "%Y-%m")
            start_date = dt.date().replace(day=1)
            # Find end of month
            if dt.month == 12:
                end_date = dt.date().replace(year=dt.year+1, month=1, day=1)
            else:
                end_date = dt.date().replace(month=dt.month+1, day=1)
            
            # Fetch summary
            tx_rows = await conn.fetch("""
                SELECT 
                    t.direction,
                    SUM(t.amount) as total
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 AND t.date >= $2 AND t.date < $3
                GROUP BY t.direction
            """, user_id, start_date, end_date)
            
            cashflow = {"in": 0, "out": 0}
            for row in tx_rows:
                if row['direction'] == 'credit':
                    cashflow['in'] = float(row['total'] or 0)
                else:
                    cashflow['out'] = float(row['total'] or 0)
                    
            # Top merchants
            merchant_rows = await conn.fetch("""
                SELECT merchant_name, SUM(amount) as total
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 AND t.date >= $2 AND t.date < $3 AND t.direction = 'debit'
                GROUP BY merchant_name
                ORDER BY total DESC
                LIMIT 5
            """, user_id, start_date, end_date)
            top_merchants = [{"merchant_name": r['merchant_name'], "total": float(r['total'])} for r in merchant_rows]
            
            # Top categories
            cat_rows = await conn.fetch("""
                SELECT c.name as category_name, SUM(t.amount) as total
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                LEFT JOIN categories c ON t.category_id = c.id
                WHERE a.user_id = $1 AND t.date >= $2 AND t.date < $3 AND t.direction = 'debit'
                GROUP BY c.name
                ORDER BY total DESC
                LIMIT 3
            """, user_id, start_date, end_date)
            top_categories = [{"category_name": r['category_name'] or "Outros", "total": float(r['total'])} for r in cat_rows]

            # Maior gasto
            biggest_tx_row = await conn.fetchrow("""
                SELECT merchant_name, amount, date
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 AND t.date >= $2 AND t.date < $3 AND t.direction = 'debit'
                ORDER BY amount DESC
                LIMIT 1
            """, user_id, start_date, end_date)
            biggest_tx = dict(biggest_tx_row) if biggest_tx_row else None
            if biggest_tx:
                biggest_tx['amount'] = float(biggest_tx['amount'])
                biggest_tx['date'] = str(biggest_tx['date'])

            context = {
                "month": month,
                "cashflow": cashflow,
                "top_merchants": top_merchants,
                "top_categories": top_categories,
                "biggest_transaction": biggest_tx
            }

            prompt = f"""
Você é o assistente do app Finance OS, criando o "Replay Financeiro" do mês estilo Spotify Wrapped para {user_name}.
Dados do mês {month}:
{json.dumps(context, ensure_ascii=False)}

Crie uma narrativa envolvente, dividida em telas (slides). Para cada slide, dê um título curto e um texto divertido/emocional sem julgamentos. Use números reais, mas foque na história.

Retorne APENAS um JSON válido no formato:
{{
  "slides": [
    {{
      "title": "A grande revelação do mês",
      "text": "Você focou bastante em X...",
      "highlight_value": "R$ 1.230"
    }},
    ...
  ]
}}
"""
            llm_response = await self.llm.generate(prompt)
            # Remove possible markdown fences from LLM output
            if llm_response.startswith("```json"):
                llm_response = llm_response.replace("```json", "", 1)
                if llm_response.endswith("```"):
                    llm_response = llm_response[:-3]
            elif llm_response.startswith("```"):
                llm_response = llm_response.replace("```", "", 1)
                if llm_response.endswith("```"):
                    llm_response = llm_response[:-3]
            
            try:
                narrative = json.loads(llm_response.strip())
            except json.JSONDecodeError:
                narrative = {"slides": [{"title": "Resumo do Mês", "text": llm_response, "highlight_value": ""}]}

            # Criar a tabela se não existir
            await conn.execute("""
                CREATE TABLE IF NOT EXISTS monthly_replays (
                    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                    user_id UUID REFERENCES users(id),
                    month TEXT NOT NULL,
                    narrative JSONB NOT NULL,
                    data JSONB NOT NULL,
                    created_at TIMESTAMPTZ DEFAULT NOW(),
                    UNIQUE(user_id, month)
                )
            """)

            # Upsert na tabela monthly_replays
            await conn.execute("""
                INSERT INTO monthly_replays (user_id, month, narrative, data)
                VALUES ($1, $2, $3, $4)
                ON CONFLICT (user_id, month) 
                DO UPDATE SET narrative = EXCLUDED.narrative, data = EXCLUDED.data, created_at = NOW()
            """, user_id, month, json.dumps(narrative), json.dumps(context))

            return {
                "user_id": user_id,
                "month": month,
                "narrative": narrative,
                "data": context
            }

        except Exception as e:
            logger.error(f"Erro em MonthlyReplayAgent: {str(e)}")
            return {"error": str(e)}
        finally:
            await conn.close()