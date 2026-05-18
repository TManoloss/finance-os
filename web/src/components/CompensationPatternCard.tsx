"use client";

import React from "react";
import { Brain, Activity, TrendingUp, AlertCircle } from "lucide-react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { ResponsiveContainer, BarChart, Bar, XAxis, YAxis, Tooltip, Cell } from "recharts";

interface StressEvent {
  data_salario: string;
  valor_salario: number;
  gasto_48h_pos: number;
  impacto_percentual: number;
}

interface CompensationData {
  autocorrelation: number;
  weekly_median: number;
  num_weeks: number;
  compensation_score: string;
  stress_events: StressEvent[];
  weekly_data: number[];
  insights: {
    pattern_detected: boolean;
    pattern_type: string;
    strength: string;
    narrative: string;
    stress_insight: string;
  };
}

interface CompensationPatternCardProps {
  data: CompensationData;
}

export default function CompensationPatternCard({ data }: CompensationPatternCardProps) {
  if (!data) return null;

  const chartData = data.weekly_data.map((value, i) => ({
    name: `W${i + 1}`,
    value: value,
  }));

  return (
    <div className="border-2 border-black bg-background shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] overflow-hidden">
      <div className="p-6 bg-accent-primary border-b-2 border-black flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Brain className="w-6 h-6 text-white" />
          <h2 className="text-xl font-black text-white uppercase tracking-tighter">COMPENSATION_PATTERN</h2>
        </div>
        <div className="px-3 py-1 bg-black text-white text-[10px] font-black uppercase tracking-widest">
          BEHAVIORAL_CORE_V3
        </div>
      </div>

      <div className="p-8 space-y-8">
        {/* Weekly Volatility Chart */}
        <div className="space-y-4">
          <div className="flex items-center gap-2 text-text-secondary font-black text-[10px] uppercase tracking-widest">
            <TrendingUp className="w-3 h-3" /> WEEKLY_SPENDING_VOLATILITY
          </div>
          <div className="h-[200px] w-full border-2 border-black bg-elevated/10 p-4">
             <ResponsiveContainer width="100%" height="100%">
                <BarChart data={chartData}>
                   <XAxis dataKey="name" hide />
                   <YAxis hide />
                   <Tooltip 
                      cursor={{ fill: 'rgba(0,0,0,0.1)' }}
                      contentStyle={{ backgroundColor: "#16161d", border: "2px solid currentColor", borderRadius: "0px" }}
                      itemStyle={{ fontSize: "12px", color: '#ffffff' }}
                      formatter={(value: any) => new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(value)}
                   />
                   <Bar dataKey="value">
                      {chartData.map((entry, index) => (
                        <Cell 
                           key={`cell-${index}`} 
                           fill={entry.value > data.weekly_median ? "#FF6B6B" : "#4ECDC4"} 
                        />
                      ))}
                   </Bar>
                </BarChart>
             </ResponsiveContainer>
          </div>
        </div>

        {/* Status & Correlation */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="p-6 border-2 border-black bg-elevated relative overflow-hidden group">
            <div className="text-[10px] font-black text-text-secondary uppercase mb-2">PADRÃO_DETECTADO</div>
            <div className={`text-2xl font-black uppercase tracking-tighter ${data.insights.pattern_detected ? 'text-danger' : 'text-success'}`}>
               {data.insights.pattern_type} // {data.insights.strength}
            </div>
          </div>

          <div className="p-6 border-2 border-black bg-elevated relative overflow-hidden group">
            <div className="text-[10px] font-black text-text-secondary uppercase mb-2">AUTOCORRELAÇÃO_LAG-1</div>
            <div className="text-2xl font-black text-text-primary">
               {data.autocorrelation.toFixed(3)}
            </div>
            <div className="mt-1 text-[10px] font-bold text-text-secondary uppercase">
               VALORES NEGATIVOS INDICAM COMPENSAÇÃO
            </div>
          </div>
        </div>

        {/* Narrative Analysis */}
        <div className="space-y-4">
          <div className="flex items-center gap-2 text-text-secondary font-black text-[10px] uppercase tracking-widest">
            <Activity className="w-3 h-3" /> IA_PSYCHOLOGICAL_ANALYSIS
          </div>
          <div className="p-6 border-2 border-black bg-background shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] italic text-lg font-bold leading-relaxed border-l-8 border-l-accent-primary">
            <ReactMarkdown remarkPlugins={[remarkGfm]}>
              {data.insights.narrative}
            </ReactMarkdown>
          </div>
        </div>

        {/* Stress Insights */}
        {data.insights.stress_insight && (
          <div className="p-6 border-2 border-black bg-elevated/20 border-dashed">
             <div className="flex items-center gap-2 text-danger font-black text-[10px] uppercase tracking-widest mb-4">
                <AlertCircle className="w-4 h-4" /> POST_STRESS_SPENDING_ALERTA
             </div>
             <p className="text-sm font-bold leading-relaxed text-text-primary">
                {data.insights.stress_insight}
             </p>
          </div>
        )}
      </div>
    </div>
  );
}
