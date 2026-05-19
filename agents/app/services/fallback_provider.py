import logging
import base64
import asyncpg
from contextvars import ContextVar
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
from app.services.llm_base import LLMProvider
from app.services.groq_provider import GroqProvider
from app.services.gemini_provider import GeminiProvider
from app.config import config

# Contexto global para armazenar o ID do usuário da requisição atual
current_user_id = ContextVar("current_user_id", default=None)

def decrypt_aes_gcm(encrypted_base64: str, encryption_key: str) -> str:
    if not encrypted_base64 or not encryption_key:
        return ""
    try:
        if len(encryption_key) == 44:
            key_bytes = base64.b64decode(encryption_key)
        else:
            key_bytes = encryption_key.encode('utf-8')
            
        data = base64.b64decode(encrypted_base64)
        nonce_size = 12 
        nonce = data[:nonce_size]
        ciphertext = data[nonce_size:]
        
        aesgcm = AESGCM(key_bytes)
        decrypted = aesgcm.decrypt(nonce, ciphertext, None)
        return decrypted.decode('utf-8')
    except Exception as e:
        logging.error(f"[LLM_DECRYPT] Erro ao descriptografar chave do usuário: {e}")
        return ""

async def get_user_keys(user_id: str) -> tuple[str, str]:
    if not user_id:
        return "", ""
    try:
        conn = await asyncpg.connect(config.DATABASE_URL, statement_cache_size=0)
        try:
            row = await conn.fetchrow(
                "SELECT groq_api_key_encrypted, gemini_api_key_encrypted FROM users WHERE id = $1", 
                user_id
            )
            if row:
                groq_enc = row['groq_api_key_encrypted'] or ""
                gemini_enc = row['gemini_api_key_encrypted'] or ""
                
                groq_key = ""
                gemini_key = ""
                
                encryption_key = config.ENCRYPTION_KEY
                
                if groq_enc and encryption_key:
                    groq_key = decrypt_aes_gcm(groq_enc, encryption_key)
                if gemini_enc and encryption_key:
                    gemini_key = decrypt_aes_gcm(gemini_enc, encryption_key)
                    
                return groq_key, gemini_key
        finally:
            await conn.close()
    except Exception as e:
        logging.error(f"[LLM_FALLBACK] Erro ao carregar chaves no banco de dados para o usuário {user_id}: {e}")
    return "", ""

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
        # Tenta obter o ID do usuário do contexto atual
        user_id = current_user_id.get()
        groq_key = None
        gemini_key = None
        
        if user_id:
            try:
                groq_key, gemini_key = await get_user_keys(user_id)
                if groq_key:
                    logging.info(f"[LLM_FALLBACK] Utilizando chave Groq customizada para o usuário {user_id}")
                if gemini_key:
                    logging.info(f"[LLM_FALLBACK] Utilizando chave Gemini customizada para o usuário {user_id}")
            except Exception as e:
                logging.error(f"[LLM_FALLBACK] Falha ao carregar chaves customizadas do usuário {user_id}: {e}")

        # Determina quais chaves de API passar para os provedores
        primary_key = groq_key if self.primary_name == "groq" else gemini_key
        secondary_key = gemini_key if self.primary_name == "groq" else groq_key

        try:
            logging.info(f"[LLM_FALLBACK] Tentando provedor primário: {self.primary_name.upper()}...")
            result = await self.primary.completion(prompt, system_prompt, api_key=primary_key)
            
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
                result_fallback = await self.secondary.completion(prompt, system_prompt, api_key=secondary_key)
                
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
