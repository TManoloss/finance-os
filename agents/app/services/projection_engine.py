from app.services.base_agent import BaseAgent
from datetime import datetime, timedelta
import json
import numpy as np

class ProjectionEngine(BaseAgent):
    async def run(self, user_id: str, period: str = ""):
        end_of_month = await self.project_end_of_month(user_id)
        next_3_months = await self.project_next_3_months(user_id)
        risks = await self.detect_financial_risks(user_id, end_of_month, next_3_months)
        
        return {
            "end_of_month_projection": end_of_month,
            "next_3_months_projections": next_3_months,
            "risks_and_opportunities": risks
        }

    async def project_end_of_month(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # 1. Saldo atual
            current_balance_row = await conn.fetchrow("""
                SELECT SUM(balance) as total_balance 
                FROM connected_accounts 
                WHERE user_id = $1
            """, user_id)
            balance = float(current_balance_row['total_balance'] or 0)

            # 2. Compromissos fixos restantes (Parcelas e Assinaturas)
            # Simplificado: busca o que costuma cair depois do dia de hoje no mês
            today_day = datetime.now().day
            
            commitments_row = await conn.fetchrow("""
                SELECT SUM(amount) as total
                FROM (
                    -- Parcelas
                    SELECT SUM(amount) as amount FROM transactions t
                    JOIN connected_accounts a ON t.account_id = a.id
                    WHERE a.user_id = $1 AND is_recurring = true AND EXTRACT(DAY FROM date) > $2
                    UNION ALL
                    -- Parcelamentos do cartao (próxima fatura)
                    SELECT SUM(total_amount / installments_total) FROM installments i
                    WHERE account_id IN (SELECT id FROM connected_accounts WHERE user_id = $1)
                ) t
            """, user_id, today_day)
            fixed_remaining = float(commitments_row['total'] or 0)

            # 3. Gastos variáveis estimados
            # Média diária de gastos variáveis * dias restantes
            days_in_month = 30 # simplificado
            days_left = days_in_month - today_day
            
            avg_var_row = await conn.fetchrow("""
                SELECT AVG(daily_sum) as avg_daily
                FROM (
                    SELECT date, SUM(amount) as daily_sum
                    FROM transactions t
                    JOIN connected_accounts a ON t.account_id = a.id
                    WHERE a.user_id = $1 AND direction = 'debit' AND is_recurring = false
                      AND date > NOW() - INTERVAL '60 days'
                    GROUP BY date
                ) t
            """, user_id)
            avg_daily = float(avg_var_row['avg_daily'] or 0)
            estimated_variable = avg_daily * days_left

            # 4. Receitas esperadas
            # Salário detectado se ainda não caiu este mês
            expected_income = 0
            salary_row = await conn.fetchrow("""
                SELECT amount FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 AND direction = 'credit' 
                  AND date > NOW() - INTERVAL '45 days'
                ORDER BY amount DESC LIMIT 1
            """, user_id)
            
            if salary_row:
                # Verificar se ja caiu este mes
                already_received = await conn.fetchval("""
                    SELECT EXISTS(
                        SELECT 1 FROM transactions t
                        JOIN connected_accounts a ON t.account_id = a.id
                        WHERE a.user_id = $1 AND direction = 'credit' 
                          AND amount >= $2 * 0.9
                          AND EXTRACT(MONTH FROM date) = EXTRACT(MONTH FROM NOW())
                    )
                """, user_id, float(salary_row['amount']))
                if not already_received:
                    expected_income = float(salary_row['amount'])

            projected_final = balance + expected_income - fixed_remaining - estimated_variable

            return {
                "current_balance": balance,
                "expected_income": expected_income,
                "fixed_commitments_remaining": fixed_remaining,
                "estimated_variable_spending": estimated_variable,
                "projected_balance_end_of_month": projected_final,
                "status": "healthy" if projected_final > 0 else "at_risk"
            }
        finally:
            await conn.close()

    async def project_next_3_months(self, user_id: str):
        # Implementação simplificada baseada na média mensal
        conn = await self.get_db_connection()
        try:
            avg_monthly_spent = await conn.fetchval("""
                SELECT AVG(monthly_sum) FROM (
                    SELECT EXTRACT(MONTH FROM date) as month, SUM(amount) as monthly_sum
                    FROM transactions t
                    JOIN connected_accounts a ON t.account_id = a.id
                    WHERE a.user_id = $1 AND direction = 'debit'
                    GROUP BY month
                ) t
            """, user_id)
            avg_spent = float(avg_monthly_spent or 0)

            projections = []
            now = datetime.now()
            for i in range(1, 4):
                future_month = now + timedelta(days=30*i)
                projections.append({
                    "month": future_month.strftime("%B"),
                    "estimated_spending": avg_spent,
                    "confidence": 0.8 - (i * 0.1) # Confidence decreases as we look further
                })
            return projections
        finally:
            await conn.close()

    async def detect_financial_risks(self, user_id: str, eom: dict, next_3: list):
        risks = []
        if eom['projected_balance_end_of_month'] < 0:
            risks.append({
                "type": "negative_balance",
                "severity": "high",
                "message": "Seu saldo projetado para o fim do mês é negativo. Considere reduzir gastos variáveis imediatamente."
            })
        
        # Oportunidade: Sobra > 20% da renda
        if eom['expected_income'] > 0 and eom['projected_balance_end_of_month'] > eom['expected_income'] * 0.2:
            risks.append({
                "type": "savings_opportunity",
                "severity": "info",
                "message": "Você terá uma sobra considerável este mês. Ótima oportunidade para investir ou criar sua reserva de emergência."
            })

        return risks
