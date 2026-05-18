"use client";

import React from "react";
import { 
  BarChart, 
  Bar, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer,
  Cell
} from "recharts";
import { Calendar, Clock, Activity, Zap, TrendingUp } from "lucide-react";

interface DayProfile {
  day_name: string;
  avg_spent: number;
  top_categories: string[];
  peak_hours: number[];
  transaction_count: number;
}

interface WeeklyProfileProps {
  data: {
    day_profiles: DayProfile[];
    peak_spending_day: string;
    total_avg_weekly: number;
  };
  weekdayWeekendData?: {
    weekday_avg: number;
    weekend_avg: number;
    weekend_multiplier: number;
    insight: string;
  };
}

export default function WeeklyProfileChart({ data, weekdayWeekendData }: WeeklyProfileProps) {
  if (!data || !data.day_profiles) return null;

  const chartData = data.day_profiles.map(p => ({
    name: p.day_name.substring(0, 3).toUpperCase(),
    avg: p.avg_spent,
    fullName: p.day_name
  }));

  const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      const dayData = data.day_profiles.find(p => p.day_name.substring(0, 3).toUpperCase() === payload[0].payload.name);
      return (
        <div className="bg-background border-2 border-black p-4 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
          <p className="font-black text-xs uppercase mb-2 border-b border-black pb-1">{dayData?.day_name}</p>
          <p className="text-xl font-black text-accent-primary">
            R$ {payload[0].value.toLocaleString("pt-BR", { minimumFractionDigits: 2 })}
          </p>
          <p className="text-[10px] font-bold text-text-secondary uppercase mt-1">MEDIA_DIARIA</p>
          
          {dayData && dayData.top_categories.length > 0 && (
            <div className="mt-3">
               <p className="text-[10px] font-black uppercase text-text-secondary mb-1">TOP_CATEGORIES</p>
               <div className="flex flex-wrap gap-1">
                  {dayData.top_categories.map((cat, i) => (
                    <span key={i} className="px-2 py-0.5 bg-elevated text-[9px] font-bold border border-black uppercase">{cat}</span>
                  ))}
               </div>
            </div>
          )}
        </div>
      );
    }
    return null;
  };

  return (
    <div className="border-2 border-black bg-background shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] overflow-hidden">
      <div className="p-6 bg-accent-primary border-b-2 border-black flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Calendar className="w-6 h-6 text-white" />
          <h2 className="text-xl font-black text-white uppercase tracking-tighter">WEEKLY_TEMPORAL_PROFILE</h2>
        </div>
        <div className="px-3 py-1 bg-black text-white text-[10px] font-black uppercase tracking-widest">
          TEMPORAL_KERNEL_V1
        </div>
      </div>

      <div className="p-8 space-y-8">
        {/* Weekly Chart */}
        <div className="space-y-4">
           <div className="flex items-center gap-2 text-text-secondary font-black text-[10px] uppercase tracking-widest">
              <Activity className="w-3 h-3" /> AVG_SPENDING_BY_DAY
           </div>
           <div className="h-[250px] w-full border-2 border-black p-4 bg-elevated/20">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={chartData}>
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
                  <Tooltip content={<CustomTooltip />} cursor={{ fill: 'rgba(124, 111, 255, 0.1)' }} />
                  <Bar dataKey="avg" radius={[0, 0, 0, 0]}>
                    {chartData.map((entry, index) => (
                      <Cell 
                        key={`cell-${index}`} 
                        fill={entry.fullName === data.peak_spending_day ? '#7C6FFF' : '#2A2A3A'} 
                        stroke="#000"
                        strokeWidth={2}
                      />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
           </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
          {/* Weekday vs Weekend */}
          {weekdayWeekendData && (
            <div className="p-6 border-2 border-black bg-elevated relative overflow-hidden group">
               <div className="absolute -right-4 -bottom-4 opacity-5 group-hover:opacity-10 transition-opacity">
                  <TrendingUp className="w-24 h-24" />
               </div>
               <div className="text-[10px] font-black text-text-secondary uppercase mb-2">WEEKEND_INTENSITY</div>
               <div className="flex items-baseline gap-2">
                  <div className="text-4xl font-black text-accent-primary">x{weekdayWeekendData.weekend_multiplier.toFixed(2)}</div>
                  <div className="text-xs font-bold uppercase text-text-secondary">MULTIPLIER</div>
               </div>
               <p className="mt-4 text-xs font-bold text-text-secondary uppercase leading-relaxed italic border-l-2 border-accent-primary pl-3">
                 "{weekdayWeekendData.insight}"
               </p>
            </div>
          )}

          {/* Peak Stats */}
          <div className="grid grid-cols-1 gap-4">
             <div className="p-4 border-2 border-black bg-background flex items-center gap-4">
                <div className="w-10 h-10 border-2 border-black flex items-center justify-center bg-accent-primary text-white shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]">
                   <Clock className="w-5 h-5" />
                </div>
                <div>
                   <div className="text-[10px] font-black text-text-secondary uppercase">PEAK_SPENDING_HOUR</div>
                   <div className="text-lg font-black uppercase tracking-tighter">
                      {data.day_profiles.find(p => p.day_name === data.peak_spending_day)?.peak_hours[0]}:00
                   </div>
                </div>
             </div>
             
             <div className="p-4 border-2 border-black bg-background flex items-center gap-4">
                <div className="w-10 h-10 border-2 border-black flex items-center justify-center bg-black text-white shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]">
                   <Zap className="w-5 h-5" />
                </div>
                <div>
                   <div className="text-[10px] font-black text-text-secondary uppercase">MAX_CAPACITY_DAY</div>
                   <div className="text-lg font-black uppercase tracking-tighter">{data.peak_spending_day}</div>
                </div>
             </div>
          </div>
        </div>
      </div>
    </div>
  );
}
