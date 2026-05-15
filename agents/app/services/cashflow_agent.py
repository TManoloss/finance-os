from app.services.base_agent import BaseAgent
from datetime import datetime, timedelta
import json

class CashflowAgent(BaseAgent):
    async def run(self, user_id: str, period: str):
        # Implementation of the abstract method if needed, 
        # though this agent is mostly for specific calculations
        pass

    async def build_daily_cashflow(self, user_id: str, from_date: datetime, to_date: datetime):
        conn = await self.get_db_connection()
        try:
            # 1. Get current balance and last sync date to calculate backwards
            # Or simpler: just use all transactions to build the history
            
            # Get initial balance (current sum of all accounts for the user)
            current_balance_row = await conn.fetchrow("""
                SELECT SUM(balance) as total_balance 
                FROM connected_accounts 
                WHERE user_id = $1
            """, user_id)
            current_balance = float(current_balance_row['total_balance'] or 0)

            # Get all transactions from from_date to now to calculate historical balances
            # We need transactions AFTER to_date as well if we want to be precise from "current" balance
            # But usually we want to show a specific range.
            
            # Let's get all transactions to reconstruct history correctly
            transactions = await conn.fetch("""
                SELECT t.date, t.amount, t.direction, t.merchant_name, t.description
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1
                ORDER BY t.date DESC, t.created_at DESC
            """, user_id)

            # Reconstruct daily balances
            # We start from current_balance and go backwards
            daily_data = {}
            temp_balance = current_balance
            
            # Map of transactions by date
            tx_by_date = {}
            for tx in transactions:
                d = tx['date']
                if d not in tx_by_date:
                    tx_by_date[d] = []
                tx_by_date[d].append(tx)

            # Iterate backwards from today to from_date
            today = datetime.now().date()
            curr_date = today
            running_balance = current_balance
            
            # Ensure from_date and to_date are date objects for comparison
            from_date_obj = from_date.date() if isinstance(from_date, datetime) else from_date
            to_date_obj = to_date.date() if isinstance(to_date, datetime) else to_date

            results = []
            # We go back in time
            while curr_date >= from_date_obj:
                day_txs = tx_by_date.get(curr_date, [])
                
                day_income = sum(float(tx['amount']) for tx in day_txs if tx['direction'] == 'credit')
                day_outcome = sum(float(tx['amount']) for tx in day_txs if tx['direction'] == 'debit')
                
                # Balance at START of day = Balance at END of day - income + outcome
                balance_end = running_balance
                balance_start = balance_end - day_income + day_outcome
                
                if curr_date <= to_date_obj:
                    # Find biggest spending
                    biggest_spend = None
                    spend_txs = [tx for tx in day_txs if tx['direction'] == 'debit']
                    if spend_txs:
                        biggest = max(spend_txs, key=lambda x: float(x['amount']))
                        biggest_spend = {
                            "merchant": biggest['merchant_name'] or biggest['description'],
                            "amount": float(biggest['amount'])
                        }

                    results.append({
                        "date": curr_date.isoformat(),
                        "balance_start": balance_start,
                        "balance_end": balance_end,
                        "income": day_income,
                        "outcome": day_outcome,
                        "biggest_spending": biggest_spend,
                        "is_critical": balance_end < 500 # Threshold from GEMINI.md
                    })
                
                running_balance = balance_start
                curr_date -= timedelta(days=1)

            return sorted(results, key=lambda x: x['date'])

        finally:
            await conn.close()

    async def detect_cashflow_patterns(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # 1. Salary Day Detection
            # Biggest recurring credit in the last 3 months
            salary_row = await conn.fetchrow("""
                SELECT EXTRACT(DAY FROM date) as day, AVG(amount) as avg_amount
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 
                  AND direction = 'credit'
                  AND amount > 1000
                  AND date > NOW() - INTERVAL '90 days'
                GROUP BY day
                ORDER BY COUNT(*) DESC, avg_amount DESC
                LIMIT 1
            """, user_id)
            
            salary_day = int(salary_row['day']) if salary_row else None

            # 2. Peak Spending Days
            peak_days = await conn.fetch("""
                SELECT TO_CHAR(date, 'Day') as weekday, AVG(amount) as avg_spent
                FROM transactions t
                JOIN connected_accounts a ON t.account_id = a.id
                WHERE a.user_id = $1 AND direction = 'debit'
                GROUP BY weekday
                ORDER BY avg_spent DESC
            """, user_id)

            # 3. Low Balance Days
            # This would require building history, let's simplify for now
            
            return {
                "salary_day": salary_day,
                "peak_spending_days": [{"day": r['weekday'].strip(), "avg": float(r['avg_spent'])} for r in peak_days[:3]],
                "monthly_cashflow_cycle": f"Seu maior gasto costuma ser às {peak_days[0]['weekday'].strip()}." if peak_days else ""
            }
        finally:
            await conn.close()
