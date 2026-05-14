from app.services.base_agent import BaseAgent
from datetime import datetime, timedelta
import json
import numpy as np

class ComparisonAgent(BaseAgent):
    async def run(self, user_id: str, period_a: dict = None, period_b: dict = None):
        # Default periods if not provided (not matching BaseAgent signature exactly but let's see)
        # BaseAgent.run(user_id, period)
        if period_a is None:
            # dummy
            return {"error": "periods are required"}
        
        comp = await self.compare_periods(user_id, period_a, period_b)
        anomalies = await self.detect_spending_anomalies(user_id)
        
        return {
            "comparison": comp,
            "anomalies": anomalies
        }

    async def compare_periods(self, user_id: str, pA: dict, pB: dict):
        conn = await self.get_db_connection()
        try:
            query = """
                SELECT 
                    c.name as category,
                    SUM(CASE WHEN t.date BETWEEN $2 AND $3 THEN t.amount ELSE 0 END) as total_a,
                    SUM(CASE WHEN t.date BETWEEN $4 AND $5 THEN t.amount ELSE 0 END) as total_b
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                LEFT JOIN categories c ON t.category_id = c.id
                WHERE a.user_id = $1 AND direction = 'debit'
                GROUP BY c.name
            """
            rows = await conn.fetch(query, user_id, 
                                   datetime.fromisoformat(pA['start']), datetime.fromisoformat(pA['end']),
                                   datetime.fromisoformat(pB['start']), datetime.fromisoformat(pB['end']))
            
            items = []
            total_a = 0
            total_b = 0
            
            for r in rows:
                ta = float(r['total_a'])
                tb = float(r['total_b'])
                total_a += ta
                total_b += tb
                
                diff = tb - ta
                diff_pct = (tb / ta - 1) * 100 if ta > 0 else 0
                
                items.append({
                    "category": r['category'] or "Sem Categoria",
                    "total_a": ta,
                    "total_b": tb,
                    "diff": diff,
                    "diff_pct": diff_pct
                })
            
            return {
                "categories": sorted(items, key=lambda x: abs(x['diff']), reverse=True),
                "total_a": total_a,
                "total_b": total_b,
                "total_diff": total_b - total_a,
                "total_diff_pct": (total_b / total_a - 1) * 100 if total_a > 0 else 0
            }
        finally:
            await conn.close()

    async def detect_spending_anomalies(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # Buscar transações dos últimos 6 meses para calcular média e desvio padrão por merchant
            rows = await conn.fetch("""
                SELECT merchant_name, amount, date
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 AND direction = 'debit'
                  AND date > NOW() - INTERVAL '180 days'
                  AND merchant_name IS NOT NULL
            """, user_id)
            
            if not rows:
                return []

            merchants = {}
            for r in rows:
                m = r['merchant_name']
                if m not in merchants: merchants[m] = []
                merchants[m].append(float(r['amount']))
            
            anomalies = []
            for m, amounts in merchants.items():
                if len(amounts) < 3: continue
                
                mean = np.mean(amounts)
                std = np.std(amounts)
                
                # Se STD é zero, ignorar (gastos fixos)
                if std < 1: continue
                
                last_amount = amounts[-1]
                z_score = (last_amount - mean) / std
                
                if z_score > 2: # 2 desvios padrão acima da média
                    anomalies.append({
                        "merchant": m,
                        "last_amount": last_amount,
                        "avg_amount": mean,
                        "z_score": z_score,
                        "severity": "high" if z_score > 3 else "medium"
                    })
            
            return sorted(anomalies, key=lambda x: x['z_score'], reverse=True)
        finally:
            await conn.close()
