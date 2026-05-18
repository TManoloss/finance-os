from fastapi import FastAPI, BackgroundTasks
from app.config import config
from app.models.transaction import TransactionClassifyRequest
from app.models.chat import ChatRequest
from app.services.classifier import ClassifierService
from app.services.daily_agent import DailyAgent
from app.services.weekly_agent import WeeklyAgent
from app.services.monthly_agent import MonthlyAgent
from app.services.chat_agent import ChatAgent
from app.services.cashflow_agent import CashflowAgent
from app.services.behavioral_agent import BehavioralAgent
from app.services.comparison_agent import ComparisonAgent
from app.services.invisible_spending_agent import InvisibleSpendingAgent
from app.services.projection_engine import ProjectionEngine
from app.services.health_score_agent import HealthScoreAgent
from app.services.merchant_intelligence_agent import MerchantIntelligenceAgent
from app.services.seasonality_agent import SeasonalityAgent
from app.services.narrative_report_agent import NarrativeReportAgent
from app.services.goals_agent import GoalsAgent
from agents.personal_inflation import PersonalInflationAgent
from agents.silent_growth import SilentGrowthAgent
from datetime import datetime

app = FastAPI(title="Finance OS Agents Service")
classifier_service = ClassifierService()
daily_agent = DailyAgent()
weekly_agent = WeeklyAgent()
monthly_agent = MonthlyAgent()
chat_agent = ChatAgent()
cashflow_agent = CashflowAgent()
behavioral_agent = BehavioralAgent()
comparison_agent = ComparisonAgent()
invisible_agent = InvisibleSpendingAgent()
projection_engine = ProjectionEngine()
health_score_agent = HealthScoreAgent()
merchant_agent = MerchantIntelligenceAgent()
seasonality_agent = SeasonalityAgent()
narrative_agent = NarrativeReportAgent()
goals_agent = GoalsAgent()
personal_inflation_agent = PersonalInflationAgent()
silent_growth_agent = SilentGrowthAgent()

@app.get("/health")
async def health_check():
    return {
        "status": "ok",
        "provider": config.LLM_PROVIDER
    }

@app.post("/classify")
async def classify_transaction(tx: TransactionClassifyRequest):
    result = await classifier_service.classify(tx)
    return result

import logging
import asyncio

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

async def run_agent_task(agent_func, user_id, agent_name):
    logger.info(f"Iniciando tarefa do agente {agent_name} para o usuário {user_id}")
    try:
        result = await agent_func(user_id)
        if "error" in result:
            logger.error(f"Erro na tarefa do agente {agent_name}: {result['error']}")
        else:
            logger.info(f"Tarefa do agente {agent_name} concluída com sucesso para o usuário {user_id}")
    except Exception as e:
        logger.error(f"Exceção não tratada na tarefa do agente {agent_name}: {str(e)}")

@app.post("/agents/daily/{user_id}")
async def run_daily_agent(user_id: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(run_agent_task, daily_agent.run, user_id, "diário")
    return {"message": "Processamento do agente diário iniciado"}

@app.post("/agents/weekly/{user_id}")
async def run_weekly_agent(user_id: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(run_agent_task, weekly_agent.run, user_id, "semanal")
    return {"message": "Processamento do agente semanal iniciado"}

@app.post("/agents/monthly/{user_id}")
async def run_monthly_agent(user_id: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(run_agent_task, monthly_agent.run, user_id, "mensal")
    return {"message": "Processamento do agente mensal iniciado"}

@app.post("/chat")
async def chat(req: ChatRequest):
    result = await chat_agent.run(req)
    return result

@app.get("/reports/cashflow/{user_id}")
async def get_cashflow_timeline(user_id: str, from_date: str, to_date: str):
    try:
        from_dt = datetime.fromisoformat(from_date)
        to_dt = datetime.fromisoformat(to_date)
        timeline = await cashflow_agent.build_daily_cashflow(user_id, from_dt, to_dt)
        patterns = await cashflow_agent.detect_cashflow_patterns(user_id)
        return {
            "timeline": timeline,
            "patterns": patterns
        }
    except Exception as e:
        logger.error(f"Erro ao gerar timeline de cashflow: {str(e)}")
        return {"error": str(e)}

@app.get("/reports/behavioral/{user_id}")
async def get_behavioral_insights(user_id: str):
    try:
        result = await behavioral_agent.run(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar insights comportamentais: {str(e)}")
        return {"error": str(e)}

@app.get("/reports/comparison/{user_id}")
async def get_period_comparison(user_id: str, 
                               a_start: str, a_end: str, 
                               b_start: str, b_end: str):
    try:
        period_a = {"start": a_start, "end": a_end}
        period_b = {"start": b_start, "end": b_end}
        result = await comparison_agent.run(user_id, period_a, period_b)
        return result
    except Exception as e:
        logger.error(f"Erro ao comparar períodos: {str(e)}")
        return {"error": str(e)}

@app.get("/reports/invisible-spending/{user_id}")
async def get_invisible_spending(user_id: str):
    try:
        result = await invisible_agent.run(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao detectar gastos invisíveis: {str(e)}")
        return {"error": str(e)}

@app.get("/reports/projection/{user_id}")
async def get_projections(user_id: str):
    try:
        result = await projection_engine.run(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar projeções financeiras: {str(e)}")
        return {"error": str(e)}

@app.get("/reports/health-score/{user_id}")
async def get_health_score(user_id: str):
    try:
        result = await health_score_agent.run(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao calcular health score: {str(e)}")
        return {"error": str(e)}

@app.get("/merchants/{user_id}")
async def get_top_merchants(user_id: str, months: int = 3):
    try:
        result = await merchant_agent.get_top_merchants(user_id, months)
        return result
    except Exception as e:
        logger.error(f"Erro ao buscar top merchants: {str(e)}")
        return {"error": str(e)}

@app.get("/merchants/{user_id}/{merchant_name}")
async def get_merchant_profile(user_id: str, merchant_name: str):
    try:
        result = await merchant_agent.get_merchant_profile(user_id, merchant_name)
        return result
    except Exception as e:
        logger.error(f"Erro ao buscar perfil do merchant: {str(e)}")
        return {"error": str(e)}

@app.get("/reports/upcoming-expenses/{user_id}")
async def get_upcoming_expenses(user_id: str):
    try:
        result = await seasonality_agent.run(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao prever despesas futuras: {str(e)}")
        return {"error": str(e)}

@app.get("/reports/narrative/{user_id}")
async def get_narrative_report(user_id: str, month: int, year: int):
    try:
        result = await narrative_agent.run(user_id, month, year)
        return {"narrative": result}
    except Exception as e:
        logger.error(f"Erro ao gerar relatório narrativo: {str(e)}")
        return {"error": str(e)}

@app.get("/agents/goals/suggest/{user_id}")
async def suggest_goals(user_id: str):
    try:
        result = await goals_agent.suggest_goals(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao sugerir metas: {str(e)}")
        return {"error": str(e)}

@app.post("/agents/personal-inflation/{user_id}")
async def run_personal_inflation_agent(user_id: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(run_agent_task, personal_inflation_agent.run, user_id, "inflação pessoal")
    return {"message": "Processamento do agente de inflação pessoal iniciado"}

@app.post("/agents/silent-growth/{user_id}")
async def run_silent_growth_agent(user_id: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(run_agent_task, silent_growth_agent.run, user_id, "crescimento silencioso")
    return {"message": "Processamento do agente de crescimento silencioso iniciado"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
