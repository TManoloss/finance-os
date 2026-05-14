"use client";

import { signOut } from "next-auth/react";
import { LogOut, Terminal } from "lucide-react";

export default function MobileHeader() {
  return (
    <header className="md:hidden flex items-center justify-between p-4 border-b-4 border-black bg-elevated sticky top-0 z-40">
      <div className="flex items-center gap-2 text-text-primary">
        <Terminal className="w-5 h-5" />
        <span className="font-black text-xs uppercase tracking-tighter">FINANCE_OS</span>
      </div>
      <button 
        onClick={() => signOut({ callbackUrl: "/login" })}
        className="flex items-center gap-2 bg-danger text-white px-3 py-1.5 text-[10px] font-black uppercase border-2 border-black shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] active:shadow-none active:translate-x-[1px] active:translate-y-[1px]"
      >
        <LogOut className="w-3 h-3" /> SAIR
      </button>
    </header>
  );
}
