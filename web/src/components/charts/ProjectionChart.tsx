"use client";

import { ResponsiveContainer, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, Cell } from "recharts";
import { format, addMonths } from "date-fns";
import { ptBR } from "date-fns/locale";

interface ProjectionChartProps {
  data: any;
}

export default function ProjectionChart({ data }: ProjectionChartProps) {
  if (!data || !data.next_3_months_projections) return (
    <div className="h-[300px] flex items-center justify-center text-text-secondary border-2 border-dashed border-black uppercase text-[10px] font-bold bg-elevated/50">
      DADOS_DE_PROJEÇÃO_INSUFICIENTES
    </div>
  );

  const chartData = data.next_3_months_projections.map((item: any) => ({
    name: item.month.substring(0, 3).toUpperCase(),
    gasto: item.estimated_spending,
    confianca: item.confidence * 100,
  }));

  return (
    <div className="h-[300px] w-full">
      <ResponsiveContainer width="100%" height="100%">
        <BarChart data={chartData} margin={{ top: 20, right: 30, left: 0, bottom: 0 }}>
          <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#2A2A3A" />
          <XAxis 
            dataKey="name" 
            axisLine={false} 
            tickLine={false} 
            tick={{ fill: "#8888A0", fontSize: 12, fontWeight: 'bold' }}
          />
          <YAxis 
            axisLine={false} 
            tickLine={false} 
            tick={{ fill: "#8888A0", fontSize: 10 }}
            tickFormatter={(value) => `R$ ${value}`}
          />
          <Tooltip 
            cursor={{ fill: 'rgba(255,255,255,0.05)' }}
            contentStyle={{ backgroundColor: "#111118", border: "2px solid #2A2A3A", borderRadius: "0px" }}
            itemStyle={{ fontSize: "12px", fontWeight: "bold" }}
          />
          <Legend iconType="square" wrapperStyle={{ paddingTop: "20px", fontSize: "10px", fontWeight: "bold" }} />
          <Bar dataKey="gasto" name="GASTO_ESTIMADO" fill="#7C6FFF" radius={[4, 4, 0, 0]} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
