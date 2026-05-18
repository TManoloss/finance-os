import json
import logging
from datetime import date, datetime, timedelta
from app.services.base_agent import BaseAgent

logger = logging.getLogger(__name__)

ACHIEVEMENT_PROMPT = """
Você é um especialista em gamificação e psicologia comportamental aplicada a finanças.
O usuário acabou de conquistar uma nova meta financeira ou atingiu um marco importante.
Gere uma mensagem de parabéns curta, entusiasmada e personalizada para a conquista abaixo.

Conquista: {achievement_name}
Contexto: {context}

Instruções:
1. Use um tom de celebração (use emojis 🎉, 🏆, 🌟).
2. Explique brevemente por que isso é bom para a saúde financeira dele.
3. Máximo 2 frases.
4. Responda em Português.
"""

MISSION_PROMPT = """
Você é um coach financeiro focado em mudança de comportamento.
Com base no histórico do usuário, você deve gerar missões desafiadoras mas atingíveis para o mês atual.

Histórico resumido:
{history_json}

Missões Disponíveis para Personalizar:
- "Adeus Delivery": Ficar X dias sem pedir delivery.
- "Limite de Lazer": Gastar menos de R$ Y em Lazer.
- "Mestre da Categoria": Gastar menos que a média histórica em Z categoria.
- "Sincronização Diária": Sincronizar o app todos os dias.
- "Poupador": Direcionar R$ W para investimentos.

Gere 3 missões personalizadas. Retorne obrigatoriamente um JSON:
[
  {{
    "template_id": "id_curto",
    "title": "Título da Missão",
    "description": "Descrição clara e motivadora",
    "target_value": 0.0,
    "target_type": "spending_limit | frequency | income_target"
  }}
]
"""

class GamificationAgent(BaseAgent):
    async def check_achievements(self, user_id: str):
        logger.info(f"Verificando conquistas para o usuário {user_id}")
        conn = await self.get_db_connection()
        try:
            achievements = []
            today = date.today()
            start_of_month = today.replace(day=1)
            
            # 1. Mestre do Mercado (3+ idas ao mercado no mês)
            # Vamos considerar categoria Alimentação e merchants comuns ou subcategoria se existisse
            # Aqui vamos simplificar buscando 'Mercado', 'Super', 'Atacado' na descrição
            mercado_count = await conn.fetchval("""
                SELECT COUNT(*) 
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                WHERE acc.user_id = $1 
                AND t.direction = 'debit'
                AND t.date >= $2
                AND (t.description ILIKE '%mercado%' OR t.description ILIKE '%super%' OR t.description ILIKE '%atacado%')
            """, user_id, start_of_month)
            
            if mercado_count >= 3:
                achievements.append({
                    "id": "mestre_do_mercado",
                    "name": "Mestre do Mercado",
                    "context": f"Você realizou {mercado_count} compras em supermercados este mês, priorizando o preparo em casa!"
                })

            # 2. Adeus Delivery (3+ dias consecutivos sem delivery)
            # Primeiro buscamos os dias com delivery no mês
            delivery_days = await conn.fetch("""
                SELECT DISTINCT t.date
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                WHERE acc.user_id = $1 
                AND t.direction = 'debit'
                AND t.date >= $2
                AND (t.description ILIKE '%ifood%' OR t.description ILIKE '%rappi%' OR t.description ILIKE '%uber%eats%')
            """, user_id, start_of_month)
            
            delivery_dates = {r['date'] for r in delivery_days}
            streak = 0
            max_streak = 0
            curr = start_of_month
            while curr <= today:
                if curr not in delivery_dates:
                    streak += 1
                else:
                    max_streak = max(max_streak, streak)
                    streak = 0
                curr += timedelta(days=1)
            max_streak = max(max_streak, streak)
            
            if max_streak >= 3:
                achievements.append({
                    "id": "adeus_delivery",
                    "name": "Adeus Delivery",
                    "context": f"Você conseguiu ficar {max_streak} dias seguidos sem pedir delivery. Sua carteira agradece!"
                })

            # 3. Poupador Iniciante (Primeira transação em Investimentos no mês)
            has_investment = await conn.fetchval("""
                SELECT EXISTS (
                    SELECT 1 FROM transactions t
                    JOIN connected_accounts acc ON t.account_id = acc.id
                    JOIN categories c ON t.category_id = c.id
                    WHERE acc.user_id = $1 
                    AND c.name = 'Investimentos'
                    AND t.date >= $2
                )
            """, user_id, start_of_month)
            
            if has_investment:
                achievements.append({
                    "id": "poupador_iniciante",
                    "name": "Poupador Iniciante",
                    "context": "Você realizou seu primeiro aporte em investimentos este mês. O futuro começa agora!"
                })

            # Filtrar conquistas já concedidas este mês
            new_achievements = []
            for ach in achievements:
                already_awarded = await conn.fetchval("""
                    SELECT EXISTS (
                        SELECT 1 FROM achievements_awarded 
                        WHERE user_id = $1 AND achievement_id = $2
                        AND date_trunc('month', awarded_at) = date_trunc('month', $3::timestamp)
                    )
                """, user_id, ach['id'], datetime.combine(start_of_month, datetime.min.time()))
                
                if not already_awarded:
                    # Gerar mensagem com LLM
                    prompt = ACHIEVEMENT_PROMPT.format(achievement_name=ach['name'], context=ach['context'])
                    message = await self.llm.completion(prompt)
                    
                    # Salvar no banco
                    await conn.execute("""
                        INSERT INTO achievements_awarded (user_id, achievement_id, awarded_at, context_data)
                        VALUES ($1, $2, NOW(), $3)
                    """, user_id, ach['id'], json.dumps({"name": ach['name'], "message": message, "context": ach['context']}))
                    
                    # Gerar FeedEvent
                    await conn.execute("""
                        INSERT INTO feed_events (user_id, type, title, description, severity, created_at)
                        VALUES ($1, 'achievement', $2, $3, 'info', NOW())
                    """, user_id, f"Conquista: {ach['name']}", message)
                    
                    new_achievements.append({**ach, "message": message})
            
            return new_achievements

        finally:
            await conn.close()

    async def track_missions(self, user_id: str):
        logger.info(f"Rastreando progresso de missões para o usuário {user_id}")
        conn = await self.get_db_connection()
        try:
            active_missions = await conn.fetch("""
                SELECT * FROM missions 
                WHERE user_id = $1 AND status = 'active'
            """, user_id)
            
            for mission in active_missions:
                mission_id = mission['id']
                template_id = mission['template_id']
                target_value = float(mission['target_value'])
                current_value = 0.0
                
                # Lógica de rastreamento baseada no template
                if template_id == 'limite_lazer':
                    # Gastos em Lazer no mês atual
                    spent = await conn.fetchval("""
                        SELECT SUM(t.amount)
                        FROM transactions t
                        JOIN connected_accounts acc ON t.account_id = acc.id
                        JOIN categories c ON t.category_id = c.id
                        WHERE acc.user_id = $1 AND t.direction = 'debit'
                        AND c.name = 'Lazer'
                        AND t.date >= date_trunc('month', NOW())
                    """, user_id)
                    current_value = float(spent or 0)
                    
                    # Para limite, a missão é "concluída" se o mês acabar e o valor for menor
                    # Mas podemos atualizar o valor atual
                    await conn.execute("UPDATE missions SET current_value = $1 WHERE id = $2", current_value, mission_id)
                    
                elif template_id == 'no_delivery':
                    # Dias sem delivery no mês
                    delivery_days = await conn.fetchval("""
                        SELECT COUNT(DISTINCT t.date)
                        FROM transactions t
                        JOIN connected_accounts acc ON t.account_id = acc.id
                        WHERE acc.user_id = $1 
                        AND t.direction = 'debit'
                        AND t.date >= date_trunc('month', NOW())
                        AND (t.description ILIKE '%ifood%' OR t.description ILIKE '%rappi%' OR t.description ILIKE '%uber%eats%')
                    """, user_id)
                    
                    days_in_month_so_far = (date.today() - date.today().replace(day=1)).days + 1
                    current_value = float(days_in_month_so_far - (delivery_days or 0))
                    await conn.execute("UPDATE missions SET current_value = $1 WHERE id = $2", current_value, mission_id)

                elif template_id == 'poupador':
                    # Investimentos no mês
                    invested = await conn.fetchval("""
                        SELECT SUM(t.amount)
                        FROM transactions t
                        JOIN connected_accounts acc ON t.account_id = acc.id
                        JOIN categories c ON t.category_id = c.id
                        WHERE acc.user_id = $1 AND t.direction = 'debit'
                        AND c.name = 'Investimentos'
                        AND t.date >= date_trunc('month', NOW())
                    """, user_id)
                    current_value = float(invested or 0)
                    await conn.execute("UPDATE missions SET current_value = $1 WHERE id = $2", current_value, mission_id)

                # Verificar conclusão (simplificado)
                if template_id == 'poupador' and current_value >= target_value:
                    await conn.execute("""
                        UPDATE missions SET status = 'completed', completed_at = NOW() WHERE id = $1
                    """, mission_id)
                    # FeedEvent
                    await conn.execute("""
                        INSERT INTO feed_events (user_id, type, title, description, severity, created_at)
                        VALUES ($1, 'achievement', $2, $3, 'success', NOW())
                    """, user_id, f"Missão Concluída: {mission['title']}", f"Parabéns! Você atingiu sua meta de {mission['title']}.")

            return True
        finally:
            await conn.close()

    async def generate_missions(self, user_id: str):
        logger.info(f"Gerando missões para o usuário {user_id}")
        conn = await self.get_db_connection()
        try:
            # 1. Verificar se já tem missões ativas este mês
            active_missions = await conn.fetch("""
                SELECT * FROM missions 
                WHERE user_id = $1 AND status = 'active'
                AND started_at >= date_trunc('month', NOW())
            """, user_id)
            
            if active_missions:
                return [dict(m) for m in active_missions]
            
            # 2. Buscar histórico para personalização (últimos 30 dias)
            history = await conn.fetch("""
                SELECT c.name as category, SUM(t.amount) as total
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                JOIN categories c ON t.category_id = c.id
                WHERE acc.user_id = $1 AND t.direction = 'debit'
                AND t.date >= NOW() - INTERVAL '30 days'
                GROUP BY c.name
            """, user_id)
            
            history_json = json.dumps([dict(h) for h in history], default=str)
            
            # 3. Gerar missões com LLM
            prompt = MISSION_PROMPT.format(history_json=history_json)
            response_text = await self.llm.completion(prompt)
            
            try:
                start_idx = response_text.find("[")
                end_idx = response_text.rfind("]")
                missions_data = json.loads(response_text[start_idx:end_idx+1])
            except:
                # Fallback em caso de erro no LLM
                missions_data = [
                    {
                        "template_id": "limite_lazer",
                        "title": "Limite de Lazer",
                        "description": "Mantenha seus gastos de lazer sob controle.",
                        "target_value": 300.0,
                        "target_type": "spending_limit"
                    }
                ]
            
            # 4. Salvar missões
            ends_at = (date.today().replace(day=1) + timedelta(days=32)).replace(day=1) - timedelta(days=1)
            saved_missions = []
            for m in missions_data:
                mission_id = await conn.fetchval("""
                    INSERT INTO missions (user_id, template_id, title, description, target_value, started_at, ends_at, status)
                    VALUES ($1, $2, $3, $4, $5, NOW(), $6, 'active')
                    RETURNING id
                """, user_id, m['template_id'], m['title'], m['description'], m['target_value'], datetime.combine(ends_at, datetime.max.time()))
                
                saved_missions.append({**m, "id": str(mission_id), "status": "active"})
            
            return saved_missions

        finally:
            await conn.close()

    async def run(self, user_id: str):
        # Este método run pode ser um orquestrador que roda os três
        achievements = await self.check_achievements(user_id)
        await self.track_missions(user_id)
        missions = await self.generate_missions(user_id)
        return {
            "new_achievements": achievements,
            "active_missions": missions
        }
