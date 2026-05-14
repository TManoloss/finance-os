from app.config import config
from app.services.groq_provider import GroqProvider
from app.services.gemini_provider import GeminiProvider

def get_llm_provider():
    if config.LLM_PROVIDER == "gemini":
        return GeminiProvider()
    return GroqProvider() # Default groq
