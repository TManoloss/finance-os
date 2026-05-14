"use client";

import { useState } from "react";
import { useSession } from "next-auth/react";
import { useRouter } from "next/navigation";
import { Terminal, Key, ShieldCheck, CheckCircle2, Loader2, AlertTriangle } from "lucide-react";
import api from "@/lib/api";

export default function OnboardingPluggyPage() {
  const { data: session, update } = useSession();
  const [clientId, setClientId] = useState("");
  const [clientSecret, setClientSecret] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [loading, setLoading] = useState(false);
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");
    setSuccess("");

    try {
      const response = await api.post(
        "/accounts/keys",
        {
          client_id: clientId,
          client_secret: clientSecret,
        },
        {
          headers: {
            Authorization: `Bearer ${(session as any)?.accessToken}`,
          },
        }
      );

      if (response.data.success) {
        setSuccess("SYSTEM_UPDATE: CREDENCIAIS_PLUGGY_ATIVADAS");
        
        // Atualiza a sessão local para incluir o novo pluggyClientId
        await update({
          ...session,
          user: {
            ...session?.user,
            pluggyClientId: clientId,
          },
          pluggyClientId: clientId,
        });

        setTimeout(() => {
          router.push("/dashboard");
          router.refresh();
        }, 1500);
      } else {
        setError("API_ERROR: FALHA_AO_SALVAR_CHAVES");
        setLoading(false);
      }
    } catch (err: any) {
      console.error(err);
      setError(
        err.response?.data?.error || "CORE_FAULT: ERRO_DE_CONEXAO_COM_O_SERVIDOR"
      );
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4 font-mono">
      <div className="w-full max-w-md border-4 border-black bg-background shadow-[12px_12px_0px_0px_rgba(0,0,0,1)] overflow-hidden">
        {/* Terminal Header */}
        <div className="bg-black text-white p-4 flex items-center justify-between">
           <div className="flex items-center gap-2">
              <Terminal className="w-4 h-4" />
              <span className="text-xs font-black uppercase tracking-widest">ONBOARDING_INITIALIZATION_V1.0</span>
           </div>
           <div className="flex gap-1">
              <div className="w-2 h-2 rounded-full bg-white/20"></div>
              <div className="w-2 h-2 rounded-full bg-white/20"></div>
              <div className="w-2 h-2 rounded-full bg-success animate-pulse"></div>
           </div>
        </div>

        <div className="p-8 space-y-8">
          <div className="text-center">
            <h1 className="text-3xl font-black text-text-primary uppercase tracking-tighter mb-2">CONFIG_PLUGGY</h1>
            <p className="text-[10px] font-bold text-text-secondary uppercase tracking-[0.2em]">INTEGRAÇÃO_OPEN_FINANCE_REQUERIDA</p>
          </div>

          <div className="bg-elevated/50 border-2 border-dashed border-black p-4 space-y-2">
            <p className="text-[10px] font-bold text-text-primary uppercase leading-relaxed">
              O FINANCE_OS utiliza a API da Pluggy para conectar com seus bancos de forma segura. 
              Insira suas chaves de desenvolvedor para continuar a sincronização do core.
            </p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-6">
            {error && (
              <div className="bg-danger text-white p-4 text-xs font-black uppercase border-2 border-black flex items-center gap-2">
                <AlertTriangle className="w-4 h-4 shrink-0" /> {error}
              </div>
            )}

            {success && (
              <div className="bg-success text-white p-4 text-xs font-black uppercase border-2 border-black flex items-center gap-2">
                <CheckCircle2 className="w-4 h-4 shrink-0" /> {success}
              </div>
            )}

            <div className="space-y-4">
              <div className="space-y-2">
                <label className="block text-[10px] font-black uppercase tracking-widest text-text-secondary">PLUGGY_CLIENT_ID</label>
                <div className="relative">
                   <Key className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-black" />
                   <input
                    type="text"
                    value={clientId}
                    onChange={(e) => setClientId(e.target.value)}
                    className="w-full bg-elevated border-2 border-black p-4 pl-10 text-sm font-bold uppercase focus:outline-none focus:bg-background transition-all"
                    placeholder="CLIENT_ID_HEX"
                    required
                  />
                </div>
              </div>

              <div className="space-y-2">
                <label className="block text-[10px] font-black uppercase tracking-widest text-text-secondary">PLUGGY_CLIENT_SECRET</label>
                <div className="relative">
                   <ShieldCheck className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-black" />
                   <input
                    type="password"
                    value={clientSecret}
                    onChange={(e) => setClientSecret(e.target.value)}
                    className="w-full bg-elevated border-2 border-black p-4 pl-10 text-sm font-bold uppercase focus:outline-none focus:bg-background transition-all"
                    placeholder="••••••••••••••••"
                    required
                  />
                </div>
              </div>
            </div>

            <button
              type="submit"
              disabled={loading || success !== ""}
              className="w-full bg-black text-white font-black py-4 uppercase tracking-[0.2em] hover:bg-text-secondary transition-all active:translate-x-[2px] active:translate-y-[2px] flex items-center justify-center gap-2"
            >
              {loading ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  EXECUTANDO_SYNC...
                </>
              ) : success ? (
                "SISTEMA_PRONTO"
              ) : (
                "SALVAR_E_CONTINUAR"
              )}
            </button>
          </form>
        </div>
        
        <div className="bg-elevated p-3 border-t-2 border-black flex justify-between items-center">
           <div className="text-[8px] font-black text-black/30 uppercase">MODULE: OPEN_FINANCE_SYNC</div>
           <div className="text-[8px] font-black text-black/30 uppercase">STATUS: AWAITING_KEYS</div>
        </div>
      </div>
    </div>
  );
}
