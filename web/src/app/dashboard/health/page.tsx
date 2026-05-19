import { auth } from "@/auth";
import apiServer from "@/lib/api-server";
import { Terminal, Activity, TrendingUp, TrendingDown, Target, ShieldCheck, Zap } from "lucide-react";
import HealthScoreGauge from "@/components/HealthScoreGauge";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

async function getHealthScore(token: string) {
  try {
    const resp = await apiServer.get("/reports/health-score", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data.data;
  } catch (error) {
    return null;
  }
}

export default async function HealthPage() {
  const session = await auth() as any;
  const token = session?.accessToken;
  const healthData = await getHealthScore(token);

  if (!healthData) {
    return (
      <div className="p-8 text-center">
        <h1 className="text-2xl font-black uppercase">ERRO_AO_CARREGAR_SCORE</h1>
      </div>
    );
  }

  // Mapeamento de dimensões vindo do Agente Python ou do Snapshot do banco
  const dims = healthData.dimensions || {
    cashflow: healthData.cashflow_score,
    installments: healthData.installments_score,
    consistency: healthData.consistency_score,
    subscriptions: healthData.subscriptions_score,
    diversification: healthData.diversification_score,
    trend: healthData.trend_score,
  };

  const dimensions = [
    { label: "FLUXO_DE_CAIXA", score: dims.cashflow, icon: Activity, weight: "25%" },
    { label: "PARCELAMENTOS", score: dims.installments, icon: Target, weight: "20%" },
    { label: "CONSISTÊNCIA", score: dims.consistency, icon: ShieldCheck, weight: "20%" },
    { label: "ASSINATURAS", score: dims.subscriptions, icon: Zap, weight: "15%" },
    { label: "DIVERSIFICAÇÃO", score: dims.diversification, icon: TrendingUp, weight: "10%" },
    { label: "TENDÊNCIA", score: dims.trend, icon: TrendingDown, weight: "10%" },
  ];

  return (
    <div className="grid-blueprint grid-cols-1">
      <header className="p-8 bg-elevated border-b-2 border-black">
        <div className="flex items-center gap-2 mb-1">
          <Terminal className="w-4 h-4" />
          <span className="text-[10px] font-bold tracking-widest uppercase">HEALTH_DIAGNOSTIC_V1.0</span>
        </div>
        <h1 className="text-4xl font-black text-text-primary uppercase tracking-tighter">SCORE_DE_SAÚDE_FINANCEIRA</h1>
      </header>

      <div className="grid grid-cols-1 lg:grid-cols-3 bg-background border-b-2 border-black">
        <div className="p-8 border-r-2 border-black flex flex-col items-center justify-center bg-elevated/20">
          <HealthScoreGauge score={Math.round(healthData.score)} />
          <p className="mt-8 text-center text-xs font-bold text-text-secondary uppercase max-w-[200px]">
            Seu score é baseado em padrões reais dos seus últimos 90 dias.
          </p>
        </div>

        <div className="lg:col-span-2 p-8 grid grid-cols-1 md:grid-cols-2 gap-8">
          {dimensions.map((dim) => (
            <div key={dim.label} className="border-2 border-black p-4 bg-background shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
              <div className="flex justify-between items-start mb-4">
                <div className="flex items-center gap-2">
                  <dim.icon className="w-4 h-4 text-accent-primary" />
                  <span className="text-[10px] font-black uppercase tracking-widest text-text-secondary">{dim.label}</span>
                </div>
                <span className="text-[8px] font-bold bg-black text-white px-1.5 py-0.5 uppercase">PESO: {dim.weight}</span>
              </div>
              <div className="flex items-end justify-between">
                <span className="text-3xl font-black font-mono">{Math.round(dim.score)}/100</span>
                <div className="w-24 h-2 bg-[#2A2A3A] mb-2">
                  <div 
                    className={cn(
                      "h-full transition-all",
                      dim.score >= 80 ? "bg-accent-secondary" : dim.score >= 60 ? "bg-warning" : "bg-danger"
                    )}
                    style={{ width: `${dim.score}%` }}
                  />
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      <section className="p-8 bg-background">
        <h3 className="text-xl font-black uppercase tracking-tighter mb-8 flex items-center gap-2">
          <Zap className="w-5 h-5 fill-current text-accent-primary" />
          RECOMENDAÇÕES_ESTRATÉGICAS
        </h3>
        
        {typeof healthData.recommendations === 'string' ? (
          <div className="p-8 border-2 border-black bg-elevated shadow-[6px_6px_0px_0px_rgba(0,0,0,1)]">
            <div className="text-lg font-black leading-relaxed italic markdown-content">
              <ReactMarkdown remarkPlugins={[remarkGfm]}>
                {healthData.recommendations}
              </ReactMarkdown>
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {healthData.recommendations?.map((rec: any, i: number) => (
              <div key={i} className="p-6 border-2 border-black bg-elevated hover:bg-background transition-all group relative overflow-hidden">
                <div className={cn(
                  "absolute top-0 left-0 w-1 h-full bg-black transition-colors",
                  rec.severity === 'high' ? 'group-hover:bg-danger' : 'group-hover:bg-accent-primary'
                )}></div>
                <div className="text-[10px] font-black text-text-secondary mb-2 uppercase tracking-widest">
                  DIMENSÃO: {rec.dimension}
                </div>
                <h4 className="font-black uppercase text-sm mb-3">{rec.title}</h4>
                <p className="text-sm font-bold text-text-secondary leading-relaxed">
                  {rec.description}
                </p>
                {rec.impact && (
                  <div className="mt-4 pt-4 border-t border-black/10">
                    <span className="text-[10px] font-black text-accent-secondary uppercase">IMPACTO_ESTIMADO: +{rec.impact} PONTOS</span>
                  </div>
                )}
              </div>
            ))}
            {(!healthData.recommendations || healthData.recommendations.length === 0) && (
              <div className="col-span-3 p-12 border-2 border-dashed border-black text-center text-[10px] font-bold uppercase opacity-50">
                ANALISANDO_OPORTUNIDADES_ADICIONAIS...
              </div>
            )}
          </div>
        )}
      </section>

      <footer className="p-8 border-t-2 border-black bg-elevated flex justify-between items-center text-[10px] font-black uppercase tracking-[0.2em]">
         <span>HEALTH_SCORE_ENGINE // ONLINE</span>
         <span>PROXIMO_CALCULO: 24H_CYLE</span>
      </footer>
    </div>
  );
}
