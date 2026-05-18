import json
from datetime import date, datetime, timedelta
from app.services.base_agent import BaseAgent

PIERRE_SALARY_ADVICE_PROMPT = """
Você é o Pierre, o CFO pessoal do usuário. O salário dele acaba de ser detectado e você preparou um plano de gastos.
Com base nos números abaixo, gere um briefing curto e amigável (máximo 3 frases) em português.
Tom: parceiro financeiro de confiança, direto, encorajador, como um CFO pessoal.

Dados do Plano:
- Salário Detectado: R$ {salary:.2f}
- Compromissos Fixos: R$ {fixed:.2f}
- Reserva Recomendada (10%): R$ {reserve:.2f}
- Orçamento Variável Disponível: R$ {variable:.2f}
- Limite Diário Seguro: R$ {daily_limit:.2f}

Exemplo de tom:
"Seus compromissos fixos consomem R$ 1.840. Com R$ 2.242 disponíveis para o mês, seu limite diário seguro é de R$ 58. Se mantiver esse ritmo, sobrará uma reserva de R$ 340 ao final do mês."
"""

class SalaryPlannerAgent(BaseAgent):
    async def run(self, user_id: str, period: str = ""):
        """
        Default run method to satisfy BaseAgent abstract method.
        """
        return await self.generate_salary_plan(user_id)

    async def generate_salary_plan(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # 1. Detect Salary (Most recent credit > 1000 in the last 31 days)
            salary_row = await conn.fetchrow("""
                SELECT amount, date
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 
                  AND direction = 'credit' 
                  AND amount > 1000
                  AND date > NOW() - INTERVAL '31 days'
                ORDER BY date DESC LIMIT 1
            """, user_id)

            if not salary_row:
                # Se não detectou nos últimos 31 dias, tenta pegar o último histórico para ter uma base
                salary_row = await conn.fetchrow("""
                    SELECT amount, date
                    FROM transactions t
                    JOIN connected_accounts a ON t.account_id = a.id
                    WHERE a.user_id = $1 
                      AND direction = 'credit' 
                      AND amount > 1000
                    ORDER BY date DESC LIMIT 1
                """, user_id)

            if not salary_row:
                return {"error": "No salary detected for user"}

            salary_amount = float(salary_row['amount'])
            salary_date = salary_row['date']

            # 2. Fixed Commitments
            # a) Installments for the next 30 days
            installments_val = await conn.fetchval("""
                SELECT SUM(total_amount / installments_total)
                FROM installments i
                JOIN connected_accounts a ON i.account_id = a.id
                WHERE a.user_id = $1
            """, user_id)
            fixed_installments = float(installments_val or 0)

            # b) Subscriptions and fixed categories
            # Look for categories like 'Moradia', 'Educação', 'Assinaturas'
            # We average them over 3 months or take the last month if it's more representative
            fixed_costs_val = await conn.fetchval("""
                SELECT SUM(amount) / 3.0
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                JOIN categories c ON t.category_id = c.id
                WHERE a.user_id = $1 
                  AND direction = 'debit'
                  AND c.name IN ('Moradia', 'Educação', 'Assinaturas', 'Saúde')
                  AND t.date > NOW() - INTERVAL '90 days'
            """, user_id)
            fixed_other = float(fixed_costs_val or 0)

            total_fixed = fixed_installments + fixed_other

            # 3. Variable Budget & Calculations
            recommended_reserve = salary_amount * 0.10
            available_after_fixed = salary_amount - total_fixed
            
            # If available is less than reserve, adjust
            if available_after_fixed < recommended_reserve:
                recommended_reserve = max(0, available_after_fixed * 0.5)
            
            variable_budget = available_after_fixed - recommended_reserve
            
            # Days until next salary (approx 30 days from salary_date)
            # If today is salary_date, it's 30.
            # If today is 5 days after, it's 25.
            days_passed = (date.today() - salary_date).days
            days_remaining = max(1, 30 - days_passed)
            
            safe_daily_limit = variable_budget / days_remaining if variable_budget > 0 else 0

            # 4. LLM Pierre's Advice
            plan_data = {
                "salary": salary_amount,
                "fixed": total_fixed,
                "reserve": recommended_reserve,
                "variable": variable_budget,
                "daily_limit": safe_daily_limit
            }

            advice = await self.llm.completion(
                PIERRE_SALARY_ADVICE_PROMPT.format(**plan_data),
                system_prompt="Você é o Pierre, um CFO pessoal amigável e direto."
            )

            # 5. Save to salary_plans
            valid_until = salary_date + timedelta(days=30)
            
            await conn.execute("""
                INSERT INTO salary_plans (
                    user_id, salary_detected, fixed_commitments, 
                    safe_daily_limit, plan_data, valid_until
                ) VALUES ($1, $2, $3, $4, $5, $6)
            """, user_id, salary_amount, total_fixed, safe_daily_limit, json.dumps({
                "recommended_reserve": recommended_reserve,
                "variable_budget": variable_budget,
                "days_remaining": days_remaining,
                "pierre_advice": advice.strip(),
                "fixed_breakdown": {
                    "installments": fixed_installments,
                    "fixed_categories": fixed_other
                }
            }), valid_until)

            return {
                "user_id": user_id,
                "salary_detected": salary_amount,
                "fixed_commitments": total_fixed,
                "safe_daily_limit": safe_daily_limit,
                "recommended_reserve": recommended_reserve,
                "variable_budget": variable_budget,
                "days_remaining": days_remaining,
                "pierre_advice": advice.strip()
            }
        finally:
            await conn.close()
