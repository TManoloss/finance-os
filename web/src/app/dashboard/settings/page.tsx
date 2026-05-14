"use client";

import { useEffect, useState } from "react";
import dynamic from "next/dynamic";
import { useSession } from "next-auth/react";
import api from "@/lib/api";
import { Plus, Building2, RefreshCw, Terminal, ShieldCheck, Trash2 } from "lucide-react";
import { useTheme } from "next-themes";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

// Carrega o widget da Pluggy dinamicamente para evitar erro de SSR (window is not defined)
const PluggyConnect = dynamic(
  () => import("react-pluggy-connect").then((mod) => mod.PluggyConnect),
  { ssr: false }
);

export default function SettingsPage() {
  const { data: session } = useSession() as any;
  const token = session?.accessToken;
  const { theme } = useTheme();
  const [connectToken, setConnectToken] = useState("");
  const [showWidget, setShowWidget] = useState(false);
  const [accounts, setAccounts] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (token) {
      fetchAccounts();
    }
  }, [token]);

  const fetchAccounts = async () => {
    if (!token) return;
    try {
      const resp = await api.get("/accounts", {
        headers: { Authorization: `Bearer ${token}` }
      });
      setAccounts(resp.data.data || []);
    } catch (err) {
      console.error("Erro ao buscar contas", err);
    } finally {
      setLoading(false);
    }
  };

  const handleOpenWidget = async () => {
    if (!token) return;
    try {
      const resp = await api.post("/accounts/connect-token", {}, {
        headers: { Authorization: `Bearer ${token}` }
      });
      setConnectToken(resp.data.data.accessToken);
      setShowWidget(true);
    } catch (err: any) {
      alert("ERRO_CONFIG: Falha ao inicializar handshake com Pluggy API.");
    }
  };

  const handleSync = async (itemID: string) => {
    if (!token) return;
    try {
      await api.post("/accounts/sync", { item_id: itemID }, {
        headers: { Authorization: `Bearer ${token}` }
      });
      alert("SYNC_COMMAND_SENT: Sincronização em segundo plano inicializada.");
    } catch (err) {
      alert("SYNC_ERROR: Falha no comando de sincronização.");
    }
  };

  const handleUpdateSettings = async (accountID: string, closeDay: number, dueDay: number) => {
    if (!token) return;
    try {
      await api.patch(`/accounts/${accountID}/settings`, {
        close_day: closeDay,
        due_day: dueDay
      }, {
        headers: { Authorization: `Bearer ${token}` }
      });
      fetchAccounts();
    } catch (err) {
      alert("UPDATE_ERROR: Falha ao salvar configurações da conta.");
    }
  };

  const handleDeleteAccount = async (accountID: string, institutionName: string) => {
    if (!token) return;
    if (!confirm(`Deseja realmente desconectar a conta ${institutionName}? Todos os dados de transações associados serão removidos permanentemente.`)) {
      return;
    }

    try {
      await api.delete(`/accounts/${accountID}`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      fetchAccounts();
      alert("ACCOUNT_DISCONNECTED: Fonte de dados removida com sucesso.");
    } catch (err) {
      alert("DELETE_ERROR: Falha ao remover fonte de dados.");
    }
  };

  return (
    <div className="grid-blueprint grid-cols-1">
      <header className="p-8 bg-elevated border-b-2 border-black flex flex-col md:flex-row md:items-center justify-between gap-6">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <Terminal className="w-4 h-4" />
            <span className="text-[10px] font-bold tracking-widest uppercase">CONFIG_ENVIRONMENT_V1.0</span>
          </div>
          <h1 className="text-4xl font-black text-text-primary uppercase tracking-tighter">CONFIGURACOES_DO_SISTEMA</h1>
          <p className="text-sm font-bold text-text-secondary uppercase">IDENTIFICADOR_USUARIO: {session?.user?.id || "NULL"}</p>
        </div>
      </header>

      {/* Seção Open Finance */}
      <section className="bg-background">
        <div className="p-8 border-b-2 border-black flex flex-col md:flex-row md:items-center justify-between gap-4 bg-elevated/50">
          <h2 className="text-xl font-black flex items-center gap-2 uppercase tracking-tighter">
            <Building2 className="w-5 h-5" /> FONTES_DE_DADOS_CONECTADAS
          </h2>
          <button 
            onClick={handleOpenWidget}
            className="flex items-center gap-2 bg-black text-white px-6 py-3 text-xs font-black uppercase hover:bg-text-secondary transition-all shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] active:shadow-none active:translate-x-[2px] active:translate-y-[2px]"
          >
            <Plus className="w-4 h-4" /> ADICIONAR_NOVA_FONTE
          </button>
        </div>

        <div className="p-8 grid grid-cols-1 gap-6">
          {accounts.length > 0 ? accounts.map((acc: any) => (
            <div key={acc.id} className="p-6 border-2 border-black bg-elevated flex flex-col md:flex-row md:items-center justify-between gap-6 hover:bg-background transition-colors group">
              <div className="flex items-center gap-6">
                {acc.institution_logo ? (
                  <div 
                    className="w-14 h-14 bg-white border-2 border-black flex items-center justify-center p-2 shadow-inner overflow-hidden"
                    style={{ borderColor: acc.institution_color || '#000000' }}
                  >
                    <img 
                      src={acc.institution_logo} 
                      alt={acc.institution_name} 
                      className="w-full h-full object-contain"
                    />
                  </div>
                ) : (
                  <div className="w-14 h-14 bg-background border-2 border-black flex items-center justify-center text-2xl shadow-inner uppercase font-black">
                    {acc.institution_name?.substring(0, 2)}
                  </div>
                )}
                <div>
                  <h3 className="font-black text-text-primary uppercase text-lg tracking-tight">{acc.institution_name}</h3>
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="text-[10px] font-bold text-text-secondary uppercase tracking-widest">TIPO: {acc.account_type}</span>
                    <div className="w-1 h-1 rounded-full bg-text-secondary"></div>
                    <span className="text-[10px] font-black text-text-primary uppercase font-mono">
                      {new Intl.NumberFormat('pt-BR', { style: 'currency', currency: acc.currency || 'BRL' }).format(acc.balance)}
                    </span>
                  </div>
                  {(acc.account_type === 'CREDIT' || acc.account_type === 'credit') && (
                    <div className="mt-4 flex flex-col sm:flex-row items-start sm:items-center gap-4 p-3 border-2 border-dashed border-black/20 bg-background/50">
                      <div className="flex items-center gap-2">
                        <label className="text-[8px] font-black uppercase text-text-secondary">FECHAMENTO (DIA):</label>
                        <input 
                          type="number" 
                          min="1" max="31"
                          defaultValue={acc.close_day}
                          onBlur={(e) => handleUpdateSettings(acc.id, parseInt(e.target.value), acc.due_day)}
                          className="w-12 bg-elevated border-2 border-black p-1 text-[10px] font-black text-center focus:outline-none"
                        />
                      </div>
                      <div className="flex items-center gap-2">
                        <label className="text-[8px] font-black uppercase text-text-secondary">VENCIMENTO (DIA):</label>
                        <input 
                          type="number" 
                          min="1" max="31"
                          defaultValue={acc.due_day}
                          onBlur={(e) => handleUpdateSettings(acc.id, acc.close_day, parseInt(e.target.value))}
                          className="w-12 bg-elevated border-2 border-black p-1 text-[10px] font-black text-center focus:outline-none"
                        />
                      </div>
                    </div>
                  )}
                </div>
              </div>
              
              <div className="flex items-center gap-8">
                <div className="text-right">
                  <div className="text-[10px] font-black text-text-secondary uppercase mb-1">LAST_SYNC_TIMESTAMP</div>
                  <div className="text-xs font-bold text-text-primary font-mono">
                    {acc.last_synced_at ? new Date(acc.last_synced_at).toLocaleString('pt-BR') : 'NOT_SYNCED_YET'}
                  </div>
                </div>
                <button 
                  onClick={() => handleSync(acc.pluggy_item_id)}
                  className="p-4 border-2 border-black bg-background text-text-primary hover:bg-black hover:text-white transition-all shadow-[3px_3px_0px_0px_rgba(0,0,0,1)] active:shadow-none active:translate-x-[1px] active:translate-y-[1px]"
                  title="EXECUTE_SYNC"
                >
                  <RefreshCw className="w-5 h-5" />
                </button>
                <button 
                  onClick={() => handleDeleteAccount(acc.id, acc.institution_name)}
                  className="p-4 border-2 border-black bg-background text-danger hover:bg-danger hover:text-white transition-all shadow-[3px_3px_0px_0px_rgba(0,0,0,1)] active:shadow-none active:translate-x-[1px] active:translate-y-[1px]"
                  title="DISCONNECT_SOURCE"
                >
                  <Trash2 className="w-5 h-5" />
                </button>
              </div>
            </div>
          )) : !loading && (
            <div className="py-24 border-2 border-dashed border-black bg-elevated/20 flex flex-col items-center justify-center text-center">
              <div className="w-16 h-16 border-2 border-black flex items-center justify-center mb-6 text-2xl grayscale opacity-50 font-black">?</div>
              <h3 className="text-text-primary font-black uppercase text-xl mb-2 tracking-tighter">NENHUMA_FONTE_DETECTADA</h3>
              <p className="text-text-secondary text-xs font-bold uppercase max-w-xs leading-relaxed">Conecte sua primeira conta via Open Finance para inicializar o fluxo de dados do Pierre.</p>
            </div>
          )}
        </div>
      </section>

      {/* Security Info */}
      <section className="p-8 border-t-2 border-black bg-background flex flex-col md:flex-row gap-8 items-start">
         <div className="p-6 border-2 border-black bg-success/10 flex-1">
            <div className="flex items-center gap-3 mb-4 text-success">
               <ShieldCheck className="w-6 h-6" />
               <h4 className="font-black uppercase tracking-tighter text-lg">PROTOCOLO_SEGURANCA</h4>
            </div>
            <p className="text-xs font-bold uppercase leading-relaxed text-text-secondary">Seus dados são criptografados via AES-256 e acessados apenas através de tokens efêmeros da Pluggy API (Open Finance Brasil).</p>
         </div>
         <div className="p-6 border-2 border-black bg-elevated w-full md:w-80">
            <div className="text-[10px] font-black uppercase text-text-secondary mb-2">SYSTEM_AUTH_STATUS</div>
            <div className="flex items-center gap-2">
               <div className="w-2 h-2 rounded-full bg-success animate-pulse"></div>
               <span className="text-xs font-black uppercase">TOKEN_ATUAL_VALIDO</span>
            </div>
         </div>
      </section>

      {/* Widget Pluggy Connect */}
      {showWidget && connectToken && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm p-4">
          <div className="w-full max-w-2xl bg-background border-4 border-black rounded-none shadow-[10px_10px_0px_0px_rgba(0,0,0,1)] relative">
            <div className="p-4 border-b-4 border-black bg-elevated flex items-center justify-between">
               <span className="text-xs font-black uppercase">WIDGET_CONEXAO_PLUGGY</span>
               <button 
                  onClick={() => setShowWidget(false)}
                  className="font-black hover:text-danger"
                >
                  [X]_FECHAR
                </button>
            </div>
            <PluggyConnect
              connectToken={connectToken}
              includeSandbox={true}
              onSuccess={(itemData: any) => {
                handleSync(itemData.item.id);
                setShowWidget(false);
                fetchAccounts();
              }}
              onError={(error: any) => {
                alert("ERRO_CONEXAO: O processo de autenticação falhou.");
              }}
            />
          </div>
        </div>
      )}

      <footer className="p-8 border-t-2 border-black bg-black text-white flex justify-between items-center text-[10px] font-black uppercase tracking-[0.2em]">
         <span>SYSTEM_LOG // PAGE_ID: CFG_ENV_01</span>
         <span>END_OF_CONFIG</span>
      </footer>
    </div>
  );
}
