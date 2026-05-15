"use client";

import { useState } from "react";
import { format } from "date-fns";
import { ptBR } from "date-fns/locale";
import { FileText, Loader2, BookOpen, ArrowLeft } from "lucide-react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import api from "@/lib/api";

export default function NarrativeReportView({ token }: { token: string }) {
  const [report, setReport] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const generateReport = async () => {
    setLoading(true);
    setError("");
    try {
      const now = new Date();
      const month = format(now, "MM");
      const year = format(now, "yyyy");
      const resp = await api.get(`/reports/narrative?month=${month}&year=${year}`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      
      if (resp.data.data && resp.data.data.error) {
        setError(resp.data.data.error);
        setReport(null);
      } else {
        setReport(resp.data.data);
      }
    } catch (err: any) {
      setError("FALHA_AO_GERAR_RELATORIO_NARRATIVO_SISTEMA");
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="py-32 flex flex-col items-center justify-center animate-pulse">
        <Loader2 className="w-12 h-12 mb-4 animate-spin text-accent-primary" />
        <div className="font-black uppercase tracking-widest text-xs">PIERRE_ESTÁ_ESCREVENDO_SEU_RELATÓRIO...</div>
      </div>
    );
  }

  if (report) {
    return (
      <div className="max-w-3xl mx-auto py-12 px-6 animate-in fade-in duration-1000">
        <button 
          onClick={() => setReport(null)}
          className="mb-12 flex items-center gap-2 text-[10px] font-black uppercase hover:text-accent-primary transition-colors"
        >
          <ArrowLeft className="w-3 h-3" /> VOLTAR_AOS_RELATÓRIOS_TÉCNICOS
        </button>

        <div className="space-y-12">
           <header className="border-b-2 border-black pb-8">
              <div className="text-[10px] font-black text-accent-primary uppercase tracking-[0.3em] mb-2">RELATÓRIO_NARRATIVO_MENSAL</div>
              <h2 className="text-5xl font-black uppercase tracking-tighter leading-none">
                 {report.title || "O_ESTADO_DAS_SUAS_FINANÇAS"}
              </h2>
              <p className="mt-4 text-sm font-bold text-text-secondary uppercase">
                 GERADO_EM: {format(new Date(), "dd 'DE' MMMM 'DE' yyyy", { locale: ptBR })}
              </p>
           </header>

           <div className="max-w-none">
              <div className="text-xl font-medium leading-relaxed text-text-primary markdown-content">
                 <ReactMarkdown remarkPlugins={[remarkGfm]}>
                    {report.narrative || report.content}
                 </ReactMarkdown>
              </div>
           </div>

           {report.action_items && report.action_items.length > 0 && (
             <div className="pt-12 border-t-2 border-black">
                <div className="text-[10px] font-black uppercase tracking-widest mb-6">AÇÕES_RECOMENDADAS</div>
                <div className="space-y-4">
                   {report.action_items.map((item: string, i: number) => (
                     <div key={i} className="flex gap-4 items-start p-4 bg-elevated border-l-4 border-accent-primary">
                        <div className="font-black font-mono text-accent-primary mt-1">#0{i+1}</div>
                        <p className="font-bold text-sm leading-tight">{item}</p>
                     </div>
                   ))}
                </div>
             </div>
           )}
        </div>
      </div>
    );
  }

  return (
    <div className="py-20 flex flex-col items-center justify-center text-center">
      <div className="w-24 h-24 border-2 border-black flex items-center justify-center mb-8 bg-elevated rotate-3 hover:rotate-0 transition-transform cursor-pointer" onClick={generateReport}>
         <BookOpen className="w-10 h-10" />
      </div>
      <h3 className="text-2xl font-black uppercase tracking-tighter mb-4">EXTRATO_INTELIGENTE</h3>
      <p className="max-w-md text-xs font-bold text-text-secondary uppercase leading-relaxed mb-8">
        Deixe o Pierre analisar todo o seu mês e escrever um relatório narrativo com linguagem humana sobre sua saúde financeira.
      </p>
      <button 
        onClick={generateReport}
        className="px-8 py-4 bg-black text-white font-black uppercase text-xs tracking-widest hover:bg-accent-primary transition-all shadow-[6px_6px_0px_0px_rgba(124,111,255,0.3)]"
      >
        GERAR_NARRATIVA_DO_MÊS
      </button>
      {error && <p className="mt-4 text-[10px] font-black text-danger uppercase">{error}</p>}
    </div>
  );
}
