"use client";

import { format } from "date-fns";
import { ptBR } from "date-fns/locale";
import { Calendar, AlertCircle, CheckCircle2 } from "lucide-react";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

interface UpcomingExpense {
  date: string;
  description: string;
  amount: number;
  confidence_level: "high" | "medium" | "low";
}

export default function UpcomingExpenses({ expenses }: { expenses: UpcomingExpense[] }) {
  const safeExpenses = Array.isArray(expenses) ? expenses : [];

  if (safeExpenses.length === 0) {
    return (
      <div className="p-8 border-2 border-dashed border-black text-center text-[10px] font-bold uppercase opacity-50">
        NENHUMA_DESPESA_RELEVANTE_PROJETADA
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {safeExpenses.map((expense, i) => (
        <div key={i} className="p-4 border-2 border-black bg-elevated flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className={cn(
              "p-2 border-2 border-black",
              expense.confidence_level === 'high' ? "bg-accent-secondary" : expense.confidence_level === 'medium' ? "bg-warning" : "bg-elevated"
            )}>
              <Calendar className="w-4 h-4 text-black" />
            </div>
            <div>
              <div className="font-black uppercase text-xs">{expense.description}</div>
              <div className="text-[10px] font-bold text-text-secondary uppercase">
                PREVISÃO: {format(new Date(expense.date), "dd 'DE' MMMM", { locale: ptBR })}
              </div>
            </div>
          </div>
          <div className="text-right">
            <div className="font-black font-mono text-sm">
              R$ {expense.amount.toLocaleString('pt-BR', { minimumFractionDigits: 2 })}
            </div>
            <div className={cn(
              "text-[8px] font-black uppercase",
              expense.confidence_level === 'high' ? "text-accent-secondary" : "text-warning"
            )}>
              CONFIANÇA_{expense.confidence_level.toUpperCase()}
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
