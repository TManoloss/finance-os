import { auth } from "@/auth";
import apiServer from "@/lib/api-server";
import { format } from "date-fns";
import { ptBR } from "date-fns/locale";
import { ArrowUpCircle, ArrowDownCircle, Search, Filter, ChevronLeft, ChevronRight, Terminal, Download } from "lucide-react";
import Link from "next/link";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import CategoryFilter from "@/components/CategoryFilter";

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

async function getTransactions(token: string, page = 1, categoryId?: string) {
  try {
    let url = `transactions?page=${page}&page_size=15`;
    if (categoryId && categoryId !== "all") {
      url += `&category_id=${categoryId}`;
    }
    const resp = await apiServer.get(url, {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data.data;
  } catch (error) {
    return { transactions: [], total: 0, total_pages: 0 };
  }
}

async function getCategories(token: string) {
  try {
    const resp = await apiServer.get("categories", {
      headers: { Authorization: `Bearer ${token}` }
    });
    return resp.data.data || [];
  } catch (error) {
    return [];
  }
}

export default async function TransactionsPage({
  searchParams,
}: {
  searchParams: Promise<{ page?: string; category_id?: string }>;
}) {
  const session = await auth() as any;
  const token = session?.accessToken;
  const params = await searchParams;
  const currentPage = Number(params.page) || 1;
  const currentCategoryId = params.category_id || "all";

  const [transactionsData, categories] = await Promise.all([
    getTransactions(token, currentPage, currentCategoryId),
    getCategories(token)
  ]);

  const { transactions = [], total_pages = 0, total = 0 } = transactionsData || {};
  const txList = transactions || [];

  return (
    <div className="grid-blueprint grid-cols-1">
      <header className="p-4 md:p-8 bg-elevated border-b-2 border-black flex flex-col lg:flex-row lg:items-center justify-between gap-6">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <Terminal className="w-4 h-4" />
            <span className="text-[10px] font-bold tracking-widest uppercase">TRANSACTION_MODULE_V1.0</span>
          </div>
          <h1 className="text-2xl md:text-4xl font-black text-text-primary uppercase tracking-tighter">EXTRATO_DETALHADO</h1>
          <p className="text-[10px] md:text-sm font-bold text-text-secondary uppercase tracking-widest">REGISTROS_ENCONTRADOS: {total}</p>
        </div>
        
        <div className="flex flex-col sm:flex-row flex-wrap items-stretch sm:items-center gap-3">
          <div className="relative group flex-1 sm:flex-initial">
            <div className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-black group-focus-within:text-accent-primary transition-colors">
               <Search className="w-full h-full" />
            </div>
            <input 
              type="text" 
              placeholder="PESQUISAR..." 
              className="bg-background border-2 border-black py-2.5 pl-10 pr-4 text-[10px] md:text-xs font-bold uppercase focus:outline-none focus:bg-elevated transition-all w-full sm:w-64 shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]"
            />
          </div>
          <div className="flex gap-2">
            <CategoryFilter categories={categories} />
            <button className="flex-1 sm:flex-none flex items-center justify-center gap-2 bg-black text-white px-4 py-2.5 text-[10px] md:text-xs font-black uppercase hover:bg-text-secondary transition-all shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] active:shadow-none active:translate-x-[2px] active:translate-y-[2px]">
              <Download className="w-3 h-3 md:w-4 h-4" /> CSV
            </button>
          </div>
        </div>
      </header>

      <div className="bg-background">
        {/* Desktop Table View */}
        <div className="hidden md:block overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="bg-black text-white">
                <th className="px-6 py-4 text-xs font-black uppercase tracking-widest border-r border-white/20">TIMESTAMP</th>
                <th className="px-6 py-4 text-xs font-black uppercase tracking-widest border-r border-white/20">IDENTIFICADOR_DESCRICAO</th>
                <th className="px-6 py-4 text-xs font-black uppercase tracking-widest border-r border-white/20">FONTE_DADOS</th>
                <th className="px-6 py-4 text-xs font-black uppercase tracking-widest border-r border-white/20">METADATA_CATEGORIA</th>
                <th className="px-6 py-4 text-xs font-black uppercase tracking-widest text-right">VALOR_LIQUIDO</th>
              </tr>
            </thead>
            <tbody className="divide-y-2 border-black">
              {txList.length > 0 ? txList.map((tx: any) => (
                <tr key={tx.id} className="hover:bg-elevated transition-colors group">
                  <td className="px-6 py-4 text-xs font-bold text-text-secondary border-r-2 border-black uppercase whitespace-nowrap">
                    {format(new Date(tx.date), "dd.MM.yyyy", { locale: ptBR })}
                  </td>
                  <td className="px-6 py-4 border-r-2 border-black">
                    <div className="flex items-center gap-4">
                      <div className={cn(
                        "w-6 h-6 flex items-center justify-center border-2 border-black text-[10px] font-black",
                        tx.direction === 'credit' ? "bg-accent-secondary text-white" : "bg-danger text-white"
                      )}>
                        {tx.direction === 'credit' ? 'IN' : 'OUT'}
                      </div>
                      <span className="text-sm font-black text-text-primary uppercase tracking-tight">{tx.description}</span>
                    </div>
                  </td>
                  <td className="px-6 py-4 border-r-2 border-black text-[10px] font-black uppercase text-text-secondary">
                    {tx.account_name}
                  </td>
                  <td className="px-6 py-4 border-r-2 border-black">
                    <span className="px-2 py-1 bg-black text-white text-[10px] font-black uppercase tracking-tighter">
                      {tx.category?.name || 'NULL_CATEGORY'}
                    </span>
                  </td>
                  <td className={cn(
                    "px-6 py-4 text-lg font-black font-mono text-right tabular-nums",
                    tx.direction === 'credit' ? 'text-accent-secondary' : 'text-text-primary'
                  )}>
                    {tx.direction === 'debit' ? '-' : '+'}
                    {new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(tx.amount)}
                  </td>
                </tr>
              )) : (
                <tr>
                  <td colSpan={5} className="px-6 py-24 text-center text-text-secondary uppercase font-black text-sm italic bg-elevated/20">
                    ERROR: NO_TRANSACTION_RECORDS_FOUND_IN_BUFFER
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        {/* Mobile Card View */}
        <div className="md:hidden divide-y-2 border-black">
          {txList.length > 0 ? txList.map((tx: any) => (
            <div key={tx.id} className="p-4 flex items-start justify-between gap-4 hover:bg-elevated transition-colors group">
              <div className="flex items-start gap-3 min-w-0">
                <div className={cn(
                  "w-8 h-8 flex items-center justify-center border-2 border-black text-[8px] font-black shrink-0 mt-1",
                  tx.direction === 'credit' ? "bg-accent-secondary text-white" : "bg-danger text-white"
                )}>
                  {tx.direction === 'credit' ? 'IN' : 'OUT'}
                </div>
                <div className="min-w-0">
                  <div className="text-xs font-black text-text-primary uppercase tracking-tight truncate leading-tight mb-1">{tx.description}</div>
                  <div className="flex flex-wrap gap-2 items-center">
                    <span className="text-[9px] font-bold text-text-secondary uppercase">
                      {format(new Date(tx.date), "dd.MM.yyyy", { locale: ptBR })}
                    </span>
                    <span className="text-[8px] font-black text-black/40 uppercase">
                      {tx.account_name}
                    </span>
                    <span className="px-1.5 py-0.5 bg-black text-white text-[8px] font-black uppercase tracking-tighter">
                      {tx.category?.name || 'NULL'}
                    </span>
                  </div>
                </div>
              </div>
              <div className={cn(
                "text-sm font-black font-mono tabular-nums text-right whitespace-nowrap",
                tx.direction === 'credit' ? 'text-accent-secondary' : 'text-text-primary'
              )}>
                {tx.direction === 'debit' ? '-' : '+'}
                {new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(tx.amount)}
              </div>
            </div>
          )) : (
            <div className="px-4 py-16 text-center text-text-secondary uppercase font-black text-xs italic bg-elevated/20">
              BUFFER_EMPTY
            </div>
          )}
        </div>
      </div>

      {/* Pagination */}
      <div className="p-4 md:p-8 border-t-2 border-black flex flex-col md:flex-row items-center justify-between gap-4 md:gap-6 bg-elevated">
        <p className="text-[10px] md:text-xs font-black uppercase tracking-widest text-text-secondary">
          PAG: {currentPage} / TOTAL: {total_pages || 1}
        </p>
        <div className="flex gap-3 w-full sm:w-auto">
          {currentPage > 1 ? (
            <Link 
              href={`/dashboard/transactions?page=${currentPage - 1}${currentCategoryId !== 'all' ? `&category_id=${currentCategoryId}` : ''}`}
              className="flex-1 sm:flex-none flex items-center justify-center gap-2 px-4 py-2.5 md:px-6 md:py-3 bg-background border-2 border-black text-[10px] md:text-xs font-black uppercase hover:bg-black hover:text-white transition-all shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] active:shadow-none active:translate-x-[2px] active:translate-y-[2px]"
            >
              <ChevronLeft className="w-3 h-3 md:w-4 h-4" /> PREV
            </Link>
          ) : (
            <button className="flex-1 sm:flex-none flex items-center justify-center gap-2 px-4 py-2.5 md:px-6 md:py-3 bg-background border-2 border-black text-[10px] md:text-xs font-black uppercase opacity-20 cursor-not-allowed" disabled>
              <ChevronLeft className="w-3 h-3 md:w-4 h-4" /> PREV
            </button>
          )}

          {currentPage < total_pages ? (
            <Link 
              href={`/dashboard/transactions?page=${currentPage + 1}${currentCategoryId !== 'all' ? `&category_id=${currentCategoryId}` : ''}`}
              className="flex-1 sm:flex-none flex items-center justify-center gap-2 px-4 py-2.5 md:px-6 md:py-3 bg-background border-2 border-black text-[10px] md:text-xs font-black uppercase hover:bg-black hover:text-white transition-all shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] active:shadow-none active:translate-x-[2px] active:translate-y-[2px]"
            >
              NEXT <ChevronRight className="w-3 h-3 md:w-4 h-4" />
            </Link>
          ) : (
            <button className="flex-1 sm:flex-none flex items-center justify-center gap-2 px-4 py-2.5 md:px-6 md:py-3 bg-background border-2 border-black text-[10px] md:text-xs font-black uppercase opacity-20 cursor-not-allowed" disabled>
              NEXT <ChevronRight className="w-3 h-3 md:w-4 h-4" />
            </button>
          )}
        </div>
      </div>

      <footer className="p-8 border-t-2 border-black bg-black text-white flex justify-between items-center text-[10px] font-black uppercase tracking-[0.2em]">
         <span>SYSTEM_LOG // PAGE_ID: TX_EXTR_01</span>
         <span>END_OF_REPORT</span>
      </footer>
    </div>
  );
}
