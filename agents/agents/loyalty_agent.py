import json
import logging
from datetime import date, datetime, timedelta
from app.services.base_agent import BaseAgent

logger = logging.getLogger(__name__)

LOYALTY_ANALYSIS_PROMPT = """
Você é um especialista em retenção de clientes e comportamento de consumo.
Analise os padrões de lealdade e abandono de estabelecimentos (merchants) do usuário abaixo.
Identifique mudanças significativas (shifts de lealdade) ou desperdícios com experimentações únicas (abandonment cost).

Dados de Lealdade:
{loyalty_data_json}

Diretrizes:
1. Identifique se o usuário "trocou" um estabelecimento por outro na mesma categoria (Shift).
2. Comente sobre o Customer Lifetime Value (CLV) dos estabelecimentos mais leais.
3. Gere um insight em 2-3 frases em português, com tom analítico e útil.

Exemplo de tom: "Você parece ter trocado o Pão de Açúcar pelo Carrefour nos últimos 3 meses: seu gasto médio no Carrefour subiu 80% enquanto o Pão de Açúcar foi abandonado. Essa mudança te economizou cerca de R$120/mês em ticket médio."

Retorne obrigatoriamente um JSON com:
- summary: resumo geral da saúde da lealdade do usuário.
- insights: lista de 3 observações específicas [{{"title": "...", "description": "...", "severity": "info/warning"}}]
- insight_narrative: o parágrafo narrativo solicitado.
"""

class LoyaltyAgent(BaseAgent):
    async def classify_merchant_relationships(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # Busca histórico de todos os merchants nos últimos 12 meses
            today = date.today()
            one_year_ago = today - timedelta(days=365)
            
            # Agrupar por merchant e mês para ver a constância
            merchants_data = await conn.fetch("""
                SELECT 
                    t.merchant_name,
                    c.name as category_name,
                    COUNT(t.id) as total_freq,
                    SUM(t.amount) as total_spent,
                    MIN(t.date) as first_seen,
                    MAX(t.date) as last_seen,
                    COUNT(DISTINCT TO_CHAR(t.date, 'YYYY-MM')) as active_months
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                LEFT JOIN categories c ON t.category_id = c.id
                WHERE acc.user_id = $1 AND t.direction = 'debit'
                AND t.date >= $2
                GROUP BY t.merchant_name, c.name
                HAVING COUNT(t.id) > 0
            """, user_id, one_year_ago)
            
            relationships = []
            abandonment_cost = 0.0
            
            for m in merchants_data:
                merchant = m['merchant_name']
                total_freq = m['total_freq']
                total_spent = float(m['total_spent'])
                last_seen = m['last_seen']
                active_months = m['active_months']
                months_since_first = (today - m['first_seen']).days / 30
                days_since_last = (today - last_seen).days
                
                classification = "OCASIONAL"
                
                # Regras de Classificação
                if total_freq >= 12 and active_months >= 6 and days_since_last <= 45:
                    classification = "LEAL"
                elif total_freq >= 4 and days_since_last <= 30:
                    classification = "FREQUENTE"
                elif total_freq <= 2 and months_since_first > 3 and days_since_last > 90:
                    classification = "EXPERIMENTADO"
                    abandonment_cost += total_spent
                elif (total_freq >= 5 or active_months >= 3) and days_since_last > 60:
                    classification = "ABANDONADO"
                
                # Detecção de RESGATADO (se foi abandonado no passado mas voltou recentemente)
                # Nota: Simplificação pois não estamos vendo o gap histórico completo aqui
                
                relationships.append({
                    "merchant": merchant,
                    "category": m['category_name'],
                    "classification": classification,
                    "total_freq": total_freq,
                    "total_spent": round(total_spent, 2),
                    "clv": round(total_spent, 2), # CLV simplificado para o período visto
                    "last_seen_days": days_since_last,
                    "active_months": active_months
                })
                
            return relationships, round(abandonment_cost, 2)
        finally:
            await conn.close()

    async def detect_shifts(self, relationships):
        # Agrupar por categoria e ver quem subiu e quem desceu (ABANDONADO vs NOVO/LEAL)
        # Por enquanto vamos deixar para o LLM detectar os padrões complexos, 
        # mas podemos preparar os dados.
        category_summary = {}
        for r in relationships:
            cat = r['category'] or "Outros"
            if cat not in category_summary:
                category_summary[cat] = {"abandoned": [], "loyal": [], "frequent": []}
            
            if r['classification'] == "ABANDONADO":
                category_summary[cat]["abandoned"].append(r['merchant'])
            elif r['classification'] == "LEAL":
                category_summary[cat]["loyal"].append(r['merchant'])
            elif r['classification'] == "FREQUENTE":
                category_summary[cat]["frequent"].append(r['merchant'])
                
        shifts = []
        for cat, data in category_summary.items():
            if data["abandoned"] and (data["loyal"] or data["frequent"]):
                shifts.append({
                    "category": cat,
                    "from": data["abandoned"][:2],
                    "to": (data["loyal"] + data["frequent"])[:2]
                })
        return shifts

    async def run(self, user_id: str):
        logger.info(f"Iniciando análise de lealdade para o usuário {user_id}")
        relationships, abandonment_cost = await self.classify_merchant_relationships(user_id)
        shifts = await self.detect_shifts(relationships)
        
        # Filtrar apenas os dados mais relevantes para o prompt
        # Top leais e abandonados
        top_loyal = sorted([r for r in relationships if r['classification'] in ["LEAL", "FREQUENTE"]], 
                          key=lambda x: x['total_spent'], reverse=True)[:10]
        abandoned = [r for r in relationships if r['classification'] == "ABANDONADO"][:5]
        
        loyalty_data = {
            "top_loyal": top_loyal,
            "abandoned": abandoned,
            "abandonment_cost": abandonment_cost,
            "detected_shifts": shifts
        }

        prompt = f"Analise a lealdade de merchants para o usuário {user_id}:\n{json.dumps(loyalty_data, default=str)}"
        
        try:
            response_text = await self.llm.completion(prompt, system_prompt=LOYALTY_ANALYSIS_PROMPT)
            start_idx = response_text.find("{")
            end_idx = response_text.rfind("}")
            if start_idx == -1 or end_idx == -1:
                raise ValueError("JSON not found")
            report_data = json.loads(response_text[start_idx:end_idx+1])
        except Exception as e:
            logger.error(f"Erro ao chamar LLM no loyalty agent: {e}")
            report_data = {
                "summary": "Análise de lealdade e abandonos.",
                "insights": [
                    {"title": "Mapeamento de Hábitos", "description": "Você possui estabelecimentos frequentes mapeados. Continue acompanhando suas compras repetidas.", "severity": "info"}
                ],
                "insight_narrative": "Não foi possível gerar a análise detalhada via inteligência artificial no momento, mas seus dados históricos continuam disponíveis abaixo."
            }

        # Adicionar dados brutos processados ao report_data
        report_data["abandonment_cost"] = abandonment_cost
        report_data["top_loyal"] = top_loyal
        report_data["loyalty_stats"] = {
            "leal": len([r for r in relationships if r['classification'] == "LEAL"]),
            "frequente": len([r for r in relationships if r['classification'] == "FREQUENTE"]),
            "abandonado": len([r for r in relationships if r['classification'] == "ABANDONADO"]),
            "experimentado": len([r for r in relationships if r['classification'] == "EXPERIMENTADO"])
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
            """, user_id, "loyalty_analysis", today_month, json.dumps(report_data), expires_at)
        finally:
            await conn.close()

        return report_data
