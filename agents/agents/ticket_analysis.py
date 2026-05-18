import json
import logging
from datetime import date, datetime, timedelta
from app.services.base_agent import BaseAgent

logger = logging.getLogger(__name__)

TICKET_ANALYSIS_PROMPT = """
Você é um analista financeiro sênior especializado em comportamento de consumo.
Dadas as decomposições de gastos abaixo, identifique o padrão mais interessante e gere um insight em 2-3 frases em português.
Foque no padrão menos óbvio (não necessariamente o maior em valor, mas o mais revelador sobre o comportamento do usuário).

Dados de Decomposição:
{decompositions_json}

Exemplo de tom: "Seus gastos com restaurantes cresceram 45%, mas não porque você saiu mais — você foi 10% menos vezes, mas o ticket médio subiu de R$67 para R$112. Seu padrão de jantar fora ficou mais premium."

Retorne obrigatoriamente um JSON com:
- summary: resumo geral do padrão identificado.
- top_decompositions: as 3 decomposições mais relevantes.
- insight_narrative: o parágrafo narrativo solicitado.
"""

class TicketAnalysisAgent(BaseAgent):
    async def decompose_spending_growth(self, user_id: str, category_id: str, category_name: str, months: int = 6):
        conn = await self.get_db_connection()
        try:
            today = date.today()
            half_months = months // 2
            
            period_b_start = (today - timedelta(days=half_months * 30)).replace(day=1)
            period_a_start = (period_b_start - timedelta(days=half_months * 30)).replace(day=1)
            period_b_end = today
            
            # Period A
            data_a = await conn.fetchrow("""
                SELECT 
                    COALESCE(SUM(t.amount), 0) as total,
                    COUNT(t.id) as freq
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                WHERE acc.user_id = $1 AND t.category_id = $2 AND t.direction = 'debit'
                AND t.date >= $3 AND t.date < $4
            """, user_id, category_id, period_a_start, period_b_start)
            
            # Period B
            data_b = await conn.fetchrow("""
                SELECT 
                    COALESCE(SUM(t.amount), 0) as total,
                    COUNT(t.id) as freq
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                WHERE acc.user_id = $1 AND t.category_id = $2 AND t.direction = 'debit'
                AND t.date >= $3 AND t.date < $4
            """, user_id, category_id, period_b_start, period_b_end)
            
            total_a = float(data_a['total'])
            freq_a = int(data_a['freq'])
            total_b = float(data_b['total'])
            freq_b = int(data_b['freq'])
            
            if freq_a == 0 or freq_b == 0 or total_a == 0:
                return None
                
            avg_ticket_a = total_a / freq_a
            avg_ticket_b = total_b / freq_b
            
            var_total = (total_b - total_a) / total_a
            var_freq = (freq_b - freq_a) / freq_a
            var_ticket = (avg_ticket_b - avg_ticket_a) / avg_ticket_a
            
            # Determine growth type
            tipo_crescimento = "MIXED"
            if var_total > 0:
                if var_freq > 0 and var_ticket <= 0.05:
                    tipo_crescimento = "FREQUENCY_DRIVEN"
                elif var_ticket > 0 and var_freq <= 0.05:
                    tipo_crescimento = "PRICE_DRIVEN"
                elif var_ticket > 0 and var_freq < 0:
                    tipo_crescimento = "TICKET_UP_FREQUENCY_DOWN"
            
            return {
                "category_id": category_id,
                "category_name": category_name,
                "total_a": round(total_a, 2),
                "total_b": round(total_b, 2),
                "freq_a": freq_a,
                "freq_b": freq_b,
                "avg_ticket_a": round(avg_ticket_a, 2),
                "avg_ticket_b": round(avg_ticket_b, 2),
                "var_total_percent": round(var_total * 100, 2),
                "var_freq_percent": round(var_freq * 100, 2),
                "var_ticket_percent": round(var_ticket * 100, 2),
                "tipo_crescimento": tipo_crescimento
            }
        finally:
            await conn.close()

    async def analyze_all_categories(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            categories = await conn.fetch("""
                SELECT id, name FROM categories 
                WHERE user_id = $1 OR user_id IS NULL
            """, user_id)
            
            decompositions = []
            for cat in categories:
                decomp = await self.decompose_spending_growth(user_id, str(cat['id']), cat['name'])
                if decomp:
                    decompositions.append(decomp)
            
            # Ordenar por variacao_gasto_total DESC
            decompositions.sort(key=lambda x: x['var_total_percent'], reverse=True)
            return decompositions
        finally:
            await conn.close()

    async def run(self, user_id: str):
        logger.info(f"Iniciando análise de ticket médio para o usuário {user_id}")
        decompositions = await self.analyze_all_categories(user_id)
        
        if not decompositions:
            return {
                "summary": "Dados insuficientes para analisar o ticket médio.",
                "top_decompositions": [],
                "insight_narrative": "Ainda não temos transações suficientes em períodos distintos para decompor o crescimento dos seus gastos."
            }

        # Filtrar apenas categorias com crescimento relevante para o insight
        relevant_decomps = [d for d in decompositions if abs(d['var_total_percent']) > 5]
        if not relevant_decomps:
            relevant_decomps = decompositions[:10]
        
        prompt = f"Analise o comportamento de ticket médio para o usuário {user_id}:\n{json.dumps(relevant_decomps[:10], default=str)}"
        response_text = await self.llm.completion(prompt, system_prompt=TICKET_ANALYSIS_PROMPT)
        
        try:
            start_idx = response_text.find("{")
            end_idx = response_text.rfind("}")
            if start_idx == -1 or end_idx == -1:
                raise ValueError("JSON not found")
            report_data = json.loads(response_text[start_idx:end_idx+1])
        except Exception:
            report_data = {
                "summary": "Análise de crescimento de gastos.",
                "top_decompositions": decompositions[:3],
                "insight_narrative": response_text
            }

        # Salvar no cache de relatórios
        conn = await self.get_db_connection()
        try:
            today_month = date.today().strftime("%Y-%m")
            expires_at = datetime.now() + timedelta(days=7)
            await conn.execute("""
                INSERT INTO report_cache (user_id, report_type, period_key, result_json, expires_at)
                VALUES ($1, $2, $3, $4, $5)
                ON CONFLICT (user_id, report_type, period_key) DO UPDATE
                SET result_json = EXCLUDED.result_json,
                    expires_at = EXCLUDED.expires_at,
                    computed_at = NOW()
            """, user_id, "ticket_analysis", today_month, json.dumps(report_data), expires_at)
        finally:
            await conn.close()

        return report_data
