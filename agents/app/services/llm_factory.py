from app.services.fallback_provider import FallbackProvider

def get_llm_provider():
    return FallbackProvider()
