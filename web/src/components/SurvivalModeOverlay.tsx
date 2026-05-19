"use client";

import { useState, useEffect } from "react";
import api from "@/lib/api";
import { AlertTriangle, ShieldAlert, X, ChevronRight, Ban, MessageSquareWarning } from "lucide-react";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

interface SurvivalRecommendation {
  id: string;
  title: string;
  description: string;
  priority: string;
  type: string;
}

interface SurvivalModeStatus {
  risk_score: number;
  level: string;
  is_active: boolean;
  projected_shortfall: number;
  days_until_salary: number;
  recommendations: SurvivalRecommendation[];
}

export default function SurvivalModeOverlay() {
  const [data, setData] = useState<SurvivalModeStatus | null>(null);
  const [isVisible, setIsVisible] = useState(true);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchSurvival() {
      try {
        const resp = await api.get("reports/survival-mode");
        setData(resp.data.data);
      } catch (error) {
        console.error("Erro ao buscar modo sobrevivência:", error);
      } finally {
        setLoading(false);
      }
    }
    fetchSurvival();
  }, []);

  if (loading || !data || !isVisible || data.level === "TRANQUILO") return null;

  if (data.level === "CRÍTICO") {
    return (
      <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/80 backdrop-blur-sm p-4 overflow-y-auto">
        <div className="w-full max-w-2xl bg-danger border-4 border-black p-6 shadow-[16px_16px_0px_0px_rgba(0,0,0,1)] relative animate-in fade-in zoom-in duration-300">
          <button 
            onClick={() => setIsVisible(false)}
            className="absolute top-4 right-4 p-2 hover:bg-black/10 transition-colors"
          >
            <X className="w-6 h-6 text-black" />
          </button>

          <div className="flex items-center gap-4 mb-6">
            <div className="p-3 bg-black text-danger">
              <ShieldAlert className="w-8 h-8" />
            </div>
            <div>
              <h2 className="text-2xl font-black text-black uppercase tracking-tighter">MODO_SOBREVIVÊNCIA_ATIVO</h2>
              <p className="text-sm font-bold text-black/80 uppercase tracking-tight">RISCO_CRÍTICO_DETECTADO // PROTOCOLO_DE_CONTENÇÃO</p>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-8">
            <div className="bg-black/10 p-4 border-2 border-black">
              <div className="text-[10px] font-black text-black/60 uppercase mb-1">DÉFICIT_PROJETADO</div>
              <div className="text-3xl font-black text-black font-mono">
                -{new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(data.projected_shortfall)}
              </div>
            </div>
            <div className="bg-black/10 p-4 border-2 border-black">
              <div className="text-[10px] font-black text-black/60 uppercase mb-1">DIAS_PARA_SALÁRIO</div>
              <div className="text-3xl font-black text-black font-mono">
                {data.days_until_salary}D
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <h3 className="text-sm font-black text-black uppercase flex items-center gap-2">
              <Ban className="w-4 h-4" /> RECOMENDAÇÕES_URGENTES
            </h3>
            <div className="space-y-2">
              {data.recommendations?.map((rec) => (
                <div key={rec.id} className="bg-white/90 p-3 border-2 border-black flex items-start gap-3 group hover:bg-white transition-colors">
                  <div className="mt-1 p-1 bg-black text-white shrink-0">
                    <ChevronRight className="w-3 h-3" />
                  </div>
                  <div>
                    <div className="text-xs font-black text-black uppercase">{rec.title}</div>
                    <div className="text-[10px] font-bold text-black/70 leading-tight uppercase">{rec.description}</div>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="mt-8 flex justify-end">
            <button 
              onClick={() => setIsVisible(false)}
              className="bg-black text-white px-6 py-3 font-black text-sm uppercase border-2 border-black shadow-[4px_4px_0px_0px_rgba(255,255,255,0.3)] active:shadow-none active:translate-x-[2px] active:translate-y-[2px] transition-all"
            >
              ENTENDIDO_E_MONITORANDO
            </button>
          </div>
        </div>
      </div>
    );
  }

  if (data.level === "ATENÇÃO" || data.level === "PRESSÃO") {
    return (
      <div className={cn(
        "sticky top-0 z-50 p-3 flex items-center justify-between border-b-4 border-black transition-colors duration-500",
        data.level === "ATENÇÃO" ? "bg-warning" : "bg-orange-500"
      )}>
        <div className="flex items-center gap-3">
          <AlertTriangle className="w-5 h-5 text-black" />
          <div>
            <span className="text-xs font-black text-black uppercase tracking-tighter">ALERTA_DE_ESTADO: {data.level}</span>
            <span className="hidden md:inline-block ml-4 text-[10px] font-bold text-black/70 uppercase">
              {data.days_until_salary} dias para o próximo salário // Déficit: R$ {data.projected_shortfall.toFixed(2)}
            </span>
          </div>
        </div>
        <div className="flex items-center gap-4">
          <button 
            className="text-[10px] font-black underline uppercase hover:text-black/70 transition-colors"
            onClick={() => {/* Abrir modal de detalhes se necessário */}}
          >
            VER_PROTOCOLOS
          </button>
          <button onClick={() => setIsVisible(false)}>
            <X className="w-4 h-4 text-black" />
          </button>
        </div>
      </div>
    );
  }

  return null;
}
