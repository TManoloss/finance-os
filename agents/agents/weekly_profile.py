import json
from datetime import date, datetime, timedelta
import collections
import statistics
from app.services.base_agent import BaseAgent

WEEKDAY_WEEKEND_PROMPT = """
Você é um analista financeiro sênior especializado em comportamento temporal de gastos.
Analise os dados comparativos entre dias úteis (seg-sex) e fins de semana (sab-dom) abaixo e gere 3 insights interessantes em português.

Dados:
{data_json}

Foque em padrões inesperados, como o "prêmio de fim de semana" ou comportamentos compensatórios.
Seja direto e use um tom analítico mas acessível.
"""

DAY_PROFILE_PROMPT = """
Você é um analista financeiro sênior. Baseado nos perfis dos 7 dias da semana abaixo, gere um adjetivo curto (uma palavra) para cada dia e uma narrativa da "semana financeira típica" do usuário.

Perfis:
{profiles_json}

Retorne obrigatoriamente um JSON com:
- day_adjectives: {{ "0": "...", "1": "...", ..., "6": "..." }} (0=segunda, 6=domingo)
- weekly_narrative: "um parágrafo em português descrevendo a semana típica"
"""

class WeeklyProfileAgent(BaseAgent):
    async def run(self, user_id: str):
        # Este método run executa a análise completa e salva nos dois caches
        report = await self.compare_weekday_vs_weekend(user_id)
        profiles = await self.build_day_profiles(user_id)
        
        # O resultado final combina tudo
        result = {
            "weekday_weekend": report,
            "day_profiles": profiles
        }
        return result

    async def compare_weekday_vs_weekend(self, user_id: str, months: int = 3):
        conn = await self.get_db_connection()
        try:
            start_date = date.today() - timedelta(days=months * 30)
            
            rows = await conn.fetch("""
                SELECT 
                    EXTRACT(DOW FROM t.date) as dow,
                    EXTRACT(HOUR FROM t.created_at) as hour,
                    t.amount,
                    c.name as category,
                    t.merchant_name
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                LEFT JOIN categories c ON t.category_id = c.id
                WHERE acc.user_id = $1 AND t.date >= $2 AND t.direction = 'debit'
            """, user_id, start_date)

            if not rows:
                return None

            weekday_data = [r for r in rows if r['dow'] in [1, 2, 3, 4, 5]] # 1=Mon, 5=Fri
            weekend_data = [r for r in rows if r['dow'] in [0, 6]] # 0=Sun, 6=Sat
            
            # Helper to calculate stats
            def calc_stats(data, num_days):
                if not data:
                    return {
                        "gasto_medio_por_dia": 0,
                        "ticket_medio": 0,
                        "top_categorias": [],
                        "top_merchants": [],
                        "horario_pico": "N/A"
                    }
                
                total_spent = sum(r['amount'] for r in data)
                
                # Categories
                cat_counts = collections.Counter(r['category'] for r in data if r['category'])
                top_cats = [{"name": name, "total": float(sum(r['amount'] for r in data if r['category'] == name))} 
                            for name, _ in cat_counts.most_common(5)]
                
                # Merchants
                merch_counts = collections.Counter(r['merchant_name'] for r in data if r['merchant_name'])
                top_merchs = [name for name, _ in merch_counts.most_common(5)]
                
                # Peak hour
                hour_counts = collections.Counter(r['hour'] for r in data)
                peak_hour = hour_counts.most_common(1)[0][0] if hour_counts else "N/A"
                
                return {
                    "gasto_medio_por_dia": float(total_spent / num_days),
                    "ticket_medio": float(total_spent / len(data)),
                    "top_categorias": top_cats,
                    "top_merchants": top_merchs,
                    "horario_pico": f"{int(peak_hour)}h" if peak_hour != "N/A" else "N/A"
                }

            # Estimate number of weekdays and weekends in the period
            num_days_total = (date.today() - start_date).days
            num_weekends = (num_days_total // 7) * 2
            num_weekdays = num_days_total - num_weekends
            
            weekday_stats = calc_stats(weekday_data, max(1, num_weekdays))
            weekend_stats = calc_stats(weekend_data, max(1, num_weekends))
            
            # Premium
            premium_abs = weekend_stats["gasto_medio_por_dia"] - weekday_stats["gasto_medio_por_dia"]
            premium_pct = (premium_abs / weekday_stats["gasto_medio_por_dia"] * 100) if weekday_stats["gasto_medio_por_dia"] > 0 else 0
            
            report = {
                "weekday": weekday_stats,
                "weekend": weekend_stats,
                "premium": {
                    "premium_absoluto": float(premium_abs),
                    "premium_percentual": float(premium_pct),
                    "premium_mensal_estimado": float(premium_abs * 8),
                    "premium_anual": float(premium_abs * 8 * 12)
                }
            }
            
            # Generate insights
            prompt = f"Dados comparativos:\n{json.dumps(report, default=str)}"
            insights = await self.llm.completion(prompt, system_prompt=WEEKDAY_WEEKEND_PROMPT)
            report["insights"] = insights
            
            # Save to report_cache
            expires_at = datetime.now() + timedelta(days=7)
            await conn.execute("""
                INSERT INTO report_cache (user_id, report_type, period_key, result_json, expires_at)
                VALUES ($1, $2, $3, $4, $5)
                ON CONFLICT (user_id, report_type, period_key) DO UPDATE
                SET result_json = EXCLUDED.result_json,
                    expires_at = EXCLUDED.expires_at,
                    computed_at = NOW()
            """, user_id, "weekday_weekend", date.today().strftime("%Y-%m-%W"), json.dumps(report), expires_at)
            
            return report

        finally:
            await conn.close()

    async def build_day_profiles(self, user_id: str, months: int = 3):
        conn = await self.get_db_connection()
        try:
            start_date = date.today() - timedelta(days=months * 30)
            
            rows = await conn.fetch("""
                SELECT 
                    EXTRACT(DOW FROM t.date) as dow,
                    EXTRACT(HOUR FROM t.created_at) as hour,
                    t.amount,
                    c.name as category,
                    t.merchant_name,
                    t.date
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                LEFT JOIN categories c ON t.category_id = c.id
                WHERE acc.user_id = $1 AND t.date >= $2 AND t.direction = 'debit'
            """, user_id, start_date)

            if not rows:
                return []

            profiles = []
            for dow in range(7):
                day_rows = [r for r in rows if int(r['dow']) == dow]
                
                # Group by date to find zero-gasto days
                dates_with_spending = set(r['date'] for r in day_rows)
                
                # Total occurrences of this DOW in the period
                curr = start_date
                total_dow_count = 0
                while curr <= date.today():
                    if curr.weekday() == (dow - 1) % 7: # Python weekday() 0=Mon, 6=Sun. DOW 0=Sun, 1=Mon.
                        total_dow_count += 1
                    curr += timedelta(days=1)
                
                prob_zero = (1 - (len(dates_with_spending) / total_dow_count)) * 100 if total_dow_count > 0 else 0
                
                # Stats
                amounts = [float(r['amount']) for r in day_rows]
                avg_spent = statistics.mean(amounts) if amounts else 0
                median_spent = statistics.median(amounts) if amounts else 0
                std_dev = statistics.stdev(amounts) if len(amounts) > 1 else 0
                
                # Categories
                cat_counts = collections.Counter(r['category'] for r in day_rows if r['category'])
                top_cats = [{"name": name, "pct": float(sum(r['amount'] for r in day_rows if r['category'] == name) / sum(amounts) * 100) if amounts else 0} 
                            for name, _ in cat_counts.most_common(3)]
                
                # Merchants
                merch_counts = collections.Counter(r['merchant_name'] for r in day_rows if r['merchant_name'])
                top_merchs = [name for name, _ in merch_counts.most_common(3)]
                
                # Hour
                hour_counts = collections.Counter(r['hour'] for r in day_rows)
                peak_hour = hour_counts.most_common(1)[0][0] if hour_counts else None
                
                def get_hour_label(h):
                    if h is None: return "N/A"
                    if 0 <= h < 6: return "madrugada"
                    if 6 <= h < 12: return "manhã"
                    if 12 <= h < 14: return "almoço"
                    if 14 <= h < 18: return "tarde"
                    return "noite"

                profiles.append({
                    "dow": dow,
                    "gasto_medio": float(avg_spent),
                    "gasto_mediano": float(median_spent),
                    "desvio_padrao": float(std_dev),
                    "top_3_categories": top_cats,
                    "top_3_merchants": top_merchs,
                    "horario_pico": get_hour_label(peak_hour),
                    "num_transacoes_medio": float(len(day_rows) / total_dow_count) if total_dow_count > 0 else 0,
                    "probabilidade_gasto_zero": float(prob_zero)
                })

            # LLM for adjectives and narrative
            prompt = f"Perfis dos dias:\n{json.dumps(profiles, default=str)}"
            llm_response = await self.llm.completion(prompt, system_prompt=DAY_PROFILE_PROMPT)
            
            try:
                start_idx = llm_response.find("{")
                end_idx = llm_response.rfind("}")
                clean_json = llm_response[start_idx:end_idx+1]
                llm_data = json.loads(clean_json)
                
                for p in profiles:
                    dow_str = str(p["dow"])
                    p["adjetivo"] = llm_data.get("day_adjectives", {}).get(dow_str, "comum")
                
                narrative = llm_data.get("weekly_narrative", "")
            except:
                narrative = "Não foi possível gerar a narrativa semanal."
                for p in profiles: p["adjetivo"] = "neutro"

            # Save to day_profiles_cache
            for p in profiles:
                await conn.execute("""
                    INSERT INTO day_profiles_cache (
                        user_id, day_of_week, avg_amount, median_amount, 
                        std_dev, top_categories, peak_hour_label, 
                        prob_zero_spending, adjective, last_computed_at
                    )
                    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
                    ON CONFLICT (user_id, day_of_week) DO UPDATE
                    SET avg_amount = EXCLUDED.avg_amount,
                        median_amount = EXCLUDED.median_amount,
                        std_dev = EXCLUDED.std_dev,
                        top_categories = EXCLUDED.top_categories,
                        peak_hour_label = EXCLUDED.peak_hour_label,
                        prob_zero_spending = EXCLUDED.prob_zero_spending,
                        adjective = EXCLUDED.adjective,
                        last_computed_at = NOW()
                """, user_id, p['dow'], p['gasto_medio'], p['gasto_mediano'], 
                     p['desvio_padrao'], json.dumps(p['top_3_categories']), 
                     p['horario_pico'], p['probabilidade_gasto_zero'], p['adjetivo'])

            return {"profiles": profiles, "narrative": narrative}

        finally:
            await conn.close()

    async def analyze_monday_effect(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            start_date = date.today() - timedelta(days=90)
            
            rows = await conn.fetch("""
                SELECT t.date, t.amount, EXTRACT(DOW FROM t.date) as dow
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                WHERE acc.user_id = $1 AND t.date >= $2 AND t.direction = 'debit'
            """, user_id, start_date)
            
            if not rows: return None
            
            mondays = [r for r in rows if r['dow'] == 1]
            other_weekdays = [r for r in rows if r['dow'] in [2, 3, 4, 5]]
            
            avg_monday = statistics.mean([r['amount'] for r in mondays]) if mondays else 0
            avg_others = statistics.mean([r['amount'] for r in other_weekdays]) if other_weekdays else 0
            
            diff_pct = ((avg_monday - avg_others) / avg_others * 100) if avg_others > 0 else 0
            
            return {
                "avg_monday": float(avg_monday),
                "avg_other_weekdays": float(avg_others),
                "monday_diff_percent": float(diff_pct),
                "effect_type": "compensation" if diff_pct > 15 else "containment" if diff_pct < -15 else "neutral"
            }
        finally:
            await conn.close()

    async def detect_day_anomalies(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # Stats for each DOW over last 6 months
            start_stats = date.today() - timedelta(days=180)
            rows_stats = await conn.fetch("""
                SELECT EXTRACT(DOW FROM date) as dow, SUM(amount) as daily_total, date
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                WHERE acc.user_id = $1 AND t.date >= $2 AND t.direction = 'debit'
                GROUP BY date, dow
            """, user_id, start_stats)
            
            if not rows_stats: return []
            
            dow_groups = collections.defaultdict(list)
            for r in rows_stats:
                dow_groups[int(r['dow'])].append(float(r['daily_total']))
            
            dow_metrics = {}
            for dow, values in dow_groups.items():
                if len(values) > 1:
                    dow_metrics[dow] = {
                        "mean": statistics.mean(values),
                        "std": statistics.stdev(values)
                    }
            
            # Check last 30 days
            last_30 = date.today() - timedelta(days=30)
            rows_recent = await conn.fetch("""
                SELECT EXTRACT(DOW FROM date) as dow, SUM(amount) as daily_total, date
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                WHERE acc.user_id = $1 AND t.date >= $2 AND t.direction = 'debit'
                GROUP BY date, dow
                ORDER BY date DESC
            """, user_id, last_30)
            
            anomalies = []
            for r in rows_recent:
                dow = int(r['dow'])
                total = float(r['daily_total'])
                if dow in dow_metrics:
                    m = dow_metrics[dow]
                    if total > m['mean'] + 2 * m['std']:
                        anomalies.append({
                            "date": r['date'].isoformat(),
                            "dow": dow,
                            "amount": total,
                            "expected_mean": float(m['mean']),
                            "severity": "high",
                            "type": "positive_anomaly"
                        })
                    elif total < m['mean'] - 2 * m['std'] and total > 0:
                        anomalies.append({
                            "date": r['date'].isoformat(),
                            "dow": dow,
                            "amount": total,
                            "expected_mean": float(m['mean']),
                            "severity": "medium",
                            "type": "negative_anomaly"
                        })
            
            return anomalies[:5]
        finally:
            await conn.close()

    async def find_reset_day(self, user_id: str):
        # Already somewhat covered in build_day_profiles (probabilidade_gasto_zero)
        profiles = await self.build_day_profiles(user_id)
        if not profiles: return None
        
        profiles_list = profiles["profiles"]
        # Reset day is the one with highest prob_zero or lowest gasto_medio
        reset_day = max(profiles_list, key=lambda x: x['probabilidade_gasto_zero'])
        
        if reset_day['probabilidade_gasto_zero'] > 50:
            return {
                "dia_reset": reset_day['dow'],
                "probabilidade_zero_gasto": reset_day['probabilidade_gasto_zero'],
                "gasto_medio_no_reset_day": reset_day['gasto_medio']
            }
        return None
