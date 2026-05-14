from app.services.base_agent import BaseAgent
from datetime import datetime, timedelta
import json

class SeasonalityAgent(BaseAgent):
    async def run(self, user_id: str, period: str = ""):
        upcoming = await self.predict_upcoming_large_expenses(user_id)
        annual_patterns = await self.detect_annual_patterns(user_id)
        
        return {
            "upcoming_expenses": upcoming,
            "annual_patterns": annual_patterns
        }

    async def detect_annual_patterns(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # Compara meses de anos diferentes para encontrar picos sazonais
            # Requer histórico longo
            rows = await conn.fetch("""
                SELECT 
                    month,
                    AVG(monthly_sum) as avg_spent
                FROM (
                    SELECT 
                        EXTRACT(YEAR FROM date) as year,
                        EXTRACT(MONTH FROM date) as month,
                        SUM(amount) as monthly_sum
                    FROM transactions t
                    JOIN connected_accounts a ON t.account_id = a.id
                    WHERE a.user_id = $1 AND direction = 'debit'
                    GROUP BY year, month
                ) t
                GROUP BY month
                ORDER BY month ASC
            """, user_id)
            
            return [{"month": int(r['month']), "avg_spent": float(r['avg_spent'])} for r in rows]
        finally:
            await conn.close()

    async def predict_upcoming_large_expenses(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # 1. Parcelamentos ativos (dados reais)
            installments = await conn.fetch("""
                SELECT merchant_name, (total_amount / installments_total) as amount, next_due_date
                FROM installments i
                JOIN connected_accounts a ON i.account_id = a.id
                WHERE a.user_id = $1 AND next_due_date > NOW()
                ORDER BY next_due_date ASC
            """, user_id)

            # 2. Despesas detectadas como anuais (IPVA, IPTU, Seguros)
            annual_hints = await conn.fetch("""
                SELECT merchant_name, amount, date
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 AND direction = 'debit'
                  AND (merchant_name ILIKE '%IPVA%' OR merchant_name ILIKE '%IPTU%' OR merchant_name ILIKE '%SEGURO%')
                  AND date > NOW() - INTERVAL '400 days'
            """, user_id)

            results = []
            for inst in installments:
                results.append({
                    "type": "installment",
                    "merchant": inst['merchant_name'],
                    "amount": float(inst['amount']),
                    "due_date": inst['next_due_date'].isoformat(),
                    "confidence": 1.0
                })
            
            for hint in annual_hints:
                # Projeta para o próximo ano
                next_date = hint['date'] + timedelta(days=365)
                if datetime.now().date() < next_date:
                    results.append({
                        "type": "seasonal_predicted",
                        "merchant": hint['merchant_name'],
                        "amount": float(hint['amount']),
                        "due_date": next_date.isoformat(),
                        "confidence": 0.7
                    })

            return sorted(results, key=lambda x: x['due_date'])
        finally:
            await conn.close()
