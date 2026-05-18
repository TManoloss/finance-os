import json
import collections
import statistics
from datetime import date, datetime, timedelta
from app.services.base_agent import BaseAgent

IMPULSE_REPORT_PROMPT = """
Você é um analista financeiro sênior especializado em comportamento de consumo.
Analise os dados de compras por impulso abaixo e gere um relatório narrativo em português.

Dados:
{impulse_data_json}

Inclua na sua análise:
1. O impacto real das compras por impulso no orçamento (valor e percentual).
2. Padrões de horário ou categorias que disparam esse comportamento.
3. Uma observação sobre o "proxy de arrependimento" (merchants de impulso que nunca mais viram uma transação).

Tom: Analítico, empático e sem julgamento moral. Foque em conscientização.
Retorne um JSON com:
- summary: "resumo executivo do comportamento"
- categories_at_risk: ["lista", "de", "categorias"]
- primary_triggers: ["padrões detectados, ex: madrugadas, sextas-feiras"]
- recommendation: "uma dica prática para reduzir o impulso sem sacrificar o prazer"
- narrative: "o parágrafo completo de análise"
"""

COMPENSATION_REPORT_PROMPT = """
Você é um analista financeiro especializado em psicologia do consumo.
Analise o padrão de compensação financeira do usuário baseado nos dados semanais abaixo.

Dados:
{compensation_data_json}

Se houver um padrão claro de compensação (semanas de contenção seguidas de semanas de gasto excessivo), descreva o ciclo típico com números reais.
Se não houver padrão, indique que os gastos são independentes.

Diretrizes:
- Use o coeficiente de autocorrelação lag-1 para fundamentar sua análise.
- Mencione eventos de "gasto pós-estresse" se detectados.
- Tom: Analítico e neutro.

Retorne um JSON com:
- pattern_detected: boolean
- pattern_type: "compensation", "momentum" ou "random"
- strength: "baixa", "média" ou "alta"
- narrative: "o parágrafo completo de análise em português"
- stress_insight: "insight específico sobre gastos pós-saldo-baixo, se houver"
"""

class BehavioralNuancesAgent(BaseAgent):
    async def run(self, user_id: str):
        # Default run calls both
        impulse = await self.run_impulse(user_id)
        compensation = await self.run_compensation(user_id)
        return {
            "impulse": impulse,
            "compensation": compensation
        }

    async def run_impulse(self, user_id: str, months: int = 3):
        conn = await self.get_db_connection()
        try:
            start_date = date.today() - timedelta(days=months * 30)
            
            # 1. Buscar histórico completo para linha de base
            all_txs = await conn.fetch("""
                SELECT 
                    t.id, t.amount, t.date, t.created_at, t.merchant_name, 
                    c.name as category, EXTRACT(HOUR FROM t.created_at) as hour,
                    EXTRACT(DOW FROM t.date) as dow
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                LEFT JOIN categories c ON t.category_id = c.id
                WHERE acc.user_id = $1 AND t.direction = 'debit'
                ORDER BY t.date ASC, t.created_at ASC
            """, user_id)

            if not all_txs:
                return {"error": "Sem transações suficientes para análise"}

            # 2. Calcular linha de base por categoria
            cat_stats = collections.defaultdict(lambda: {
                "amounts": [], "hours": collections.Counter(), "dows": collections.Counter(), "merchants": set()
            })
            
            # Usar apenas transações ANTERIORES ao período de análise para a linha de base seria ideal, 
            # mas vamos simplificar usando tudo e depois analisando o período recente.
            for tx in all_txs:
                cat = tx['category'] or "Outros"
                cat_stats[cat]["amounts"].append(float(tx['amount']))
                cat_stats[cat]["hours"][int(tx['hour'])] += 1
                cat_stats[cat]["dows"][int(tx['dow'])] += 1
                cat_stats[cat]["merchants"].add(tx['merchant_name'])

            category_baselines = {}
            for cat, data in cat_stats.items():
                category_baselines[cat] = {
                    "avg_ticket": statistics.mean(data["amounts"]),
                    "common_hours": [h for h, c in data["hours"].most_common(3)],
                    "common_dows": [d for d, c in data["dows"].most_common(3)],
                    "known_merchants": data["merchants"]
                }

            # 3. Analisar transações recentes para pontuação de impulso
            recent_txs = [tx for tx in all_txs if tx['date'] >= start_date]
            if not recent_txs:
                return {"error": "Sem transações recentes para análise"}

            impulse_txs_details = []
            
            for i, tx in enumerate(recent_txs):
                score = 0
                cat = tx['category'] or "Outros"
                baseline = category_baselines.get(cat)
                
                if not baseline:
                    continue

                # - Horário incomum (+0.2)
                if int(tx['hour']) not in baseline["common_hours"]:
                    score += 0.2
                
                # - Merchant novo (+0.2)
                # (Nota: isso é um pouco impreciso se o merchant apareceu apenas no período recente, 
                # mas serve para o propósito)
                prev_txs_same_merchant = [t for t in all_txs if t['merchant_name'] == tx['merchant_name'] and t['created_at'] < tx['created_at']]
                if not prev_txs_same_merchant:
                    score += 0.2
                
                # - Valor atípico (> 2x o ticket médio) (+0.2)
                if float(tx['amount']) > 2 * baseline["avg_ticket"]:
                    score += 0.2
                
                # - Dia incomum (+0.1)
                if int(tx['dow']) not in baseline["common_dows"]:
                    score += 0.1
                
                # - Sequência rápida (< 30 min após outra) (+0.15)
                if i > 0:
                    prev_tx = recent_txs[i-1]
                    diff = tx['created_at'] - prev_tx['created_at']
                    if diff.total_seconds() < 1800: # 30 min
                        score += 0.15
                
                # - Horário de madrugada (22h-06h) (+0.15)
                h = int(tx['hour'])
                if h >= 22 or h < 6:
                    score += 0.15

                impulse_txs_details.append({
                    "id": str(tx['id']),
                    "merchant": tx['merchant_name'],
                    "amount": float(tx['amount']),
                    "date": tx['date'].isoformat(),
                    "category": cat,
                    "score": round(score, 2)
                })

            # 4. Agregações
            impulse_list = [t for t in impulse_txs_details if t['score'] > 0.6]
            planned_list = [t for t in impulse_txs_details if t['score'] < 0.3]
            uncertain_list = [t for t in impulse_txs_details if 0.3 <= t['score'] <= 0.6]
            
            total_spent = sum(t['amount'] for t in impulse_txs_details)
            impulse_total = sum(t['amount'] for t in impulse_list)
            
            # Regret proxy: merchants de impulso (score > 0.6) que nunca mais apareceram
            impulse_merchants = set(t['merchant'] for t in impulse_list)
            regret_merchants = []
            for m in impulse_merchants:
                # Ver se houve alguma transação DEPOIS da transação de impulso
                # Para simplificar: se m apareceu apenas 1 vez em todo o histórico
                m_count = sum(1 for t in all_txs if t['merchant_name'] == m)
                if m_count == 1:
                    regret_merchants.append(m)

            report_data = {
                "impulse_count": len(impulse_list),
                "planned_count": len(planned_list),
                "uncertain_count": len(uncertain_list),
                "impulse_total_amount": float(impulse_total),
                "impulse_percent_of_spending": float(impulse_total / total_spent * 100) if total_spent > 0 else 0,
                "impulse_avg_ticket": float(impulse_total / len(impulse_list)) if impulse_list else 0,
                "planned_avg_ticket": float(sum(t['amount'] for t in planned_list) / len(planned_list)) if planned_list else 0,
                "regret_merchants_count": len(regret_merchants),
                "top_impulse_categories": collections.Counter(t['category'] for t in impulse_list).most_common(3),
                "recent_impulse_transactions": impulse_list[-10:]
            }

            # 5. LLM Insights
            prompt = f"Dados de impulso:\n{json.dumps(report_data, default=str)}"
            llm_response = await self.llm.completion(prompt, system_prompt=IMPULSE_REPORT_PROMPT)
            
            try:
                start_idx = llm_response.find("{")
                end_idx = llm_response.rfind("}")
                llm_json = json.loads(llm_response[start_idx:end_idx+1])
                report_data["insights"] = llm_json
            except:
                report_data["insights"] = {"narrative": llm_response}

            # 6. Salvar no cache
            expires_at = datetime.now() + timedelta(days=7)
            await conn.execute("""
                INSERT INTO report_cache (user_id, report_type, period_key, result_json, expires_at)
                VALUES ($1, $2, $3, $4, $5)
                ON CONFLICT (user_id, report_type, period_key) DO UPDATE
                SET result_json = EXCLUDED.result_json,
                    expires_at = EXCLUDED.expires_at,
                    computed_at = NOW()
            """, user_id, "impulse_analysis", date.today().strftime("%Y-%m-%W"), 
                 json.dumps(report_data), expires_at)

            return report_data
        finally:
            await conn.close()

    async def run_compensation(self, user_id: str, months: int = 4):
        conn = await self.get_db_connection()
        try:
            start_date = date.today() - timedelta(days=months * 30)
            
            # 1. Agrupar gastos por semana
            rows = await conn.fetch("""
                SELECT 
                    date_trunc('week', t.date) as week,
                    SUM(t.amount) as total
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                WHERE acc.user_id = $1 AND t.date >= $2 AND t.direction = 'debit'
                GROUP BY 1
                ORDER BY 1 ASC
            """, user_id, start_date)

            if len(rows) < 4:
                return {"error": "Semanas insuficientes para análise de compensação (mínimo 4)"}

            weekly_totals = [float(r['total']) for r in rows]
            mediana = statistics.median(weekly_totals)
            
            # 2. Autocorrelação lag-1
            # (x_i - mean) * (x_{i+1} - mean)
            mean = statistics.mean(weekly_totals)
            numerator = 0
            denominator = sum((x - mean) ** 2 for x in weekly_totals)
            
            for i in range(len(weekly_totals) - 1):
                numerator += (weekly_totals[i] - mean) * (weekly_totals[i+1] - mean)
            
            autocorr = (numerator / denominator) if denominator > 0 else 0
            
            # 3. Detectar post-stress spending
            # Saldo baixo seguido de gasto alto após salário
            stress_events = await conn.fetch("""
                WITH daily_balances AS (
                    SELECT 
                        t.date,
                        SUM(CASE WHEN t.direction = 'credit' THEN t.amount ELSE -t.amount END) as net_flow
                    FROM transactions t
                    JOIN connected_accounts acc ON t.account_id = acc.id
                    WHERE acc.user_id = $1
                    GROUP BY t.date
                ),
                running_balance AS (
                    -- Esta é uma estimativa simplificada
                    SELECT 
                        date,
                        SUM(net_flow) OVER (ORDER BY date) as est_balance
                    FROM daily_balances
                )
                SELECT t.date, t.amount, t.merchant_name, rb.est_balance
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                JOIN running_balance rb ON t.date = rb.date
                WHERE acc.user_id = $1 
                AND t.direction = 'credit' 
                AND t.amount > 1000 -- Provável salário
                ORDER BY t.date DESC
                LIMIT 5
            """, user_id)
            
            stress_insights = []
            for s in stress_events:
                # Ver gastos nas 48h após esse 'salário'
                sal_date = s['date']
                post_spending = await conn.fetchval("""
                    SELECT SUM(amount)
                    FROM transactions t
                    JOIN connected_accounts acc ON t.account_id = acc.id
                    WHERE acc.user_id = $1 AND direction = 'debit'
                    AND date >= $2 AND date <= $2 + interval '2 days'
                """, user_id, sal_date)
                
                stress_insights.append({
                    "data_salario": sal_date.isoformat(),
                    "valor_salario": float(s['amount']),
                    "gasto_48h_pos": float(post_spending or 0),
                    "impacto_percentual": float(post_spending / s['amount'] * 100) if s['amount'] > 0 else 0
                })

            report_data = {
                "autocorrelation": round(autocorr, 3),
                "weekly_median": float(mediana),
                "num_weeks": len(weekly_totals),
                "compensation_score": "high" if autocorr < -0.3 else "medium" if autocorr < -0.1 else "low",
                "stress_events": stress_insights,
                "weekly_data": [float(r['total']) for r in rows]
            }

            # 4. LLM Narrative
            prompt = f"Dados de compensação:\n{json.dumps(report_data, default=str)}"
            llm_response = await self.llm.completion(prompt, system_prompt=COMPENSATION_REPORT_PROMPT)
            
            try:
                start_idx = llm_response.find("{")
                end_idx = llm_response.rfind("}")
                llm_json = json.loads(llm_response[start_idx:end_idx+1])
                report_data["insights"] = llm_json
            except:
                report_data["insights"] = {"narrative": llm_response}

            # 5. Salvar no cache
            expires_at = datetime.now() + timedelta(days=7)
            await conn.execute("""
                INSERT INTO report_cache (user_id, report_type, period_key, result_json, expires_at)
                VALUES ($1, $2, $3, $4, $5)
                ON CONFLICT (user_id, report_type, period_key) DO UPDATE
                SET result_json = EXCLUDED.result_json,
                    expires_at = EXCLUDED.expires_at,
                    computed_at = NOW()
            """, user_id, "compensation_analysis", date.today().strftime("%Y-%m-%W"), 
                 json.dumps(report_data), expires_at)

            return report_data
        finally:
            await conn.close()
