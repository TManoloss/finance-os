"use client";

import { useState } from "react";
import api from "@/lib/api";
import { 
  Zap, 
  ShoppingBag, 
  Trash2, 
  TrendingUp, 
  ArrowRight, 
  Terminal,
  Calculator,
  AlertCircle
} from "lucide-react";
import { 
  ResponsiveContainer, 
  BarChart, 
  Bar, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  Cell 
} from "recharts";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export default function SimulatorPage() {
  const [activeTab, setActiveTab] = useState<"purchase" | "cut">("purchase");
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<any>(null);

  // Form states
  const [purchaseAmount, setPurchaseAmount] = useState("");
  const [installments, setInstallments] = useState("1");
  const [cutAmount, setCutAmount] = useState("");
  const [cutName, setCutName] = useState("");

  const handleSimulatePurchase = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const resp = await api.post("simulator/purchase", {
        amount: parseFloat(purchaseAmount),
        installments: parseInt(installments),
      });
      setResult(resp.data.data);
    } catch (error) {
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const handleSimulateCut = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const resp = await api.post("simulator/cut-subscription", {
        monthly_amount: parseFloat(cutAmount),
        merchant_name: cutName,
      });
      setResult(resp.data.data);
    } catch (error) {
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="grid-blueprint grid-cols-1">
      <header className="p-8 bg-elevated border-b-2 border-black">
        <div className="flex items-center gap-2 mb-1">
          <Terminal className="w-4 h-4" />
          <span className="text-[10px] font-bold tracking-widest uppercase">SIMULATION_CORE_V1.1</span>
        </div>
        <h1 className="text-4xl font-black text-text-primary uppercase tracking-tighter">SIMULADOR "E_SE?"</h1>
      </header>

      <div className="grid grid-cols-1 lg:grid-cols-2 bg-background">
        {/* Left Side: Controls */}
        <div className="p-8 border-r-2 border-black">
          <div className="flex gap-2 mb-8">
            <button 
              onClick={() => { setActiveTab("purchase"); setResult(null); }}
              className={cn(
                "flex-1 p-4 border-2 border-black font-black text-xs uppercase transition-all flex items-center justify-center gap-2",
                activeTab === "purchase" ? "bg-black text-white" : "bg-elevated hover:bg-background"
              )}
            >
              <ShoppingBag className="w-4 h-4" /> COMPRA_PARCELADA
            </button>
            <button 
              onClick={() => { setActiveTab("cut"); setResult(null); }}
              className={cn(
                "flex-1 p-4 border-2 border-black font-black text-xs uppercase transition-all flex items-center justify-center gap-2",
                activeTab === "cut" ? "bg-black text-white" : "bg-elevated hover:bg-background"
              )}
            >
              <Trash2 className="w-4 h-4" /> CORTAR_ASSINATURA
            </button>
          </div>

          {activeTab === "purchase" ? (
            <form onSubmit={handleSimulatePurchase} className="space-y-6">
              <div className="space-y-2">
                <label className="text-[10px] font-black text-text-secondary uppercase">VALOR_DA_COMPRA (R$)</label>
                <input 
                  type="number" 
                  step="0.01"
                  required
                  value={purchaseAmount}
                  onChange={(e) => setPurchaseAmount(e.target.value)}
                  className="w-full p-4 bg-elevated border-2 border-black font-mono font-bold text-xl outline-none focus:bg-background transition-colors"
                  placeholder="0,00"
                />
              </div>
              <div className="space-y-2">
                <label className="text-[10px] font-black text-text-secondary uppercase">NUMERO_DE_PARCELAS</label>
                <select 
                  value={installments}
                  onChange={(e) => setInstallments(e.target.value)}
                  className="w-full p-4 bg-elevated border-2 border-black font-bold outline-none focus:bg-background transition-colors"
                >
                  {[1, 2, 3, 4, 5, 6, 10, 12, 18, 24].map(n => (
                    <option key={n} value={n}>{n}x</option>
                  ))}
                </select>
              </div>
              <button 
                type="submit"
                disabled={loading}
                className="w-full p-6 bg-accent-primary text-white font-black uppercase text-lg shadow-[6px_6px_0px_0px_rgba(0,0,0,1)] active:translate-x-[2px] active:translate-y-[2px] active:shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] transition-all disabled:opacity-50"
              >
                {loading ? "PROCESSANDO..." : "EXECUTAR_SIMULAÇÃO"}
              </button>
            </form>
          ) : (
            <form onSubmit={handleSimulateCut} className="space-y-6">
              <div className="space-y-2">
                <label className="text-[10px] font-black text-text-secondary uppercase">NOME_DO_SERVIÇO</label>
                <input 
                  type="text" 
                  required
                  value={cutName}
                  onChange={(e) => setCutName(e.target.value)}
                  className="w-full p-4 bg-elevated border-2 border-black font-bold outline-none focus:bg-background transition-colors"
                  placeholder="EX: NETFLIX, ACADEMIA..."
                />
              </div>
              <div className="space-y-2">
                <label className="text-[10px] font-black text-text-secondary uppercase">VALOR_MENSAL (R$)</label>
                <input 
                  type="number" 
                  step="0.01"
                  required
                  value={cutAmount}
                  onChange={(e) => setCutAmount(e.target.value)}
                  className="w-full p-4 bg-elevated border-2 border-black font-mono font-bold text-xl outline-none focus:bg-background transition-colors"
                  placeholder="0,00"
                />
              </div>
              <button 
                type="submit"
                disabled={loading}
                className="w-full p-6 bg-accent-primary text-white font-black uppercase text-lg shadow-[6px_6px_0px_0px_rgba(0,0,0,1)] active:translate-x-[2px] active:translate-y-[2px] active:shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] transition-all disabled:opacity-50"
              >
                {loading ? "PROCESSANDO..." : "EXECUTAR_SIMULAÇÃO"}
              </button>
            </form>
          )}
        </div>

        {/* Right Side: Results */}
        <div className="p-8 bg-elevated/20 flex flex-col">
          {result ? (
            <div className="space-y-8 animate-in fade-in slide-in-from-right-4 duration-500">
              <h3 className="text-xl font-black uppercase tracking-tighter flex items-center gap-2">
                <Calculator className="w-5 h-5" />
                IMPACTO_ESTIMADO
              </h3>

              <div className="grid grid-cols-2 gap-4">
                <div className="p-4 border-2 border-black bg-background">
                  <div className="text-[10px] font-black text-text-secondary uppercase mb-1">MENSAL</div>
                  <div className="text-2xl font-black font-mono">
                    R$ {result.monthly_impact?.toLocaleString('pt-BR', { minimumFractionDigits: 2 })}
                  </div>
                </div>
                <div className="p-4 border-2 border-black bg-background">
                  <div className="text-[10px] font-black text-text-secondary uppercase mb-1">ANUAL</div>
                  <div className="text-2xl font-black font-mono">
                    R$ {result.annual_impact?.toLocaleString('pt-BR', { minimumFractionDigits: 2 })}
                  </div>
                </div>
              </div>

              {result.projections && (
                <div className="space-y-4">
                  <div className="text-[10px] font-black text-text-secondary uppercase">PROJEÇÃO_DE_SALDO_RESTANTE</div>
                  <div className="h-[200px] w-full">
                    <ResponsiveContainer width="100%" height="100%">
                      <BarChart data={result.projections}>
                        <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#2A2A3A" />
                        <XAxis 
                          dataKey="month" 
                          axisLine={false} 
                          tickLine={false} 
                          tick={{ fill: "#8888A0", fontSize: 10, fontWeight: 'bold' }}
                        />
                        <YAxis hide />
                        <Tooltip 
                          cursor={{ fill: 'rgba(0,0,0,0.1)' }}
                          content={({ active, payload }) => {
                            if (active && payload && payload.length) {
                              return (
                                <div className="bg-black text-white p-2 text-[10px] font-bold">
                                  SALDO: R$ {payload[0].value?.toLocaleString('pt-BR')}
                                </div>
                              );
                            }
                            return null;
                          }}
                        />
                        <Bar dataKey="balance" radius={[4, 4, 0, 0]}>
                          {result.projections.map((entry: any, index: number) => (
                            <Cell key={`cell-${index}`} fill={entry.balance >= 0 ? "#4ECDC4" : "#FF6B6B"} />
                          ))}
                        </Bar>
                      </BarChart>
                    </ResponsiveContainer>
                  </div>
                </div>
              )}

              {result.risk_alerts && result.risk_alerts.length > 0 && (
                <div className="space-y-3">
                  {result.risk_alerts.map((alert: any, i: number) => (
                    <div key={i} className="flex gap-3 p-4 border-2 border-danger bg-danger/5">
                      <AlertCircle className="w-5 h-5 text-danger shrink-0" />
                      <div>
                        <div className="text-[10px] font-black text-danger uppercase mb-1">ALERTA_DE_RISCO</div>
                        <p className="text-xs font-bold leading-tight">{alert.message}</p>
                      </div>
                    </div>
                  ))}
                </div>
              )}

              {result.opportunity_cost && (
                <div className="p-6 border-2 border-accent-primary bg-accent-primary/5">
                  <div className="text-[10px] font-black text-accent-primary uppercase mb-2">CUSTO_DE_OPORTUNIDADE (5 ANOS)</div>
                  <p className="text-sm font-bold leading-relaxed italic mb-4">
                    "{result.opportunity_cost.narrative}"
                  </p>
                  <div className="flex items-center justify-between font-mono font-black text-accent-primary">
                    <span>SE_INVESTIDO:</span>
                    <span>R$ {result.opportunity_cost.invested_value?.toLocaleString('pt-BR')}</span>
                  </div>
                </div>
              )}
            </div>
          ) : (
            <div className="flex-1 flex flex-col items-center justify-center text-center opacity-30 grayscale">
              <Calculator className="w-16 h-16 mb-4" />
              <div className="font-black uppercase text-sm">AGUARDANDO_INPUTS...</div>
              <p className="text-[10px] font-bold uppercase mt-2">Os resultados da simulação aparecerão aqui.</p>
            </div>
          )}
        </div>
      </div>

      <footer className="p-8 border-t-2 border-black bg-elevated flex justify-between items-center text-[8px] md:text-[10px] font-black uppercase tracking-[0.2em]">
         <span>SIMULATION_ENGINE // STATUS: READY</span>
         <span>© 2026 // FINANCE_OS_CORE</span>
      </footer>
    </div>
  );
}
