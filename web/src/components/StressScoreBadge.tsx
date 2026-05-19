"use client";

import { useState, useEffect } from "react";
import api from "@/lib/api";
import { AlertCircle, Brain, TrendingUp, TrendingDown, Minus } from "lucide-react";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

interface StressScore {
  score: number;
  level: string;
  trend: string;
  components: {
    burn_rate: number;
    salary_consumed: number;
    credit_usage: number;
    installments_ratio: number;
    recent_volatility: number;
  };
}

export default function StressScoreBadge() {
  const [data, setData] = useState<StressScore | null>(null);
  const [isOpen, setIsOpen] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchStress() {
      try {
        const resp = await api.get("reports/stress-score");
        setData(resp.data.data);
      } catch (error) {
        console.error("Erro ao buscar score de stress:", error);
      } finally {
        setLoading(false);
      }
    }
    fetchStress();
  }, []);

  if (loading || !data) return null;

  const getLevelColor = (level: string) => {
    switch (level.toLowerCase()) {
      case "tranquilo": return "bg-accent-secondary";
      case "atenção": return "bg-warning";
      case "pressão": return "bg-orange-500";
      case "crítico": return "bg-danger";
      default: return "bg-text-secondary";
    }
  };

  const getTrendIcon = (trend: string) => {
    switch (trend.toLowerCase()) {
      case "melhorando": return <TrendingDown className="w-3 h-3 text-accent-secondary" />;
      case "piorando": return <TrendingUp className="w-3 h-3 text-danger" />;
      default: return <Minus className="w-3 h-3 text-text-secondary" />;
    }
  };

  return (
    <div className="relative">
      <button 
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 px-3 py-1.5 bg-elevated border-2 border-black shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] hover:translate-x-[1px] hover:translate-y-[1px] hover:shadow-none transition-all active:translate-x-[2px] active:translate-y-[2px]"
      >
        <div className={cn("w-2 h-2 rounded-full animate-pulse", getLevelColor(data.level))}></div>
        <span className="text-[10px] font-black uppercase tracking-tighter text-text-primary">
          STRESS: {data.level}
        </span>
        {getTrendIcon(data.trend)}
      </button>

      {isOpen && (
        <>
          <div 
            className="fixed inset-0 z-50" 
            onClick={() => setIsOpen(false)}
          ></div>
          <div className="absolute top-full right-0 mt-2 w-64 bg-elevated border-4 border-black p-4 z-[60] shadow-[8px_8px_0px_0px_rgba(0,0,0,1)]">
            <div className="flex items-center gap-2 mb-4 pb-2 border-b-2 border-black">
              <Brain className="w-4 h-4 text-accent-primary" />
              <h4 className="text-xs font-black uppercase tracking-tight">ANALISE_DE_STRESS</h4>
            </div>

            <div className="space-y-3">
              <StressComponent 
                label="TAXA_DE_BURN" 
                value={data.components.burn_rate} 
                description="Velocidade de consumo do saldo"
              />
              <StressComponent 
                label="CONSUMO_SALARIO" 
                value={data.components.salary_consumed} 
                description="Percentual do salário já gasto"
              />
              <StressComponent 
                label="USO_CREDITO" 
                value={data.components.credit_usage} 
                description="Crédito como % do gasto total"
              />
              <StressComponent 
                label="PARCELAS_RENDA" 
                value={data.components.installments_ratio} 
                description="Impacto estrutural das parcelas"
              />
              <StressComponent 
                label="VOLATILIDADE" 
                value={data.components.recent_volatility} 
                description="Desvio padrão dos gastos diários"
              />
            </div>

            <div className="mt-4 pt-2 border-t-2 border-black">
              <div className="text-[10px] font-bold text-text-secondary uppercase">SCORE_FINAL</div>
              <div className="flex items-end justify-between">
                <span className="text-2xl font-black font-mono">{Math.round(data.score)}/100</span>
                <span className={cn("text-[8px] font-black px-1.5 py-0.5 uppercase text-white", getLevelColor(data.level))}>
                  {data.level}
                </span>
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  );
}

function StressComponent({ label, value, description }: { label: string, value: number, description: string }) {
  // Score de stress aqui: quanto MAIOR o valor, MAIOR o stress (dependendo de como o backend calculou)
  // Mas no FASE13 diz: "Score de 0 (crítico) a 100 (tranquilo)"
  // Então se o score final é 100, os componentes também devem ser "bons".
  // Vou assumir que value é de 0 a 100 onde 100 é BOM.
  
  const getBarColor = (val: number) => {
    if (val >= 80) return "bg-accent-secondary";
    if (val >= 50) return "bg-warning";
    return "bg-danger";
  };

  return (
    <div className="group">
      <div className="flex justify-between items-center mb-1">
        <span className="text-[8px] font-black uppercase tracking-widest">{label}</span>
        <span className="text-[10px] font-mono font-bold">{Math.round(value)}%</span>
      </div>
      <div className="h-1.5 w-full bg-black/20 border border-black overflow-hidden">
        <div 
          className={cn("h-full transition-all duration-500", getBarColor(value))} 
          style={{ width: `${value}%` }}
        ></div>
      </div>
      <p className="text-[7px] font-bold text-text-secondary uppercase mt-1 hidden group-hover:block leading-none">
        {description}
      </p>
    </div>
  );
}
