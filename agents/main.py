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
from agents.installment_timeline_agent import InstallmentTimelineAgent
from agents.memory_prediction_agent import MemoryPredictionAgent
from agents.future_leakage_agent import FutureLeakageAgent
from agents.gamification_agent import GamificationAgent
from agents.salary_agent import SalaryPlannerAgent
from agents.dependency_map_agent import DependencyMapAgent
from agents.monthly_replay_agent import MonthlyReplayAgent
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
installment_timeline_agent = InstallmentTimelineAgent()
memory_prediction_agent = MemoryPredictionAgent()
future_leakage_agent = FutureLeakageAgent()
gamification_agent = GamificationAgent()
salary_agent = SalaryPlannerAgent()
dependency_map_agent = DependencyMapAgent()
monthly_replay_agent = MonthlyReplayAgent()

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
from starlette.middleware.base import BaseHTTPMiddleware
from app.services.fallback_provider import current_user_id

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class UserContextMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request, call_next):
        user_id = None
        # Tenta extrair das partes do caminho (UUID4)
        path_parts = request.url.path.strip("/").split("/")
        for part in path_parts:
            if len(part) == 36 and part.count("-") == 4:
                user_id = part
                break
        
        # Tenta extrair dos parâmetros de consulta
        if not user_id:
            user_id = request.query_params.get("user_id")
            
        token = None
        if user_id:
            token = current_user_id.set(user_id)
            logger.info(f"[MIDDLEWARE] Contexto de usuário definido: {user_id}")
            
        try:
            return await call_next(request)
        finally:
            if token:
                current_user_id.reset(token)

app.add_middleware(UserContextMiddleware)

async def run_agent_task(agent_func, user_id, agent_name):
    logger.info(f"Iniciando tarefa do agente {agent_name} para o usuário {user_id}")
    token = current_user_id.set(user_id)
    try:
        result = await agent_func(user_id)
        if result and isinstance(result, dict) and "error" in result:
            logger.error(f"Erro na tarefa do agente {agent_name}: {result['error']}")
        else:
            logger.info(f"Tarefa do agente {agent_name} concluída com sucesso para o usuário {user_id}")
    except Exception as e:
        logger.error(f"Exceção não tratada na tarefa do agente {agent_name}: {str(e)}")
    finally:
        current_user_id.reset(token)

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
    token = current_user_id.set(req.user_id)
    try:
        result = await chat_agent.run(req)
        return result
    finally:
        current_user_id.reset(token)

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

@app.post("/reports/personal-inflation/{user_id}")
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

@app.post("/reports/meal-cost/{user_id}")
async def get_meal_cost(user_id: str):
    try:
        result = await specific_costs_agent.calculate_real_meal_cost(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar custo real por refeição: {str(e)}")
        return {"error": str(e)}

@app.post("/reports/convenience-index/{user_id}")
async def get_convenience_index(user_id: str):
    try:
        result = await specific_costs_agent.calculate_convenience_spending(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar índice de conveniência: {str(e)}")
        return {"error": str(e)}

@app.post("/reports/ticket-analysis/{user_id}")
async def get_ticket_analysis(user_id: str):
    try:
        result = await ticket_analysis_agent.run(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar análise de ticket médio: {str(e)}")
        return {"error": str(e)}

@app.post("/reports/loyalty/{user_id}")
async def get_loyalty_analysis(user_id: str):
    try:
        result = await loyalty_agent.run(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar análise de lealdade: {str(e)}")
        return {"error": str(e)}

@app.post("/reports/stress/{user_id}")
async def get_stress_score_report(user_id: str):
    try:
        result = await stress_agent.calculate_stress_score(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao calcular score de stress: {str(e)}")
        return {"error": str(e)}

@app.post("/reports/survival-mode/{user_id}")
async def get_survival_mode_report(user_id: str):
    try:
        result = await stress_agent.evaluate_survival_mode(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao avaliar modo sobrevivência: {str(e)}")
        return {"error": str(e)}

@app.post("/reports/memory/{user_id}")
async def get_financial_memory(user_id: str):
    try:
        result = await memory_prediction_agent.generate_memory_insights(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar memória financeira: {str(e)}")
        return {"error": str(e)}

@app.post("/reports/dangerous-days/{user_id}")
async def get_dangerous_days(user_id: str):
    try:
        result = await memory_prediction_agent.generate_preventive_alerts(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar alertas de dias perigosos: {str(e)}")
        return {"error": str(e)}

@app.post("/reports/prediction/{user_id}")
async def get_behavioral_prediction(user_id: str):
    try:
        result = await future_leakage_agent.predict_future(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar previsão comportamental: {str(e)}")
        return {"error": str(e)}

@app.post("/reports/micro-spending/{user_id}")
async def get_micro_spending(user_id: str):
    try:
        result = await future_leakage_agent.analyze_micro_transactions(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao analisar pequenos vazamentos: {str(e)}")
        return {"error": str(e)}

@app.post("/reports/timeline/{user_id}")
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

@app.post("/reports/lifestyle-drift/{user_id}")
async def get_lifestyle_drift_report(user_id: str):
    try:
        result = await timeline_drift_agent.detect_lifestyle_drift(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao detectar lifestyle drift: {str(e)}")
        return {"error": str(e)}

@app.post("/reports/installment-timeline/{user_id}")
async def get_installment_timeline_report(user_id: str):
    try:
        result = await installment_timeline_agent.generate_installment_timeline(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar timeline de parcelamentos: {str(e)}")
        return {"error": str(e)}

@app.post("/reports/gamification/{user_id}")
async def get_gamification_report(user_id: str):
    try:
        # Executa o rastreamento de missões, checagem de conquistas e geração de novas missões se nulas
        try:
            await gamification_agent.run(user_id)
        except Exception as run_err:
            logger.error(f"Erro ao executar agente de gamificação para o usuário {user_id}: {str(run_err)}")
            # Continuamos para tentar ler o que já existe no banco caso a geração falhe

        conn = await gamification_agent.get_db_connection()
        try:
            # Pegar conquistas do mês atual
            achievements = await conn.fetch("""
                SELECT achievement_id, awarded_at, context_data
                FROM achievements_awarded
                WHERE user_id = $1 AND awarded_at >= date_trunc('month', NOW())
                ORDER BY awarded_at DESC
            """, user_id)
            
            # Pegar missões ativas
            missions = await conn.fetch("""
                SELECT id, template_id, title, description, target_value, current_value, status, ends_at
                FROM missions
                WHERE user_id = $1 AND status = 'active'
                ORDER BY ends_at ASC
            """, user_id)
            
            return {
                "achievements": [dict(a) for a in achievements],
                "missions": [dict(m) for m in missions]
            }
        finally:
            await conn.close()
    except Exception as e:
        logger.error(f"Erro ao buscar relatório de gamificação: {str(e)}")
        return {"error": str(e)}

@app.post("/reports/salary-plan/{user_id}")
async def get_salary_plan(user_id: str):
    try:
        conn = await salary_agent.get_db_connection()
        try:
            row = await conn.fetchrow("""
                SELECT salary_detected, fixed_commitments, safe_daily_limit, plan_data, valid_until, generated_at
                FROM salary_plans
                WHERE user_id = $1
                ORDER BY generated_at DESC LIMIT 1
            """, user_id)
            if row:
                result = dict(row)
                if isinstance(result['plan_data'], str):
                    result['plan_data'] = json.loads(result['plan_data'])
                return result
            return {"error": "No salary plan found"}
        finally:
            await conn.close()
    except Exception as e:
        logger.error(f"Erro ao buscar planejamento de salário: {str(e)}")
        return {"error": str(e)}

@app.get("/reports/dependency-map/{user_id}")
async def get_dependency_map(user_id: str):
    try:
        result = await dependency_map_agent.run(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar dependency map: {str(e)}")
        return {"error": str(e)}

@app.get("/merchants/{user_id}")
async def get_top_merchants(user_id: str, months: int = 3):
    try:
        result = await merchant_agent.get_top_merchants(user_id, period_months=months)
        return result
    except Exception as e:
        logger.error(f"Erro ao obter principais estabelecimentos: {str(e)}")
        return {"error": str(e)}

@app.get("/merchants/{user_id}/{merchant_name}")
async def get_merchant_profile(user_id: str, merchant_name: str):
    try:
        result = await merchant_agent.get_merchant_profile(user_id, merchant_name)
        return result
    except Exception as e:
        logger.error(f"Erro ao obter perfil de estabelecimento: {str(e)}")
        return {"error": str(e)}

@app.post("/agents/monthly-replay/{user_id}")
async def generate_monthly_replay(user_id: str, month: str):
    try:
        result = await monthly_replay_agent.run(user_id, month)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar monthly replay: {str(e)}")
        return {"error": str(e)}

@app.get("/reports/monthly-replay/{user_id}")
async def get_monthly_replay(user_id: str, month: str):
    # This also generates if not found (since it runs the agent)
    try:
        result = await monthly_replay_agent.run(user_id, month)
        return result
    except Exception as e:
        logger.error(f"Erro ao buscar monthly replay: {str(e)}")
        return {"error": str(e)}

@app.post("/reports/weekly-profile/{user_id}")
async def get_weekly_profile(user_id: str):
    try:
        result = await weekly_profile_agent.run(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar perfil semanal: {str(e)}")
        return {"error": str(e)}

@app.post("/reports/weekday-weekend/{user_id}")
async def get_weekday_weekend(user_id: str):
    try:
        result = await weekly_profile_agent.compare_weekday_vs_weekend(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar dia util vs fim de semana: {str(e)}")
        return {"error": str(e)}

@app.post("/reports/impulse/{user_id}")
async def get_impulse_report(user_id: str):
    try:
        result = await behavioral_nuances_agent.run_impulse(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar relatório de impulsividade: {str(e)}")
        return {"error": str(e)}

@app.post("/reports/compensation/{user_id}")
async def get_compensation_report(user_id: str):
    try:
        result = await behavioral_nuances_agent.run_compensation(user_id)
        return result
    except Exception as e:
        logger.error(f"Erro ao gerar relatório de compensação: {str(e)}")
        return {"error": str(e)}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
