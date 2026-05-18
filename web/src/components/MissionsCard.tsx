"use client";

import { Target, CheckCircle2, Circle } from "lucide-react";

interface Mission {
  id: string;
  mission_key: string;
  name: string;
  description: string;
  status: string;
  progress: number;
  target_value: number;
  reward_points: number;
}

export default function MissionsCard({ missions }: { missions: Mission[] }) {
  if (!missions || missions.length === 0) {
    return (
      <div className="p-6 border-2 border-black bg-elevated shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
        <h3 className="text-[10px] font-bold text-text-secondary uppercase tracking-widest mb-4">MISSOES_ATIVAS</h3>
        <p className="text-xs font-bold text-text-secondary opacity-50 uppercase tracking-tighter italic">AGUARDANDO_NOVAS_DIRETRIZES...</p>
      </div>
    );
  }

  return (
    <div className="p-6 border-2 border-black bg-elevated shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
      <h3 className="text-[10px] font-bold text-text-secondary uppercase tracking-widest mb-6 flex items-center gap-2">
        <Target className="w-3 h-3 text-accent-primary" /> MISSOES_OPERACIONAIS_ATIVAS
      </h3>
      
      <div className="space-y-6">
        {missions.map((mission) => {
          const progressPercent = Math.min(100, Math.round((mission.progress / mission.target_value) * 100));
          const isCompleted = mission.status === 'completed';

          return (
            <div key={mission.id} className="space-y-2 group">
              <div className="flex justify-between items-start gap-4">
                <div className="flex gap-3">
                  {isCompleted ? (
                    <CheckCircle2 className="w-4 h-4 text-green-500 mt-1 flex-shrink-0" />
                  ) : (
                    <Circle className="w-4 h-4 text-accent-primary mt-1 flex-shrink-0 animate-pulse" />
                  )}
                  <div>
                    <div className="text-[10px] font-black uppercase tracking-tighter text-text-primary">
                      {mission.name}
                    </div>
                    <div className="text-[8px] font-bold uppercase text-text-secondary leading-tight">
                      {mission.description}
                    </div>
                  </div>
                </div>
                <div className="text-[10px] font-black font-mono text-accent-primary whitespace-nowrap">
                   +{mission.reward_points} PTS
                </div>
              </div>

              <div className="space-y-1">
                <div className="flex justify-between text-[8px] font-black uppercase tracking-widest text-text-secondary">
                  <span>PROGRESSO</span>
                  <span>{progressPercent}%</span>
                </div>
                <div className="h-2 border border-black bg-background p-[1px] relative overflow-hidden">
                  <div 
                    className={`h-full transition-all duration-1000 ease-out ${isCompleted ? 'bg-green-500' : 'bg-accent-primary'}`}
                    style={{ width: `${progressPercent}%` }}
                  >
                    <div className="absolute inset-0 bg-white/20 animate-[shimmer_2s_infinite]"></div>
                  </div>
                </div>
              </div>
            </div>
          );
        })}
      </div>
      
      <style jsx>{`
        @keyframes shimmer {
          0% { transform: translateX(-100%); }
          100% { transform: translateX(100%); }
        }
      `}</style>
    </div>
  );
}
