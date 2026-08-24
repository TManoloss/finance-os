import json
from app.services.base_agent import BaseAgent
from app.models.chat import ChatRequest

CHAT_SYSTEM_PROMPT = """
Você é o Pierre, o explicador opcional do FinanceOS.
Sua missão é responder perguntas abertas e explicar resultados que já foram calculados pelo FinanceOS.

Diretrizes de Resposta:
1. Use somente os resultados calculados no contexto fornecido.
2. Não calcule métricas novas, não invente valores e não trate inferências como fatos.
3. Se o contexto não contiver a resposta, diga explicitamente que não há dados suficientes.
4. O FinanceOS não envia, recebe ou paga dinheiro e não executa investimentos.
5. Use formatação MARKDOWN para tornar a resposta visual e organizada:
   - Use **negrito** para destacar valores ou estabelecimentos.
   - Use TABELAS para comparar gastos ou listar itens.
   - Use listas com marcadores para dicas ou recomendações.
6. Seja conciso, útil e profissional.
7. Fale sempre em português do Brasil.

Resultados calculados do Usuário (JSON):
{context}
"""


def build_fallback_response(context: dict) -> str:
    parts = [
        "A explicação por IA está indisponível agora, mas seus cálculos continuam disponíveis."
    ]
    intelligence = context.get("intelligence") or {}
    score = intelligence.get("overall_health_score")
    if isinstance(score, (int, float)):
        status = {
            "excellent": "excelente",
            "good": "boa",
            "fair": "regular",
            "attention": "em atenção",
            "critical": "crítica",
        }.get(intelligence.get("health_status"), "sem classificação")
        parts.append(
            f"Sua saúde financeira está em **{score:.0f}/100** ({status})."
        )

    replay = context.get("monthly_replay") or {}
    parts.extend(str(item) for item in replay.get("insights", []) if item)
    guidance = replay.get("next_month_guidance")
    if guidance:
        parts.append(str(guidance))

    return "\n\n".join(parts)


class ChatAgent(BaseAgent):
    async def run(self, chat_req: ChatRequest):
        context = chat_req.context or {}
        system_prompt = CHAT_SYSTEM_PROMPT.format(
            context=json.dumps(context, default=str)
        )
        history = "".join(
            f"{message.role}: {message.content}\n" for message in chat_req.history[-10:]
        )
        try:
            response_text = await self.llm.completion(
                f"{history}user: {chat_req.message}", system_prompt=system_prompt
            )
            return {"response": response_text, "source": "calculated_results"}
        except Exception:
            return {
                "response": build_fallback_response(context),
                "source": "deterministic_fallback",
                "fallback": True,
            }
