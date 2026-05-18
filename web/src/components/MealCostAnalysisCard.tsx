"use client";

import React from "react";
import { Utensils, Activity, TrendingUp, BarChart3 } from "lucide-react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { ResponsiveContainer, BarChart, Bar, XAxis, YAxis, Tooltip, Cell } from "recharts";

interface MealData {
  total_spent: float;
  total_meals_estimated: number;
  avg_cost_per_meal: number;
  by_channel: {
    delivery: { total: number; count: number; percent: number };
    supermercado: { total: number; count: number; percent: number };
    restaurante: { total: number; count: number; percent: number };
  };
  period_days: number;
  insight: string;
}

interface MealCostAnalysisCardProps {
  data: MealData;
}

export default function MealCostAnalysisCard({ data }: MealCostAnalysisCardProps) {
  if (!data) return null;

  const chartData = [
    { name: "DELIVERY", value: data.by_channel.delivery?.total || 0, color: "#FF6B6B" },
    { name: "MERCADO", value: data.by_channel.supermercado?.total || 0, color: "#4ECDC4" },
    { name: "RESTAURANTE", value: data.by_channel.restaurante?.total || 0, color: "#7C6FFF" },
  ];

  return (
    <div className="border-2 border-black bg-background shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] overflow-hidden">
      <div className="p-6 bg-success border-b-2 border-black flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Utensils className="w-6 h-6 text-white" />
          <h2 className="text-xl font-black text-white uppercase tracking-tighter">MEAL_COST_EFFICIENCY</h2>
        </div>
        <div className="px-3 py-1 bg-black text-white text-[10px] font-black uppercase tracking-widest">
          SPECIFIC_COSTS_V1
        </div>
      </div>

      <div className="p-8 space-y-8">
        {/* Main Stat */}
        <div className="p-8 border-2 border-black bg-elevated relative overflow-hidden group">
           <div className="absolute -right-4 -bottom-4 opacity-5 group-hover:opacity-10 transition-opacity">
             <BarChart3 className="w-32 h-32" />
           </div>
           <div className="text-[10px] font-black text-text-secondary uppercase mb-2">CUSTO_MEDIO_POR_REFEICAO</div>
           <div className="flex items-baseline gap-2">
              <div className="text-5xl font-black text-success">
                 R$ {data.avg_cost_per_meal.toLocaleString("pt-BR", { minimumFractionDigits: 2 })}
              </div>
              <div className="text-sm font-bold text-text-secondary uppercase">/ REFEIÇÃO</div>
           </div>
           <p className="mt-4 text-xs font-bold text-text-secondary uppercase">
             ESTIMATIVA BASEADA EM {data.total_meals_estimated} REFEIÇÕES EM {data.period_days} DIAS
           </p>
        </div>

        {/* Channel Breakdown */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
           <div className="space-y-4">
              <div className="flex items-center gap-2 text-text-secondary font-black text-[10px] uppercase tracking-widest">
                <TrendingUp className="w-3 h-3" /> SPENDING_BY_CHANNEL
              </div>
              <div className="h-[200px] w-full">
                 <ResponsiveContainer width="100%" height="100%">
                    <BarChart data={chartData} layout="vertical" margin={{ left: -20 }}>
                       <XAxis type="number" hide />
                       <YAxis 
                          dataKey="name" 
                          type="category" 
                          axisLine={false} 
                          tickLine={false} 
                          tick={{ fontSize: 10, fontWeight: 'black', fill: '#8888A0' }}
                          width={100}
                       />
                       <Tooltip 
                          cursor={{ fill: 'rgba(0,0,0,0.1)' }}
                          contentStyle={{ 
                            backgroundColor: "#16161d", 
                            border: "2px solid currentColor", 
                            borderRadius: "0px"
                          }}
                          itemStyle={{ fontSize: "12px", color: '#ffffff' }}
                          formatter={(value: any) => new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(value)}
                       />
                       <Bar dataKey="value" radius={[0, 4, 4, 0]}>
                          {chartData.map((entry, index) => (
                             <Cell key={`cell-${index}`} fill={entry.color} />
                          ))}
                       </Bar>
                    </BarChart>
                 </ResponsiveContainer>
              </div>
           </div>

           <div className="grid grid-cols-1 gap-4">
              {Object.entries(data.by_channel).map(([key, channel]) => (
                 <div key={key} className="p-4 border-2 border-black bg-background flex items-center justify-between group hover:bg-elevated transition-colors">
                    <div>
                       <div className="text-[10px] font-black text-text-secondary uppercase mb-1">{key}</div>
                       <div className="text-lg font-black tracking-tighter">
                          R$ {channel.total.toLocaleString("pt-BR", { minimumFractionDigits: 2 })}
                       </div>
                    </div>
                    <div className="text-right">
                       <div className="text-[10px] font-black text-text-secondary uppercase mb-1">TRANSAÇÕES</div>
                       <div className="text-lg font-black tracking-tighter">{channel.count}X</div>
                    </div>
                 </div>
              ))}
           </div>
        </div>

        {/* Narrative Insight */}
        <div className="space-y-4">
          <div className="flex items-center gap-2 text-text-secondary font-black text-[10px] uppercase tracking-widest">
            <Activity className="w-3 h-3" /> IA_SPECIFIC_COST_INSIGHT
          </div>
          <div className="p-6 border-2 border-black bg-background shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] italic text-lg font-bold leading-relaxed border-l-8 border-l-success">
            <ReactMarkdown remarkPlugins={[remarkGfm]}>
              {data.insight}
            </ReactMarkdown>
          </div>
        </div>
      </div>
    </div>
  );
}
