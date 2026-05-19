import google.generativeai as genai
from app.services.llm_base import LLMProvider
from app.config import config
import logging

class GeminiProvider(LLMProvider):
    def __init__(self):
        genai.configure(api_key=config.GOOGLE_API_KEY)
        self.model = genai.GenerativeModel('gemini-2.0-flash')

    async def completion(self, prompt: str, system_prompt: str = None, api_key: str = None) -> str:
        try:
            full_prompt = f"{system_prompt}\n\n{prompt}" if system_prompt else prompt
            
            model = self.model
            if api_key:
                genai.configure(api_key=api_key)
                model = genai.GenerativeModel('gemini-2.0-flash')

            try:
                response = await model.generate_content_async(full_prompt)
            finally:
                if api_key:
                    genai.configure(api_key=config.GOOGLE_API_KEY)
            
            # Tenta obter o texto, tratando possíveis bloqueios de segurança
            try:
                if response.text:
                    return response.text
                return "O modelo retornou uma resposta vazia."
            except ValueError:
                # Se o response.text lançar ValueError, é provável que a resposta tenha sido bloqueada
                if response.candidates:
                    return f"A resposta foi bloqueada pelos filtros de segurança. Motivo: {response.candidates[0].finish_reason}"
                return "A resposta foi bloqueada e não há candidatos disponíveis."

        except BaseException as e:
            logging.error(f"[Gemini] Erro crítico na geração de conteúdo: {str(e)}")
            if "429" in str(e) or "quota" in str(e).lower():
                return "ERRO_SISTEMA: Limite de requisições excedido (Quota 429). Por favor, aguarde um momento."
            return f"ERRO_SISTEMA: Falha ao processar requisição de IA. ({str(e)})"
