import os
from dotenv import load_dotenv

load_dotenv()

class Config:
    GROQ_API_KEY = os.getenv("GROQ_API_KEY")
    GOOGLE_API_KEY = os.getenv("GOOGLE_API_KEY")
    DATABASE_URL = os.getenv("DATABASE_URL")
    GO_BACKEND_URL = os.getenv("GO_BACKEND_URL", "http://localhost:8080")
    LLM_PROVIDER = os.getenv("LLM_PROVIDER", "groq").lower()
    PORT = int(os.getenv("PORT", 8000))

config = Config()
