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
  ShoppingBag
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

  // useEffect para evitar erro de hidratação com next-themes
  useEffect(() => {
    setMounted(true);
  }, []);

  const menuItems = [
    { name: "DASHBOARD", href: "/dashboard", icon: LayoutDashboard },
    { name: "TRANSACOES", href: "/dashboard/transactions", icon: ArrowLeftRight },
    { name: "ESTABELECIMENTOS", href: "/dashboard/merchants", icon: ShoppingBag },
    { name: "SAUDE", href: "/dashboard/health", icon: Activity },
    { name: "SIMULADOR", href: "/dashboard/simulator", icon: Zap },
    { name: "CARTOES", href: "/dashboard/cards", icon: CreditCard },
    { name: "RELATORIOS", href: "/dashboard/reports", icon: FileText },
    { name: "CHAT", href: "/dashboard/chat", icon: MessageCircle },
    { name: "CONFIG", href: "/dashboard/settings", icon: Settings },
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
      <div className="p-6 flex items-center justify-between border-b-2 border-border-subtle">
        {!isCollapsed && <span className="text-text-primary font-bold text-xl uppercase tracking-tighter">FINANCE_OS</span>}
        {isCollapsed && <span className="text-text-primary font-bold text-xl mx-auto">F_OS</span>}
      </div>

      <nav className="flex-1 space-y-0">
        {menuItems.map((item) => {
          const isActive = pathname === item.href;
          return (
            <Link 
              key={item.href}
              href={item.href}
              className={cn(
                "flex items-center p-4 border-b-2 border-border-subtle transition-colors group",
                isActive ? "bg-elevated text-text-primary" : "text-text-secondary hover:bg-elevated hover:text-text-primary"
              )}
            >
              <item.icon className={cn("w-5 h-5", !isCollapsed && "mr-4")} />
              {!isCollapsed && <span className="font-bold text-xs tracking-widest">{item.name}</span>}
              {isActive && !isCollapsed && <div className="ml-auto w-2 h-2 bg-text-primary"></div>}
            </Link>
          );
        })}
      </nav>

      <div className="mt-auto flex flex-col">
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
            {!isCollapsed && <span className="font-bold text-xs tracking-widest uppercase">{theme === "dark" ? "MODO_CLARO" : "MODO_ESCURO"}</span>}
          </button>
        )}

        {!isCollapsed && (
          <div className="p-4 border-t-2 border-border-subtle bg-elevated">
            <div className="text-[10px] uppercase font-bold text-text-secondary mb-1">SYSTEM_STATUS</div>
            <p className="text-[10px] text-text-primary font-bold">PIERRE_ONLINE: OK</p>
          </div>
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
          {!isCollapsed && <span className="font-bold text-xs tracking-widest uppercase">LOGOUT</span>}
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
