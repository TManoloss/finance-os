"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import api from "@/lib/api";
import { Terminal, ShieldCheck, User, AtSign, Lock } from "lucide-react";
import Link from "next/link";

export default function RegisterPage() {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      const resp = await api.post("auth/register", {
        name,
        email,
        password,
      });

      if (resp.data.success) {
        router.push("/login?registered=true");
      }
    } catch (err: any) {
      setError(err.response?.data?.error || "ERR_CORE: FALHA_AO_REGISTRAR_OPERADOR");
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
              <span className="text-xs font-black uppercase tracking-widest">USER_REGISTRATION_V1.0</span>
           </div>
           <div className="flex gap-1">
              <div className="w-2 h-2 rounded-full bg-white/20"></div>
              <div className="w-2 h-2 rounded-full bg-white/20"></div>
              <div className="w-2 h-2 rounded-full bg-white/50"></div>
           </div>
        </div>

        <div className="p-8 space-y-8">
          <div className="text-center">
            <h1 className="text-4xl font-black text-text-primary uppercase tracking-tighter mb-2">NOVO_OPERADOR</h1>
            <p className="text-[10px] font-bold text-text-secondary uppercase tracking-[0.2em]">INITIALIZE_IDENTITY_BUFFER</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-6">
            {error && (
              <div className="bg-danger text-white p-4 text-xs font-black uppercase border-2 border-black">
                {error}
              </div>
            )}

            <div className="space-y-4">
              <div className="space-y-1">
                <label className="block text-[10px] font-black uppercase tracking-widest text-text-secondary">NAME_IDENTIFIER</label>
                <div className="relative">
                   <User className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-black" />
                   <input
                    type="text"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    className="w-full bg-elevated border-2 border-black p-4 pl-10 text-sm font-bold uppercase focus:outline-none focus:bg-background transition-all"
                    placeholder="NOME_COMPLETO"
                    required
                  />
                </div>
              </div>

              <div className="space-y-1">
                <label className="block text-[10px] font-black uppercase tracking-widest text-text-secondary">EMAIL_ADDRESS</label>
                <div className="relative">
                   <AtSign className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-black" />
                   <input
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="w-full bg-elevated border-2 border-black p-4 pl-10 text-sm font-bold uppercase focus:outline-none focus:bg-background transition-all"
                    placeholder="EMAIL@DOMAIN.COM"
                    required
                  />
                </div>
              </div>

              <div className="space-y-1">
                <label className="block text-[10px] font-black uppercase tracking-widest text-text-secondary">SECURITY_KEY</label>
                <div className="relative">
                   <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-black" />
                   <input
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="w-full bg-elevated border-2 border-black p-4 pl-10 text-sm font-bold uppercase focus:outline-none focus:bg-background transition-all"
                    placeholder="••••••••"
                    required
                  />
                </div>
              </div>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-black text-white font-black py-4 uppercase tracking-[0.3em] hover:bg-text-secondary transition-all active:translate-x-[2px] active:translate-y-[2px]"
            >
              {loading ? "PROCESSING..." : "REGISTER_SYSTEM"}
            </button>
          </form>

          <div className="pt-6 border-t-2 border-dashed border-black/20 flex items-center justify-between">
            <div className="flex items-center gap-2 text-success">
               <ShieldCheck className="w-4 h-4" />
               <span className="text-[8px] font-black uppercase">CORE_ENCRYPTION_READY</span>
            </div>
            <Link href="/login" className="text-[8px] font-black uppercase text-text-secondary hover:text-black underline">BACK_TO_LOGIN</Link>
          </div>
        </div>
        
        <div className="bg-elevated p-3 border-t-2 border-black flex justify-center">
           <div className="text-[8px] font-black text-black/30 uppercase">BUILD_592.10.X_KNL</div>
        </div>
      </div>
    </div>
  );
}
