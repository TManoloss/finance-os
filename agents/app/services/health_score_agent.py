from app.services.base_agent import BaseAgent
from datetime import datetime, timedelta
import json
import numpy as np

class HealthScoreAgent(BaseAgent):
    async def run(self, user_id: str, period: str = ""):
        score_data = await self.calculate_health_score(user_id)
        recommendations = await self.get_score_recommendations(user_id, score_data)
        
        # Save snapshot
        await self.save_score_snapshot(user_id, score_data)
        
        return {
            "score": score_data["total_score"],
            "dimensions": score_data["dimensions"],
            "recommendations": recommendations
        }

    async def calculate_health_score(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # 1. Cashflow (25%)
            # (Revenue - Spending) / Revenue
            cashflow_data = await conn.fetchrow("""
                SELECT 
                    SUM(CASE WHEN direction = 'credit' THEN amount ELSE 0 END) as income,
                    SUM(CASE WHEN direction = 'debit' THEN amount ELSE 0 END) as spent
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 AND date > NOW() - INTERVAL '90 days'
            """, user_id)
            
            income = float(cashflow_data['income'] or 0)
            spent = float(cashflow_data['spent'] or 0)
            
            cashflow_ratio = (income - spent) / income if income > 0 else 0
            cashflow_score = min(max(cashflow_ratio * 300, 0), 100) # 30% saving = 100 pts

            # 2. Installments Control (20%)
            # installments / income
            installments_sum = await conn.fetchval("""
                SELECT SUM(total_amount / installments_total)
                FROM installments i
                JOIN connected_accounts a ON i.account_id = a.id
                WHERE a.user_id = $1
            """, user_id)
            inst_sum = float(installments_sum or 0)
            inst_ratio = inst_sum / (income / 3) if income > 0 else 1
            inst_score = max(0, 100 - (inst_ratio * 200)) # 50% income committed = 0 pts

            # 3. Diversification (10%)
            # Use HHI index for categories
            cat_shares = await conn.fetch("""
                SELECT category_id, SUM(amount) as total
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 AND direction = 'debit' AND date > NOW() - INTERVAL '30 days'
                GROUP BY category_id
            """, user_id)
            
            total_spent_month = sum(float(c['total']) for c in cat_shares)
            hhi = sum((float(c['total'])/total_spent_month)**2 for c in cat_shares) if total_spent_month > 0 else 1
            div_score = (1 - hhi) * 100

            # Dimensões (simplificadas para o MVP do score)
            dimensions = {
                "cashflow": cashflow_score,
                "installments": inst_score,
                "diversification": div_score,
                "consistency": 75.0, # Mocked for now
                "subscriptions": 80.0, # Mocked for now
                "trend": 70.0 # Mocked for now
            }
            
            total_score = (
                dimensions["cashflow"] * 0.25 +
                dimensions["installments"] * 0.20 +
                dimensions["consistency"] * 0.20 +
                dimensions["subscriptions"] * 0.15 +
                dimensions["diversification"] * 0.10 +
                dimensions["trend"] * 0.10
            )

            return {
                "total_score": round(total_score, 2),
                "dimensions": dimensions
            }
        finally:
            await conn.close()

    async def get_score_recommendations(self, user_id: str, score_data: dict):
        # Find 2 worst dimensions
        worst = sorted(score_data["dimensions"].items(), key=lambda x: x[1])[:2]
        
        prompt = f"""
        O usuário tem um Health Score financeiro de {score_data['total_score']}/100.
        As piores dimensões são: {worst[0][0]} ({worst[0][1]:.1f}) e {worst[1][0]} ({worst[1][1]:.1f}).
        
        Gere 2 recomendações práticas e curtas em português para melhorar esses pontos específicos.
        """
        response = await self.llm.completion(prompt)
        return response

    async def save_score_snapshot(self, user_id: str, data: dict):
        conn = await self.get_db_connection()
        try:
            await conn.execute("""
                INSERT INTO health_score_snapshots (
                    user_id, score, cashflow_score, installments_score, 
                    consistency_score, subscriptions_score, diversification_score, 
                    trend_score, period_month
                ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
                ON CONFLICT (user_id, period_month) DO UPDATE SET
                    score = EXCLUDED.score,
                    created_at = NOW()
            """, user_id, data["total_score"], 
               data["dimensions"]["cashflow"], data["dimensions"]["installments"],
               data["dimensions"]["consistency"], data["dimensions"]["subscriptions"],
               data["dimensions"]["diversification"], data["dimensions"]["trend"],
               datetime.now().date().replace(day=1))
        finally:
            await conn.close()
