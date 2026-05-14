"use client";

import { useState, useEffect } from "react";
import { formatDistanceToNow } from "date-fns";
import { ptBR } from "date-fns/locale";
import api from "@/lib/api";
import { 
  Bell, 
  AlertTriangle, 
  Info, 
  TrendingUp, 
  CreditCard, 
  DollarSign, 
  CheckCircle2, 
  Zap,
  ShoppingBag
} from "lucide-react";

interface FeedEvent {
  id: string;
  type: string;
  title: string;
  description: string;
  amount: number | null;
  severity: string;
  created_at: string;
  read_at: string | null;
}

const getEventIcon = (type: string, severity: string) => {
  switch (type) {
    case "salary_detected": return <DollarSign className="w-5 h-5 text-[#4ECDC4]" />;
    case "unusual_spending": return <AlertTriangle className="w-5 h-5 text-[#FF6B6B]" />;
    case "duplicate_charge": return <Zap className="w-5 h-5 text-[#FFD93D]" />;
    case "milestone": return <CheckCircle2 className="w-5 h-5 text-[#7C6FFF]" />;
    case "new_merchant": return <ShoppingBag className="w-5 h-5 text-[#8888A0]" />;
    default: return severity === "alert" ? <AlertTriangle className="w-5 h-5 text-[#FF6B6B]" /> : <Info className="w-5 h-5 text-[#7C6FFF]" />;
  }
};

const getSeverityStyles = (severity: string) => {
  switch (severity) {
    case "alert": return "border-l-4 border-l-[#FF6B6B] bg-[#FF6B6B]/5";
    case "warning": return "border-l-4 border-l-[#FFD93D] bg-[#FFD93D]/5";
    default: return "border-l-4 border-l-[#7C6FFF] bg-[#7C6FFF]/5";
  }
};

export default function ActivityFeed({ events: initialEvents }: { events: FeedEvent[] }) {
  const [events, setEvents] = useState(initialEvents);

  useEffect(() => {
    setEvents(initialEvents);
  }, [initialEvents]);

  const handleMarkRead = async (id: string) => {
    try {
      await api.patch(`/feed/${id}/read`);
      setEvents(prev => prev.map(e => e.id === id ? { ...e, read_at: new Date().toISOString() } : e));
    } catch (error) {
      console.error("Erro ao marcar como lido:", error);
    }
  };

  if (!events || events.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <div className="bg-[#111118] p-4 rounded-full mb-4">
          <Bell className="w-8 h-8 text-[#2A2A3A]" />
        </div>
        <p className="text-[#8888A0] text-sm">Nenhuma atividade recente por aqui.</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {events.map((event) => (
        <div 
          key={event.id}
          className={`p-4 rounded-xl border border-[#2A2A3A] transition-all hover:bg-[#1A1A24] cursor-pointer relative ${getSeverityStyles(event.severity)} ${!event.read_at ? 'ring-1 ring-[#7C6FFF]/30' : ''}`}
          onClick={() => !event.read_at && handleMarkRead(event.id)}
        >
          <div className="flex gap-4">
            <div className={`p-2 rounded-lg bg-[#1A1A24] h-fit`}>
              {getEventIcon(event.type, event.severity)}
            </div>
            <div className="flex-1">
              <div className="flex justify-between items-start mb-1">
                <h4 className="text-[#F0F0F5] font-semibold text-sm">{event.title}</h4>
                <span className="text-[#8888A0] text-[10px]">
                  {formatDistanceToNow(new Date(event.created_at), { addSuffix: true, locale: ptBR })}
                </span>
              </div>
              <p className="text-[#8888A0] text-sm leading-relaxed">{event.description}</p>
              {event.amount && (
                <p className={`text-sm font-bold mt-2 ${event.severity === 'alert' ? 'text-[#FF6B6B]' : 'text-[#F0F0F5]'}`}>
                  R$ {event.amount.toLocaleString('pt-BR', { minimumFractionDigits: 2 })}
                </p>
              )}
            </div>
          </div>
          {!event.read_at && (
            <div className="absolute top-2 right-2 w-2 h-2 bg-[#7C6FFF] rounded-full shadow-[0_0_8px_#7C6FFF]" />
          )}
        </div>
      ))}
    </div>
  );
}
