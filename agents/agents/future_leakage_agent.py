import json
import random
import collections
from datetime import date, datetime, timedelta
from app.services.base_agent import BaseAgent

FUTURE_PREDICTION_PROMPT = """
Você é um analista financeiro preditivo focado em antecipar cenários futuros para o usuário.
Com base nas projeções matemáticas de saldo e riscos de cheque especial (Monte Carlo), gere uma análise narrativa.

Dados de Projeção:
{projection_data_json}

Sua tarefa é explicar para o usuário qual é a tendência para o fim do mês e o quão seguro ele está nos próximos 7 dias.

Diretrizes:
1. Tom: Preventivo, calmo e baseado em dados.
2. Seja específico sobre a chance de fechar o mês no azul ou vermelho.
3. Explique sutilmente o risco de saldo negativo se houver.

Retorne um JSON com:
- month_end_forecast: "previsão curta do fechamento do mês"
- risk_assessment: "avaliação do risco para os próximos 7 dias"
- key_insight: "um insight principal sobre o futuro financeiro imediato"
- narrative: "o parágrafo completo de análise em português"
"""

MICRO_SPENDING_PROMPT = """
Você é um "detetive financeiro" focado em pequenos vazamentos (micro-gastos).
Analise o impacto acumulado de transações abaixo de R$ 30.

Dados de Micro-gastos:
{micro_spending_json}

Sua tarefa é revelar ao usuário como esses pequenos valores se somam e impactam o orçamento anual.

Diretrizes:
1. Tom: Revelador, curioso (estilo "você sabia?") e sem julgamento.
2. Mostre o impacto anual desses gastos e o equivalente em investimento.
3. Identifique padrões (ex: gastos recorrentes no mesmo merchant).

Retorne um JSON com:
- monthly_impact_summary: "resumo do impacto mensal"
- annual_projection_narrative: "narrativa do impacto anual e custo de oportunidade"
- detected_habits: ["lista de hábitos de micro-gastos detectados"]
- detective_note: "uma nota final do 'detetive' sobre onde está o maior vazamento"
"""

class FutureLeakageAgent(BaseAgent):
    async def run(self, user_id: str):
        # Para compatibilidade com o padrão run
        prediction = await self.predict_future(user_id)
        micro = await self.analyze_micro_transactions(user_id)
        return {
            "future_prediction": prediction,
            "micro_spending": micro
        }

    async def predict_future(self, user_id: str):
        # Esta função agrega as previsões para o relatório
        month_end = await self.predict_month_end_balance(user_id)
        overdraft = await self.predict_overdraft_risk(user_id)
        
        projection_data = {
            **month_end,
            **overdraft
        }

        # Narrative
        prompt = f"Dados de Projeção:\n{json.dumps(projection_data, default=str)}"
        llm_response = await self.llm.completion(prompt, system_prompt=FUTURE_PREDICTION_PROMPT)
        
        try:
            start_idx = llm_response.find("{")
            end_idx = llm_response.rfind("}")
            result = json.loads(llm_response[start_idx:end_idx+1])
        except:
            result = {"narrative": llm_response}
        
        result["data"] = projection_data
        return result

    async def predict_month_end_balance(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # 1. Obter saldo atual (liquidez)
            current_balance = await conn.fetchval("""
                SELECT COALESCE(SUM(balance), 0)
                FROM connected_accounts
                WHERE user_id = $1 AND account_type IN ('CHECKING', 'SAVINGS')
            """, user_id)
            curr_bal = float(current_balance)

            # 2. Calcular gasto médio diário (últimos 30 dias)
            daily_burn = await conn.fetchval("""
                SELECT COALESCE(SUM(amount), 0) / 30.0
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 AND direction = 'debit' AND date > NOW() - INTERVAL '30 days'
            """, user_id)
            avg_burn = float(daily_burn or 0)

            # 3. Obter compromissos fixos restantes no mês
            upcoming_installments = await conn.fetchval("""
                SELECT COALESCE(SUM(total_amount / installments_total), 0)
                FROM installments i
                JOIN connected_accounts a ON i.account_id = a.id
                WHERE a.user_id = $1
            """, user_id)
            
            subscriptions_sum = await conn.fetchval("""
                SELECT COALESCE(SUM(amount), 0)
                FROM (
                    SELECT DISTINCT ON (merchant_name) amount
                    FROM transactions t
                    JOIN connected_accounts a ON t.account_id = a.id
                    WHERE a.user_id = $1 AND is_recurring = true
                    ORDER BY merchant_name, date DESC
                ) s
            """, user_id)
            
            fixed_commitments = float(upcoming_installments) + float(subscriptions_sum)

            # 4. Projeção de fim de mês
            today = date.today()
            next_month = (today.replace(day=28) + timedelta(days=4)).replace(day=1)
            days_remaining = (next_month - today).days
            
            predicted_spending = avg_burn * days_remaining
            projected_balance = curr_bal - predicted_spending - fixed_commitments

            return {
                "current_balance": curr_bal,
                "avg_daily_burn": round(avg_burn, 2),
                "fixed_commitments": round(fixed_commitments, 2),
                "days_remaining": days_remaining,
                "projected_month_end_balance": round(projected_balance, 2)
            }
        finally:
            await conn.close()

    async def predict_overdraft_risk(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            current_balance = await conn.fetchval("""
                SELECT COALESCE(SUM(balance), 0)
                FROM connected_accounts
                WHERE user_id = $1 AND account_type IN ('CHECKING', 'SAVINGS')
            """, user_id)
            curr_bal = float(current_balance)

            daily_burn = await conn.fetchval("""
                SELECT COALESCE(SUM(amount), 0) / 30.0
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 AND direction = 'debit' AND date > NOW() - INTERVAL '30 days'
            """, user_id)
            avg_burn = float(daily_burn or 0)

            # Para simulação precisamos da volatilidade
            daily_stats = await conn.fetch("""
                SELECT date, SUM(amount) as daily_total
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 AND direction = 'debit' AND date > NOW() - INTERVAL '30 days'
                GROUP BY date
            """, user_id)
            
            if len(daily_stats) > 1:
                import statistics
                amounts = [float(s['daily_total']) for s in daily_stats]
                std_dev = statistics.stdev(amounts)
            else:
                std_dev = avg_burn * 0.5 # fallback

            # Simulação Monte Carlo para 7 dias
            overdraft_count = 0
            for _ in range(1000):
                sim_bal = curr_bal
                for d in range(1, 8):
                    sim_spend = max(0, random.normalvariate(avg_burn, std_dev))
                    sim_bal -= sim_spend
                    # Pequena chance de cair compromisso fixo
                    if random.random() < (7/30):
                        sim_bal -= (avg_burn * 2) # estimativa simples
                    
                    if sim_bal < 0:
                        overdraft_count += 1
                        break
            
            overdraft_probability = overdraft_count / 1000.0
            return {
                "overdraft_probability_7d": round(overdraft_probability, 4),
                "risk_level": "alto" if overdraft_probability > 0.3 else ("médio" if overdraft_probability > 0.1 else "baixo")
            }
        finally:
            await conn.close()

    async def analyze_micro_transactions(self, user_id: str, threshold: float = 30.0):
        conn = await self.get_db_connection()
        try:
            rows = await conn.fetch("""
                SELECT amount, merchant_name, description, date, category_id, c.name as category_name
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                LEFT JOIN categories c ON t.category_id = c.id
                WHERE a.user_id = $1 
                AND direction = 'debit' 
                AND amount < $2
                AND date > NOW() - INTERVAL '30 days'
            """, user_id, threshold)

            if not rows:
                return {"message": "Nenhum micro-gasto detectado no período."}

            total_amount = sum(float(r['amount']) for r in rows)
            count = len(rows)
            
            by_merchant = collections.defaultdict(float)
            merchant_counts = collections.defaultdict(int)
            by_category = collections.defaultdict(float)
            
            for r in rows:
                m_name = r['merchant_name'] or r['description'] or "Desconhecido"
                by_merchant[m_name] += float(r['amount'])
                merchant_counts[m_name] += 1
                by_category[r['category_name'] or "Outros"] += float(r['amount'])

            habits = []
            for m, c in merchant_counts.items():
                if c >= 4:
                    habits.append({
                        "merchant": m,
                        "frequency": c,
                        "total_spent": round(by_merchant[m], 2),
                        "avg_spent": round(by_merchant[m] / c, 2)
                    })
            
            annual_impact = total_amount * 12
            selic_rate = 0.1075 
            invested_5y = total_amount * ((1 + selic_rate/12)**60 - 1) / (selic_rate/12)

            micro_spending_data = {
                "monthly_total": round(total_amount, 2),
                "transaction_count": count,
                "avg_micro_tx": round(total_amount / count, 2) if count > 0 else 0,
                "top_merchants": sorted([{"name": k, "value": round(v, 2)} for k, v in by_merchant.items()], key=lambda x: x['value'], reverse=True)[:5],
                "top_categories": sorted([{"name": k, "value": round(v, 2)} for k, v in by_category.items()], key=lambda x: x['value'], reverse=True)[:5],
                "detected_habits": habits,
                "annual_impact": round(annual_impact, 2),
                "invested_5y_equivalent": round(invested_5y, 2)
            }

            prompt = f"Dados de Micro-gastos:\n{json.dumps(micro_spending_data, default=str)}"
            llm_response = await self.llm.completion(prompt, system_prompt=MICRO_SPENDING_PROMPT)
            
            try:
                start_idx = llm_response.find("{")
                end_idx = llm_response.rfind("}")
                result = json.loads(llm_response[start_idx:end_idx+1])
            except:
                result = {"narrative": llm_response}
            
            result["data"] = micro_spending_data
            return result
        finally:
            await conn.close()
