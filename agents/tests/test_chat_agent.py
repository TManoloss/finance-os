import asyncio
import unittest

from app.models.chat import ChatMessage, ChatRequest
from app.services.chat_agent import ChatAgent


class FakeLLM:
    def __init__(
        self,
        *,
        fail=False,
        response="Explicação baseada nos resultados calculados.",
    ):
        self.fail = fail
        self.response = response
        self.prompt = None
        self.system_prompt = None

    async def completion(self, prompt, system_prompt=None):
        self.prompt = prompt
        self.system_prompt = system_prompt
        if self.fail:
            raise RuntimeError("LLM indisponível")
        return self.response


def make_agent(llm):
    agent = ChatAgent.__new__(ChatAgent)
    agent.llm = llm
    agent.db_calls = 0

    async def unexpected_db_call():
        agent.db_calls += 1
        raise AssertionError("ChatAgent não deve consultar o banco diretamente")

    agent.get_db_connection = unexpected_db_call
    return agent


def calculated_context():
    return {
        "intelligence": {
            "overall_health_score": 78,
            "health_status": "attention",
        },
        "monthly_replay": {
            "insights": ["Gastos reduziram 12,5% em relação ao mês anterior."],
            "next_month_guidance": "Reserve R$ 300 para compromissos antes de gastar.",
        },
    }


class ChatAgentContractTest(unittest.TestCase):
    def test_uses_calculated_context_without_database(self):
        llm = FakeLLM()
        agent = make_agent(llm)
        context = calculated_context()
        request = ChatRequest(
            user_id="user-1",
            message="Explique meu fechamento mensal.",
            history=[ChatMessage(role="user", content="Olá")],
            context=context,
        )

        result = asyncio.run(agent.run(request))

        self.assertEqual(result["source"], "calculated_results")
        self.assertEqual(result["response"], "Explicação baseada nos resultados calculados.")
        self.assertEqual(agent.db_calls, 0)
        self.assertIn('"overall_health_score": 78', llm.system_prompt)
        self.assertIn("Reserve R$ 300", llm.system_prompt)
        self.assertIn("Explique meu fechamento mensal.", llm.prompt)

    def test_llm_failure_returns_deterministic_fallback(self):
        agent = make_agent(FakeLLM(fail=True))
        request = ChatRequest(
            user_id="user-1",
            message="O que devo observar no próximo mês?",
            context=calculated_context(),
        )

        result = asyncio.run(agent.run(request))

        self.assertEqual(result["source"], "deterministic_fallback")
        self.assertTrue(result["fallback"])
        self.assertEqual(agent.db_calls, 0)
        self.assertIn("78/100", result["response"])
        self.assertIn("Gastos reduziram 12,5%", result["response"])
        self.assertIn("Reserve R$ 300", result["response"])

    def test_provider_error_string_returns_safe_deterministic_fallback(self):
        provider_error = (
            "ERRO_CRITICO_LLM: Ambos os provedores falharam em cascata! "
            "Primário (Groq): model_not_found | Secundário (Gemini): erro interno"
        )
        agent = make_agent(FakeLLM(response=provider_error))
        request = ChatRequest(
            user_id="user-1",
            message="Explique meu fechamento mensal.",
            context=calculated_context(),
        )

        result = asyncio.run(agent.run(request))

        self.assertEqual(result["source"], "deterministic_fallback")
        self.assertTrue(result["fallback"])
        self.assertEqual(agent.db_calls, 0)
        response = result["response"].lower()
        for forbidden in ("groq", "gemini", "model_not_found", "erro interno"):
            self.assertNotIn(forbidden, response)
        self.assertIn("78/100", result["response"])
        self.assertIn("Gastos reduziram 12,5%", result["response"])
        self.assertIn("Reserve R$ 300", result["response"])


if __name__ == "__main__":
    unittest.main()
