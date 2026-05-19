"use client";

import { useState } from "react";
import { Zap, Loader2, CheckCircle2 } from "lucide-react";
import api from "@/lib/api";
import { useSession } from "next-auth/react";

export default function TriggerReportButtons() {
  const { data: session } = useSession();
  const [loading, setLoading] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const handleTrigger = async (type: string) => {
    setLoading(type);
    setSuccess(null);
    try {
      await api.post(`reports/trigger/${type}`, {}, {
        headers: { Authorization: `Bearer ${(session as any)?.accessToken}` }
      });
      setSuccess(type);
      setTimeout(() => setSuccess(null), 3000);
    } catch (error) {
      console.error("Erro ao disparar agente:", error);
    } finally {
      setLoading(null);
    }
  };

  const types = [
    { id: "daily", label: "DIÁRIO" },
    { id: "weekly", label: "SEMANAL" },
    { id: "monthly", label: "MENSAL" },
  ];

  return (
    <div className="flex flex-wrap gap-3">
      {types.map((t) => (
        <button
          key={t.id}
          onClick={() => handleTrigger(t.id)}
          disabled={!!loading}
          className="flex items-center gap-2 bg-black text-white px-4 py-2 text-[10px] font-black uppercase hover:bg-accent-primary transition-all shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] active:shadow-none active:translate-x-[2px] active:translate-y-[2px] disabled:opacity-50"
        >
          {loading === t.id ? (
            <Loader2 className="w-3 h-3 animate-spin" />
          ) : success === t.id ? (
            <CheckCircle2 className="w-3 h-3 text-success" />
          ) : (
            <Zap className="w-3 h-3" />
          )}
          GERAR_{t.label}
        </button>
      ))}
    </div>
  );
}
