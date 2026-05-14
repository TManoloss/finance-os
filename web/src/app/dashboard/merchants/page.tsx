import { auth } from "@/auth";
import api from "@/lib/api";
import { Terminal, ShoppingBag, TrendingUp, TrendingDown, ArrowRight, BarChart3, Calendar, Clock } from "lucide-react";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

async function getTopMerchants(token: string) {
  try {
    const resp = await api.get("/merchants?limit=20", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data.data || [];
  } catch (error) {
    return [];
  }
}

export default async function MerchantsPage() {
  const session = await auth() as any;
  const token = session?.accessToken;
  const merchants = await getTopMerchants(token);

  return (
    <div className="grid-blueprint grid-cols-1">
      <header className="p-8 bg-elevated border-b-2 border-black">
        <div className="flex items-center gap-2 mb-1">
          <Terminal className="w-4 h-4" />
          <span className="text-[10px] font-bold tracking-widest uppercase">MERCHANT_INTELLIGENCE_V1.0</span>
        </div>
        <h1 className="text-4xl font-black text-text-primary uppercase tracking-tighter">INTELIGÊNCIA_DE_ESTABELECIMENTOS</h1>
      </header>

      <div className="bg-background min-h-screen">
        <div className="p-8 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {merchants.length > 0 ? merchants.map((m: any, i: number) => (
            <div key={i} className="border-2 border-black bg-background hover:bg-elevated transition-all group shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
              <div className="p-6">
                <div className="flex justify-between items-start mb-6">
                   <div className="w-12 h-12 border-2 border-black bg-black text-white flex items-center justify-center font-black text-xl">
                      {(m.merchant || m.merchant_name || "?").charAt(0)}
                   </div>
                   <div className="text-right">
                      <div className="text-[8px] font-black text-text-secondary uppercase">TOTAL_GASTO</div>
                      <div className="text-xl font-black font-mono text-danger">
                        R$ {m.total.toLocaleString('pt-BR', { minimumFractionDigits: 2 })}
                      </div>
                   </div>
                </div>

                <h3 className="text-lg font-black uppercase tracking-tight mb-1 truncate">{m.merchant || m.merchant_name}</h3>
                <div className="flex items-center gap-2 text-[10px] font-bold text-text-secondary uppercase mb-6">
                   <ShoppingBag className="w-3 h-3" /> {m.count} OPERAÇÕES REALIZADAS
                </div>

                <div className="grid grid-cols-2 gap-4 border-t-2 border-black/5 pt-6">
                   <div>
                      <div className="text-[8px] font-black text-text-secondary uppercase mb-1">TICKET_MÉDIO</div>
                      <div className="font-mono font-bold text-sm">R$ {(m.avg_amount || m.avg_ticket || 0).toLocaleString('pt-BR')}</div>
                   </div>
                   <div className="text-right">
                      <div className="text-[8px] font-black text-text-secondary uppercase mb-1">ÚLTIMA_COMPRA</div>
                      <div className="font-mono font-bold text-xs">{(m.last_purchase || "").split('T')[0]}</div>
                   </div>
                </div>

                {m.trend && (
                  <div className={cn(
                    "mt-6 p-2 border-2 border-black text-[10px] font-black uppercase flex items-center justify-center gap-2",
                    m.trend === 'up' ? "bg-danger/10 text-danger" : "bg-accent-secondary/10 text-accent-secondary"
                  )}>
                    {m.trend === 'up' ? <TrendingUp className="w-3 h-3" /> : <TrendingDown className="w-3 h-3" />}
                    TENDÊNCIA_DE_GASTO: {m.trend === 'up' ? 'CRESCENTE' : 'REDUZINDO'}
                  </div>
                )}
              </div>
            </div>
          )) : (
            <div className="col-span-3 py-32 text-center opacity-30 grayscale flex flex-col items-center">
              <ShoppingBag className="w-16 h-16 mb-4" />
              <div className="font-black uppercase">BUFFER_DE_ESTABELECIMENTOS_VAZIO</div>
            </div>
          )}
        </div>
      </div>

      <footer className="p-8 border-t-2 border-black bg-elevated flex justify-between items-center text-[8px] md:text-[10px] font-black uppercase tracking-[0.2em]">
         <span>MERCHANT_ANALYSIS // STABLE</span>
         <span>PROXIMO_RECALCULO: SYNC_EVENT</span>
      </footer>
    </div>
  );
}
