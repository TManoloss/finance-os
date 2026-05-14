from app.services.base_agent import BaseAgent
from datetime import datetime, timedelta
import json

class InvisibleSpendingAgent(BaseAgent):
    async def run(self, user_id: str, period: str = ""):
        forgotten = await self.detect_forgotten_subscriptions(user_id)
        duplicates = await self.detect_duplicate_charges(user_id)
        increases = await self.detect_price_increases(user_id)
        
        total_waste_monthly = sum(f['monthly_amount'] for f in forgotten) + \
                             sum(d['amount'] for d in duplicates) + \
                             sum(i['monthly_increase'] for i in increases)

        summary_prompt = f"""
        Analise os seguintes gastos invisíveis encontrados para o usuário:
        Assinaturas possivelmente esquecidas: {json.dumps(forgotten)}
        Cobranças duplicadas: {json.dumps(duplicates)}
        Aumentos de preços detectados: {json.dumps(increases)}
        Total desperdiçado mensalmente: R$ {total_waste_monthly:.2f}

        Gere um parágrafo curto e direto em português alertando o usuário sobre quanto ele está perdendo e onde deve focar para economizar.
        """
        summary = await self.llm.completion(summary_prompt)

        return {
            "forgotten_subscriptions": forgotten,
            "duplicate_charges": duplicates,
            "price_increases": increases,
            "total_monthly_waste": total_waste_monthly,
            "summary": summary
        }

    async def detect_forgotten_subscriptions(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # Busca merchants com cobranças mensais recorrentes (pelo menos 3 meses)
            # e verifica se houve "uso" (outras transações menores ou em datas diferentes)
            rows = await conn.fetch("""
                WITH recurring AS (
                    SELECT merchant_name, amount, COUNT(*) as occurrences, AVG(amount) as avg_amount
                    FROM transactions t
                    JOIN connected_accounts a ON t.account_id = a.id
                    WHERE a.user_id = $1 AND direction = 'debit'
                      AND date > NOW() - INTERVAL '120 days'
                    GROUP BY merchant_name, amount
                    HAVING COUNT(*) >= 3
                )
                SELECT r.merchant_name, r.avg_amount
                FROM recurring r
                LEFT JOIN transactions t ON t.merchant_name = r.merchant_name 
                    AND t.amount != r.avg_amount
                    AND t.date > NOW() - INTERVAL '90 days'
                GROUP BY r.merchant_name, r.avg_amount
                HAVING COUNT(t.id) = 0 -- Se não há outros gastos com esse merchant, pode ser assinatura esquecida
            """, user_id)
            
            return [{
                "merchant": r['merchant_name'],
                "monthly_amount": float(r['avg_amount']),
                "annual_cost": float(r['avg_amount']) * 12,
                "reason": "Cobrança recorrente sem outros pontos de contato detectados."
            } for r in rows]
        finally:
            await conn.close()

    async def detect_duplicate_charges(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # Mesma descrição, mesmo valor, mesmo dia (ou +/- 2 dias)
            rows = await conn.fetch("""
                SELECT t1.merchant_name, t1.amount, t1.date as date1, t2.date as date2
                FROM transactions t1
                JOIN transactions t2 ON t1.merchant_name = t2.merchant_name 
                    AND t1.amount = t2.amount 
                    AND t1.id < t2.id
                    AND t1.date BETWEEN t2.date - INTERVAL '2 days' AND t2.date + INTERVAL '2 days'
                JOIN connected_accounts a ON t1.account_id = a.id
                WHERE a.user_id = $1 AND t1.direction = 'debit'
                  AND t1.date > NOW() - INTERVAL '30 days'
            """, user_id)
            
            return [{
                "merchant": r['merchant_name'],
                "amount": float(r['amount']),
                "dates": [r['date1'].isoformat(), r['date2'].isoformat()],
                "severity": "high" if float(r['amount']) > 50 else "medium"
            } for r in rows]
        finally:
            await conn.close()

    async def detect_price_increases(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # Compara valor da última ocorrência de um merchant vs ocorrências anteriores
            rows = await conn.fetch("""
                WITH history AS (
                    SELECT merchant_name, amount, date,
                           LAG(amount) OVER (PARTITION BY merchant_name ORDER BY date) as prev_amount
                    FROM transactions t
                    JOIN connected_accounts a ON t.account_id = a.id
                    WHERE a.user_id = $1 AND direction = 'debit'
                      AND date > NOW() - INTERVAL '180 days'
                )
                SELECT merchant_name, amount as current_amount, prev_amount, date
                FROM history
                WHERE prev_amount IS NOT NULL AND amount > prev_amount * 1.05 -- Aumento > 5%
                AND date > NOW() - INTERVAL '45 days'
            """, user_id)
            
            return [{
                "merchant": r['merchant_name'],
                "old_amount": float(r['prev_amount']),
                "new_amount": float(r['current_amount']),
                "monthly_increase": float(r['current_amount']) - float(r['prev_amount']),
                "date": r['date'].isoformat()
            } for r in rows]
        finally:
            await conn.close()
