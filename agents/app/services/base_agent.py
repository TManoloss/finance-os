from abc import ABC, abstractmethod
import asyncpg
from app.config import config
from app.services.llm_factory import get_llm_provider

class BaseAgent(ABC):
    def __init__(self):
        self.llm = get_llm_provider()

    async def get_db_connection(self):
        return await asyncpg.connect(config.DATABASE_URL, statement_cache_size=0)

    @abstractmethod
    async def run(self, user_id: str, period: str):
        pass

    async def save_report(self, user_id: str, agent_type: str, period_start, period_end, summary: str, insights: str):
        conn = await self.get_db_connection()
        try:
            await conn.execute("""
                INSERT INTO agent_reports (user_id, agent_type, period_start, period_end, summary_markdown, insights)
                VALUES ($1, $2, $3, $4, $5, $6)
            """, user_id, agent_type, period_start, period_end, summary, insights)
        finally:
            await conn.close()
