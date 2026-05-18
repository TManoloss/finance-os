"use client";

import React from "react";
import { Zap, TrendingUp, BarChart3, AlertCircle } from "lucide-react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

interface GrowthTrend {
  category_name: string;
  taxa_crescimento_mensal_percent: number;
  slope: number;
  gasto_ha_n_meses: number;
  gasto_atual: number;
  variacao_absoluta: number;
  variacao_total_percent: number;
  projecao_12_meses: number;
  data_points: number[];
}

interface SilentGrowthProps {
  data: {
    top_category: GrowthTrend;
    alert: string;
    all_trends: GrowthTrend[];
    computed_at: string;
  };
}

export default function SilentGrowthCard({ data }: SilentGrowthProps) {
  if (!data || !data.top_category) return null;

  return (
    <div className="border-2 border-black bg-background shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] overflow-hidden">
      <div className="p-6 bg-black border-b-2 border-black flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Zap className="w-6 h-6 text-accent-primary" />
          <h2 className="text-xl font-black text-white uppercase tracking-tighter">SILENT_GROWTH_WATCH</h2>
        </div>
        <div className="px-3 py-1 bg-accent-primary text-white text-[10px] font-black uppercase tracking-widest">
          PATTERN_DETECTION_ACTIVE
        </div>
      </div>

      <div className="p-8 space-y-8">
        {/* Main Alert Card */}
        <div className="p-8 border-2 border-black bg-danger/10 shadow-[4px_4px_0px_0px_rgba(255,107,107,0.3)] relative group">
          <div className="absolute top-4 right-4">
             <AlertCircle className="w-8 h-8 text-danger animate-pulse" />
          </div>
          <div className="flex items-center gap-2 text-danger font-black text-[10px] uppercase tracking-widest mb-4">
            <Zap className="w-3 h-3" /> CRITICAL_GROWTH_WARNING
          </div>
          <div className="text-2xl font-black leading-tight tracking-tighter italic markdown-content">
             <ReactMarkdown remarkPlugins={[remarkGfm]}>
                {data.alert}
             </ReactMarkdown>
          </div>
        </div>

        {/* Top Category Details */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
           <div className="p-6 border-2 border-black bg-elevated">
              <div className="text-[10px] font-black text-text-secondary uppercase mb-2">CATEGORIA_EM_FOCO</div>
              <div className="text-2xl font-black uppercase tracking-tighter text-accent-primary">
                 {data.top_category.category_name}
              </div>
           </div>
           <div className="p-6 border-2 border-black bg-elevated">
              <div className="text-[10px] font-black text-text-secondary uppercase mb-2">TAXA_CRESCIMENTO_MENSAL</div>
              <div className="text-2xl font-black uppercase tracking-tighter text-danger">
                 +{data.top_category.taxa_crescimento_mensal_percent}%
              </div>
           </div>
           <div className="p-6 border-2 border-black bg-elevated">
              <div className="text-[10px] font-black text-text-secondary uppercase mb-2">PROJECAO_12_MESES_EXTRA</div>
              <div className="text-2xl font-black uppercase tracking-tighter">
                 R$ {(data.top_category.slope * 12).toLocaleString("pt-BR", { minimumFractionDigits: 2 })}
              </div>
           </div>
        </div>

        {/* Other Trends */}
        {data.all_trends && data.all_trends.length > 1 && (
          <div className="space-y-4">
            <div className="flex items-center gap-2 text-text-secondary font-black text-[10px] uppercase tracking-widest">
              <BarChart3 className="w-3 h-3" /> OTHER_CONSISTENT_TRENDS
            </div>
            <div className="overflow-x-auto">
              <table className="w-full border-2 border-black">
                <thead className="bg-elevated text-[10px] font-black uppercase tracking-widest border-b-2 border-black">
                  <tr>
                    <th className="p-4 text-left">CATEGORIA</th>
                    <th className="p-4 text-center">CRESCIMENTO_%</th>
                    <th className="p-4 text-center">GASTO_ATUAL</th>
                    <th className="p-4 text-right">TREND_SLOPE</th>
                  </tr>
                </thead>
                <tbody className="divide-y-2 divide-black">
                  {data.all_trends.slice(1).map((trend, i) => (
                    <tr key={i} className="hover:bg-elevated/50 transition-colors">
                      <td className="p-4 font-bold uppercase text-xs">{trend.category_name}</td>
                      <td className="p-4 text-center font-black text-danger text-xs">+{trend.taxa_crescimento_mensal_percent}%</td>
                      <td className="p-4 text-center font-bold text-xs">R$ {trend.gasto_atual.toFixed(2)}</td>
                      <td className="p-4 text-right font-bold text-xs text-text-secondary">+R$ {trend.slope}/mês</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
