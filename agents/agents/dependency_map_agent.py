from app.services.base_agent import BaseAgent
from datetime import datetime, timedelta
import json
import logging

logger = logging.getLogger(__name__)

class DependencyMapAgent(BaseAgent):
    async def run(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # Query transactions from the last 6 months
            six_months_ago = datetime.now() - timedelta(days=180)
            
            rows = await conn.fetch("""
                SELECT 
                    c.id as category_id,
                    c.name as category_name,
                    t.merchant_name,
                    SUM(t.amount) as total_amount,
                    COUNT(t.id) as tx_count,
                    MAX(t.date) as last_seen
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                LEFT JOIN categories c ON t.category_id = c.id
                WHERE a.user_id = $1 AND t.direction = 'debit' AND t.date >= $2
                GROUP BY c.id, c.name, t.merchant_name
                ORDER BY c.name, total_amount DESC
            """, user_id, six_months_ago.date())

            if not rows:
                return {"user_id": user_id, "dependency_map": []}

            # Build hierarchical structure
            categories_map = {}
            for row in rows:
                cat_id = str(row['category_id']) if row['category_id'] else "uncategorized"
                cat_name = row['category_name'] or "Outros"
                
                if cat_id not in categories_map:
                    categories_map[cat_id] = {
                        "category_name": cat_name,
                        "merchants": [],
                        "total_category_spending": 0
                    }
                
                merchant_total = float(row['total_amount'] or 0)
                tx_count = row['tx_count']
                
                # Determine dependency level based on frequency and amount
                # Heuristic: 
                # > 24 tx in 6 months (approx 1/week) = Crítica
                # > 6 tx in 6 months (approx 1/month) = Alta
                # else Normal
                if tx_count >= 24:
                    dep_level = "Crítica"
                elif tx_count >= 6:
                    dep_level = "Alta"
                else:
                    dep_level = "Normal"

                categories_map[cat_id]["merchants"].append({
                    "merchant_name": row['merchant_name'],
                    "total_amount": merchant_total,
                    "tx_count": tx_count,
                    "dependency_level": dep_level,
                    "last_seen": str(row['last_seen'])
                })
                categories_map[cat_id]["total_category_spending"] += merchant_total

            # Format the output map
            dependency_map = []
            for cat_id, data in categories_map.items():
                dependency_map.append({
                    "category_id": cat_id,
                    "category_name": data["category_name"],
                    "total_amount": data["total_category_spending"],
                    "merchants": data["merchants"]
                })
                
            # Optional: use LLM to summarize dependencies
            prompt = f"""
Você é um analista financeiro. Analise o mapa de dependência do usuário (onde ele concentra seus gastos):
{json.dumps(dependency_map, ensure_ascii=False)}

Gere 1 ou 2 frases resumindo a dependência crítica do usuário, apontando se ele é muito dependente de poucos lugares.
Seja direto, profissional e não use julgamentos morais.
"""
            llm_insight = await self.llm.generate(prompt)

            return {
                "user_id": user_id,
                "dependency_map": dependency_map,
                "insight": llm_insight
            }

        except Exception as e:
            logger.error(f"Erro em DependencyMapAgent: {str(e)}")
            return {"error": str(e)}
        finally:
            await conn.close()
