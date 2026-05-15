"use client";

import { Terminal } from "lucide-react";

export default function Loading() {
  return (
    <div className="grid-blueprint grid-cols-1 min-h-screen bg-background animate-pulse">
      <header className="p-8 bg-elevated border-b-2 border-black">
        <div className="flex items-center gap-2 mb-1">
          <Terminal className="w-4 h-4 text-text-secondary" />
          <span className="text-[10px] font-bold tracking-widest uppercase text-text-secondary">SYSTEM_INITIALIZING...</span>
        </div>
        <div className="h-10 w-64 bg-black/20 mb-2"></div>
        <div className="h-4 w-40 bg-black/10"></div>
      </header>

      <div className="p-8">
        <div className="bg-elevated border-2 border-black">
          <div className="p-6 border-b-2 border-black bg-elevated/50 flex flex-col md:flex-row md:items-center justify-between gap-4 h-24">
            <div className="h-6 w-48 bg-black/10"></div>
            <div className="h-10 w-32 bg-black/20 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]"></div>
          </div>
          <div className="p-8">
            <div className="space-y-6">
              {[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
                <div key={i} className="h-16 w-full bg-black/5 border-2 border-black/10"></div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
