import { auth } from "@/auth";
import apiServer from "@/lib/api-server";
import { format } from "date-fns";
import { ptBR } from "date-fns/locale";
import { FileText, Calendar, Zap, Terminal, Activity, BookOpen, TrendingUp, Clock, MousePointer2, Award, Target, Anchor } from "lucide-react";
import TriggerReportButtons from "@/components/TriggerReportButtons";
import NarrativeReportView from "@/components/NarrativeReportView";
import PersonalInflationCard from "@/components/PersonalInflationCard";
import SilentGrowthCard from "@/components/SilentGrowthCard";
import WeeklyProfileChart from "@/components/WeeklyProfileChart";
import MonthlyCycleChart from "@/components/MonthlyCycleChart";
import ImpulseAnalysisCard from "@/components/ImpulseAnalysisCard";
import CompensationPatternCard from "@/components/CompensationPatternCard";
import MealCostAnalysisCard from "@/components/MealCostAnalysisCard";
import ConvenienceIndexCard from "@/components/ConvenienceIndexCard";
import TicketAnalysisCard from "@/components/TicketAnalysisCard";
import LoyaltyAnalysisCard from "@/components/LoyaltyAnalysisCard";
import AchievementsFeed from "@/components/AchievementsFeed";
import MissionsCard from "@/components/MissionsCard";
import InstallmentTimeline from "@/components/InstallmentTimeline";
import { DependencyTreemap } from "@/components/DependencyTreemap";
import { SpendingHeatmap } from "@/components/SpendingHeatmap";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

async function getReports(token: string) {
  try {
    const resp = await apiServer.get("reports", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data?.data || [];
  } catch (error) {
    return [];
  }
}

async function getTicketAnalysis(token: string) {
  try {
    const resp = await apiServer.get("reports/ticket-analysis", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data?.data;
  } catch (error) {
    return null;
  }
}

async function getLoyalty(token: string) {
  try {
    const resp = await apiServer.get("reports/loyalty", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data?.data;
  } catch (error) {
    return null;
  }
}

async function getPersonalInflation(token: string) {
  try {
    const resp = await apiServer.get("reports/personal-inflation", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data?.data;
  } catch (error) {
    return null;
  }
}

async function getSilentGrowth(token: string) {
  try {
    const resp = await apiServer.get("reports/silent-growth", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data?.data;
  } catch (error) {
    return null;
  }
}

async function getWeeklyProfile(token: string) {
  try {
    const resp = await apiServer.get("reports/weekly-profile", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data?.data;
  } catch (error) {
    return null;
  }
}

async function getWeekdayWeekend(token: string) {
  try {
    const resp = await apiServer.get("reports/weekday-weekend", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data?.data;
  } catch (error) {
    return null;
  }
}

async function getSalaryEffect(token: string) {
  try {
    const resp = await apiServer.get("reports/salary-effect", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data?.data;
  } catch (error) {
    return null;
  }
}

async function getMonthlyWeeks(token: string) {
  try {
    const resp = await apiServer.get("reports/monthly-weeks", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data?.data;
  } catch (error) {
    return null;
  }
}

async function getImpulse(token: string) {
  try {
    const resp = await apiServer.get("reports/impulse", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data?.data;
  } catch (error) {
    return null;
  }
}

async function getCompensation(token: string) {
  try {
    const resp = await apiServer.get("reports/compensation-pattern", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data?.data;
  } catch (error) {
    return null;
  }
}

async function getMealCost(token: string) {
  try {
    const resp = await apiServer.get("reports/meal-cost", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data?.data;
  } catch (error) {
    return null;
  }
}

async function getConvenience(token: string) {
  try {
    const resp = await apiServer.get("reports/convenience-index", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data?.data;
  } catch (error) {
    return null;
  }
}

async function getGamification(token: string) {
  try {
    const resp = await apiServer.get("reports/gamification", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data?.data;
  } catch (error) {
    return null;
  }
}

async function getSalaryPlan(token: string) {
  try {
    const resp = await apiServer.get("reports/salary-plan", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data?.data;
  } catch (error) {
    return null;
  }
}

async function getInstallmentTimeline(token: string) {
  try {
    const resp = await apiServer.get("reports/installment-timeline", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data?.data;
  } catch (error) {
    return null;
  }
}

async function getDependencyMap(token: string) {
  try {
    const resp = await apiServer.get("reports/dependency-map", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data?.data;
  } catch (error) {
    return null;
  }
}

async function getSpendingHeatmap(token: string) {
  try {
    const resp = await apiServer.get("reports/spending-heatmap", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data?.data;
  } catch (error) {
    return null;
  }
}

export default async function ReportsPage() {
  const session = await auth() as any;
  const token = session?.accessToken;
  const [
    reports, 
    inflation, 
    growth, 
    weeklyProfile, 
    weekdayWeekend, 
    salaryEffect, 
    monthlyWeeks,
    impulse,
    compensation,
    mealCost,
    convenience,
    ticketAnalysis,
    loyalty,
    gamification,
    salaryPlan,
    installmentTimeline,
    dependencyMap,
    spendingHeatmap
  ] = await Promise.all([
    getReports(token),
    getPersonalInflation(token),
    getSilentGrowth(token),
    getWeeklyProfile(token),
    getWeekdayWeekend(token),
    getSalaryEffect(token),
    getMonthlyWeeks(token),
    getImpulse(token),
    getCompensation(token),
    getMealCost(token),
    getConvenience(token),
    getTicketAnalysis(token),
    getLoyalty(token),
    getGamification(token),
    getSalaryPlan(token),
    getInstallmentTimeline(token),
    getDependencyMap(token),
    getSpendingHeatmap(token)
  ]);

  return (
    <div className="grid-blueprint grid-cols-1">
      <header className="p-8 bg-elevated border-b-2 border-black flex flex-col md:flex-row md:items-center justify-between gap-6">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <Terminal className="w-4 h-4" />
            <span className="text-[10px] font-bold tracking-widest uppercase">ANALYTICS_KERNEL_V1.0</span>
          </div>
          <h1 className="text-4xl font-black text-text-primary uppercase tracking-tighter">RELATORIOS_DE_IA</h1>
          <p className="text-sm font-bold text-text-secondary uppercase">TOTAL_ANALISES_DISPONIVEIS: {reports.length}</p>
        </div>
        <TriggerReportButtons />
      </header>

      <div className="bg-background">
         <div className="border-b-2 border-black bg-elevated/10">
            <NarrativeReportView token={token} />
         </div>
      </div>

      {/* Gamification & Rewards Section */}
      {(gamification || salaryPlan) && (
        <div className="p-8 md:p-12 bg-background border-b-2 border-black space-y-8">
           <div className="flex items-center gap-2 text-accent-primary font-black text-[10px] uppercase tracking-[0.3em]">
            <Award className="w-4 h-4" /> REWARDS_AND_ACHIEVEMENTS_V1.0
          </div>
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            <div className="lg:col-span-2">
              <AchievementsFeed achievements={gamification?.awarded_achievements} />
            </div>
            <div>
              <MissionsCard missions={gamification?.active_missions} />
            </div>
          </div>

          {salaryPlan && (
            <div className="p-6 border-2 border-black bg-elevated shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
              <div className="flex items-center gap-2 mb-4">
                <Target className="w-4 h-4 text-accent-primary" />
                <h3 className="text-[10px] font-black uppercase tracking-widest">PLANO_DE_SALARIO_INTELIGENTE</h3>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                 <div className="space-y-4">
                    <div className="text-sm font-bold leading-relaxed markdown-content">
                       <ReactMarkdown remarkPlugins={[remarkGfm]}>
                          {salaryPlan.narrative}
                       </ReactMarkdown>
                    </div>
                 </div>
                 <div className="space-y-4">
                    <div className="text-[10px] font-black uppercase text-text-secondary border-b border-black pb-2">DISTRIBUICAO_RECOMENDADA</div>
                    <div className="space-y-3">
                       {salaryPlan.allocation && Object.entries(salaryPlan.allocation).map(([key, val]: [string, any]) => (
                         <div key={key} className="flex justify-between items-center">
                            <span className="text-[10px] font-bold uppercase">{key}</span>
                            <span className="text-xs font-black font-mono text-accent-primary">{new Intl.NumberFormat('pt-BR', { style: 'percent' }).format(val / 100)}</span>
                         </div>
                       ))}
                    </div>
                 </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Installment Timeline Section */}
      {installmentTimeline && (
        <div className="p-8 md:p-12 bg-background border-b-2 border-black">
          <InstallmentTimeline data={installmentTimeline} />
        </div>
      )}

      {/* Advanced Analytic Cards Section */}
      {(inflation || growth || weeklyProfile || monthlyWeeks || impulse || compensation || mealCost || convenience || ticketAnalysis || loyalty) && (
        <div className="p-8 md:p-12 bg-background border-b-2 border-black space-y-12">
          <div className="flex items-center gap-2 text-accent-primary font-black text-[10px] uppercase tracking-[0.3em]">
            <TrendingUp className="w-4 h-4" /> ADVANCED_PATTERN_INSIGHTS_V2.5
          </div>
          
          <div className="grid grid-cols-1 xl:grid-cols-2 gap-12">
            {inflation && <PersonalInflationCard data={inflation} />}
            {growth && <SilentGrowthCard data={growth} />}
            {weeklyProfile && <WeeklyProfileChart data={weeklyProfile} weekdayWeekendData={weekdayWeekend} />}
            {monthlyWeeks && <MonthlyCycleChart data={monthlyWeeks} salaryEffectData={salaryEffect} />}
            {impulse && <ImpulseAnalysisCard data={impulse} />}
            {compensation && <CompensationPatternCard data={compensation} />}
            {mealCost && <MealCostAnalysisCard data={mealCost} />}
            {convenience && <ConvenienceIndexCard data={convenience} />}
            {ticketAnalysis && <TicketAnalysisCard data={ticketAnalysis} />}
            {loyalty && <LoyaltyAnalysisCard data={loyalty} />}
          </div>
        </div>
      )}

      {/* Visual Premium Section */}
      {(spendingHeatmap || dependencyMap) && (
        <div className="p-8 md:p-12 bg-background border-b-2 border-black space-y-12">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 text-accent-primary font-black text-[10px] uppercase tracking-[0.3em]">
              <Target className="w-4 h-4" /> VISUAL_PREMIUM_V1.0
            </div>
            <a 
              href="/dashboard/reports/replay/2026-05" 
              className="text-xs font-bold uppercase border-2 border-black px-4 py-2 hover:bg-black hover:text-white transition-colors"
            >
              Replay Financeiro (Abril 2026)
            </a>
          </div>
          
          <div className="grid grid-cols-1 gap-12">
            {spendingHeatmap && <SpendingHeatmap data={spendingHeatmap} />}
            {dependencyMap && <DependencyTreemap data={dependencyMap} />}
          </div>
        </div>
      )}

      <div className="bg-background min-h-screen">
        {reports.length > 0 ? reports.map((report: any) => (
          <div key={report.id} className="border-b-2 border-black last:border-b-0 hover:bg-elevated/20 transition-colors">
            <div className="p-8 md:p-12 space-y-10">
              {/* Report Meta */}
              <div className="flex flex-col md:flex-row md:items-center justify-between gap-6">
                <div className="flex items-center gap-6">
                   <div className="w-16 h-16 border-2 border-black flex items-center justify-center bg-accent-primary text-white shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
                      <FileText className="w-8 h-8" />
                   </div>
                   <div>
                      <div className="text-[10px] font-black uppercase tracking-widest text-text-secondary mb-1">REPORT_TYPE_ID</div>
                      <h2 className="text-2xl font-black uppercase tracking-tighter">
                         {report.agent_type === 'daily' ? 'DAILY_FINANCIAL_SNAPSHOT' : 'SPECIAL_ANALYTIC_REPORT'}
                      </h2>
                   </div>
                </div>
                <div className="flex flex-col md:items-end">
                   <div className="text-[10px] font-black uppercase text-text-secondary mb-1">GENERATION_TIMESTAMP</div>
                   <div className="flex items-center gap-2 font-bold text-xs uppercase">
                      <Calendar className="w-3 h-3" />
                      {format(new Date(report.created_at), "dd.MM.yyyy HH:mm", { locale: ptBR })}
                   </div>
                </div>
              </div>

              {/* Summary Block */}
              <div className="space-y-4">
                 <div className="flex items-center gap-2 text-text-secondary font-black text-[10px] uppercase tracking-widest">
                    <Activity className="w-3 h-3" /> EXECUTIVE_SUMMARY
                 </div>
                 <div className="p-8 border-2 border-black bg-background shadow-[6px_6px_0px_0px_rgba(0,0,0,1)] relative overflow-hidden">
                    <div className="absolute top-0 right-0 p-2 opacity-5">
                       <FileText className="w-32 h-32" />
                    </div>
                    <div className="text-xl font-black leading-relaxed italic relative z-10 markdown-content">
                       <ReactMarkdown remarkPlugins={[remarkGfm]}>
                          {report.summary_markdown}
                       </ReactMarkdown>
                    </div>
                 </div>
              </div>

              {/* Insights Grid */}
              {report.insights && Array.isArray(report.insights) && report.insights.length > 0 && (
                <div className="space-y-6">
                  <div className="flex items-center gap-2 text-accent-primary font-black text-[10px] uppercase tracking-widest">
                    <Zap className="w-4 h-4 fill-current" /> DETAILED_HEURISTIC_INSIGHTS
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {report.insights.map((insight: string, i: number) => (
                      <div key={i} className="p-6 border-2 border-black bg-elevated hover:bg-background transition-all group relative overflow-hidden">
                        <div className="absolute top-0 left-0 w-1 h-full bg-black group-hover:bg-accent-primary transition-colors"></div>
                        <div className="text-[10px] font-black text-text-secondary mb-4">INSIGHT_REF_#0{i+1}</div>
                        <div className="text-sm font-bold leading-relaxed markdown-content">
                          <ReactMarkdown remarkPlugins={[remarkGfm]}>
                             {insight}
                          </ReactMarkdown>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        )) : (
          <div className="py-32 flex flex-col items-center justify-center text-center">
            <div className="w-24 h-24 border-2 border-black flex items-center justify-center mb-8 text-4xl grayscale opacity-20 animate-pulse bg-elevated">?</div>
            <h3 className="text-text-primary font-black uppercase text-2xl mb-4 tracking-tighter">ANALYTIC_BUFFER_EMPTY</h3>
            <p className="text-text-secondary text-xs font-bold uppercase max-w-sm leading-relaxed px-8">
              O núcleo de processamento do Pierre ainda não gerou relatórios para sua conta. Continue registrando operações para inicializar o fluxo.
            </p>
          </div>
        )}
      </div>

      <footer className="p-8 border-t-2 border-black bg-black text-white flex justify-between items-center text-[10px] font-black uppercase tracking-[0.2em]">
         <span>CORE_ANALYTICS // STATUS_STABLE</span>
         <span>END_OF_TRANSMISSION</span>
      </footer>
    </div>
  );
}
