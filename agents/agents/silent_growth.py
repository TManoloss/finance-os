import json
import numpy as np
from datetime import date, datetime, timedelta
from app.services.base_agent import BaseAgent
import logging

logger = logging.getLogger(__name__)

SILENT_GROWTH_SYSTEM_PROMPT = """
Você é um analista financeiro especialista em detectar padrões de gastos despercebidos.
Seu objetivo é alertar o usuário sobre categorias que estão crescendo rapidamente, mesmo que o valor absoluto ainda não seja o maior.
"""

SILENT_GROWTH_PROMPT = """
Dado o crescimento silencioso da categoria abaixo, escreva um alerta em 2 frases.
Tom: direto, sem alarme exagerado. Mostre os números e a projeção.

Categoria: {category_name}
Crescimento: {growth_data}

Exemplo: "Seus gastos com 'Assinaturas' cresceram R$87/mês nos últimos 6 meses — um aumento de 340% que provavelmente passou despercebido. Se continuar, serão R$1.044 a mais por ano nessa categoria."

Retorne obrigatoriamente apenas o texto do alerta.
"""

class SilentGrowthAgent(BaseAgent):
    async def run(self, user_id: str):
        return await self.detect_silent_growth_categories(user_id)

    async def detect_silent_growth_categories(self, user_id: str, months: int = 6):
        conn = await self.get_db_connection()
        try:
            # 1. Buscar gastos mensais por categoria nos últimos N meses
            today = date.today()
            current_month_start = today.replace(day=1)
            
            # Calcular início da janela
            start_date = current_month_start
            for _ in range(months):
                if start_date.month == 1:
                    start_date = start_date.replace(year=start_date.year - 1, month=12)
                else:
                    start_date = start_date.replace(month=start_date.month - 1)
            
            rows = await conn.fetch("""
                SELECT 
                    c.name as category_name,
                    date_trunc('month', t.date)::date as month,
                    SUM(t.amount) as total_amount
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                JOIN categories c ON t.category_id = c.id
                WHERE acc.user_id = $1
                AND t.date >= $2 AND t.date < $3
                AND t.direction = 'debit'
                GROUP BY c.name, month
                ORDER BY c.name, month
            """, user_id, start_date, current_month_start)

            if not rows:
                return {"message": "Dados insuficientes para análise de crescimento silencioso."}

            # Gerar lista de todos os meses na janela para mapeamento de X
            all_months = []
            curr = start_date
            while curr < current_month_start:
                all_months.append(curr)
                if curr.month == 12:
                    curr = curr.replace(year=curr.year + 1, month=1)
                else:
                    curr = curr.replace(month=curr.month + 1)
            
            month_to_idx = {m: i for i, m in enumerate(all_months)}
            
            # Organizar dados por categoria
            category_data = {}
            for r in rows:
                cat = r['category_name']
                if cat not in category_data:
                    category_data[cat] = []
                
                m = r['month']
                if m in month_to_idx:
                    category_data[cat].append((month_to_idx[m], float(r['total_amount'])))

            # 2. Filtrar categorias que NÃO são as maiores em valor absoluto
            cat_totals = {cat: sum(amount for _, amount in data) for cat, data in category_data.items()}
            sorted_cats_by_total = sorted(cat_totals.items(), key=lambda x: x[1], reverse=True)
            top_3_cats = {cat for cat, _ in sorted_cats_by_total[:3]}
            
            # 3. Aplicar regressão linear
            results = []
            for cat, data in category_data.items():
                if cat in top_3_cats:
                    continue
                
                if len(data) < 3: # Mínimo 3 meses com dados
                    continue
                
                x = np.array([idx for idx, _ in data])
                y = np.array([amount for _, amount in data])
                
                # Regressão linear simples: y = ax + b
                A = np.vstack([x, np.ones(len(x))]).T
                slope, intercept = np.linalg.lstsq(A, y, rcond=None)[0]
                
                # Calcular R-squared (coeficiente de determinação)
                y_pred = slope * x + intercept
                ss_res = np.sum((y - y_pred) ** 2)
                ss_tot = np.sum((y - np.mean(y)) ** 2)
                r_squared = 1 - (ss_res / ss_tot) if ss_tot > 0 else 0
                
                # Filtrar apenas tendências consistentes e crescentes
                if r_squared > 0.5 and slope > 0:
                    gasto_primeiro_mes = data[0][1]
                    gasto_atual = data[-1][1]
                    variacao_absoluta = gasto_atual - gasto_primeiro_mes
                    
                    avg_amount = np.mean(y)
                    taxa_crescimento_mensal_percent = (slope / avg_amount * 100) if avg_amount > 0 else 0
                    
                    # Projeção de 12 meses a partir do último mês
                    projecao_12_meses = gasto_atual + (slope * 12)
                    
                    # Variação total no período analisado
                    variacao_total_percent = ((gasto_atual - gasto_primeiro_mes) / gasto_primeiro_mes * 100) if gasto_primeiro_mes > 0 else 0

                    results.append({
                        "category_name": cat,
                        "taxa_crescimento_mensal_percent": round(taxa_crescimento_mensal_percent, 2),
                        "slope": round(slope, 2),
                        "gasto_ha_n_meses": round(gasto_primeiro_mes, 2),
                        "gasto_atual": round(gasto_atual, 2),
                        "variacao_absoluta": round(variacao_absoluta, 2),
                        "variacao_total_percent": round(variacao_total_percent, 2),
                        "projecao_12_meses": round(projecao_12_meses, 2),
                        "r_squared": round(r_squared, 4),
                        "data_points": [round(amount, 2) for _, amount in data]
                    })

            # Ordenar por taxa de crescimento mensal %
            results = sorted(results, key=lambda x: x['taxa_crescimento_mensal_percent'], reverse=True)

            if not results:
                return {"message": "Nenhuma tendência de crescimento silencioso consistente detectada."}

            top_cat = results[0]
            
            # 4. Gerar alerta via LLM
            growth_data_str = (
                f"R${top_cat['slope']}/mês, "
                f"subiu de R${top_cat['gasto_ha_n_meses']} para R${top_cat['gasto_atual']} em {len(all_months)} meses "
                f"({top_cat['variacao_total_percent']}% de aumento). "
                f"Projeção anual extra: R${top_cat['slope'] * 12}."
            )
            
            prompt = SILENT_GROWTH_PROMPT.format(
                category_name=top_cat['category_name'],
                growth_data=growth_data_str
            )
            
            alert_text = await self.llm.completion(prompt, system_prompt=SILENT_GROWTH_SYSTEM_PROMPT)
            
            report_data = {
                "top_category": top_cat,
                "alert": alert_text.strip(),
                "all_trends": results[:5], # Top 5 tendências
                "computed_at": datetime.now().isoformat()
            }

            # 5. Salvar no cache
            expires_at = datetime.now() + timedelta(days=7)
            period_key = current_month_start.strftime("%Y-%m")
            
            await conn.execute("""
                INSERT INTO report_cache (user_id, report_type, period_key, result_json, expires_at)
                VALUES ($1, $2, $3, $4, $5)
                ON CONFLICT (user_id, report_type, period_key) DO UPDATE
                SET result_json = EXCLUDED.result_json,
                    expires_at = EXCLUDED.expires_at,
                    computed_at = NOW()
            """, user_id, "silent_growth", period_key, json.dumps(report_data), expires_at)

            return report_data

        finally:
            await conn.close()
