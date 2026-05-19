"use client";

import { useState, useEffect } from "react";
import { ResponsiveContainer, AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip } from "recharts";
import { format } from "date-fns";
import { ptBR } from "date-fns/locale";
import { useTheme } from "next-themes";

interface AreaChartProps {
  data: any[];
}

export default function AreaChartComponent({ data }: AreaChartProps) {
  const [isMounted, setIsMounted] = useState(false);
  const { theme } = useTheme();

  useEffect(() => {
    setIsMounted(true);
  }, []);

  if (!isMounted) return <div className="h-[300px] w-full bg-elevated animate-pulse border-2 border-black" />;

  const chartData = data.map(item => ({
    date: item.date ? format(new Date(item.date), "dd/MM", { locale: ptBR }) : '?',
    recebido: item.total_received,
    gasto: item.total_spent,
  }));

  const axisColor = theme === 'dark' ? '#ffffff' : '#000000';
  const gridColor = theme === 'dark' ? 'rgba(255,255,255,0.1)' : 'rgba(0,0,0,0.1)';

  return (
    <div className="h-[300px] w-full">
      <ResponsiveContainer width="100%" height="100%" minWidth={0}>
        <AreaChart data={chartData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
          <CartesianGrid strokeDasharray="0" vertical={true} stroke={gridColor} strokeWidth={1} />
          <XAxis 
            dataKey="date" 
            axisLine={{ stroke: axisColor, strokeWidth: 2 }} 
            tickLine={{ stroke: axisColor, strokeWidth: 2 }} 
            tick={{ fill: axisColor, fontSize: 10, fontWeight: 'bold' }}
            dy={10}
          />
          <YAxis 
            axisLine={{ stroke: axisColor, strokeWidth: 2 }} 
            tickLine={{ stroke: axisColor, strokeWidth: 2 }} 
            tick={{ fill: axisColor, fontSize: 10, fontWeight: 'bold' }}
          />
          <Tooltip 
            contentStyle={{ 
              backgroundColor: theme === 'dark' ? "#16161d" : "#f4f1ea", 
              border: `2px solid ${axisColor}`, 
              borderRadius: "0px", 
              fontWeight: 'bold',
              color: axisColor
            }}
            itemStyle={{ fontSize: "12px", color: 'inherit' }}
          />
          <Area 
            type="stepAfter" 
            dataKey="recebido" 
            stroke="#008000" 
            fill="#008000"
            fillOpacity={0.3} 
            strokeWidth={3}
          />
          <Area 
            type="stepAfter" 
            dataKey="gasto" 
            stroke="#d00000" 
            fill="#d00000"
            fillOpacity={0.3} 
            strokeWidth={3}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
