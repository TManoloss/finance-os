import logging
from app.services.llm_base import LLMProvider
from app.services.groq_provider import GroqProvider
from app.services.gemini_provider import GeminiProvider
from app.config import config

class FallbackProvider(LLMProvider):
    def __init__(self):
        self.primary_name = config.LLM_PROVIDER
        
        # Inicializa ambos os provedores
        self.groq = GroqProvider()
        self.gemini = GeminiProvider()
        
        if self.primary_name == "gemini":
            self.primary = self.gemini
            self.secondary = self.groq
            self.secondary_name = "groq"
        else:
            self.primary_name = "groq"
            self.primary = self.groq
            self.secondary = self.gemini
            self.secondary_name = "gemini"

    async def completion(self, prompt: str, system_prompt: str = None) -> str:
        try:
            logging.info(f"[LLM_FALLBACK] Tentando provedor primário: {self.primary_name.upper()}...")
            result = await self.primary.completion(prompt, system_prompt)
            
            # Se o provedor retornar alguma mensagem de erro explícita em vez de lançar exceção
            if result and ("ERRO_SISTEMA" in result or "quota" in result.lower() or "limite de requisições" in result.lower()):
                raise ValueError(f"Provedor {self.primary_name} reportou erro interno: {result}")
                
            return result
        except BaseException as e:
            logging.warning(
                f"[LLM_FALLBACK] Falha no provedor primário {self.primary_name.upper()} "
                f"(erro/limite de tokens/timeout): {str(e)}. "
                f"Acionando provedor de contingência secundário: {self.secondary_name.upper()}..."
            )
            try:
                result_fallback = await self.secondary.completion(prompt, system_prompt)
                
                # Se o secundário também retornar erro estruturado
                if result_fallback and ("ERRO_SISTEMA" in result_fallback or "quota" in result_fallback.lower()):
                    raise ValueError(f"Provedor secundário {self.secondary_name} também falhou: {result_fallback}")
                    
                logging.info(f"[LLM_FALLBACK] Recuperação bem-sucedida usando {self.secondary_name.upper()}!")
                return result_fallback
            except BaseException as fallback_err:
                critical_msg = (
                    f"ERRO_CRITICO_LLM: Ambos os provedores falharam em cascata! "
                    f"Primário ({self.primary_name}): {str(e)} | "
                    f"Secundário ({self.secondary_name}): {str(fallback_err)}"
                )
                logging.error(f"[LLM_FALLBACK] {critical_msg}")
                return critical_msg
