"use client";

import React from "react";
import { Coffee, Activity, TrendingUp, Info } from "lucide-react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

interface ConveniencePair {
  name: string;
  spent: number;
  count: number;
  premium: number;
  premium_percent: number;
}

interface ConvenienceData {
  pairs: ConveniencePair[];
  total_monthly_premium: number;
  total_annual_premium: number;
  percent_of_income: number;
  insight: string;
}

interface ConvenienceIndexCardProps {
  data: ConvenienceData;
}

export default function ConvenienceIndexCard({ data }: ConvenienceIndexCardProps) {
  if (!data) return null;

  return (
    <div className="border-2 border-black bg-background shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] overflow-hidden">
      <div className="p-6 bg-accent-secondary border-b-2 border-black flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Coffee className="w-6 h-6 text-white" />
          <h2 className="text-xl font-black text-white uppercase tracking-tighter">CONVENIENCE_INDEX</h2>
        </div>
        <div className="px-3 py-1 bg-black text-white text-[10px] font-black uppercase tracking-widest">
          SPECIFIC_COSTS_V1
        </div>
      </div>

      <div className="p-8 space-y-8">
        {/* Main Stats */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="p-6 border-2 border-black bg-elevated relative overflow-hidden group">
            <div className="absolute -right-4 -bottom-4 opacity-5 group-hover:opacity-10 transition-opacity">
              <Info className="w-24 h-24" />
            </div>
            <div className="text-[10px] font-black text-text-secondary uppercase mb-2">PREMIO_DE_CONVENIENCIA_MENSAL</div>
            <div className="text-4xl font-black text-accent-secondary">
               R$ {data.total_monthly_premium.toLocaleString("pt-BR", { minimumFractionDigits: 2 })}
            </div>
            <div className="mt-4 text-xs font-bold text-text-secondary uppercase">
               OU R$ {data.total_annual_premium.toLocaleString("pt-BR", { minimumFractionDigits: 0 })} POR ANO
            </div>
          </div>

          <div className="p-6 border-2 border-black bg-elevated relative overflow-hidden group">
            <div className="absolute -right-4 -bottom-4 opacity-5 group-hover:opacity-10 transition-opacity">
              <TrendingUp className="w-24 h-24" />
            </div>
            <div className="text-[10px] font-black text-text-secondary uppercase mb-2">PERCENTUAL_DA_RENDA</div>
            <div className="text-4xl font-black text-text-primary">
               {data.percent_of_income.toFixed(1)}%
            </div>
            <div className="mt-4 text-xs font-bold text-text-secondary uppercase text-warning">
               PAGO APENAS PELA PRATICIDADE
            </div>
          </div>
        </div>

        {/* Narrative Insight */}
        <div className="space-y-4">
          <div className="flex items-center gap-2 text-text-secondary font-black text-[10px] uppercase tracking-widest">
            <Activity className="w-3 h-3" /> IA_CONVENIENCE_INSIGHT
          </div>
          <div className="p-6 border-2 border-black bg-background shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] italic text-lg font-bold leading-relaxed border-l-8 border-l-accent-secondary">
            <ReactMarkdown remarkPlugins={[remarkGfm]}>
              {data.insight}
            </ReactMarkdown>
          </div>
        </div>

        {/* Convenience Pairs */}
        <div className="space-y-4">
          <div className="flex items-center gap-2 text-text-secondary font-black text-[10px] uppercase tracking-widest">
            <TrendingUp className="w-3 h-3" /> CATEGORY_CONVENIENCE_PAIRS
          </div>
          <div className="grid grid-cols-1 gap-4">
            {data.pairs.map((pair, i) => (
              <div key={i} className="p-6 border-2 border-black bg-elevated hover:bg-background transition-all group relative overflow-hidden">
                <div className="flex flex-col md:flex-row md:items-center justify-between gap-6 relative z-10">
                   <div className="flex items-center gap-4">
                      <div className="w-12 h-12 border-2 border-black bg-accent-secondary flex items-center justify-center text-white text-xl font-black">
                         {i + 1}
                      </div>
                      <div>
                         <div className="text-[10px] font-black text-text-secondary uppercase mb-1">PAIR_IDENTIFIER</div>
                         <h3 className="text-lg font-black uppercase tracking-tighter">{pair.name}</h3>
                      </div>
                   </div>

                   <div className="grid grid-cols-2 md:grid-cols-3 gap-8">
                      <div>
                         <div className="text-[10px] font-black text-text-secondary uppercase mb-1">GASTO_TOTAL</div>
                         <div className="text-xl font-black tracking-tighter">
                            R$ {pair.spent.toLocaleString("pt-BR", { minimumFractionDigits: 2 })}
                         </div>
                      </div>
                      <div>
                         <div className="text-[10px] font-black text-text-secondary uppercase mb-1">PREMIO_ESTIMADO</div>
                         <div className="text-xl font-black tracking-tighter text-warning">
                            R$ {pair.premium.toLocaleString("pt-BR", { minimumFractionDigits: 2 })}
                         </div>
                      </div>
                      <div className="hidden md:block text-right">
                         <div className="text-[10px] font-black text-text-secondary uppercase mb-1">FATOR_CONVENIÊNCIA</div>
                         <div className="text-xl font-black tracking-tighter">
                            +{pair.premium_percent}%
                         </div>
                      </div>
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
