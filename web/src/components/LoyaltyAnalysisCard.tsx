"use client";

import React from "react";
import { Heart, UserCheck, Activity, Trash2, ArrowRight } from "lucide-react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

interface LoyaltyMerchant {
  merchant: string;
  category: string;
  classification: string;
  total_freq: number;
  total_spent: number;
  clv: number;
  last_seen_days: number;
}

interface LoyaltyData {
  summary: string;
  insights: { title: string; description: string; severity: string }[];
  insight_narrative: string;
  abandonment_cost: number;
  loyalty_stats: {
    leal: number;
    frequente: number;
    abandonado: number;
    experimentado: number;
  };
  top_loyal: LoyaltyMerchant[];
}

interface LoyaltyAnalysisCardProps {
  data: LoyaltyData;
}

export default function LoyaltyAnalysisCard({ data }: LoyaltyAnalysisCardProps) {
  if (!data) return null;

  const classificationColors: Record<string, string> = {
    LEAL: "bg-accent-teal text-black",
    FREQUENTE: "bg-accent-primary text-white",
    ABANDONADO: "bg-danger text-white",
    EXPERIMENTADO: "bg-elevated text-text-secondary",
  };

  return (
    <div className="border-2 border-black bg-background shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] overflow-hidden">
      <div className="p-6 bg-accent-primary border-b-2 border-black flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Heart className="w-6 h-6 text-white" />
          <h2 className="text-xl font-black text-white uppercase tracking-tighter">LOYALTY_ABANDONMENT_MAP</h2>
        </div>
        <div className="px-3 py-1 bg-black text-white text-[10px] font-black uppercase tracking-widest">
          BEHAVIORAL_CORE_V4
        </div>
      </div>

      <div className="p-8 space-y-8">
        {/* Main Stats */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="p-6 border-2 border-black bg-elevated/20">
             <div className="text-[10px] font-black text-text-secondary uppercase mb-2">ABANDONMENT_COST</div>
             <div className="text-3xl font-black text-danger">R$ {data.abandonment_cost.toLocaleString("pt-BR", { minimumFractionDigits: 2 })}</div>
             <div className="mt-2 text-[10px] font-bold text-text-secondary uppercase">GASTO EM LUGARES NÃO REPETIDOS</div>
          </div>
          <div className="p-6 border-2 border-black bg-accent-teal/10 col-span-2">
             <div className="text-[10px] font-black text-text-secondary uppercase mb-4">LOYALTY_DISTRIBUTION</div>
             <div className="grid grid-cols-4 gap-4">
                {Object.entries(data.loyalty_stats).map(([key, val]) => (
                  <div key={key}>
                    <div className="text-[10px] font-black uppercase mb-1">{key}</div>
                    <div className="text-2xl font-black">{val}</div>
                  </div>
                ))}
             </div>
          </div>
        </div>

        {/* Top Loyal Merchants */}
        <div className="space-y-4">
          <div className="flex items-center gap-2 text-text-secondary font-black text-[10px] uppercase tracking-widest">
            <UserCheck className="w-3 h-3" /> TOP_LEAL_ESTABLISHMENTS
          </div>
          <div className="grid grid-cols-1 gap-2">
            {data.top_loyal.map((m, i) => (
              <div key={i} className="flex items-center justify-between p-4 border-2 border-black bg-elevated/10 group hover:bg-background transition-colors">
                <div className="flex items-center gap-4">
                   <div className={`px-2 py-0.5 text-[10px] font-black uppercase border border-black ${classificationColors[m.classification] || classificationColors.LEAL}`}>
                      {m.classification}
                   </div>
                   <div>
                      <div className="text-sm font-black uppercase">{m.merchant}</div>
                      <div className="text-[10px] font-bold text-text-secondary uppercase">{m.category}</div>
                   </div>
                </div>
                <div className="text-right">
                   <div className="text-sm font-black">R$ {m.clv.toLocaleString("pt-BR", { minimumFractionDigits: 2 })}</div>
                   <div className="text-[10px] font-bold text-text-secondary uppercase">{m.total_freq}X VISITAS</div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Narrative Analysis */}
        <div className="space-y-4">
          <div className="flex items-center gap-2 text-text-secondary font-black text-[10px] uppercase tracking-widest">
            <Activity className="w-3 h-3" /> IA_LOYALTY_INSIGHT
          </div>
          <div className="p-6 border-2 border-black bg-background shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] italic text-lg font-bold leading-relaxed markdown-content border-l-8 border-l-accent-primary">
            <ReactMarkdown remarkPlugins={[remarkGfm]}>
              {data.insight_narrative}
            </ReactMarkdown>
          </div>
        </div>

        {/* Heuristic Insights */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
           {data.insights.map((insight, i) => (
             <div key={i} className={`p-4 border-2 border-black bg-background relative overflow-hidden ${insight.severity === 'warning' ? 'border-l-4 border-l-danger' : 'border-l-4 border-l-accent-teal'}`}>
                <div className="text-[10px] font-black uppercase mb-2 text-text-secondary">{insight.title}</div>
                <p className="text-xs font-bold leading-tight">{insight.description}</p>
             </div>
           ))}
        </div>
      </div>
    </div>
  );
}
