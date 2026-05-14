"use client";

import { useState, useEffect } from "react";
import { ResponsiveContainer, PieChart, Pie, Cell, Tooltip, Legend } from "recharts";
import { useTheme } from "next-themes";

interface DonutChartProps {
  data: any[];
}

export default function DonutChartComponent({ data }: DonutChartProps) {
  const [isMounted, setIsMounted] = useState(false);
  const { theme } = useTheme();

  useEffect(() => {
    setIsMounted(true);
  }, []);

  const chartData = data.map(item => ({
    name: item.category_name.toUpperCase(),
    value: item.total,
  }));

  // Cores de alto contraste para o estilo Blueprint
  const lightColors = ["#0000ff", "#d00000", "#008000", "#ffa500", "#7c6fff", "#444444"];
  const darkColors = ["#7C6FFF", "#FF6B6B", "#4ECDC4", "#FFD93D", "#A78BFA", "#F0F0F5"];
  
  const technicalColors = theme === 'dark' ? darkColors : lightColors;

  if (!isMounted) return <div className="h-[300px] w-full bg-elevated animate-pulse border-2 border-black" />;

  return (
    <div className="h-[300px] w-full">
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
            stroke={theme === 'dark' ? "#0A0A0F" : "#f4f1ea"}
            strokeWidth={2}
          >
            {chartData.map((entry, index) => (
              <Cell key={`cell-${index}`} fill={technicalColors[index % technicalColors.length]} />
            ))}
          </Pie>
          <Tooltip 
            contentStyle={{ 
              backgroundColor: theme === 'dark' ? "#16161d" : "#f4f1ea", 
              border: "2px solid currentColor", 
              borderRadius: "0px", 
              fontWeight: 'bold',
              color: theme === 'dark' ? "#ffffff" : "#000000"
            }}
            itemStyle={{ fontSize: "12px", color: 'inherit' }}
            formatter={(value: any) => new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(value)}
          />
          <Legend 
            verticalAlign="bottom" 
            height={36} 
            formatter={(value) => <span className="text-[10px] text-text-primary font-black uppercase tracking-tighter">{value}</span>}
          />
        </PieChart>
      </ResponsiveContainer>
    </div>
  );
}
