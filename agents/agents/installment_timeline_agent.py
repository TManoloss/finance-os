import json
from datetime import date, timedelta
from dateutil.relativedelta import relativedelta
from app.services.base_agent import BaseAgent

INSTALLMENT_TIMELINE_PROMPT = """
Você é um analista financeiro. Com base na projeção de parcelamentos para os próximos 12 meses, gere um resumo narrativo sobre o "Peso dos Parcelamentos" ao longo do tempo.
Destaque o alívio financeiro conforme as parcelas terminam e identifique o "Mês Farol" (último mês com parcelas ativas no horizonte de 12 meses).

Dados da Projeção:
{projection_data_json}

Tom esperado: motivador, focado em liberdade financeira e alívio de caixa. Máximo 3 frases.
"""

class InstallmentTimelineAgent(BaseAgent):
    async def run(self, user_id: str, months: int = 12):
        return await self.generate_installment_timeline(user_id, months)

    async def generate_installment_timeline(self, user_id: str, months: int = 12):
        conn = await self.get_db_connection()
        try:
            today = date.today().replace(day=1)
            period_key = today.strftime("%Y-%m")
            
            # Verificar cache
            cached = await conn.fetchrow("""
                SELECT result_json FROM report_cache 
                WHERE user_id = $1 AND report_type = $2 AND period_key = $3 AND expires_at > NOW()
            """, user_id, "installment_timeline", period_key)
            
            if cached:
                return json.loads(cached['result_json'])

            # Buscar todos os parcelamentos ativos do usuário
            # Consideramos ativos aqueles onde installment_current < installments_total
            rows = await conn.fetch("""
                SELECT 
                    i.id,
                    i.merchant_name,
                    i.total_amount,
                    i.installments_total,
                    i.installment_current,
                    i.next_due_date,
                    (i.total_amount / i.installments_total) as installment_value
                FROM installments i
                JOIN connected_accounts a ON i.account_id = a.id
                WHERE a.user_id = $1 AND i.installment_current < i.installments_total
            """, user_id)

            timeline = []
            
            total_liberated_since_today = 0.0
            last_month_with_installments = None

            for i in range(months):
                target_month = today + relativedelta(months=i)
                month_str = target_month.strftime("%Y-%m")
                
                active_this_month = []
                ending_this_month = []
                total_amount_month = 0.0
                
                for row in rows:
                    next_due = row['next_due_date']
                    months_diff = (target_month.year - next_due.year) * 12 + (target_month.month - next_due.month)
                    
                    current_installment_in_target_month = row['installment_current'] + months_diff
                    
                    if 0 <= months_diff < (row['installments_total'] - row['installment_current']):
                        total_amount_month += float(row['installment_value'])
                        active_this_month.append({
                            "merchant": row['merchant_name'],
                            "value": float(row['installment_value']),
                            "current": current_installment_in_target_month + 1,
                            "total": row['installments_total']
                        })
                        
                        if current_installment_in_target_month + 1 == row['installments_total']:
                            ending_this_month.append({
                                "merchant": row['merchant_name'],
                                "value": float(row['installment_value']),
                                "total_paid": float(row['total_amount'])
                            })
                        
                        last_month_with_installments = month_str

                new_freedom = 0.0
                if i > 0:
                    prev_month_total = timeline[i-1]["total_amount"]
                    if total_amount_month < prev_month_total:
                        new_freedom = prev_month_total - total_amount_month
                
                total_liberated_since_today += new_freedom

                timeline.append({
                    "month": month_str,
                    "active_installments": len(active_this_month),
                    "total_amount": round(total_amount_month, 2),
                    "ending_this_month": ending_this_month,
                    "new_freedom": round(new_freedom, 2),
                    "cumulative_freedom": round(total_liberated_since_today, 2)
                })

            projection_summary = {
                "timeline": timeline[:6],
                "lighthouse_month": last_month_with_installments,
                "total_installments_active": len(rows)
            }
            
            narrative = await self.llm.completion(
                INSTALLMENT_TIMELINE_PROMPT.format(projection_data_json=json.dumps(projection_summary)),
                system_prompt="Você é um assistente financeiro motivador."
            )

            result = {
                "timeline": timeline,
                "lighthouse_month": last_month_with_installments,
                "narrative": narrative.strip()
            }

            # Salvar no report_cache
            expires_at = datetime.now() + timedelta(days=7)
            await conn.execute("""
                INSERT INTO report_cache (user_id, report_type, period_key, result_json, expires_at)
                VALUES ($1, $2, $3, $4, $5)
                ON CONFLICT (user_id, report_type, period_key) DO UPDATE
                SET result_json = EXCLUDED.result_json,
                    expires_at = EXCLUDED.expires_at,
                    computed_at = NOW()
            """, user_id, "installment_timeline", period_key, json.dumps(result), expires_at)

            # Salvar relatório legível (opcional, mas bom manter o padrão)
            await self.save_report(
                user_id, 
                "installment_timeline", 
                today, 
                today + relativedelta(months=months),
                narrative.strip(),
                json.dumps(result)
            )

            return result
        finally:
            await conn.close()
