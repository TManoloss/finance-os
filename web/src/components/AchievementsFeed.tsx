"use client";

import { Award, Calendar } from "lucide-react";
import { format } from "date-fns";
import { ptBR } from "date-fns/locale";

interface Achievement {
  id: string;
  achievement_key: string;
  name: string;
  description: string;
  icon: string;
  awarded_at: string;
}

export default function AchievementsFeed({ achievements }: { achievements: any[] }) {
  if (!achievements || achievements.length === 0) {
    return (
      <div className="p-6 border-2 border-black bg-elevated shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
        <h3 className="text-[10px] font-bold text-text-secondary uppercase tracking-widest mb-4">CONQUISTAS_BLOQUEADAS</h3>
        <p className="text-xs font-bold text-text-secondary opacity-50 uppercase tracking-tighter italic">NENHUMA_CONQUISTA_REGISTRADA_AINDA...</p>
      </div>
    );
  }

  const getAchievementDetails = (ach: any) => {
    let name = ach.name;
    let description = ach.description;
    let icon = ach.icon || "🏆";

    // Se vier do backend estruturado com context_data
    if (ach.context_data) {
      let ctx = ach.context_data;
      if (typeof ctx === 'string') {
        try {
          ctx = JSON.parse(ctx);
        } catch (e) {}
      }
      if (ctx && typeof ctx === 'object') {
        if (ctx.name) name = ctx.name;
        if (ctx.message) description = ctx.message;
        else if (ctx.context) description = ctx.context;
      }
    }

    // Se name continuar vazio, tentar usar o achievement_id de forma amigável
    const achievementKey = ach.achievement_id || ach.achievement_key || "";
    if (!name && achievementKey) {
      const mappings: Record<string, {name: string, desc: string, icon: string}> = {
        mestre_do_mercado: { name: "Mestre do Mercado", desc: "Você priorizou o preparo de refeições em casa!", icon: "🛒" },
        adeus_delivery: { name: "Adeus Delivery", desc: "Ficou vários dias consecutivos sem pedir delivery!", icon: "🥗" },
        poupador_iniciante: { name: "Poupador Iniciante", desc: "Primeiro aporte do mês em investimentos concluído!", icon: "💰" }
      };
      const match = mappings[achievementKey];
      if (match) {
        name = match.name;
        description = description || match.desc;
        icon = match.icon;
      } else {
        name = achievementKey.replace(/_/g, ' ').toUpperCase();
      }
    }

    return {
      name: name || "Nova Conquista!",
      description: description || "Parabéns por atingir mais um marco financeiro!",
      icon: icon || "🏆"
    };
  };

  return (
    <div className="p-6 border-2 border-black bg-elevated shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
      <h3 className="text-[10px] font-bold text-text-secondary uppercase tracking-widest mb-6 flex items-center gap-2">
        <Award className="w-3 h-3 text-accent-primary" /> CONQUISTAS_DESBLOQUEADAS
      </h3>
      
      <div className="flex overflow-x-auto pb-4 gap-6 scrollbar-hide">
        {achievements.map((achievement) => {
          const details = getAchievementDetails(achievement);
          return (
            <div 
              key={achievement.id || achievement.achievement_id} 
              className="flex-shrink-0 w-48 p-4 border border-black bg-background relative group hover:border-accent-primary transition-colors"
            >
              <div className="absolute top-0 right-0 p-1 opacity-10 group-hover:opacity-20 transition-opacity">
                <Award className="w-12 h-12" />
              </div>
              
              <div className="w-10 h-10 border border-black flex items-center justify-center bg-accent-primary text-white mb-4 shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]">
                <span className="text-xl">{details.icon}</span>
              </div>

              <div className="text-[10px] font-black uppercase tracking-tighter text-text-primary mb-1 truncate">
                {details.name}
              </div>
              
              <div className="text-[8px] font-bold uppercase text-text-secondary mb-3 line-clamp-2 h-6">
                {details.description}
              </div>

              <div className="flex items-center gap-1 text-[8px] font-black text-accent-primary uppercase">
                <Calendar className="w-2 h-2" />
                {achievement.awarded_at && !isNaN(new Date(achievement.awarded_at).getTime())
                  ? format(new Date(achievement.awarded_at), "dd.MM.yyyy", { locale: ptBR })
                  : "--"}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
