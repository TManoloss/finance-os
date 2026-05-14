import { auth } from "@/auth";
import api from "@/lib/api";
import { CreditCard, Calendar, Clock, ArrowRight, Terminal, ShieldAlert } from "lucide-react";
import { format } from "date-fns";
import { ptBR } from "date-fns/locale";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

async function getInstallments(token: string) {
  try {
    const resp = await api.get("/cards/installments", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data.data || [];
  } catch (error) {
    return [];
  }
}

async function getAccounts(token: string) {
  try {
    const resp = await api.get("/accounts", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data.data || [];
  } catch (error) {
    return [];
  }
}

export default async function CardsPage() {
  const session = await auth() as any;
  const token = session?.accessToken;
  const installments = await getInstallments(token);
  const accounts = await getAccounts(token);
  const creditAccounts = accounts.filter((acc: any) => acc.account_type === 'CREDIT' || acc.account_type === 'credit');

  return (
    <div className="grid-blueprint grid-cols-1">
      <header className="p-8 bg-elevated border-b-2 border-black flex flex-col md:flex-row md:items-center justify-between gap-6">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <Terminal className="w-4 h-4" />
            <span className="text-[10px] font-bold tracking-widest uppercase">CREDIT_MONITOR_V1.0</span>
          </div>
          <h1 className="text-4xl font-black text-text-primary uppercase tracking-tighter">GESTAO_DE_CREDITO</h1>
          <p className="text-sm font-bold text-text-secondary uppercase">CONTAS_MONITORADAS: {creditAccounts.length}</p>
        </div>
      </header>

      {/* Credit Cards Grid */}
      <section className="p-8 bg-background border-b-2 border-black">
        <h2 className="text-xs font-black uppercase tracking-[0.2em] mb-8 flex items-center gap-2">
          <div className="w-2 h-2 bg-text-primary"></div> CARTOES_ATIVOS
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {creditAccounts.length > 0 ? creditAccounts.map((acc: any) => (
            <div key={acc.id} className="relative group">
              <div className="absolute inset-0 bg-black translate-x-2 translate-y-2 group-hover:translate-x-1 group-hover:translate-y-1 transition-all"></div>
              <div className="relative p-6 border-2 border-black bg-elevated aspect-[1.6/1] flex flex-col justify-between overflow-hidden">
                <div className="flex justify-between items-start">
                  <div className="w-10 h-8 border-2 border-black rounded bg-background flex items-center justify-center opacity-50">
                     <div className="w-6 h-4 border border-black/20"></div>
                  </div>
                  <CreditCard className="w-6 h-6" />
                </div>
                <div>
                  <div className="text-xs font-black uppercase tracking-widest text-text-secondary mb-1">{acc.institution_name}</div>
                  <div className="text-xl font-black tracking-tighter uppercase mb-4">**** **** **** 4521</div>
                  <div className="flex justify-between items-end">
                    <div>
                      <div className="text-[8px] font-black uppercase text-text-secondary">SALDO_DEVEDOR</div>
                      <div className="text-lg font-black font-mono">
                        {new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(acc.balance)}
                      </div>
                    </div>
                    <div className="text-[10px] font-black bg-black text-white px-2 py-1 uppercase">ACTIVE</div>
                  </div>
                </div>
              </div>
            </div>
          )) : (
            <div className="col-span-full py-12 border-2 border-dashed border-black bg-elevated/20 flex flex-col items-center justify-center text-center">
              <p className="text-xs font-black uppercase text-text-secondary">NO_CREDIT_LINES_DETECTED</p>
            </div>
          )}
        </div>
      </section>

      {/* Installments Table */}
      <section className="bg-background">
        <div className="p-8 border-b-2 border-black bg-elevated flex items-center justify-between">
          <h2 className="text-xl font-black uppercase tracking-tighter flex items-center gap-2">
            <Clock className="w-5 h-5" /> COMPRAS_PARCELADAS_EM_ABERTO
          </h2>
          <div className="text-[10px] font-black uppercase bg-danger text-white px-3 py-1">ALERTA_DE_FLUXO</div>
        </div>
        
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-black text-white">
                <th className="px-6 py-4 text-[10px] font-black uppercase tracking-widest border-r border-white/20">MERCHANT_ID</th>
                <th className="px-6 py-4 text-[10px] font-black uppercase tracking-widest border-r border-white/20">PROGR_PAGAMENTO</th>
                <th className="px-6 py-4 text-[10px] font-black uppercase tracking-widest border-r border-white/20">PROX_VENCIMENTO</th>
                <th className="px-6 py-4 text-[10px] font-black uppercase tracking-widest text-right">VALOR_PARCELA</th>
              </tr>
            </thead>
            <tbody className="divide-y-2 border-black">
              {installments.length > 0 ? installments.map((ins: any) => (
                <tr key={ins.id} className="hover:bg-elevated transition-colors group">
                  <td className="px-6 py-4">
                    <div className="text-sm font-black text-text-primary uppercase">{ins.merchant_name}</div>
                    <div className="text-[8px] font-bold text-text-secondary uppercase">REF_ID: {ins.id.substring(0, 8)}</div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="w-full max-w-[120px] space-y-2">
                      <div className="flex justify-between text-[10px] font-black uppercase">
                        <span>{ins.installment_current}/{ins.installments_total}</span>
                        <span>{Math.round((ins.installment_current / ins.installments_total) * 100)}%</span>
                      </div>
                      <div className="h-2 border-2 border-black bg-background overflow-hidden">
                        <div 
                          className="h-full bg-accent-primary transition-all" 
                          style={{ width: `${(ins.installment_current / ins.installments_total) * 100}%` }}
                        ></div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 text-xs font-bold uppercase text-text-secondary">
                    {format(new Date(ins.next_due_date), "dd.MM.yyyy", { locale: ptBR })}
                  </td>
                  <td className="px-6 py-4 text-right">
                    <div className="text-lg font-black font-mono">
                      {new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(ins.amount)}
                    </div>
                    <div className="text-[8px] font-black uppercase text-danger">TOTAL_DEVIDO: {new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(ins.total_amount)}</div>
                  </td>
                </tr>
              )) : (
                <tr>
                  <td colSpan={4} className="px-6 py-24 text-center">
                    <div className="flex flex-col items-center gap-4">
                      <ShieldAlert className="w-8 h-8 opacity-20" />
                      <p className="text-xs font-black uppercase text-text-secondary italic">NO_INSTALLMENTS_DETECTED_IN_THE_CURRENT_CYCLE</p>
                    </div>
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </section>

      <footer className="p-8 border-t-2 border-black bg-elevated flex justify-between items-center text-[10px] font-black uppercase tracking-[0.2em]">
         <span>CREDIT_LOG // PAGE_ID: CRD_MON_01</span>
         <span>SECURE_SESSION_ACTIVE</span>
      </footer>
    </div>
  );
}
