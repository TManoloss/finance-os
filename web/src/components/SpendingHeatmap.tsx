"use client";

import React, { useMemo } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { format, parseISO, startOfYear, eachDayOfInterval, endOfYear, getDay, differenceInDays } from 'date-fns';
import { ptBR } from 'date-fns/locale';

interface HeatmapDay {
  date: string; // YYYY-MM-DD
  total: number;
}

export function SpendingHeatmap({ data }: { data: HeatmapDay[] }) {
  const { weeks, maxTotal, monthLabels } = useMemo(() => {
    if (!data || data.length === 0) return { weeks: [], maxTotal: 0, monthLabels: [] };

    // Find the date range
    const today = new Date();
    // Show approx last 52 weeks
    const startDate = new Date(today);
    startDate.setDate(today.getDate() - 364);

    // Create a map for quick lookup
    const dataMap = new Map(data.map(d => [d.date, d.total]));

    let currentMax = 0;
    for (const d of data) {
      if (d.total > currentMax) currentMax = d.total;
    }

    const allDays = eachDayOfInterval({ start: startDate, end: today });
    
    // Group by weeks
    const weeksArr: { date: Date; total: number; intensity: number }[][] = [];
    let currentWeek: { date: Date; total: number; intensity: number }[] = [];
    
    // Fill initial empty days if start is not Sunday (0)
    const firstDayOfWeek = getDay(startDate);
    for (let i = 0; i < firstDayOfWeek; i++) {
        currentWeek.push({ date: new Date(0), total: 0, intensity: 0 }); // Dummy
    }

    const mLabels: { text: string, offset: number }[] = [];
    let currentMonth = startDate.getMonth();

    allDays.forEach(date => {
      const dateStr = format(date, 'yyyy-MM-dd');
      const total = dataMap.get(dateStr) || 0;
      
      // Calculate intensity 0 to 4
      let intensity = 0;
      if (total > 0) {
        intensity = Math.ceil((total / currentMax) * 4);
        if (intensity === 0) intensity = 1; // At least 1 if > 0
      }

      currentWeek.push({ date, total, intensity });

      if (currentWeek.length === 7) {
        weeksArr.push(currentWeek);
        
        // Month label logic
        if (date.getMonth() !== currentMonth) {
            mLabels.push({ text: format(date, 'MMM', { locale: ptBR }), offset: weeksArr.length - 1 });
            currentMonth = date.getMonth();
        }

        currentWeek = [];
      }
    });

    if (currentWeek.length > 0) {
        // Pad the rest of the week
        while (currentWeek.length < 7) {
            currentWeek.push({ date: new Date(0), total: 0, intensity: 0 });
        }
        weeksArr.push(currentWeek);
    }

    return { weeks: weeksArr, maxTotal: currentMax, monthLabels: mLabels };
  }, [data]);

  const getColor = (intensity: number) => {
    switch (intensity) {
      case 0: return 'bg-[#1A1A24]';
      case 1: return 'bg-[#3A2D7D]'; // Lightest purple
      case 2: return 'bg-[#5040B2]';
      case 3: return 'bg-[#6A5AE0]';
      case 4: return 'bg-[#7C6FFF]'; // Main accent purple
      default: return 'bg-[#1A1A24]';
    }
  };

  if (!data || data.length === 0) return null;

  return (
    <Card className="bg-[#111118] border-[#2A2A3A]">
      <CardHeader>
        <CardTitle className="text-[#F0F0F5]">Frequência de Gastos</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col w-full overflow-x-auto pb-4">
            <div className="flex text-xs text-[#8888A0] mb-2 relative h-4">
                {monthLabels.map((lbl, idx) => (
                    <span key={idx} style={{ position: 'absolute', left: `${lbl.offset * 14}px` }}>
                        {lbl.text}
                    </span>
                ))}
            </div>
            <div className="flex gap-1">
                <div className="flex flex-col gap-1 text-xs text-[#8888A0] pr-2 pt-4 justify-between h-[105px]">
                    <span>Seg</span>
                    <span>Qua</span>
                    <span>Sex</span>
                </div>
                <div className="flex gap-1">
                    {weeks.map((week, wIdx) => (
                    <div key={wIdx} className="flex flex-col gap-1">
                        {week.map((day, dIdx) => (
                        <div
                            key={dIdx}
                            className={`w-3 h-3 rounded-sm ${day.date.getTime() === 0 ? 'bg-transparent' : getColor(day.intensity)}`}
                            title={day.date.getTime() !== 0 ? `${format(day.date, 'dd/MM/yyyy')}: R$ ${day.total.toFixed(2)}` : ''}
                        />
                        ))}
                    </div>
                    ))}
                </div>
            </div>
            <div className="flex justify-end items-center mt-4 gap-2 text-xs text-[#8888A0]">
                <span>Menos</span>
                <div className="flex gap-1">
                    <div className="w-3 h-3 rounded-sm bg-[#1A1A24]" />
                    <div className="w-3 h-3 rounded-sm bg-[#3A2D7D]" />
                    <div className="w-3 h-3 rounded-sm bg-[#5040B2]" />
                    <div className="w-3 h-3 rounded-sm bg-[#6A5AE0]" />
                    <div className="w-3 h-3 rounded-sm bg-[#7C6FFF]" />
                </div>
                <span>Mais</span>
            </div>
        </div>
      </CardContent>
    </Card>
  );
}
