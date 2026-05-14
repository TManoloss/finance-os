"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { 
  LayoutDashboard, 
  ArrowLeftRight, 
  CreditCard, 
  FileText, 
  MessageCircle,
  Settings
} from "lucide-react";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export default function BottomNav() {
  const pathname = usePathname();

  const menuItems = [
    { name: "DASH", href: "/dashboard", icon: LayoutDashboard },
    { name: "TRAN", href: "/dashboard/transactions", icon: ArrowLeftRight },
    { name: "CARD", href: "/dashboard/cards", icon: CreditCard },
    { name: "CHAT", href: "/dashboard/chat", icon: MessageCircle },
    { name: "CONF", href: "/dashboard/settings", icon: Settings },
  ];

  return (
    <nav className="md:hidden fixed bottom-0 left-0 right-0 bg-background border-t-4 border-black z-50 flex items-center justify-around h-16 pb-safe">
      {menuItems.map((item) => {
        const isActive = pathname === item.href;
        return (
          <Link 
            key={item.href}
            href={item.href}
            className={cn(
              "flex flex-col items-center justify-center flex-1 h-full transition-all active:bg-elevated",
              isActive ? "bg-black text-white" : "text-text-secondary"
            )}
          >
            <item.icon className={cn("w-5 h-5", isActive ? "text-white" : "text-black")} />
            <span className="text-[7px] font-black uppercase mt-1 tracking-tighter">{item.name}</span>
          </Link>
        );
      })}
    </nav>
  );
}
