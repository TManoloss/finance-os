"use client";

import React from "react";
import { BarChart3, TrendingUp, Activity, ArrowUpRight, ArrowDownRight } from "lucide-react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

interface Decomposition {
  category: string;
  total_a: number;
  total_b: number;
  freq_a: number;
  freq_b: number;
  avg_ticket_a: number;
  avg_ticket_b: number;
  var_total: number;
  var_freq: number;
  var_ticket: number;
  tipo_crescimento: string;
}

interface TicketAnalysisData {
  summary: string;
  top_decompositions: Decomposition[];
  insight_narrative: string;
}

interface TicketAnalysisCardProps {
  data: TicketAnalysisData;
}

export default function TicketAnalysisCard({ data }: TicketAnalysisCardProps) {
  if (!data || !data.top_decompositions) return null;

  return (
    <div className="border-2 border-black bg-background shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] overflow-hidden">
      <div className="p-6 bg-accent-primary border-b-2 border-black flex items-center justify-between">
        <div className="flex items-center gap-3">
          <BarChart3 className="w-6 h-6 text-white" />
          <h2 className="text-xl font-black text-white uppercase tracking-tighter">TICKET_DECOMPOSITION_ANALYSIS</h2>
        </div>
        <div className="px-3 py-1 bg-black text-white text-[10px] font-black uppercase tracking-widest">
          BEHAVIORAL_CORE_V4
        </div>
      </div>

      <div className="p-8 space-y-8">
        {/* Decompositions List */}
        <div className="space-y-6">
          <div className="flex items-center gap-2 text-text-secondary font-black text-[10px] uppercase tracking-widest">
            <TrendingUp className="w-3 h-3" /> CATEGORY_GROWTH_DECOMPOSITION
          </div>
          
          <div className="grid grid-cols-1 gap-6">
            {data.top_decompositions.map((item, i) => (
              <div key={i} className="p-6 border-2 border-black bg-elevated/20 space-y-4">
                <div className="flex items-center justify-between">
                  <h3 className="text-lg font-black uppercase tracking-tight">{item.category}</h3>
                  <div className={`px-2 py-1 text-[10px] font-black uppercase border-2 border-black ${item.var_total > 0 ? 'bg-danger text-white' : 'bg-accent-teal text-black'}`}>
                    {item.var_total > 0 ? '+' : ''}{(item.var_total * 100).toFixed(1)}% TOTAL
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                  <div className="space-y-2">
                    <div className="flex justify-between text-[10px] font-black text-text-secondary uppercase">
                      <span>Frequência</span>
                      <span className={item.var_freq > 0 ? 'text-danger' : 'text-accent-teal'}>
                        {item.var_freq > 0 ? '+' : ''}{(item.var_freq * 100).toFixed(1)}%
                      </span>
                    </div>
                    <div className="h-4 border-2 border-black bg-background overflow-hidden relative">
                      <div 
                        className={`h-full ${item.var_freq > 0 ? 'bg-danger' : 'bg-accent-teal'}`} 
                        style={{ width: `${Math.min(100, Math.abs(item.var_freq) * 100)}%` }}
                      />
                    </div>
                    <div className="flex justify-between text-[10px] font-bold text-text-secondary">
                      <span>{item.freq_a}x → {item.freq_b}x</span>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <div className="flex justify-between text-[10px] font-black text-text-secondary uppercase">
                      <span>Ticket Médio</span>
                      <span className={item.var_ticket > 0 ? 'text-danger' : 'text-accent-teal'}>
                        {item.var_ticket > 0 ? '+' : ''}{(item.var_ticket * 100).toFixed(1)}%
                      </span>
                    </div>
                    <div className="h-4 border-2 border-black bg-background overflow-hidden relative">
                      <div 
                        className={`h-full ${item.var_ticket > 0 ? 'bg-danger' : 'bg-accent-teal'}`} 
                        style={{ width: `${Math.min(100, Math.abs(item.var_ticket) * 100)}%` }}
                      />
                    </div>
                    <div className="flex justify-between text-[10px] font-bold text-text-secondary">
                      <span>R$ {item.avg_ticket_a.toFixed(2)} → R$ {item.avg_ticket_b.toFixed(2)}</span>
                    </div>
                  </div>
                </div>

                <div className="pt-2 border-t border-black/10">
                  <span className="text-[10px] font-black text-accent-primary uppercase tracking-widest">
                    GROWTH_DRIVER: {item.tipo_crescimento}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Narrative Analysis */}
        <div className="space-y-4">
          <div className="flex items-center gap-2 text-text-secondary font-black text-[10px] uppercase tracking-widest">
            <Activity className="w-3 h-3" /> IA_BEHAVIORAL_INSIGHT
          </div>
          <div className="p-6 border-2 border-black bg-background shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] italic text-lg font-bold leading-relaxed markdown-content border-l-8 border-l-accent-primary">
            <ReactMarkdown remarkPlugins={[remarkGfm]}>
              {data.insight_narrative}
            </ReactMarkdown>
          </div>
        </div>
      </div>
    </div>
  );
}
