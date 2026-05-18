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
from agents.weekly_profile import WeeklyProfileAgent
from agents.monthly_cycle import MonthlyCycleAgent
from agents.behavioral_nuances import BehavioralNuancesAgent
from agents.specific_costs import SpecificCostsAgent
from agents.ticket_analysis import TicketAnalysisAgent
from agents.stress_agent import StressAgent
from agents.loyalty_agent import LoyaltyAgent
from agents.cfo_agent import CFOAgent
from agents.timeline_drift_agent import TimelineDriftAgent
from agents.memory_prediction_agent import MemoryPredictionAgent
from agents.future_leakage_agent import FutureLeakageAgent
from datetime import datetime

app = FastAPI(title="Finance OS Agents Service")
classifier_service = ClassifierService()
daily_agent = DailyAgent()
weekly_agent = WeeklyAgent()
monthly_agent = MonthlyAgent()
chat_agent = ChatAgent()
cfo_agent = CFOAgent()
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
weekly_profile_agent = WeeklyProfileAgent()
monthly_cycle_agent = MonthlyCycleAgent()
behavioral_nuances_agent = BehavioralNuancesAgent()
specific_costs_agent = SpecificCostsAgent()
ticket_analysis_agent = TicketAnalysisAgent()
loyalty_agent = LoyaltyAgent()
stress_agent = StressAgent()
timeline_drift_agent = TimelineDriftAgent()
memory_prediction_agent = MemoryPredictionAgent()
future_leakage_agent = FutureLeakageAgent()

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

@app.get("/reports/stress-score/{user_id}")
async def get_stress_score(user_id: str):
    try:
        result = await stress_agent.calculate_stress_score(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao calcular score de stress: {str(e)}")
        return {"error": str(e)}

@app.get("/reports/survival-mode/{user_id}")
async def get_survival_mode(user_id: str):
    try:
        result = await stress_agent.evaluate_survival_mode(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao avaliar modo sobrevivência: {str(e)}")
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

@app.get("/reports/weekly-profile/{user_id}")
async def get_weekly_profile(user_id: str):
    try:
        result = await weekly_profile_agent.build_day_profiles(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar perfil semanal: {str(e)}")
        return {"error": str(e)}

@app.get("/reports/weekday-weekend/{user_id}")
async def get_weekday_weekend(user_id: str):
    try:
        result = await weekly_profile_agent.compare_weekday_vs_weekend(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar relatório dia útil vs fds: {str(e)}")
        return {"error": str(e)}

@app.get("/reports/impulse/{user_id}")
async def get_impulse_report(user_id: str):
    try:
        result = await behavioral_nuances_agent.run_impulse(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar relatório de impulso: {str(e)}")
        return {"error": str(e)}

@app.get("/reports/compensation/{user_id}")
async def get_compensation_report(user_id: str):
    try:
        result = await behavioral_nuances_agent.run_compensation(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar relatório de compensação: {str(e)}")
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

@app.post("/agents/salary-effect/{user_id}")
async def run_salary_effect_agent(user_id: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(run_agent_task, monthly_cycle_agent.run_salary_effect, user_id, "efeito salário")
    return {"message": "Processamento do agente de efeito salário iniciado"}

@app.post("/agents/monthly-weeks/{user_id}")
async def run_monthly_weeks_agent(user_id: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(run_agent_task, monthly_cycle_agent.run_monthly_weeks, user_id, "semanas mensais")
    return {"message": "Processamento do agente de semanas mensais iniciado"}

@app.post("/agents/weekly-profile/{user_id}")
async def run_weekly_profile_agent(user_id: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(run_agent_task, weekly_profile_agent.run, user_id, "perfil semanal")
    return {"message": "Processamento do agente de perfil semanal iniciado"}

@app.post("/agents/weekday-weekend/{user_id}")
async def run_weekday_weekend_agent(user_id: str, background_tasks: BackgroundTasks):
    # Reutiliza o mesmo agente pois ele calcula ambos
    background_tasks.add_task(run_agent_task, weekly_profile_agent.run, user_id, "dia útil vs fds")
    return {"message": "Processamento do agente de dia útil vs fim de semana iniciado"}

@app.post("/agents/impulse/{user_id}")
async def run_impulse_agent(user_id: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(run_agent_task, behavioral_nuances_agent.run_impulse, user_id, "análise de impulso")
    return {"message": "Processamento do agente de análise de impulso iniciado"}

@app.post("/agents/compensation/{user_id}")
async def run_compensation_agent(user_id: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(run_agent_task, behavioral_nuances_agent.run_compensation, user_id, "padrão de compensação")
    return {"message": "Processamento do agente de padrão de compensação iniciado"}

@app.post("/agents/meal-cost/{user_id}")
async def run_meal_cost_agent(user_id: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(run_agent_task, specific_costs_agent.calculate_real_meal_cost, user_id, "custo real por refeição")
    return {"message": "Processamento do agente de custo real por refeição iniciado"}

@app.post("/agents/convenience-index/{user_id}")
async def run_convenience_index_agent(user_id: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(run_agent_task, specific_costs_agent.calculate_convenience_spending, user_id, "índice de conveniência")
    return {"message": "Processamento do agente de índice de conveniência iniciado"}

@app.post("/agents/ticket-analysis/{user_id}")
async def run_ticket_analysis_agent(user_id: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(run_agent_task, ticket_analysis_agent.run, user_id, "análise de ticket médio")
    return {"message": "Processamento do agente de análise de ticket médio iniciado"}

@app.post("/agents/loyalty/{user_id}")
async def run_loyalty_agent(user_id: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(run_agent_task, loyalty_agent.run, user_id, "análise de lealdade")
    return {"message": "Processamento do agente de análise de lealdade iniciado"}

@app.post("/agents/timeline/{user_id}")
async def run_timeline_agent(user_id: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(run_agent_task, timeline_drift_agent.build_financial_timeline, user_id, "timeline financeira")
    return {"message": "Processamento da timeline financeira iniciado"}

@app.post("/agents/lifestyle-drift/{user_id}")
async def run_lifestyle_drift_agent(user_id: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(run_agent_task, timeline_drift_agent.detect_lifestyle_drift, user_id, "lifestyle drift")
    return {"message": "Processamento de lifestyle drift iniciado"}

@app.post("/chat/explain")
async def explain_spending(req: ChatRequest):
    # O user_id vem no ChatRequest (que tem o mesmo formato)
    result = await cfo_agent.explain_period_spending(req.user_id, req.message)
    return result

@app.post("/agents/cfo/proactive-insight/{user_id}")
async def run_proactive_insight(user_id: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(run_agent_task, cfo_agent.generate_proactive_insights, user_id, "insight proativo CFO")
    return {"message": "Geração de insight proativo iniciada"}

@app.post("/agents/stress-score/{user_id}")
async def run_stress_score_agent(user_id: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(run_agent_task, stress_agent.calculate_stress_score, user_id, "score de stress")
    return {"message": "Processamento do agente de score de stress iniciado"}

@app.post("/agents/survival-mode/{user_id}")
async def run_survival_mode_agent(user_id: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(run_agent_task, stress_agent.evaluate_survival_mode, user_id, "modo sobrevivência")
    return {"message": "Processamento do agente de modo sobrevivência iniciado"}

@app.post("/agents/financial-memory/{user_id}")
async def run_financial_memory_agent(user_id: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(run_agent_task, memory_prediction_agent.generate_memory_insights, user_id, "memória financeira")
    return {"message": "Processamento do agente de memória financeira iniciado"}

@app.post("/agents/dangerous-days/{user_id}")
async def run_dangerous_days_agent(user_id: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(run_agent_task, memory_prediction_agent.generate_preventive_alerts, user_id, "dias perigosos")
    return {"message": "Processamento do agente de dias perigosos iniciado"}

@app.post("/agents/behavioral-prediction/{user_id}")
async def run_behavioral_prediction_agent(user_id: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(run_agent_task, future_leakage_agent.predict_future, user_id, "previsão comportamental")
    return {"message": "Processamento do agente de previsão comportamental iniciado"}

@app.post("/agents/micro-spending/{user_id}")
async def run_micro_spending_agent(user_id: str, background_tasks: BackgroundTasks):
    background_tasks.add_task(run_agent_task, future_leakage_agent.analyze_micro_transactions, user_id, "pequenos vazamentos")
    return {"message": "Processamento do agente de pequenos vazamentos iniciado"}

@app.get("/reports/personal-inflation/{user_id}")
async def get_personal_inflation(user_id: str):
    try:
        result = await personal_inflation_agent.run(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar inflação pessoal: {str(e)}")
        return {"error": str(e)}

@app.post("/reports/silent-growth/{user_id}")
async def get_silent_growth(user_id: str):
    try:
        result = await silent_growth_agent.run(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar crescimento silencioso: {str(e)}")
        return {"error": str(e)}

@app.post("/reports/salary-effect/{user_id}")
async def get_salary_effect(user_id: str):
    try:
        result = await monthly_cycle_agent.run_salary_effect(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar efeito salário: {str(e)}")
        return {"error": str(e)}

@app.post("/reports/monthly-weeks/{user_id}")
async def get_monthly_weeks(user_id: str):
    try:
        result = await monthly_cycle_agent.run_monthly_weeks(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar semanas mensais: {str(e)}")
        return {"error": str(e)}

@app.get("/reports/meal-cost/{user_id}")
async def get_meal_cost(user_id: str):
    try:
        result = await specific_costs_agent.calculate_real_meal_cost(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar custo real por refeição: {str(e)}")
        return {"error": str(e)}

@app.get("/reports/convenience-index/{user_id}")
async def get_convenience_index(user_id: str):
    try:
        result = await specific_costs_agent.calculate_convenience_spending(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar índice de conveniência: {str(e)}")
        return {"error": str(e)}

@app.get("/reports/ticket-analysis/{user_id}")
async def get_ticket_analysis(user_id: str):
    try:
        result = await ticket_analysis_agent.run(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar análise de ticket médio: {str(e)}")
        return {"error": str(e)}

@app.get("/reports/loyalty/{user_id}")
async def get_loyalty_analysis(user_id: str):
    try:
        result = await loyalty_agent.run(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar análise de lealdade: {str(e)}")
        return {"error": str(e)}

@app.get("/reports/financial-memory/{user_id}")
async def get_financial_memory(user_id: str):
    try:
        result = await memory_prediction_agent.generate_memory_insights(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar memória financeira: {str(e)}")
        return {"error": str(e)}

@app.get("/reports/dangerous-days/{user_id}")
async def get_dangerous_days(user_id: str):
    try:
        result = await memory_prediction_agent.generate_preventive_alerts(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar alertas de dias perigosos: {str(e)}")
        return {"error": str(e)}

@app.get("/reports/behavioral-prediction/{user_id}")
async def get_behavioral_prediction(user_id: str):
    try:
        result = await future_leakage_agent.predict_future(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar previsão comportamental: {str(e)}")
        return {"error": str(e)}

@app.get("/reports/micro-spending/{user_id}")
async def get_micro_spending(user_id: str):
    try:
        result = await future_leakage_agent.analyze_micro_transactions(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao analisar pequenos vazamentos: {str(e)}")
        return {"error": str(e)}

@app.get("/reports/timeline/{user_id}")
async def get_financial_timeline(user_id: str, limit: int = 50):
    try:
        conn = await timeline_drift_agent.get_db_connection()
        try:
            rows = await conn.fetch("""
                SELECT event_type, event_date, title, narrative, event_data, created_at
                FROM financial_timeline_events
                WHERE user_id = $1
                ORDER BY event_date DESC, created_at DESC
                LIMIT $2
            """, user_id, limit)
            return [dict(r) for r in rows]
        finally:
            await conn.close()
    except Exception as e:
        logger.error(f"Erro ao buscar timeline financeira: {str(e)}")
        return {"error": str(e)}

@app.get("/reports/lifestyle-drift/{user_id}")
async def get_lifestyle_drift(user_id: str):
    try:
        result = await timeline_drift_agent.detect_lifestyle_drift(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao detectar lifestyle drift: {str(e)}")
        return {"error": str(e)}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
