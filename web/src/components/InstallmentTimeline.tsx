"use client";

import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Cell } from "recharts";
import { Clock, Info, Anchor } from "lucide-react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

interface InstallmentMonth {
  month: string;
  total: number;
}

interface InstallmentTimelineData {
  timeline: InstallmentMonth[];
  lighthouse_month?: string;
  lighthouse_narrative?: string;
  total_remaining?: number;
}

export default function InstallmentTimeline({ data }: { data: InstallmentTimelineData }) {
  if (!data || !data.timeline || data.timeline.length === 0) {
    return (
      <div className="p-8 border-2 border-black bg-elevated shadow-[6px_6px_0px_0px_rgba(0,0,0,1)]">
        <h3 className="text-[10px] font-bold text-text-secondary uppercase tracking-widest mb-4">CRONOGRAMA_DE_PARCELAS</h3>
        <p className="text-xs font-bold text-text-secondary opacity-50 uppercase tracking-tighter italic">DADOS_DE_PARCELAMENTO_NAO_DISPONIVEIS...</p>
      </div>
    );
  }

  // Encontra o maior valor para destacar (Lighthouse Month geralmente é quando as parcelas diminuem drasticamente)
  // Mas aqui vamos destacar o Lighthouse Month se ele vier explicitamente
  const maxVal = Math.max(...data.timeline.map(d => d.total));

  return (
    <div className="border-2 border-black bg-background shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] overflow-hidden">
      <div className="p-6 border-b-2 border-black bg-elevated flex justify-between items-center">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <Clock className="w-3 h-3 text-accent-primary" />
            <span className="text-[10px] font-black uppercase tracking-widest">TIMELINE_DE_COMPROMISSOS_V1.2</span>
          </div>
          <h3 className="text-xl font-black uppercase tracking-tighter">PROJECAO_DE_PARCELAS_12_MESES</h3>
        </div>
        <div className="text-right">
          <div className="text-[8px] font-black text-text-secondary uppercase mb-1">DEBITO_TOTAL_RESIDUVAL</div>
          <div className="text-lg font-black font-mono text-red-500">
            {new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(data.total_remaining || 0)}
          </div>
        </div>
      </div>

      <div className="p-8 space-y-8">
        {/* Chart Section */}
        <div className="h-64 w-full">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={data.timeline} margin={{ top: 20, right: 30, left: 20, bottom: 5 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#1e1e1e" vertical={false} />
              <XAxis 
                dataKey="month" 
                stroke="#444" 
                fontSize={10} 
                fontWeight="bold" 
                tickFormatter={(val) => val.split('-')[1] + '/' + val.split('-')[0].substring(2)}
              />
              <YAxis 
                stroke="#444" 
                fontSize={10} 
                fontWeight="bold"
                tickFormatter={(val) => `R$${val}`}
              />
              <Tooltip 
                contentStyle={{ backgroundColor: '#111', border: '2px solid #000', borderRadius: '0' }}
                itemStyle={{ color: '#7c6fff', fontWeight: 'bold', textTransform: 'uppercase', fontSize: '10px' }}
                labelStyle={{ color: '#888', marginBottom: '4px', fontSize: '10px', fontWeight: 'black' }}
                cursor={{ fill: 'rgba(124, 111, 255, 0.05)' }}
              />
              <Bar dataKey="total" radius={[2, 2, 0, 0]}>
                {data.timeline.map((entry, index) => (
                  <Cell 
                    key={`cell-${index}`} 
                    fill={entry.month === data.lighthouse_month ? '#7c6fff' : '#1e1e1e'} 
                    stroke={entry.month === data.lighthouse_month ? '#7c6fff' : '#2a2a2a'}
                    strokeWidth={2}
                  />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>

        {/* Pierre's Narrative - The Lighthouse Month */}
        {data.lighthouse_narrative && (
          <div className="p-6 border-2 border-black bg-elevated/50 relative overflow-hidden group">
            <div className="absolute top-0 right-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity">
               <Anchor className="w-24 h-24" />
            </div>
            
            <div className="flex items-center gap-2 mb-4">
              <div className="w-8 h-8 border border-black bg-accent-primary flex items-center justify-center shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]">
                 <Anchor className="w-4 h-4 text-white" />
              </div>
              <div>
                <div className="text-[10px] font-black uppercase tracking-widest text-accent-primary">MES_FAROL_IDENTIFICADO</div>
                <div className="text-xs font-bold uppercase tracking-tighter italic">A_LUZ_NO_FIM_DO_TUNEL</div>
              </div>
            </div>

            <div className="text-sm font-bold text-text-primary leading-relaxed markdown-content relative z-10">
               <ReactMarkdown remarkPlugins={[remarkGfm]}>
                  {data.lighthouse_narrative}
               </ReactMarkdown>
            </div>

            <div className="mt-6 flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-text-secondary opacity-50">
               <Info className="w-3 h-3" /> PIERRE_OPERATIONAL_INSIGHT_V4.0
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
