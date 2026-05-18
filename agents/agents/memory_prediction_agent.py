import json
import collections
from datetime import date, datetime, timedelta
from app.services.base_agent import BaseAgent

FINANCIAL_MEMORY_PROMPT = """
Você é um analista financeiro sênior especializado em memória histórica e padrões anuais.
Analise a comparação entre o mês atual e o mesmo mês do ano anterior para o usuário.

Dados de Comparação:
{comparison_data_json}

Sua tarefa é gerar insights que ajudem o usuário a entender sua evolução financeira de longo prazo.
Considere:
1. Mudanças significativas no volume total de gastos.
2. Alterações nas principais categorias (o que era prioridade ano passado e o que é agora).
3. "Vícios persistentes" ou "vitórias financeiras" (categorias que foram reduzidas).

Tom: Nostálgico, analítico e orientativo.
Retorne um JSON com:
- summary: "resumo da evolução em relação ao ano passado"
- top_wins: ["conquistas detectadas, ex: redução em delivery"]
- persistent_habits: ["hábitos que se mantêm iguais"]
- structural_changes: ["mudanças no perfil de gastos, ex: nova moradia, novos hobbies"]
- narrative: "o parágrafo completo de análise em português"
"""

DANGEROUS_DAYS_PROMPT = """
Você é um assistente financeiro preditivo focado em prevenção de gastos.
Com base nos padrões históricos de "dias perigosos" (momentos de maior gasto), gere alertas preventivos para a próxima semana.

Padrões Detectados:
{patterns_json}

Próximos Alertas:
{upcoming_alerts_json}

Sua tarefa é criar uma mensagem de alerta que seja útil e não intrusiva, ajudando o usuário a ficar atento nos momentos de maior vulnerabilidade financeira.

Tom: Preventivo, amigável e focado em conscientização (nudge).
Retorne um JSON com:
- risk_level: "baixo", "médio" ou "alto" para a próxima semana
- key_alerts: ["lista de alertas específicos para os próximos dias"]
- preventive_tip: "uma dica prática para os momentos de maior risco"
- narrative: "uma breve explicação do porquê esses momentos são considerados perigosos"
"""

class MemoryPredictionAgent(BaseAgent):
    async def run(self, user_id: str):
        memory = await self.generate_memory_insights(user_id)
        dangerous = await self.generate_preventive_alerts(user_id)
        return {
            "financial_memory": memory,
            "dangerous_days": dangerous
        }

    async def get_same_period_last_year(self, user_id: str, reference_date: date = None):
        if reference_date is None:
            reference_date = date.today()
        
        current_month_start = reference_date.replace(day=1)
        last_year_month_start = current_month_start.replace(year=current_month_start.year - 1)
        last_year_month_end = (last_year_month_start + timedelta(days=32)).replace(day=1) - timedelta(days=1)

        conn = await self.get_db_connection()
        try:
            # Gastos do mês atual (até hoje)
            current_spending = await conn.fetch("""
                SELECT COALESCE(SUM(amount), 0) as total, c.name as category
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                LEFT JOIN categories c ON t.category_id = c.id
                WHERE acc.user_id = $1 
                AND t.date >= $2 AND t.date <= $3
                AND t.direction = 'debit'
                GROUP BY c.name
            """, user_id, current_month_start, reference_date)

            # Gastos do mesmo mês no ano passado (período completo)
            last_year_spending = await conn.fetch("""
                SELECT COALESCE(SUM(amount), 0) as total, c.name as category
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                LEFT JOIN categories c ON t.category_id = c.id
                WHERE acc.user_id = $1 
                AND t.date >= $2 AND t.date <= $3
                AND t.direction = 'debit'
                GROUP BY c.name
            """, user_id, last_year_month_start, last_year_month_end)

            return {
                "current_month": {
                    "total": sum(float(r['total']) for r in current_spending),
                    "by_category": {r['category'] or "Outros": float(r['total']) for r in current_spending}
                },
                "last_year_month": {
                    "total": sum(float(r['total']) for r in last_year_spending),
                    "by_category": {r['category'] or "Outros": float(r['total']) for r in last_year_spending}
                }
            }
        finally:
            await conn.close()

    async def generate_memory_insights(self, user_id: str):
        comparison_data = await self.get_same_period_last_year(user_id)
        
        if comparison_data["last_year_month"]["total"] == 0:
            return {"message": "Histórico insuficiente do ano passado para comparação."}

        prompt = f"Dados de Comparação:\n{json.dumps(comparison_data, default=str)}"
        llm_response = await self.llm.completion(prompt, system_prompt=FINANCIAL_MEMORY_PROMPT)
        
        try:
            start_idx = llm_response.find("{")
            end_idx = llm_response.rfind("}")
            result = json.loads(llm_response[start_idx:end_idx+1])
            result["data"] = comparison_data
        except:
            result = {"narrative": llm_response, "data": comparison_data}

        # Salvar no cache
        conn = await self.get_db_connection()
        try:
            expires_at = datetime.now() + timedelta(days=7)
            await conn.execute("""
                INSERT INTO report_cache (user_id, report_type, period_key, result_json, expires_at)
                VALUES ($1, $2, $3, $4, $5)
                ON CONFLICT (user_id, report_type, period_key) DO UPDATE
                SET result_json = EXCLUDED.result_json,
                    expires_at = EXCLUDED.expires_at,
                    computed_at = NOW()
            """, user_id, "financial_memory", date.today().strftime("%Y-%m-%W"), 
                 json.dumps(result), expires_at)
        finally:
            await conn.close()

        return result

    async def identify_dangerous_patterns(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # Analisar transações do último ano para encontrar padrões de dia e hora
            start_date = date.today() - timedelta(days=365)
            
            rows = await conn.fetch("""
                SELECT 
                    EXTRACT(DOW FROM date) as dow,
                    EXTRACT(HOUR FROM created_at) as hour,
                    COALESCE(SUM(amount), 0) as total_amount,
                    COUNT(*) as tx_count
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                WHERE acc.user_id = $1 
                AND date >= $2
                AND direction = 'debit'
                GROUP BY 1, 2
                ORDER BY 3 DESC
            """, user_id, start_date)

            if not rows:
                return []

            # Filtrar top 5 combinações de DOW e Hour
            # Consideramos "perigoso" se o valor total gasto for alto ou se houver muitas transações
            dangerous_patterns = []
            for r in rows[:10]: # Pegamos os 10 primeiros e refinamos
                dangerous_patterns.append({
                    "dow": int(r['dow']),
                    "hour": int(r['hour']),
                    "avg_amount": float(r['total_amount'] / r['tx_count']) if r['tx_count'] > 0 else 0,
                    "total_amount": float(r['total_amount']),
                    "count": r['tx_count']
                })
            
            # Ordenar por valor total e pegar top 5
            dangerous_patterns.sort(key=lambda x: x['total_amount'], reverse=True)
            return dangerous_patterns[:5]
        finally:
            await conn.close()

    async def generate_preventive_alerts(self, user_id: str):
        patterns = await self.identify_dangerous_patterns(user_id)
        
        if not patterns:
            return {"message": "Padrões insuficientes para gerar alertas preventivos."}

        # Identificar próximos alertas para os próximos 7 dias
        upcoming_alerts = []
        today = date.today()
        dow_names = ["Domingo", "Segunda", "Terça", "Quarta", "Quinta", "Sexta", "Sábado"]
        
        for i in range(1, 8):
            future_date = today + timedelta(days=i)
            future_dow = int(future_date.strftime("%w")) # 0=Sunday, 1=Monday...
            
            for p in patterns:
                if p['dow'] == future_dow:
                    upcoming_alerts.append({
                        "date": future_date.isoformat(),
                        "dow_name": dow_names[future_dow],
                        "hour": p['hour'],
                        "reason": f"Historicamente você gasta mais às {dow_names[future_dow]}s por volta das {p['hour']}h"
                    })

        prompt = f"Padrões Detectados:\n{json.dumps(patterns, default=str)}\n\nPróximos Alertas:\n{json.dumps(upcoming_alerts, default=str)}"
        llm_response = await self.llm.completion(prompt, system_prompt=DANGEROUS_DAYS_PROMPT)
        
        try:
            start_idx = llm_response.find("{")
            end_idx = llm_response.rfind("}")
            result = json.loads(llm_response[start_idx:end_idx+1])
            result["patterns"] = patterns
            result["upcoming"] = upcoming_alerts
        except:
            result = {"narrative": llm_response, "patterns": patterns, "upcoming": upcoming_alerts}

        # Salvar no cache
        conn = await self.get_db_connection()
        try:
            expires_at = datetime.now() + timedelta(days=7)
            await conn.execute("""
                INSERT INTO report_cache (user_id, report_type, period_key, result_json, expires_at)
                VALUES ($1, $2, $3, $4, $5)
                ON CONFLICT (user_id, report_type, period_key) DO UPDATE
                SET result_json = EXCLUDED.result_json,
                    expires_at = EXCLUDED.expires_at,
                    computed_at = NOW()
            """, user_id, "dangerous_days", date.today().strftime("%Y-%m-%W"), 
                 json.dumps(result), expires_at)
        finally:
            await conn.close()

        return result
