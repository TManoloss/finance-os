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

      <div className="p-8 space-y-8">
        {/* Metric Cards Skeleton */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="p-6 bg-elevated border-2 border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
              <div className="h-3 w-20 bg-black/10 mb-4"></div>
              <div className="h-8 w-32 bg-black/20"></div>
            </div>
          ))}
        </div>

        {/* Charts Skeleton */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          <div className="lg:col-span-2 p-8 bg-elevated border-2 border-black h-80">
            <div className="h-6 w-48 bg-black/10 mb-8"></div>
            <div className="w-full h-48 bg-black/5"></div>
          </div>
          <div className="p-8 bg-elevated border-2 border-black h-80">
            <div className="h-6 w-32 bg-black/10 mb-8"></div>
            <div className="flex justify-center">
              <div className="w-40 h-40 rounded-full border-8 border-black/5"></div>
            </div>
          </div>
        </div>

        {/* Transactions Skeleton */}
        <div className="bg-elevated border-2 border-black">
          <div className="p-6 border-b-2 border-black bg-elevated/50 h-16"></div>
          <div className="p-0">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="p-4 border-b-2 border-black flex items-center justify-between h-20">
                <div className="flex items-center gap-4">
                  <div className="w-10 h-10 bg-black/10 rounded-none"></div>
                  <div className="space-y-2">
                    <div className="h-4 w-32 bg-black/10"></div>
                    <div className="h-3 w-20 bg-black/5"></div>
                  </div>
                </div>
                <div className="h-6 w-24 bg-black/10"></div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
