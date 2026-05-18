"use client";

import { useState, useEffect, ReactNode } from "react";
import api from "@/lib/api";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export default function SurvivalModeWrapper({ children }: { children: ReactNode }) {
  const [level, setLevel] = useState<string>("TRANQUILO");

  useEffect(() => {
    async function fetchSurvival() {
      try {
        const resp = await api.get("/reports/survival-mode");
        setLevel(resp.data.data.level);
      } catch (error) {
        console.error("Erro ao buscar modo sobrevivência:", error);
      }
    }
    fetchSurvival();
  }, []);

  const isCritical = level === "CRÍTICO";
  const isPressure = level === "PRESSÃO";

  return (
    <div className={cn(
      "flex-1 flex flex-col min-w-0 transition-all duration-700",
      isCritical && "survival-mode-critical",
      isPressure && "survival-mode-pressure"
    )}>
      {children}
      
      <style jsx global>{`
        .survival-mode-critical .grid-blueprint > div:nth-child(2) > div {
          border-left: 4px solid #FF6B6B !important;
        }
        .survival-mode-critical [class*="bg-surface"] {
          border-left: 4px solid #FF6B6B !important;
        }
        .survival-mode-critical .low-priority {
          opacity: 0.4;
          filter: grayscale(0.5);
        }
        .survival-mode-pressure .grid-blueprint > div:nth-child(2) > div {
          border-left: 4px solid #F59E0B !important;
        }
      `}</style>
    </div>
  );
}
