from abc import ABC, abstractmethod

class LLMProvider(ABC):
    @abstractmethod
    async def completion(self, prompt: str, system_prompt: str = None) -> str:
        pass
