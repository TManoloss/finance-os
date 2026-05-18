"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { 
  LayoutDashboard, 
  ArrowLeftRight, 
  CreditCard, 
  FileText, 
  Settings, 
  ChevronLeft, 
  ChevronRight,
  MessageCircle,
  LogOut,
  Sun,
  Moon,
  Activity,
  Zap,
  ShoppingBag,
  Terminal,
  Cpu,
  Shield
} from "lucide-react";
import { signOut } from "next-auth/react";
import { useTheme } from "next-themes";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export default function Sidebar() {
  const [isCollapsed, setIsCollapsed] = useState(false);
  const [mounted, setMounted] = useState(false);
  const pathname = usePathname();
  const { theme, setTheme } = useTheme();

  useEffect(() => {
    setMounted(true);
  }, []);

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

  const toggleTheme = () => {
    setTheme(theme === "dark" ? "light" : "dark");
  };

  return (
    <aside 
      className={cn(
        "hidden md:flex bg-background border-r-2 border-border-subtle transition-all duration-300 flex-col",
        isCollapsed ? "w-20" : "w-64"
      )}
    >
      <div className="p-6 flex items-center justify-between border-b-2 border-border-subtle bg-elevated/50">
        {!isCollapsed && (
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-accent-primary animate-pulse"></div>
            <span className="text-text-primary font-black text-xl uppercase tracking-tighter">FINANCE_OS</span>
          </div>
        )}
        {isCollapsed && <span className="text-text-primary font-black text-xl mx-auto">F_OS</span>}
      </div>

      <nav className="flex-1 overflow-y-auto overflow-x-hidden scrollbar-thin scrollbar-thumb-border-subtle">
        {menuGroups.map((group) => (
          <div key={group.name} className="border-b-2 border-border-subtle last:border-b-0">
            {!isCollapsed && (
              <div className="px-4 py-2 bg-elevated/30 flex items-center gap-2">
                <group.icon className="w-3 h-3 text-text-secondary" />
                <span className="text-[8px] font-black text-text-secondary uppercase tracking-[0.2em]">{group.name}</span>
              </div>
            )}
            <div className="flex flex-col">
              {group.items.map((item) => {
                const isActive = pathname === item.href;
                return (
                  <Link 
                    key={item.href}
                    href={item.href}
                    className={cn(
                      "flex items-center p-4 transition-colors group relative",
                      isActive ? "bg-elevated/80 text-text-primary" : "text-text-secondary hover:bg-elevated/50 hover:text-text-primary"
                    )}
                  >
                    <item.icon className={cn("w-5 h-5", !isCollapsed && "mr-4")} />
                    {!isCollapsed && <span className="font-bold text-[10px] tracking-widest uppercase">{item.name}</span>}
                    {isActive && (
                      <div className={cn(
                        "absolute bg-accent-primary transition-all",
                        isCollapsed ? "inset-y-0 right-0 w-1" : "right-4 w-1.5 h-1.5"
                      )}></div>
                    )}
                  </Link>
                );
              })}
            </div>
          </div>
        ))}
      </nav>

      <div className="mt-auto flex flex-col bg-elevated/30">
        {mounted && (
          <button 
            onClick={toggleTheme}
            className={cn(
              "flex items-center p-4 border-t-2 border-border-subtle text-text-primary hover:bg-elevated transition-colors group",
              isCollapsed ? "justify-center" : "justify-start"
            )}
            title={theme === "dark" ? "MODO_CLARO" : "MODO_ESCURO"}
          >
            {theme === "dark" ? <Sun className={cn("w-5 h-5", !isCollapsed && "mr-4")} /> : <Moon className={cn("w-5 h-5", !isCollapsed && "mr-4")} />}
            {!isCollapsed && <span className="font-bold text-[10px] tracking-widest uppercase">{theme === "dark" ? "MODO_CLARO" : "MODO_ESCURO"}</span>}
          </button>
        )}

        <button 
          onClick={handleLogout}
          className={cn(
            "flex items-center p-4 border-t-2 border-border-subtle text-danger hover:bg-danger hover:text-white transition-colors group",
            isCollapsed ? "justify-center" : "justify-start"
          )}
          title="SAIR"
        >
          <LogOut className={cn("w-5 h-5", !isCollapsed && "mr-4")} />
          {!isCollapsed && <span className="font-bold text-[10px] tracking-widest uppercase">LOGOUT</span>}
        </button>

        <button 
          onClick={() => setIsCollapsed(!isCollapsed)}
          className="p-4 border-t-2 border-border-subtle flex items-center justify-center bg-elevated hover:bg-background transition-colors"
        >
          {isCollapsed ? <ChevronRight className="w-5 h-5" /> : <ChevronLeft className="w-5 h-5" />}
        </button>
      </div>
    </aside>
  );
}
