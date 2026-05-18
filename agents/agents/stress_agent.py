import json
import statistics
from datetime import date, datetime, timedelta
from app.services.base_agent import BaseAgent

STRESS_CONTEXT_PROMPT = """
Você é um analista financeiro sênior. Com base nos dados de stress financeiro abaixo, gere exatamente UMA frase contextual sobre o momento financeiro do usuário.
Tom: direto, sem drama, empático, sem usar a palavra "stress" ou "estresse". Seja específico se houver dados claros.

Dados de Stress:
{stress_data_json}

Exemplos de tom esperado:
- "Você consumiu 68% do salário em 12 dias — ritmo acelerado."
- "Combinação de parcelas e cartão consome 71% da renda este mês."
- "O volume de compras parceladas está pressionando seu fluxo de caixa semanal."
"""

SURVIVAL_RECOMMENDATIONS_PROMPT = """
O sistema detectou que o usuário está em MODO CRÍTICO ou PRESSÃO financeira (Modo Sobrevivência).
Analise os dados abaixo e gere uma lista de ações prioritárias para evitar um colapso financeiro este mês.

Dados:
{survival_data_json}

Retorne um JSON com:
- recommendations: [
    {{
        "title": "Título da ação",
        "description": "Descrição detalhada com números se possível",
        "impact": "Alto/Médio/Baixo"
    }}
  ]
- daily_spending_limit: float (valor sugerido por dia até o próximo salário)

Considere:
- Assinaturas que podem ser pausadas/canceladas.
- Categorias de gastos variáveis onde o corte é mais imediato.
- Projeção de saldo negativo e como reverter.
"""

class StressAgent(BaseAgent):
    async def run(self, user_id: str, period: str = ""):
        # Default run: calculates both and returns a summary
        stress = await self.calculate_stress_score(user_id)
        survival = await self.evaluate_survival_mode(user_id)
        return {
            "stress_score": stress,
            "survival_mode": survival
        }

    async def calculate_stress_score(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # 1. Burn rate (últimos 7 dias) - 30%
            # Gasto médio diário nos últimos 7 dias vs média histórica
            burn_data = await conn.fetchrow("""
                WITH recent AS (
                    SELECT COALESCE(SUM(amount), 0) / 7.0 as daily_avg
                    FROM transactions t
                    JOIN connected_accounts a ON t.account_id = a.id
                    WHERE a.user_id = $1 AND direction = 'debit' AND date > NOW() - INTERVAL '7 days'
                ),
                historical AS (
                    SELECT COALESCE(SUM(amount), 0) / 90.0 as daily_avg
                    FROM transactions t
                    JOIN connected_accounts a ON t.account_id = a.id
                    WHERE a.user_id = $1 AND direction = 'debit' AND date > NOW() - INTERVAL '90 days'
                )
                SELECT recent.daily_avg as recent_avg, historical.daily_avg as hist_avg
                FROM recent, historical
            """, user_id)
            
            recent_avg = float(burn_data['recent_avg'] or 0)
            hist_avg = float(burn_data['hist_avg'] or 1) # avoid div by zero
            burn_ratio = recent_avg / hist_avg
            burn_score = max(0, 100 - (burn_ratio - 1) * 100) if burn_ratio > 1 else 100

            # 2. Salary consumption (25%)
            # Identificar último salário (crédito recorrente grande)
            salary_data = await conn.fetchval("""
                SELECT amount
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 AND direction = 'credit' AND amount > 1000
                ORDER BY date DESC LIMIT 1
            """, user_id)
            salary = float(salary_data or 3000) # Fallback to 3000
            
            spent_current_month = await conn.fetchval("""
                SELECT SUM(amount)
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 AND direction = 'debit' 
                AND date >= date_trunc('month', NOW())::date
            """, user_id)
            spent_current = float(spent_current_month or 0)
            
            # Fração consumida vs dias passados no mês
            day_of_month = datetime.now().day
            days_in_month = 30 # simplified
            expected_consumption_ratio = day_of_month / days_in_month
            actual_consumption_ratio = spent_current / salary if salary > 0 else 1
            
            salary_score = max(0, 100 - (actual_consumption_ratio - expected_consumption_ratio) * 200) if actual_consumption_ratio > expected_consumption_ratio else 100

            # 3. Crédito como % do gasto total (20%)
            credit_spent = await conn.fetchval("""
                SELECT SUM(amount)
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 AND direction = 'debit' AND a.account_type = 'CREDIT' AND date > NOW() - INTERVAL '30 days'
            """, user_id)
            total_spent_30d = await conn.fetchval("""
                SELECT SUM(amount)
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 AND direction = 'debit' AND date > NOW() - INTERVAL '30 days'
            """, user_id)
            
            credit_ratio = float(credit_spent or 0) / float(total_spent_30d or 1)
            credit_score = max(0, 100 - (credit_ratio - 0.4) * 200) if credit_ratio > 0.4 else 100 # >60% = high stress

            # 4. Parcelamentos vs renda (15%)
            installments_sum = await conn.fetchval("""
                SELECT SUM(total_amount / installments_total)
                FROM installments i
                JOIN connected_accounts a ON i.account_id = a.id
                WHERE a.user_id = $1
            """, user_id)
            inst_monthly = float(installments_sum or 0)
            inst_income_ratio = inst_monthly / salary if salary > 0 else 0
            inst_score = max(0, 100 - (inst_income_ratio - 0.2) * 400) if inst_income_ratio > 0.2 else 100 # >45% = critical

            # 5. Volatilidade recente (10%)
            daily_spending = await conn.fetch("""
                SELECT date, SUM(amount) as total
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 AND direction = 'debit' AND date > NOW() - INTERVAL '14 days'
                GROUP BY date
            """, user_id)
            
            if len(daily_spending) > 1:
                amounts = [float(d['total']) for d in daily_spending]
                std_dev = statistics.stdev(amounts)
                # Normalizar: se std_dev > 0.5 * média diária, stress aumenta
                avg_daily = sum(amounts) / 14.0
                vol_ratio = std_dev / avg_daily if avg_daily > 0 else 0
                vol_score = max(0, 100 - vol_ratio * 50)
            else:
                vol_score = 100

            # Final Score
            total_score = (
                burn_score * 0.30 +
                salary_score * 0.25 +
                credit_score * 0.20 +
                inst_score * 0.15 +
                vol_score * 0.10
            )
            
            level = "tranquilo"
            if total_score < 40: level = "crítico"
            elif total_score < 60: level = "pressão"
            elif total_score < 80: level = "atenção"

            # Trend (comparar com snapshot de 7 dias atrás)
            prev_snapshot = await conn.fetchrow("""
                SELECT score FROM stress_score_snapshots
                WHERE user_id = $1 AND computed_at < NOW() - INTERVAL '6 days'
                ORDER BY computed_at DESC LIMIT 1
            """, user_id)
            
            trend = "estável"
            if prev_snapshot:
                diff = total_score - float(prev_snapshot['score'])
                if diff > 5: trend = "melhorando"
                elif diff < -5: trend = "piorando"

            components = {
                "burn_rate": round(burn_score, 2),
                "salary_consumption": round(salary_score, 2),
                "credit_usage": round(credit_score, 2),
                "installments": round(inst_score, 2),
                "volatility": round(vol_score, 2)
            }

            # Generate context with Claude
            stress_data = {
                "score": round(total_score, 2),
                "level": level,
                "trend": trend,
                "salary_consumption_percent": round(actual_consumption_ratio * 100, 1),
                "credit_ratio": round(credit_ratio * 100, 1),
                "installments_income_ratio": round(inst_income_ratio * 100, 1)
            }
            context_sentence = await self.llm.completion(
                STRESS_CONTEXT_PROMPT.format(stress_data_json=json.dumps(stress_data)),
                system_prompt="Você é um assistente financeiro direto e conciso."
            )

            # Save snapshot
            await conn.execute("""
                INSERT INTO stress_score_snapshots (user_id, score, level, components, trend)
                VALUES ($1, $2, $3, $4, $5)
            """, user_id, total_score, level, json.dumps(components), trend)

            return {
                "score": round(total_score, 2),
                "level": level,
                "trend": trend,
                "components": components,
                "insight": context_sentence.strip()
            }
        finally:
            await conn.close()

    async def evaluate_survival_mode(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # 1. Saldo projetado (35%)
            # Calcular saldo atual total (liquidez)
            current_balance = await conn.fetchval("""
                SELECT SUM(balance)
                FROM connected_accounts
                WHERE user_id = $1 AND account_type IN ('CHECKING', 'SAVINGS')
            """, user_id)
            curr_bal = float(current_balance or 0)
            
            # Estimar dias até o próximo salário
            salary_days = await conn.fetch("""
                SELECT date FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 AND direction = 'credit' AND amount > 1000
                ORDER BY date DESC LIMIT 3
            """, user_id)
            
            days_until_salary = 15 # default
            if len(salary_days) > 0:
                last_salary_date = salary_days[0]['date']
                # Ver se cai mensalmente
                next_expected = last_salary_date + timedelta(days=30)
                days_until_salary = (next_expected - date.today()).days
                if days_until_salary < 0: days_until_salary = 1 # Já passou, deve cair logo
            
            # Gasto médio diário (últimos 30 dias)
            daily_burn = await conn.fetchval("""
                SELECT COALESCE(SUM(amount), 0) / 30.0
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 AND direction = 'debit' AND date > NOW() - INTERVAL '30 days'
            """, user_id)
            burn = float(daily_burn or 100)
            
            projected_spending = burn * max(0, days_until_salary)
            # Buscar próximas parcelas de cartão
            upcoming_installments = await conn.fetchval("""
                SELECT SUM(total_amount / installments_total)
                FROM installments i
                JOIN connected_accounts a ON i.account_id = a.id
                WHERE a.user_id = $1
            """, user_id)
            upcoming_inst = float(upcoming_installments or 0) # Simplificação: assume que cai antes do salário
            
            projected_shortfall = curr_bal - projected_spending - upcoming_inst
            
            # Pontuação Saldo Projetado
            if projected_shortfall < 0: bal_score = 0
            elif projected_shortfall < 500: bal_score = 30
            elif projected_shortfall < 2000: bal_score = 70
            else: bal_score = 100

            # 2. Velocidade de gasto (25%)
            velocity_data = await conn.fetchrow("""
                WITH current_week AS (
                    SELECT COALESCE(SUM(amount), 0) as total
                    FROM transactions t
                    JOIN connected_accounts a ON t.account_id = a.id
                    WHERE a.user_id = $1 AND direction = 'debit' AND date > NOW() - INTERVAL '7 days'
                ),
                hist_week AS (
                    SELECT COALESCE(SUM(amount), 0) / 4.0 as total
                    FROM transactions t
                    JOIN connected_accounts a ON t.account_id = a.id
                    WHERE a.user_id = $1 AND direction = 'debit' 
                    AND date > NOW() - INTERVAL '35 days' AND date <= NOW() - INTERVAL '7 days'
                )
                SELECT current_week.total as curr, hist_week.total as hist
                FROM current_week, hist_week
            """, user_id)
            
            curr_v = float(velocity_data['curr'] or 0)
            hist_v = float(velocity_data['hist'] or 1)
            v_ratio = curr_v / hist_v
            if v_ratio > 1.5: vel_score = 0
            elif v_ratio > 1.2: vel_score = 40
            elif v_ratio > 1.0: vel_score = 70
            else: vel_score = 100

            # 3. Uso de crédito (20%)
            # Fatura em aberto vs limite (precisamos estimar o limite se não tiver)
            # Por enquanto vamos usar total_spent_credit / salary
            salary_ref_val = await conn.fetchval("""
                SELECT amount FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 AND direction = 'credit' AND amount > 1000
                ORDER BY date DESC LIMIT 1
            """, user_id)
            sal_ref = float(salary_ref_val or 3000)
            
            open_credit = await conn.fetchval("""
                SELECT SUM(amount) FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 AND direction = 'debit' AND a.account_type = 'CREDIT'
                AND date >= date_trunc('month', NOW())::date
            """, user_id)
            credit_usage = float(open_credit or 0) / (sal_ref * 1.5) # Limite estimado = 1.5x salário
            
            if credit_usage > 0.8: cred_score = 0
            elif credit_usage > 0.6: cred_score = 40
            elif credit_usage > 0.3: cred_score = 80
            else: cred_score = 100

            # 4. Proximidade do salário (10%)
            if days_until_salary > 15: prox_score = 0
            elif days_until_salary > 7: prox_score = 50
            elif days_until_salary > 3: prox_score = 80
            else: prox_score = 100

            # 5. Recorrência de saldo baixo (10%)
            low_bal_count = await conn.fetchval("""
                SELECT COUNT(*) FROM (
                    SELECT date, SUM(net_flow) OVER (ORDER BY date) as bal
                    FROM (
                        SELECT date, SUM(CASE WHEN direction = 'credit' THEN amount ELSE -amount END) as net_flow
                        FROM transactions t
                        JOIN connected_accounts a ON t.account_id = a.id
                        WHERE a.user_id = $1 AND date > NOW() - INTERVAL '90 days'
                        GROUP BY date
                    ) s
                ) b WHERE bal < 200
            """, user_id)
            lbc = int(low_bal_count or 0)
            if lbc > 10: lbc_score = 0
            elif lbc > 3: lbc_score = 50
            else: lbc_score = 100

            # Weighted average
            risk_score = (
                bal_score * 0.35 +
                vel_score * 0.25 +
                cred_score * 0.20 +
                prox_score * 0.10 +
                lbc_score * 0.10
            )
            
            level = "TRANQUILO"
            if risk_score < 20: level = "CRITICO"
            elif risk_score < 45: level = "PRESSAO"
            elif risk_score < 70: level = "ATENCAO"
            
            is_active = risk_score < 20

            # Recommendations if critical/pressure
            recommendations = []
            daily_limit = 0
            if risk_score < 70:
                survival_data = {
                    "risk_score": risk_score,
                    "level": level,
                    "projected_shortfall": projected_shortfall,
                    "days_until_salary": days_until_salary,
                    "current_balance": curr_bal,
                    "burn_rate": burn
                }
                rec_response = await self.llm.completion(
                    SURVIVAL_RECOMMENDATIONS_PROMPT.format(survival_data_json=json.dumps(survival_data)),
                    system_prompt="Você é um consultor financeiro focado em sobrevivência e resgate de contas."
                )
                try:
                    start_idx = rec_response.find("{")
                    end_idx = rec_response.rfind("}")
                    rec_json = json.loads(rec_response[start_idx:end_idx+1])
                    recommendations = rec_json.get("recommendations", [])
                    daily_limit = rec_json.get("daily_spending_limit", 0)
                except:
                    recommendations = [{"title": "Análise Indisponível", "description": rec_response, "impact": "N/A"}]

            # Save snapshot
            await conn.execute("""
                INSERT INTO survival_mode_snapshots (
                    user_id, risk_score, level, is_active, 
                    projected_shortfall, days_until_salary, top_risks
                ) VALUES ($1, $2, $3, $4, $5, $6, $7)
            """, user_id, risk_score, level, is_active, 
               projected_shortfall, days_until_salary, json.dumps(recommendations))

            return {
                "risk_score": round(risk_score, 2),
                "level": level,
                "is_active": is_active,
                "projected_shortfall": round(projected_shortfall, 2),
                "days_until_salary": days_until_salary,
                "recommendations": recommendations,
                "daily_limit": daily_limit
            }
        finally:
            await conn.close()
