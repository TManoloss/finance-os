"use client";

import React, { useState } from "react";
import { Clock, Star, Zap, ShoppingBag, CreditCard, TrendingUp, TrendingDown, Filter } from "lucide-react";
import { format } from "date-fns";
import { ptBR } from "date-fns/locale";

interface TimelineEvent {
  id: string;
  event_type: string;
  event_date: string;
  title: string;
  narrative: string;
  event_data: any;
}

interface FinancialTimelineProps {
  events: TimelineEvent[];
}

export default function FinancialTimeline({ events }: FinancialTimelineProps) {
  const [filter, setFilter] = useState<string>("all");

  if (!events || events.length === 0) return null;

  const filteredEvents = events.filter(event => {
    if (filter === "all") return true;
    if (filter === "milestones") return ["best_month", "worst_month", "spending_peak", "debt_free", "streak_no_delivery", "streak_positive_balance"].includes(event.event_type);
    if (filter === "subscriptions") return ["new_subscription", "cancel_subscription"].includes(event.event_type);
    if (filter === "records") return ["installment_start", "installment_end", "salary_change", "lifestyle_drift"].includes(event.event_type);
    return true;
  });

  const getEventIcon = (type: string) => {
    switch (type) {
      case "new_subscription":
      case "cancel_subscription":
        return <Zap className="w-4 h-4" />;
      case "installment_start":
      case "installment_end":
        return <CreditCard className="w-4 h-4" />;
      case "salary_change":
        return <TrendingUp className="w-4 h-4" />;
      case "new_merchant_habit":
      case "abandoned_habit":
        return <ShoppingBag className="w-4 h-4" />;
      case "best_month":
      case "worst_month":
      case "spending_peak":
      case "debt_free":
      case "streak_no_delivery":
      case "streak_positive_balance":
        return <Star className="w-4 h-4" />;
      default:
        return <Clock className="w-4 h-4" />;
    }
  };

  const getEventColor = (type: string) => {
    if (["best_month", "debt_free", "streak_no_delivery", "streak_positive_balance", "installment_end"].includes(type)) return "bg-success";
    if (["worst_month", "spending_peak", "lifestyle_drift"].includes(type)) return "bg-danger";
    return "bg-accent-primary";
  };

  return (
    <div className="border-2 border-black bg-background shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] overflow-hidden">
      <div className="p-6 bg-black border-b-2 border-black flex items-center justify-between">
        <div className="flex items-center gap-3 text-white">
          <Clock className="w-6 h-6" />
          <h2 className="text-xl font-black uppercase tracking-tighter">FINANCIAL_LIFE_TIMELINE</h2>
        </div>
        <div className="flex gap-2">
          {["all", "milestones", "subscriptions", "records"].map((f) => (
            <button
              key={f}
              onClick={() => setFilter(f)}
              className={`px-3 py-1 text-[8px] font-black uppercase tracking-widest border-2 border-white transition-colors ${
                filter === f ? "bg-white text-black" : "text-white hover:bg-white/10"
              }`}
            >
              {f}
            </button>
          ))}
        </div>
      </div>

      <div className="p-8 relative">
        {/* Timeline Line */}
        <div className="absolute left-[47px] top-8 bottom-8 w-1 bg-black hidden md:block" />

        <div className="space-y-12">
          {filteredEvents.map((event, i) => (
            <div key={event.id} className="relative flex flex-col md:flex-row gap-8 items-start">
              {/* Date & Icon Column */}
              <div className="flex items-center gap-4 md:w-32 flex-shrink-0">
                <div className="text-[10px] font-black uppercase tracking-tighter text-text-secondary md:text-right flex-1">
                  {format(new Date(event.event_date), "MMM yyyy", { locale: ptBR })}
                </div>
                <div className={`w-8 h-8 border-2 border-black flex items-center justify-center text-white z-10 shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] ${getEventColor(event.event_type)}`}>
                  {getEventIcon(event.event_type)}
                </div>
              </div>

              {/* Content Card */}
              <div className="flex-1 p-6 border-2 border-black bg-elevated shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-none transition-all">
                <div className="text-[10px] font-black text-accent-primary uppercase mb-1">{event.event_type}</div>
                <h3 className="text-lg font-black uppercase tracking-tighter mb-2">{event.title}</h3>
                <p className="text-sm font-bold text-text-secondary italic leading-relaxed">
                  "{event.narrative}"
                </p>
              </div>
            </div>
          ))}
        </div>

        {filteredEvents.length === 0 && (
          <div className="py-12 text-center">
            <div className="text-xs font-black text-text-secondary uppercase tracking-widest">
              NO_EVENTS_FOUND_FOR_FILTER
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
