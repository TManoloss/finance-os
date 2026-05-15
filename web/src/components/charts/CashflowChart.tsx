"use client";

import { useState, useEffect } from "react";
import { ResponsiveContainer, AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ReferenceLine, Scatter, ScatterChart, ComposedChart, Line } from "recharts";
import { format } from "date-fns";
import { ptBR } from "date-fns/locale";

interface CashflowChartProps {
  data: any[];
}

export default function CashflowChart({ data }: CashflowChartProps) {
  const [isMounted, setIsMounted] = useState(false);

  useEffect(() => {
    setIsMounted(true);
  }, []);

  if (!isMounted || !data || data.length === 0) {
    return <div className="h-[300px] w-full bg-surface-elevated animate-pulse rounded-xl border border-border" />;
  }

  const chartData = data.map(item => ({
    displayDate: format(new Date(item.date), "dd/MM", { locale: ptBR }),
    fullDate: item.date,
    saldo: item.balance_end,
    income: item.income,
    outcome: item.outcome,
    isCritical: item.is_critical,
    biggestSpend: item.biggest_spending?.merchant,
    biggestAmount: item.biggest_spending?.amount,
  }));

  // Find critical points to show markers
  const criticalPoints = chartData.filter(d => d.isCritical);

  return (
    <div className="h-[300px] w-full">
      <ResponsiveContainer width="100%" height="100%" minWidth={0}>

        <ComposedChart data={chartData} margin={{ top: 20, right: 20, left: 0, bottom: 20 }}>
          <defs>
            <linearGradient id="colorBalance" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#7C6FFF" stopOpacity={0.3}/>
              <stop offset="95%" stopColor="#7C6FFF" stopOpacity={0}/>
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#2A2A3A" />
          <XAxis 
            dataKey="displayDate" 
            axisLine={false}
            tickLine={false}
            tick={{ fill: "#8888A0", fontSize: 12 }}
            minTickGap={30}
          />
          <YAxis 
            axisLine={false}
            tickLine={false}
            tick={{ fill: "#8888A0", fontSize: 12 }}
            tickFormatter={(value) => `R$ ${value}`}
          />
          <Tooltip 
            content={({ active, payload }) => {
              if (active && payload && payload.length) {
                const d = payload[0].payload;
                return (
                  <div className="bg-[#111118] border border-[#2A2A3A] p-3 rounded-lg shadow-xl">
                    <p className="text-[#8888A0] text-xs mb-1">{format(new Date(d.fullDate), "dd 'de' MMMM", { locale: ptBR })}</p>
                    <p className="text-[#F0F0F5] font-bold text-lg mb-2">R$ {d.saldo.toLocaleString('pt-BR', { minimumFractionDigits: 2 })}</p>
                    <div className="space-y-1">
                      {d.income > 0 && <p className="text-[#4ECDC4] text-xs">↑ Recebido: R$ {d.income.toLocaleString('pt-BR')}</p>}
                      {d.outcome > 0 && <p className="text-[#FF6B6B] text-xs">↓ Gasto: R$ {d.outcome.toLocaleString('pt-BR')}</p>}
                      {d.biggestSpend && (
                        <p className="text-[#8888A0] text-[10px] italic mt-2">
                          Maior gasto: {d.biggestSpend} (R$ {d.biggestAmount.toLocaleString('pt-BR')})
                        </p>
                      )}
                    </div>
                  </div>
                );
              }
              return null;
            }}
          />
          <Area 
            type="monotone" 
            dataKey="saldo" 
            stroke="#7C6FFF" 
            strokeWidth={3}
            fillOpacity={1} 
            fill="url(#colorBalance)" 
          />
          
          {/* Markers for critical balance */}
          <Scatter 
            data={criticalPoints} 
            fill="#FF6B6B" 
            shape="circle"
          />

          <ReferenceLine y={500} stroke="#FF6B6B" strokeDasharray="3 3" label={{ position: 'right', value: 'Crítico', fill: '#FF6B6B', fontSize: 10 }} />
        </ComposedChart>
      </ResponsiveContainer>
    </div>
  );
}
