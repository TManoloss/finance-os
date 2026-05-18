import { auth } from "@/auth";
import api from "@/lib/api";
import AreaChartComponent from "@/components/charts/AreaChartComponent";
import CashflowChart from "@/components/charts/CashflowChart";
import DonutChartComponent from "@/components/charts/DonutChartComponent";
import ProjectionChart from "@/components/charts/ProjectionChart";
import UpcomingExpenses from "@/components/UpcomingExpenses";
import ActivityFeed from "@/components/ActivityFeed";
import { ArrowRight, Zap, Database, Terminal, TrendingUp, BarChart3 } from "lucide-react";
import { format, startOfMonth, endOfMonth, startOfQuarter, endOfQuarter, startOfYear, endOfYear, subMonths } from "date-fns";
import { ptBR } from "date-fns/locale";
import Link from "next/link";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

async function getSummary(token: string, from?: string, to?: string) {
  if (!token) return null;
  try {
    let url = "/transactions/summary";
    if (from && to) {
      url += `?from_date=${from}&to_date=${to}`;
    }
    const resp = await api.get(url, {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data.data;
  } catch (error) {
    return null;
  }
}

async function getReports(token: string) {
  if (!token) return [];
  try {
    const resp = await api.get("/reports", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data.data || [];
  } catch (error) {
    return [];
  }
}

async function getCashflow(token: string, from: string, to: string) {
  if (!token) return [];
  try {
    const resp = await api.get(`/reports/cashflow?from=${from}&to=${to}`, {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data.data?.daily_balances || [];
  } catch (error) {
    return [];
  }
}

async function getFeed(token: string) {
  if (!token) return [];
  try {
    const resp = await api.get("/feed?page_size=5", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data.data || [];
  } catch (error) {
    return [];
  }
}

async function getLatestTransactions(token: string) {
  if (!token) return [];
  try {
    const resp = await api.get("/transactions?page=1&page_size=10", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data.data.transactions || [];
  } catch (error) {
    return [];
  }
}

async function getProjections(token: string) {
  if (!token) return null;
  try {
    const resp = await api.get("/reports/projection", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data.data;
  } catch (error) {
    return null;
  }
}

async function getUpcomingExpenses(token: string) {
  if (!token) return [];
  try {
    const resp = await api.get("/reports/upcoming-expenses", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data.data || [];
  } catch (error) {
    return [];
  }
}

export default async function DashboardPage({
  searchParams,
}: {
  searchParams: Promise<{ [key: string]: string | string[] | undefined }>
}) {
  const session = await auth() as any;
  if (!session) return null;

  const params = await searchParams;
  const period = (params.period as string) || "month";
  const token = session?.accessToken;

  let fromDate: string | undefined;
  let toDate: string | undefined;

  const now = new Date();
  if (period === "month") {
    fromDate = format(startOfMonth(now), "yyyy-MM-dd");
    toDate = format(endOfMonth(now), "yyyy-MM-dd");
  } else if (period === "quarter") {
    fromDate = format(startOfQuarter(now), "yyyy-MM-dd");
    toDate = format(endOfQuarter(now), "yyyy-MM-dd");
  } else if (period === "semester") {
    fromDate = format(subMonths(now, 6), "yyyy-MM-dd");
    toDate = format(now, "yyyy-MM-dd");
  } else if (period === "year") {
    fromDate = format(startOfYear(now), "yyyy-MM-dd");
    toDate = format(endOfYear(now), "yyyy-MM-dd");
  }

  const [
    summary,
    ,
    latestTransactions,
    cashflowData,
    feedEvents,
    projectionData,
    upcomingExpensesRaw
  ] = await Promise.all([
    getSummary(token, fromDate, toDate),
    getReports(token),
    getLatestTransactions(token),
    fromDate && toDate ? getCashflow(token, fromDate, toDate) : Promise.resolve([]),
    getFeed(token),
    getProjections(token),
    getUpcomingExpenses(token)
  ]);

  const upcomingExpenses = Array.isArray(upcomingExpensesRaw) ? upcomingExpensesRaw : [];

  const checkingBalance = summary?.checking_balance || 0;
  const currentInvoice = summary?.current_invoice || 0;
  const closedInvoice = summary?.closed_invoice || 0;
  const monthInstallments = summary?.month_installments || 0;
  const todaySpent = summary?.today_spent || 0;
  const weeklySpent = summary?.weekly_spent || 0;
  const totalReceived = summary?.total_received || 0;
  
  // Saldo Disponível Real = Dinheiro em Conta - (Fatura Fechada + Parcelas do Mês)
  // Ignoramos a "Fatura em Aberto" aqui porque ela só vence no mês que vem
  const balance = checkingBalance - (closedInvoice + monthInstallments);

  const stats = [
    { label: "SALDO_EM_CONTA", value: checkingBalance, color: "text-accent-secondary" },
    { label: "FATURA_FECHADA", value: closedInvoice + monthInstallments, color: "text-danger" },
    { label: "FATURA_EM_ABERTO", value: currentInvoice, color: "text-warning" },
    { label: "TOTAL_RECEBIDO_MES", value: totalReceived, color: "text-accent-secondary" },
  ];

  const quickStats = [
    { label: "GASTO_HOJE", value: todaySpent, color: "text-danger", icon: Zap },
    { label: "GASTO_SEMANA", value: weeklySpent, color: "text-danger", icon: TrendingUp },
  ];

  const periods = [
    { id: "month", label: "MÊS" },
    { id: "quarter", label: "TRIMESTRE" },
    { id: "semester", label: "SEMESTRE" },
    { id: "year", label: "ANO" },
    { id: "all", label: "TUDO" },
  ];

  return (
    <div className="grid-blueprint grid-cols-1">
      {/* Header Area */}
      <header className="p-4 md:p-8 bg-elevated flex flex-col md:flex-row justify-between items-start md:items-end gap-4">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <Terminal className="w-4 h-4" />
            <span className="text-[10px] font-bold tracking-widest uppercase">FINANCE_CORE_V1.1</span>
          </div>
          <h1 className="text-2xl md:text-4xl font-black text-text-primary uppercase tracking-tighter">
            OPERADOR: {session?.user?.name || "USUARIO"}
          </h1>
          <div className="mt-2 flex items-center gap-4">
            <div>
              <p className="text-[10px] font-black text-text-secondary uppercase">SALDO_TOTAL_LIQUIDO</p>
              <p className={cn("text-2xl font-black font-mono", balance >= 0 ? "text-accent-secondary" : "text-danger")}>
                {new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(balance)}
              </p>
            </div>
            <div className="h-10 w-0.5 bg-black/10 hidden md:block"></div>
            <p className="text-[10px] md:text-sm font-bold text-text-secondary hidden md:block">ESTADO_ATUAL: {new Date().toLocaleDateString('pt-BR', { day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit' }).replace(',', ' /')}</p>
          </div>
        </div>
        <div className="flex flex-wrap gap-2">
          {periods.map((p) => (
            <Link
              key={p.id}
              href={`/dashboard?period=${p.id}`}
              className={cn(
                "px-3 py-1.5 md:px-4 md:py-2 border-2 border-black font-bold text-[10px] md:text-xs transition-all",
                period === p.id ? "bg-black text-white" : "bg-background hover:bg-elevated"
              )}
            >
              {p.label}
            </Link>
          ))}
        </div>
      </header>

      {/* Stats Grid - Desktop: 6 in line, Tablet: 3+3, Mobile: 2+2+2 */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 bg-background">
        {stats.map((card) => (
          <div key={card.label} className="p-4 md:p-6 border-b-2 border-r-2 last:border-r-0 border-t-2 border-black hover:bg-elevated transition-colors">
            <div className="text-[8px] font-black text-text-secondary uppercase tracking-widest mb-2 flex items-center gap-2">
              <div className="w-1 h-1 bg-text-secondary"></div>
              {card.label}
            </div>
            <div className={cn("text-lg md:text-xl font-black font-mono tabular-nums", card.color)}>
              {new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(card.value)}
            </div>
          </div>
        ))}
        {quickStats.map((card) => (
          <div key={card.label} className="p-4 md:p-6 border-b-2 border-r-2 last:border-r-0 border-t-2 border-black flex flex-col justify-center hover:bg-elevated transition-colors">
            <div className="text-[8px] font-black text-text-secondary uppercase tracking-widest mb-2 flex items-center gap-2">
              <card.icon className="w-3 h-3" />
              {card.label}
            </div>
            <div className={cn("text-lg md:text-xl font-black font-mono tabular-nums", card.color)}>
              {new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(card.value)}
            </div>
          </div>
        ))}
      </div>

      {/* Main Content Area */}
      <div className="p-4 md:p-5 flex flex-col gap-4">
        {/* Cashflow — full width */}
        <div className="bg-surface border border-border-subtle p-3.5">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-xs font-black uppercase tracking-tighter flex items-center gap-2">
              <Database className="w-4 h-4" />
              FLUXO_DE_CAIXA_REAL_TIME
            </h3>
            <span className="text-[8px] font-bold bg-black text-white px-2 py-0.5">STATUS: MONITORANDO</span>
          </div>
          <div className="h-[250px] md:h-[300px]">
            {cashflowData && cashflowData.length > 0 ? (
              <CashflowChart data={cashflowData} />
            ) : summary?.by_day && summary.by_day.length > 0 ? (
              <AreaChartComponent data={summary.by_day} />
            ) : (
              <div className="h-full flex items-center justify-center text-text-secondary border-2 border-dashed border-black uppercase text-[10px] font-bold bg-elevated/50">
                NO_DATA_STREAM_FOUND
              </div>
            )}
          </div>
        </div>

        {/* Bottom row - Desktop: 2fr 3fr, Others: Stacked */}
        <div className="grid grid-cols-1 lg:grid-cols-[2fr_3fr] gap-4">
          {/* Donut Chart */}
          <div className="bg-surface border border-border-subtle p-3.5">
            <h3 className="text-xs font-black uppercase tracking-tighter mb-4">
              DISTRIBUICAO_CATEGORIAS
            </h3>
            <div className="h-[250px] md:h-[300px]">
              {summary?.by_category && summary.by_category.length > 0 ? (
                <DonutChartComponent data={summary.by_category} />
              ) : (
                <div className="h-full flex items-center justify-center text-text-secondary border-2 border-dashed border-black uppercase text-[10px] font-bold bg-elevated/50">
                  PENDING_CATEGORIZATION
                </div>
              )}
            </div>
          </div>

          {/* Activity Feed */}
          <div className="bg-surface border border-border-subtle p-3.5">
            <ActivityFeed events={feedEvents} />
          </div>
        </div>
      </div>


      {/* New Section: Top Merchants & Investments Preview */}
      <div className="grid grid-cols-1 lg:grid-cols-2 bg-background border-t-2 border-black">
        <div className="p-4 md:p-8 border-b-2 lg:border-b-0 lg:border-r-2 border-black">
           <h3 className="text-lg md:text-xl font-black uppercase tracking-tighter mb-6 md:mb-8 flex items-center gap-2">
              <TrendingUp className="w-5 h-5" />
              PRINCIPAIS_DESTINOS_DE_GASTOS
           </h3>
           <div className="space-y-3 md:space-y-4">
              {summary?.top_merchants?.map((m: any, i: number) => (
                <div key={i} className="flex justify-between items-center p-3 md:p-4 border-2 border-black bg-elevated">
                  <div>
                    <div className="font-black uppercase text-xs md:text-sm">{m.merchant_name}</div>
                    <div className="text-[8px] md:text-[10px] font-bold text-text-secondary">{m.count} OPERAÇÕES</div>
                  </div>
                  <div className="font-mono font-black text-sm md:text-base text-danger">
                    {new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(m.total)}
                  </div>
                </div>
              ))}
              {(!summary?.top_merchants || summary.top_merchants.length === 0) && (
                <div className="p-6 md:p-8 border-2 border-dashed border-black text-center text-[10px] font-bold uppercase">SEM_DADOS_DE_MERCANTES</div>
              )}
           </div>
        </div>

        <div className="p-4 md:p-8">
           <div className="flex justify-between items-center mb-6 md:mb-8">
              <h3 className="text-lg md:text-xl font-black uppercase tracking-tighter flex items-center gap-2">
                <BarChart3 className="w-5 h-5" />
                CENTRO_DE_INVESTIMENTOS
              </h3>
              <Link href="/dashboard/investments" className="text-[10px] font-black underline">FULL</Link>
           </div>
           <div className="p-8 md:p-12 border-2 border-black bg-elevated flex flex-col items-center justify-center text-center">
              <TrendingUp className="w-8 h-8 md:w-12 md:h-12 mb-4 text-accent-secondary opacity-50" />
              <div className="font-black uppercase text-xs md:text-sm mb-2">MODULO_EM_DESENVOLVIMENTO</div>
              <p className="text-[8px] md:text-[10px] font-bold text-text-secondary max-w-xs leading-tight">
                Acompanhe o crescimento do seu patrimonio, dividendos e performance de ativos em tempo real.
              </p>
           </div>
        </div>
      </div>

      {/* Phase 11: Projections & Seasonality */}
      <div className="grid grid-cols-1 lg:grid-cols-3 bg-background border-t-2 border-black">
        <div className="lg:col-span-2 p-4 md:p-8 border-b-2 lg:border-b-0 lg:border-r-2 border-black">
          <h3 className="text-lg md:text-xl font-black uppercase tracking-tighter mb-6 md:mb-8 flex items-center gap-2">
            <TrendingUp className="w-5 h-5" />
            PROJEÇÃO_ESTRATÉGICA_90_DIAS
          </h3>
          <ProjectionChart data={projectionData} />
          {projectionData?.message && (
            <p className="mt-4 text-[10px] font-bold text-text-secondary uppercase italic">
              INSIGHT_IA: {projectionData.message}
            </p>
          )}
        </div>

        <div className="p-4 md:p-8">
          <h3 className="text-lg md:text-xl font-black uppercase tracking-tighter mb-6 md:mb-8 flex items-center gap-2">
            <Zap className="w-5 h-5 fill-current text-accent-primary" />
            PRÓXIMAS_DESPESAS_RELEVANTES
          </h3>
          <UpcomingExpenses expenses={upcomingExpenses} />
        </div>
      </div>

      {/* Latest Transactions Table-ish */}
      <section className="bg-background border-t-2 border-black">
        <div className="p-4 md:p-8 border-b-2 border-black flex items-center justify-between bg-elevated">
          <h3 className="text-lg md:text-xl font-black uppercase tracking-tighter">ULTIMAS_OPERACOES</h3>
          <Link 
            href="/dashboard/transactions" 
            className="px-3 py-1.5 md:px-4 md:py-2 border-2 border-black bg-background font-bold text-[10px] md:text-xs hover:bg-black hover:text-white transition-all flex items-center gap-2"
          >
            RELATORIO <ArrowRight className="w-3 h-3" />
          </Link>
        </div>
        <div className="divide-y-2 border-black">
          {latestTransactions.length > 0 ? latestTransactions.map((tx: any) => (
            <div key={tx.id} className="p-4 md:p-6 flex flex-col sm:flex-row sm:items-center justify-between hover:bg-elevated transition-colors group gap-4 sm:gap-0">
              <div className="flex items-center gap-4 md:gap-6">
                <div className={cn(
                  "w-10 h-10 md:w-12 md:h-12 flex items-center justify-center border-2 border-black font-bold text-base md:text-lg shrink-0",
                  tx.direction === 'credit' ? "bg-accent-secondary text-white" : "bg-danger text-white"
                )}>
                  {tx.direction === 'credit' ? '+' : '-'}
                </div>
                <div className="min-w-0">
                  <div className="text-sm md:text-lg font-black text-text-primary uppercase tracking-tight truncate">{tx.description}</div>
                  <div className="text-[8px] md:text-[10px] font-bold text-text-secondary uppercase">
                    {format(new Date(tx.date), "dd/MM/yyyy", { locale: ptBR })} | {tx.account_name} | {tx.category?.name || 'NULL'}
                  </div>
                </div>
              </div>
              <div className={cn(
                "text-lg md:text-2xl font-black font-mono tabular-nums",
                tx.direction === 'credit' ? "text-accent-secondary" : "text-text-primary"
              )}>
                {new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(tx.amount)}
              </div>
            </div>
          )) : (
            <div className="p-12 md:p-16 text-center text-text-secondary uppercase font-bold text-[10px] md:text-xs italic bg-elevated/20">
              BUFFER_EMPTY: NENHUMA_OPERACAO_RECENTE
            </div>
          )}
        </div>
      </section>
      
      <footer className="p-4 md:p-8 border-t-2 border-black bg-elevated flex flex-col sm:flex-row justify-between items-center text-[8px] md:text-[10px] font-black uppercase tracking-[0.2em] gap-2">
         <span>FINANCE_OS // V1.1.0</span>
         <span>© 2026 // ALL_RIGHTS_RESERVED</span>
      </footer>
    </div>
  );
}
