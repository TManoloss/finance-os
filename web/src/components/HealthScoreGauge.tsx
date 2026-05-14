"use client";

import { ResponsiveContainer, PieChart, Pie, Cell } from "recharts";

interface HealthScoreGaugeProps {
  score: number;
}

export default function HealthScoreGauge({ score }: HealthScoreGaugeProps) {
  const data = [
    { value: score },
    { value: 100 - score },
  ];

  const getColor = (s: number) => {
    if (s >= 80) return "#4ECDC4"; // Teal
    if (s >= 60) return "#FFD93D"; // Yellow
    return "#FF6B6B"; // Red
  };

  return (
    <div className="relative w-full h-[200px] flex items-center justify-center">
      <ResponsiveContainer width="100%" height="100%">
        <PieChart>
          <Pie
            data={data}
            cx="50%"
            cy="100%"
            startAngle={180}
            endAngle={0}
            innerRadius={60}
            outerRadius={80}
            paddingAngle={0}
            dataKey="value"
            stroke="none"
          >
            <Cell fill={getColor(score)} />
            <Cell fill="#2A2A3A" />
          </Pie>
        </PieChart>
      </ResponsiveContainer>
      <div className="absolute bottom-0 text-center">
        <span className="text-4xl font-black text-text-primary font-mono">{score}</span>
        <p className="text-[10px] font-bold text-text-secondary uppercase tracking-widest">PONTOS_DE_SAUDE</p>
      </div>
    </div>
  );
}
