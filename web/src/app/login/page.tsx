"use client";

import { useState, useEffect, Suspense } from "react";
import { signIn } from "next-auth/react";
import { useRouter, useSearchParams } from "next/navigation";
import { Terminal, ShieldCheck, Lock, AtSign, CheckCircle2 } from "lucide-react";
import Link from "next/link";

function LoginForm() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [loading, setLoading] = useState(false);
  const router = useRouter();
  const searchParams = useSearchParams();

  useEffect(() => {
    if (searchParams.get("registered") === "true") {
      setSuccess("CORE_UPDATE: OPERADOR_REGISTRADO_COM_SUCESSO");
    }
  }, [searchParams]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");
    setSuccess("");

    const result = await signIn("credentials", {
      email,
      password,
      redirect: false,
    });

    if (result?.error) {
      setError("AUTH_ERROR: CREDENCIAIS_INVALIDAS");
      setLoading(false);
    } else {
      router.push("/dashboard");
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-3 md:p-4 font-mono">
      <div className="w-full max-w-md border-4 border-black bg-background shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] md:shadow-[12px_12px_0px_0px_rgba(0,0,0,1)] overflow-hidden">
        {/* Terminal Header */}
        <div className="bg-black text-white p-3 md:p-4 flex items-center justify-between">
           <div className="flex items-center gap-2">
              <Terminal className="w-4 h-4" />
              <span className="text-[10px] font-black uppercase tracking-widest">SYSTEM_AUTH_V1.0</span>
           </div>
           <div className="flex gap-1">
              <div className="w-2 h-2 rounded-full bg-white/20"></div>
              <div className="w-2 h-2 rounded-full bg-white/20"></div>
              <div className="w-2 h-2 rounded-full bg-white/50"></div>
           </div>
        </div>

        <div className="p-6 md:p-8 space-y-8 md:space-y-10">
          <div className="text-center">
            <h1 className="text-3xl md:text-4xl font-black text-text-primary uppercase tracking-tighter mb-2">FINANCE_OS</h1>
            <p className="text-[9px] md:text-[10px] font-bold text-text-secondary uppercase tracking-[0.2em]">CORE_ENGINE_LOGIN_REQUIRED</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-6 md:space-y-8">
            {error && (
              <div className="bg-danger text-white p-3 md:p-4 text-[10px] md:text-xs font-black uppercase border-2 border-black">
                {error}
              </div>
            )}

            {success && (
              <div className="bg-success text-white p-3 md:p-4 text-[10px] md:text-xs font-black uppercase border-2 border-black flex items-center gap-2">
                <CheckCircle2 className="w-4 h-4" /> {success}
              </div>
            )}

            <div className="space-y-4 md:space-y-6">
              <div className="space-y-2">
                <label className="block text-[10px] font-black uppercase tracking-widest text-text-secondary">INPUT_IDENTITY</label>
                <div className="relative">
                   <AtSign className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-black" />
                   <input
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="w-full bg-elevated border-2 border-black p-3 md:p-4 pl-10 text-sm font-bold uppercase focus:outline-none focus:bg-background transition-all"
                    placeholder="USER@DOMAIN.COM"
                    required
                  />
                </div>
              </div>

              <div className="space-y-2">
                <label className="block text-[10px] font-black uppercase tracking-widest text-text-secondary">SECURE_CREDENTIAL</label>
                <div className="relative">
                   <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-black" />
                   <input
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="w-full bg-elevated border-2 border-black p-3 md:p-4 pl-10 text-sm font-bold uppercase focus:outline-none focus:bg-background transition-all"
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
              {loading ? "INITIALIZING..." : "EXECUTE_LOGIN"}
            </button>
          </form>

          <div className="pt-6 border-t-2 border-dashed border-black/20 flex items-center justify-between">
            <div className="flex items-center gap-2 text-success">
               <ShieldCheck className="w-4 h-4" />
               <span className="text-[8px] font-black uppercase">ENCRYPTED_AUTH</span>
            </div>
            <Link href="/register" className="text-[8px] font-black uppercase text-text-secondary hover:text-black underline">REQUEST_ACCESS</Link>
          </div>
        </div>
        
        <div className="bg-elevated p-3 border-t-2 border-black flex justify-center">
           <div className="text-[8px] font-black text-black/30 uppercase">BUILD_592.10.X_KNL</div>
        </div>
      </div>
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center bg-background font-mono">
        <div className="font-black uppercase tracking-widest text-xs animate-pulse">CARREGANDO_SUBSISTEMA_DE_AUTENTICAÇÃO...</div>
      </div>
    }>
      <LoginForm />
    </Suspense>
  );
}
