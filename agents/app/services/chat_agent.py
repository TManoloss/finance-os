import json
from app.services.base_agent import BaseAgent
from app.models.chat import ChatRequest

CHAT_SYSTEM_PROMPT = """
Você é o Pierre, um assistente financeiro pessoal inteligente e amigável.
Sua missão é ajudar o usuário a entender suas finanças, responder perguntas sobre gastos e dar dicas de economia.

Diretrizes de Resposta:
1. Use os dados contextuais fornecidos para responder com precisão.
2. Use formatação MARKDOWN para tornar a resposta visual e organizada:
   - Use **negrito** para destacar valores ou estabelecimentos.
   - Use TABELAS para comparar gastos ou listar itens.
   - Use listas com marcadores para dicas ou recomendações.
3. Seja conciso, mas útil.
4. Use um tom amigável e informal, mas profissional.
5. Fale sempre em português do Brasil.

Contexto do Usuário (JSON):
{context}
"""

class ChatAgent(BaseAgent):
    async def run(self, chat_req: ChatRequest):
        conn = await self.get_db_connection()
        try:
            # 1. Buscar Contexto Relevante (Últimas transações e resumo do mês)
            rows = await conn.fetch("""
                SELECT t.description, t.amount, t.direction, t.date, c.name as category
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                LEFT JOIN categories c ON t.category_id = c.id
                WHERE acc.user_id = $1
                ORDER BY t.date DESC
                LIMIT 20
            """, chat_req.user_id)
            
            summary_rows = await conn.fetch("""
                SELECT c.name as category, SUM(t.amount) as total
                FROM transactions t
                JOIN connected_accounts acc ON t.account_id = acc.id
                JOIN categories c ON t.category_id = c.id
                WHERE acc.user_id = $1 AND t.direction = 'debit' AND t.date >= date_trunc('month', now())
                GROUP BY c.name
            """, chat_req.user_id)

            context_data = {
                "recent_transactions": [dict(r) for r in rows],
                "monthly_spending_by_category": [dict(r) for r in summary_rows]
            }
            
            # 2. Montar Prompt com Histórico
            system_prompt = CHAT_SYSTEM_PROMPT.format(context=json.dumps(context_data, default=str))
            
            # Simula histórico de mensagens para o LLM
            full_prompt = ""
            if chat_req.history:
                for msg in chat_req.history:
                    full_prompt += f"{msg.role}: {msg.content}\n"
            full_prompt += f"user: {chat_req.message}"

            # 3. Chamar LLM
            response_text = await self.llm.completion(full_prompt, system_prompt=system_prompt)
            
            return {"response": response_text}

        finally:
            await conn.close()
