import json
import asyncio
from datetime import datetime, timedelta
from app.services.base_agent import BaseAgent

CFO_EXPLAINER_PROMPT = """
Você é o CFO Pessoal, um analista financeiro sênior que responde perguntas sobre gastos com dados reais e específicos.
Sua missão é explicar "o que aconteceu" ou "por que gastei tanto" de forma direta e baseada em fatos.

Diretrizes:
1. Seja específico, use números reais e nomes de estabelecimentos.
2. Máximo 4 frases curtas e diretas.
3. Não invente dados. Se não souber, diga que não encontrou dados suficientes.
4. Identifique o principal culpado pelo aumento ou pela variação.
5. Use formatação markdown para destacar valores (ex: **R$ 1.234,56**).

Pergunta do Usuário: {question}

Contexto Financeiro (JSON):
{context_json}

Exemplos de resposta:
- "72% do aumento veio de delivery (+**R$ 438,00**), principalmente sextas à noite no **iFood**. Também houve aumento em transporte (+**R$ 119,00**) após o dia 12."
- "Foram 3 fatores: parcela subiu (**R$ 340,00** a mais), mais visitas ao restaurante (+6 vezes) e uma compra pontual de **R$ 520,00** no dia 18."
"""

CFO_PROACTIVE_PROMPT = """
Você é o CFO Pessoal do usuário. Sua missão é tomar iniciativa e enviar insights diários valiosos sem ser perguntado.
Analise o contexto completo e gere EXATAMENTE UM insight proativo.

Tom: Parceiro financeiro de confiança, direto, empático e focado em fatos concretos.
Máximo 3 frases.

Contexto Completo:
{context_json}

Prioridades de Insight:
1. Alerta de Survival Mode ou Stress alto (se houver).
2. Conquistas recentes para celebrar (ex: recorde sem delivery).
3. Drift de Estilo de Vida (custos subindo mais que a renda).
4. Próximos gastos sazonais ou parcelas terminando.
5. Meta em risco ou oportunidade de economia detectada.

Retorne um JSON com:
- title: Título curto do insight
- description: O insight detalhado em até 3 frases
- type: "cfo_insight"
- severity: "info", "warning" ou "alert"
"""

class CFOAgent(BaseAgent):
    async def run(self, user_id: str, period: str = ""):
        # Padrão para rodar proativamente
        return await self.generate_proactive_insights(user_id)

    async def explain_period_spending(self, user_id: str, question: str, period_months: int = 1):
        conn = await self.get_db_connection()
        try:
            # 1. Buscar dados para o período e período anterior para comparação
            end_date = datetime.now().date()
            start_date = (datetime.now() - timedelta(days=30 * period_months)).date()
            prev_start_date = (start_date - timedelta(days=30 * period_months))
            
            # Gastos por categoria (atual vs anterior)
            category_comparison = await conn.fetch("""
                WITH current_period AS (
                    SELECT c.name as category, SUM(t.amount) as total, COUNT(*) as count
                    FROM transactions t
                    JOIN connected_accounts acc ON t.account_id = acc.id
                    LEFT JOIN categories c ON t.category_id = c.id
                    WHERE acc.user_id = $1 AND t.direction = 'debit' AND t.date BETWEEN $2 AND $3
                    GROUP BY c.name
                ),
                previous_period AS (
                    SELECT c.name as category, SUM(t.amount) as total, COUNT(*) as count
                    FROM transactions t
                    JOIN connected_accounts acc ON t.account_id = acc.id
                    LEFT JOIN categories c ON t.category_id = c.id
                    WHERE acc.user_id = $1 AND t.direction = 'debit' AND t.date BETWEEN $4 AND $5
                    GROUP BY c.name
                )
                SELECT 
                    COALESCE(curr.category, prev.category) as category,
                    COALESCE(curr.total, 0) as current_total,
                    COALESCE(prev.total, 0) as previous_total,
                    COALESCE(curr.count, 0) as current_count,
                    COALESCE(prev.count, 0) as previous_count
                FROM current_period curr
                FULL OUTER JOIN previous_period prev ON curr.category = prev.category
                ORDER BY curr.total DESC NULLS LAST
            """, user_id, start_date, end_date, prev_start_date, start_date)

            # Top Merchants (atual)
            top_merchants = await conn.fetch("""
                SELECT merchant_name, SUM(amount) as total, COUNT(*) as count
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                WHERE acc.user_id = $1 AND t.direction = 'debit' AND t.date BETWEEN $2 AND $3
                GROUP BY merchant_name
                ORDER BY total DESC
                LIMIT 10
            """, user_id, start_date, end_date)

            # Transações incomuns (anomalias simples: acima da média histórica do merchant)
            anomalies = await conn.fetch("""
                WITH merchant_avg AS (
                    SELECT merchant_name, AVG(amount) as avg_amount
                    FROM transactions t
                    JOIN connected_accounts acc ON t.account_id = acc.id
                    WHERE acc.user_id = $1 AND t.direction = 'debit' AND t.date < $2
                    GROUP BY merchant_name
                )
                SELECT t.merchant_name, t.amount, t.date, ma.avg_amount
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                JOIN merchant_avg ma ON t.merchant_name = ma.merchant_name
                WHERE acc.user_id = $1 AND t.direction = 'debit' 
                AND t.date BETWEEN $2 AND $3
                AND t.amount > ma.avg_amount * 2
                LIMIT 5
            """, user_id, start_date, end_date)

            context = {
                "category_comparison": [dict(r) for r in category_comparison],
                "top_merchants": [dict(r) for r in top_merchants],
                "anomalies": [dict(r) for r in anomalies],
                "period": {"start": str(start_date), "end": str(end_date)}
            }

            # 2. Gerar Resposta com Claude
            system_prompt = CFO_EXPLAINER_PROMPT.format(question=question, context_json=json.dumps(context, default=str))
            response_text = await self.llm.completion(f"Explique baseado nos dados: {question}", system_prompt=system_prompt)
            
            return {
                "explanation": response_text,
                "context": context
            }

        finally:
            await conn.close()

    async def generate_proactive_insights(self, user_id: str):
        conn = await self.get_db_connection()
        try:
            # 1. Coletar contexto de múltiplos agentes/tabelas
            
            # Stress e Survival
            stress_data = await conn.fetchrow("""
                SELECT score, level, trend, components 
                FROM stress_score_snapshots 
                WHERE user_id = $1 
                ORDER BY computed_at DESC LIMIT 1
            """, user_id)
            
            survival_data = await conn.fetchrow("""
                SELECT risk_score, level, is_active, projected_shortfall 
                FROM survival_mode_snapshots 
                WHERE user_id = $1 
                ORDER BY computed_at DESC LIMIT 1
            """, user_id)
            
            # Saúde e Metas
            health_data = await conn.fetchrow("""
                SELECT score, period_month FROM health_score_snapshots 
                WHERE user_id = $1 ORDER BY period_month DESC LIMIT 1
            """, user_id)
            
            goals_in_risk = await conn.fetch("""
                SELECT name, target_amount, current_amount, status 
                FROM financial_goals 
                WHERE user_id = $1 AND status = 'active'
                -- Aqui poderíamos cruzar com projeção se tivéssemos
            """, user_id)
            
            # Conquistas Recentes
            recent_achievements = await conn.fetch("""
                SELECT achievement_id, awarded_at 
                FROM achievements_awarded 
                WHERE user_id = $1 AND awarded_at > NOW() - INTERVAL '3 days'
            """, user_id)
            
            # Resumo de Gastos do Mês
            monthly_summary = await conn.fetchrow("""
                SELECT 
                    SUM(CASE WHEN direction = 'debit' THEN amount ELSE 0 END) as total_spent,
                    SUM(CASE WHEN direction = 'credit' THEN amount ELSE 0 END) as total_received
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                WHERE acc.user_id = $1 AND t.date >= date_trunc('month', NOW())
            """, user_id)

            full_context = {
                "stress": dict(stress_data) if stress_data else None,
                "survival": dict(survival_data) if survival_data else None,
                "health_score": dict(health_data) if health_data else None,
                "goals": [dict(r) for r in goals_in_risk],
                "achievements": [dict(r) for r in recent_achievements],
                "monthly_summary": dict(monthly_summary) if monthly_summary else None,
                "today": str(datetime.now().date())
            }

            # 2. Gerar Insight com LLM
            system_prompt = CFO_PROACTIVE_PROMPT.format(context_json=json.dumps(full_context, default=str))
            response_json_str = await self.llm.completion("Gere um insight proativo do CFO.", system_prompt=system_prompt)
            
            try:
                # O LLM deve retornar um JSON
                # Limpar possível markdown do JSON
                if "```json" in response_json_str:
                    response_json_str = response_json_str.split("```json")[1].split("```")[0].strip()
                elif "```" in response_json_str:
                    response_json_str = response_json_str.split("```")[1].split("```")[0].strip()
                
                insight = json.loads(response_json_str)
            except:
                # Fallback se falhar o parse do JSON
                insight = {
                    "title": "Insight do CFO",
                    "description": response_json_str,
                    "type": "cfo_insight",
                    "severity": "info"
                }

            # 3. Salvar como FeedEvent
            await conn.execute("""
                INSERT INTO feed_events (user_id, type, title, description, severity)
                VALUES ($1, $2, $3, $4, $5)
            """, user_id, insight["type"], insight["title"], insight["description"], insight.get("severity", "info"))

            return insight

        finally:
            await conn.close()
