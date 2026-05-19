"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { 
  LayoutDashboard, 
  ArrowLeftRight, 
  CreditCard, 
  MessageCircle,
  Settings,
  Menu,
  X,
  ShoppingBag,
  Activity,
  Zap,
  FileText,
  LogOut,
  Terminal,
  Cpu,
  Shield
} from "lucide-react";
import { signOut } from "next-auth/react";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export default function BottomNav() {
  const pathname = usePathname();
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  // Fecha o menu se mudar de rota
  useEffect(() => {
    setIsMenuOpen(false);
  }, [pathname]);

  const bottomItems = [
    { name: "DASH", href: "/dashboard", icon: LayoutDashboard },
    { name: "TRAN", href: "/dashboard/transactions", icon: ArrowLeftRight },
    { name: "CARD", href: "/dashboard/cards", icon: CreditCard },
    { name: "CHAT", href: "/dashboard/chat", icon: MessageCircle },
  ];

  const menuGroups = [
    {
      name: "CORE_SYSTEM",
      icon: Terminal,
      items: [
        { name: "DASHBOARD", href: "/dashboard", icon: LayoutDashboard },
        { name: "TRANSACOES", href: "/dashboard/transactions", icon: ArrowLeftRight },
        { name: "CARTOES", href: "/dashboard/cards", icon: CreditCard },
      ]
    },
    {
      name: "INTELLIGENCE",
      icon: Cpu,
      items: [
        { name: "ESTABELECIMENTOS", href: "/dashboard/merchants", icon: ShoppingBag },
        { name: "SAUDE", href: "/dashboard/health", icon: Activity },
        { name: "SIMULADOR", href: "/dashboard/simulator", icon: Zap },
        { name: "RELATORIOS", href: "/dashboard/reports", icon: FileText },
        { name: "CHAT", href: "/dashboard/chat", icon: MessageCircle },
      ]
    },
    {
      name: "MAINTENANCE",
      icon: Shield,
      items: [
        { name: "CONFIG", href: "/dashboard/settings", icon: Settings },
      ]
    }
  ];

  const handleLogout = () => {
    signOut({ callbackUrl: "/login" });
  };

  return (
    <>
      {/* Menu Overlay Deslizante */}
      {isMenuOpen && (
        <div className="md:hidden fixed inset-0 z-40 bg-background/98 backdrop-blur-md overflow-y-auto flex flex-col justify-between p-6 animate-in slide-in-from-bottom duration-300">
          <div className="space-y-8 pt-8">
            <div className="flex items-center justify-between border-b-2 border-border-subtle pb-4">
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 bg-accent-primary animate-pulse"></div>
                <span className="text-text-primary font-black text-lg uppercase tracking-tighter">SISTEMA_MENUS</span>
              </div>
              <button 
                onClick={() => setIsMenuOpen(false)}
                className="p-2 border-2 border-border-subtle bg-elevated text-text-primary active:translate-y-[1px] rounded"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="space-y-6">
              {menuGroups.map((group) => (
                <div key={group.name} className="space-y-2">
                  <div className="flex items-center gap-2 px-1">
                    <group.icon className="w-3 h-3 text-text-secondary" />
                    <span className="text-[9px] font-black text-text-secondary uppercase tracking-[0.2em]">{group.name}</span>
                  </div>
                  <div className="grid grid-cols-2 gap-2">
                    {group.items.map((item) => {
                      const isActive = pathname === item.href;
                      return (
                        <Link 
                          key={item.href}
                          href={item.href}
                          onClick={() => setIsMenuOpen(false)}
                          className={cn(
                            "flex items-center gap-3 p-3.5 border-2 transition-all rounded font-bold text-xs uppercase tracking-wider active:scale-[0.98]",
                            isActive 
                              ? "bg-accent-primary/10 border-accent-primary text-accent-primary" 
                              : "bg-elevated border-border-subtle text-text-primary hover:border-text-primary"
                          )}
                        >
                          <item.icon className="w-4 h-4 shrink-0" />
                          <span>{item.name}</span>
                        </Link>
                      );
                    })}
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="border-t-2 border-border-subtle pt-6 mt-8">
            <button 
              onClick={handleLogout}
              className="w-full flex items-center justify-center gap-3 p-4 border-2 border-danger/40 bg-danger/10 text-danger hover:bg-danger hover:text-white transition-all font-black text-xs uppercase tracking-widest rounded active:scale-[0.98]"
            >
              <LogOut className="w-4 h-4" />
              <span>ENCERRAR_SESSÃO</span>
            </button>
          </div>
        </div>
      )}

      {/* Barra de Navegação Inferior Principal */}
      <nav className="md:hidden fixed bottom-0 left-0 right-0 bg-background border-t-2 border-border-subtle z-50 flex items-center justify-around h-16 pb-safe shadow-[0_-4px_12px_rgba(0,0,0,0.08)]">
        {bottomItems.map((item) => {
          const isActive = pathname === item.href && !isMenuOpen;
          return (
            <Link 
              key={item.href}
              href={item.href}
              className={cn(
                "flex flex-col items-center justify-center flex-1 h-full transition-all active:bg-elevated relative",
                isActive ? "text-accent-primary" : "text-text-secondary"
              )}
            >
              <item.icon className="w-5 h-5 shrink-0" />
              <span className="text-[8px] font-black uppercase mt-1 tracking-tighter">{item.name}</span>
              {isActive && (
                <div className="absolute top-0 w-8 h-1 bg-accent-primary rounded-b"></div>
              )}
            </Link>
          );
        })}

        {/* Botão de Menu para abrir o Drawer com tudo */}
        <button 
          onClick={() => setIsMenuOpen(!isMenuOpen)}
          className={cn(
            "flex flex-col items-center justify-center flex-1 h-full transition-all active:bg-elevated relative",
            isMenuOpen ? "text-accent-primary" : "text-text-secondary"
          )}
        >
          {isMenuOpen ? <X className="w-5 h-5 shrink-0" /> : <Menu className="w-5 h-5 shrink-0" />}
          <span className="text-[8px] font-black uppercase mt-1 tracking-tighter">MENUS</span>
          {isMenuOpen && (
            <div className="absolute top-0 w-8 h-1 bg-accent-primary rounded-b"></div>
          )}
        </button>
      </nav>
    </>
  );
}
