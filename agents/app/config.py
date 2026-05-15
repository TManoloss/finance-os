import os
from dotenv import load_dotenv
from pathlib import Path

# Busca o .env na raiz do projeto (dois níveis acima de agents/app/)
env_path = Path(__file__).resolve().parent.parent.parent / '.env'
load_dotenv(dotenv_path=env_path)

class Config:
    GROQ_API_KEY = os.getenv("GROQ_API_KEY")
    GOOGLE_API_KEY = os.getenv("GOOGLE_API_KEY")
    DATABASE_URL = os.getenv("DATABASE_URL")
    GO_BACKEND_URL = os.getenv("GO_BACKEND_URL", "http://localhost:8080")
    LLM_PROVIDER = os.getenv("LLM_PROVIDER", "groq").lower()
    # Usa AGENTS_PORT se definido, caso contrário PORT, default 8000
    PORT = int(os.getenv("AGENTS_PORT", os.getenv("PORT", 8000)))

config = Config()
