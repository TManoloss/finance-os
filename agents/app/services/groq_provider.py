from groq import AsyncGroq
from app.services.llm_base import LLMProvider
from app.config import config

class GroqProvider(LLMProvider):
    def __init__(self):
        self.client = AsyncGroq(api_key=config.GROQ_API_KEY)

    async def completion(self, prompt: str, system_prompt: str = None) -> str:
        messages = []
        if system_prompt:
            messages.append({"role": "system", "content": system_prompt})
        messages.append({"role": "user", "content": prompt})

        response = await self.client.chat.completions.create(
            model="llama-3.3-70b-versatile",
            messages=messages,
            temperature=0.1,
            max_tokens=1024
        )
        return response.choices[0].message.content
