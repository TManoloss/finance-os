"use client";

import React from "react";
import { 
  AreaChart, 
  Area, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer,
} from "recharts";
import { Activity, DollarSign, TrendingUp, AlertCircle, BarChart3 } from "lucide-react";

interface WeekData {
  week_number: number;
  total_spent: number;
  avg_spent: number;
  transaction_count: number;
}

interface MonthlyCycleProps {
  data: {
    weeks: WeekData[];
    peak_week: number;
    spending_trend: string;
  };
  salaryEffectData?: {
    salary_day: number;
    spending_increase_percent: number;
    days_until_normal: number;
    insight: string;
  };
}

export default function MonthlyCycleChart({ data, salaryEffectData }: MonthlyCycleProps) {
  if (!data || !data.weeks) return null;

  const chartData = data.weeks.map(w => ({
    name: `SEM ${w.week_number}`,
    total: w.total_spent,
    avg: w.avg_spent
  }));

  const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      return (
        <div className="bg-background border-2 border-black p-4 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
          <p className="font-black text-xs uppercase mb-2 border-b border-black pb-1">{payload[0].payload.name}</p>
          <p className="text-xl font-black text-accent-primary">
            R$ {payload[0].value.toLocaleString("pt-BR", { minimumFractionDigits: 2 })}
          </p>
          <p className="text-[10px] font-bold text-text-secondary uppercase mt-1">TOTAL_WEEKLY_SPENT</p>
        </div>
      );
    }
    return null;
  };

  return (
    <div className="border-2 border-black bg-background shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] overflow-hidden">
      <div className="p-6 bg-black border-b-2 border-black flex items-center justify-between">
        <div className="flex items-center gap-3">
          <BarChart3 className="w-6 h-6 text-white" />
          <h2 className="text-xl font-black text-white uppercase tracking-tighter">MONTHLY_WEEKLY_CYCLE</h2>
        </div>
        <div className="px-3 py-1 bg-accent-primary text-white text-[10px] font-black uppercase tracking-widest">
          CYCLE_ANALYSIS_V1
        </div>
      </div>

      <div className="p-8 space-y-8">
        {/* Weekly Area Chart */}
        <div className="space-y-4">
           <div className="flex items-center gap-2 text-text-secondary font-black text-[10px] uppercase tracking-widest">
              <Activity className="w-3 h-3" /> SPENDING_VELOCITY_BY_WEEK
           </div>
           <div className="h-[250px] w-full border-2 border-black p-4 bg-elevated/20">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={chartData}>
                  <defs>
                    <linearGradient id="colorTotal" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="#7C6FFF" stopOpacity={0.3}/>
                      <stop offset="95%" stopColor="#7C6FFF" stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="#2A2A3A" vertical={false} />
                  <XAxis 
                    dataKey="name" 
                    axisLine={{ stroke: '#000', strokeWidth: 2 }}
                    tickLine={false}
                    tick={{ fill: '#8888A0', fontWeight: 'bold', fontSize: 10 }}
                  />
                  <YAxis 
                    axisLine={{ stroke: '#000', strokeWidth: 2 }}
                    tickLine={false}
                    tick={{ fill: '#8888A0', fontWeight: 'bold', fontSize: 10 }}
                  />
                  <Tooltip content={<CustomTooltip />} />
                  <Area 
                    type="monotone" 
                    dataKey="total" 
                    stroke="#7C6FFF" 
                    strokeWidth={4}
                    fillOpacity={1} 
                    fill="url(#colorTotal)" 
                  />
                </AreaChart>
              </ResponsiveContainer>
           </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
          {/* Salary Effect Card */}
          {salaryEffectData && (
            <div className="p-6 border-2 border-black bg-elevated relative overflow-hidden group">
               <div className="absolute -right-4 -bottom-4 opacity-5 group-hover:opacity-10 transition-opacity">
                  <DollarSign className="w-24 h-24" />
               </div>
               <div className="text-[10px] font-black text-text-secondary uppercase mb-2">SALARY_CONSUMPTION_IMPACT</div>
               <div className="flex items-baseline gap-2">
                  <div className="text-4xl font-black text-accent-primary">+{salaryEffectData.spending_increase_percent.toFixed(0)}%</div>
                  <div className="text-xs font-bold uppercase text-text-secondary">SPIKE_POST_PAYDAY</div>
               </div>
               <div className="mt-4 flex items-center gap-2 text-[10px] font-black uppercase text-text-secondary">
                  <AlertCircle className="w-3 h-3 text-accent-primary" />
                  STABILIZATION_IN: {salaryEffectData.days_until_normal} DAYS
               </div>
               <p className="mt-4 text-xs font-bold text-text-secondary uppercase leading-relaxed italic border-l-2 border-black pl-3">
                 "{salaryEffectData.insight}"
               </p>
            </div>
          )}

          {/* Trend & Peak Card */}
          <div className="p-6 border-2 border-black bg-background flex flex-col justify-between">
             <div className="space-y-4">
                <div className="flex items-center justify-between border-b border-black pb-2">
                   <div className="text-[10px] font-black text-text-secondary uppercase">MONTHLY_PEAK</div>
                   <div className="text-sm font-black uppercase">WEEK_0{data.peak_week}</div>
                </div>
                <div className="flex items-center justify-between border-b border-black pb-2">
                   <div className="text-[10px] font-black text-text-secondary uppercase">GENERAL_TREND</div>
                   <div className={`text-sm font-black uppercase ${data.spending_trend === 'decreasing' ? 'text-success' : 'text-danger'}`}>
                      {data.spending_trend === 'decreasing' ? 'STABILIZING' : 'ACCELERATING'}
                   </div>
                </div>
             </div>

             <div className="mt-6 p-4 bg-accent-primary border-2 border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] flex items-center gap-3">
                <TrendingUp className="w-5 h-5 text-white" />
                <div className="text-[10px] font-black text-white uppercase leading-none">
                   VELOCITY_STABILITY_INDEX_CALCULATED
                </div>
             </div>
          </div>
        </div>
      </div>
    </div>
  );
}
