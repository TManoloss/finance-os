"use client";

import React from "react";
import { TrendingUp, AlertTriangle, Activity, Percent } from "lucide-react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

interface CategoryInflation {
  category: string;
  variation_percent: number;
  impact_reais: number;
  current_amount: number;
  previous_amount: number;
}

interface PersonalInflationProps {
  data: {
    personal_inflation_rate: number;
    ipca_comparison: string;
    top_categories_inflation: CategoryInflation[];
    insights: string;
  };
}

export default function PersonalInflationCard({ data }: PersonalInflationProps) {
  if (!data) return null;

  return (
    <div className="border-2 border-black bg-background shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] overflow-hidden">
      <div className="p-6 bg-accent-primary border-b-2 border-black flex items-center justify-between">
        <div className="flex items-center gap-3">
          <TrendingUp className="w-6 h-6 text-white" />
          <h2 className="text-xl font-black text-white uppercase tracking-tighter">INFLACAO_PESSOAL_DETECTION</h2>
        </div>
        <div className="px-3 py-1 bg-black text-white text-[10px] font-black uppercase tracking-widest">
          CORE_ANALYSIS_V2
        </div>
      </div>

      <div className="p-8 space-y-8">
        {/* Main Stats */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="p-6 border-2 border-black bg-elevated relative overflow-hidden group">
            <div className="absolute -right-4 -bottom-4 opacity-5 group-hover:opacity-10 transition-opacity">
              <Percent className="w-24 h-24" />
            </div>
            <div className="text-[10px] font-black text-text-secondary uppercase mb-2">TAXA_PONDERADA_MENSAL</div>
            <div className="text-4xl font-black text-accent-primary">{data.personal_inflation_rate.toFixed(2)}%</div>
            <div className="mt-4 text-xs font-bold text-text-secondary uppercase">{data.ipca_comparison}</div>
          </div>

          <div className="p-6 border-2 border-black bg-elevated relative overflow-hidden group">
            <div className="absolute -right-4 -bottom-4 opacity-5 group-hover:opacity-10 transition-opacity">
              <Activity className="w-24 h-24" />
            </div>
            <div className="text-[10px] font-black text-text-secondary uppercase mb-2">STATUS_DE_ALERTA</div>
            <div className="flex items-center gap-2 text-2xl font-black uppercase tracking-tighter">
              {data.personal_inflation_rate > 5 ? (
                <>
                  <AlertTriangle className="w-6 h-6 text-danger" />
                  <span className="text-danger">RISCO_ALTO</span>
                </>
              ) : (
                <>
                  <Activity className="w-6 h-6 text-success" />
                  <span className="text-success">NOMINAL</span>
                </>
              )}
            </div>
          </div>
        </div>

        {/* Narrative Insight */}
        <div className="space-y-4">
          <div className="flex items-center gap-2 text-text-secondary font-black text-[10px] uppercase tracking-widest">
            <Activity className="w-3 h-3" /> IA_NARRATIVE_INSIGHT
          </div>
          <div className="p-6 border-2 border-black bg-background shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] italic text-lg font-bold leading-relaxed markdown-content">
            <ReactMarkdown remarkPlugins={[remarkGfm]}>
              {data.insights}
            </ReactMarkdown>
          </div>
        </div>

        {/* Categories Breakdown */}
        <div className="space-y-4">
          <div className="flex items-center gap-2 text-text-secondary font-black text-[10px] uppercase tracking-widest">
            <TrendingUp className="w-3 h-3" /> CATEGORY_SENSITIVITY_ANALYSIS
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {data.top_categories_inflation.map((cat, i) => (
              <div key={i} className="p-4 border-2 border-black bg-elevated hover:bg-background transition-all group">
                <div className="flex justify-between items-start mb-4">
                  <div className="text-sm font-black uppercase tracking-tight">{cat.category}</div>
                  <div className={`text-xs font-black ${cat.variation_percent > 0 ? 'text-danger' : 'text-success'}`}>
                    {cat.variation_percent > 0 ? '+' : ''}{cat.variation_percent.toFixed(1)}%
                  </div>
                </div>
                <div className="space-y-1">
                  <div className="text-[10px] font-bold text-text-secondary uppercase">IMPACTO_REAL</div>
                  <div className="text-lg font-black tracking-tighter">
                    R$ {cat.impact_reais.toLocaleString("pt-BR", { minimumFractionDigits: 2 })}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
