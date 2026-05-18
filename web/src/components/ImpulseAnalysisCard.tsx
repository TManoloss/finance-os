"use client";

import React from "react";
import { Zap, AlertTriangle, Activity, MousePointer2 } from "lucide-react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { ResponsiveContainer, PieChart, Pie, Cell, Tooltip, Legend } from "recharts";

interface ImpulseData {
  impulse_count: number;
  planned_count: number;
  uncertain_count: number;
  impulse_total_amount: number;
  impulse_percent_of_spending: number;
  impulse_avg_ticket: number;
  planned_avg_ticket: number;
  regret_merchants_count: number;
  top_impulse_categories: [string, number][];
  recent_impulse_transactions: any[];
  insights: {
    summary: string;
    categories_at_risk: string[];
    primary_triggers: string[];
    recommendation: string;
    narrative: string;
  };
}

interface ImpulseAnalysisCardProps {
  data: ImpulseData;
}

export default function ImpulseAnalysisCard({ data }: ImpulseAnalysisCardProps) {
  if (!data) return null;

  const chartData = [
    { name: "PLANEJADO", value: data.planned_count, color: "#4ECDC4" },
    { name: "IMPULSO", value: data.impulse_count, color: "#FF6B6B" },
    { name: "INCERTO", value: data.uncertain_count, color: "#FFD93D" },
  ];

  return (
    <div className="border-2 border-black bg-background shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] overflow-hidden">
      <div className="p-6 bg-accent-primary border-b-2 border-black flex items-center justify-between">
        <div className="flex items-center gap-3">
          <MousePointer2 className="w-6 h-6 text-white" />
          <h2 className="text-xl font-black text-white uppercase tracking-tighter">IMPULSE_SPENDING_DETECTION</h2>
        </div>
        <div className="px-3 py-1 bg-black text-white text-[10px] font-black uppercase tracking-widest">
          BEHAVIORAL_CORE_V3
        </div>
      </div>

      <div className="p-8 space-y-8">
        {/* Main Stats & Chart */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
          <div className="h-[250px] w-full border-2 border-black bg-elevated/20 p-4">
             <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={chartData}
                    cx="50%"
                    cy="50%"
                    innerRadius={60}
                    outerRadius={80}
                    paddingAngle={2}
                    dataKey="value"
                    stroke="#0A0A0F"
                    strokeWidth={2}
                  >
                    {chartData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip 
                    contentStyle={{ 
                      backgroundColor: "#16161d", 
                      border: "2px solid currentColor", 
                      borderRadius: "0px", 
                      fontWeight: 'bold'
                    }}
                    itemStyle={{ fontSize: "12px", color: '#ffffff' }}
                  />
                  <Legend 
                    verticalAlign="bottom" 
                    height={36} 
                    formatter={(value) => <span className="text-[10px] text-text-primary font-black uppercase tracking-tighter">{value}</span>}
                  />
                </PieChart>
             </ResponsiveContainer>
          </div>

          <div className="space-y-6">
            <div className="p-6 border-2 border-black bg-elevated relative overflow-hidden group">
              <div className="absolute -right-4 -bottom-4 opacity-5 group-hover:opacity-10 transition-opacity">
                <AlertTriangle className="w-24 h-24" />
              </div>
              <div className="text-[10px] font-black text-text-secondary uppercase mb-2">TAXA_DE_IMPULSIVIDADE</div>
              <div className="text-4xl font-black text-danger">{data.impulse_percent_of_spending.toFixed(1)}%</div>
              <div className="mt-4 text-xs font-bold text-text-secondary uppercase">
                DO GASTO TOTAL EM DEBITO
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
               <div className="p-4 border-2 border-black bg-background">
                  <div className="text-[10px] font-black text-text-secondary uppercase mb-1">TOTAL_IMPULSO</div>
                  <div className="text-xl font-black tracking-tighter text-text-primary">
                    R$ {data.impulse_total_amount.toLocaleString("pt-BR", { minimumFractionDigits: 2 })}
                  </div>
               </div>
               <div className="p-4 border-2 border-black bg-background">
                  <div className="text-[10px] font-black text-text-secondary uppercase mb-1">ARREPENDIMENTO_PROXY</div>
                  <div className="text-xl font-black tracking-tighter text-accent-primary">
                    {data.regret_merchants_count} LOJAS
                  </div>
               </div>
            </div>
          </div>
        </div>

        {/* Narrative Analysis */}
        <div className="space-y-4">
          <div className="flex items-center gap-2 text-text-secondary font-black text-[10px] uppercase tracking-widest">
            <Activity className="w-3 h-3" /> IA_BEHAVIORAL_NARRATIVE
          </div>
          <div className="p-6 border-2 border-black bg-background shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] italic text-lg font-bold leading-relaxed markdown-content border-l-8 border-l-accent-primary">
            <ReactMarkdown remarkPlugins={[remarkGfm]}>
              {data.insights.narrative}
            </ReactMarkdown>
          </div>
        </div>

        {/* Categories & Triggers */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
           <div className="space-y-4">
              <div className="flex items-center gap-2 text-text-secondary font-black text-[10px] uppercase tracking-widest">
                <Zap className="w-3 h-3" /> CATEGORIAS_DE_RISCO
              </div>
              <div className="space-y-2">
                 {data.top_impulse_categories.map(([cat, count], i) => (
                   <div key={i} className="flex items-center justify-between p-3 border-2 border-black bg-elevated">
                      <span className="text-xs font-black uppercase tracking-tight">{cat}</span>
                      <span className="px-2 py-0.5 bg-black text-white text-[10px] font-bold">{count}X</span>
                   </div>
                 ))}
              </div>
           </div>

           <div className="space-y-4">
              <div className="flex items-center gap-2 text-text-secondary font-black text-[10px] uppercase tracking-widest">
                <Activity className="w-3 h-3" /> GATILHOS_PRIMARIOS
              </div>
              <div className="p-6 border-2 border-black bg-elevated/10">
                 <ul className="space-y-2">
                    {data.insights.primary_triggers.map((trigger, i) => (
                      <li key={i} className="flex items-start gap-2">
                         <div className="mt-1.5 w-2 h-2 bg-accent-primary shrink-0" />
                         <span className="text-sm font-bold uppercase tracking-tight">{trigger}</span>
                      </li>
                    ))}
                 </ul>
              </div>
           </div>
        </div>
        
        {/* Recommendation */}
        <div className="p-6 border-2 border-black bg-black text-white">
           <div className="text-[10px] font-black text-accent-primary uppercase mb-2 tracking-[0.2em]">IA_RECOMMENDATION_SYSTEM</div>
           <p className="text-sm font-bold italic">"{data.insights.recommendation}"</p>
        </div>
      </div>
    </div>
  );
}
