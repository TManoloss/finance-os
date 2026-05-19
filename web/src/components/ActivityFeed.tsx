"use client";

import { useState, useEffect } from "react";
import api from "@/lib/api";

interface FeedEvent {
  id: string;
  type: string;
  title: string;
  description: string;
  amount: number | null;
  severity: string;
  direction?: string;
}

export default function ActivityFeed({ events: initialEvents }: { events: FeedEvent[] }) {
  const [events, setEvents] = useState<FeedEvent[]>(Array.isArray(initialEvents) ? initialEvents : []);
  const [loading, setLoading] = useState(false);

  // Sincroniza o estado interno se as props iniciais mudarem
  useEffect(() => {
    if (Array.isArray(initialEvents) && initialEvents.length > 0) {
      setEvents(initialEvents);
    }
  }, [initialEvents]);

  // Se não houver eventos iniciais, busca do servidor
  useEffect(() => {
    if (!initialEvents || initialEvents.length === 0) {
      const fetchFeed = async () => {
        setLoading(true);
        try {
          const resp = await api.get("feed?page_size=5");
          setEvents(resp.data.data || []);
        } catch (error) {
          console.error("Erro ao carregar feed:", error);
        } finally {
          setLoading(false);
        }
      };
      fetchFeed();
    }
  }, [initialEvents]);

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case "alert": return "#f44336";
      case "warning":
      case "warn": return "#ff9800";
      case "ok":
      case "info": return "#4caf50";
      default: return "#7c6fff";
    }
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <h3 style={{ fontSize: 10, fontWeight: 'bold', color: '#888', letterSpacing: '1px', marginBottom: 8 }}>INTELIGENCIA_ATIVIDADE</h3>
        {[1, 2, 3, 4].map((i) => (
          <div key={i} style={{ height: 40, background: '#1a1a1a', borderRadius: 2, animation: 'pulse 1.5s infinite' }} />
        ))}
        <style jsx>{`
          @keyframes pulse {
            0% { opacity: 0.5; }
            50% { opacity: 0.8; }
            100% { opacity: 0.5; }
          }
        `}</style>
      </div>
    );
  }

  if (!events || events.length === 0) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <h3 style={{ fontSize: 10, fontWeight: 'bold', color: '#888', letterSpacing: '1px', marginBottom: 8 }}>INTELIGENCIA_ATIVIDADE</h3>
        <div style={{ color: '#333', fontSize: 10, fontWeight: 'bold', fontStyle: 'italic', textAlign: 'center', padding: '20px 0' }}>
          AGUARDANDO_EVENTOS...
        </div>
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 0 }}>
      <h3 style={{ fontSize: 10, fontWeight: 'bold', color: '#888', letterSpacing: '1px', marginBottom: 16 }}>INTELIGENCIA_ATIVIDADE</h3>
      
      <div style={{ display: 'flex', flexDirection: 'column' }}>
        {events.map((event) => (
          <div 
            key={event.id} 
            style={{ 
              display: 'flex', 
              alignItems: 'center', 
              padding: '10px 0', 
              borderBottom: '1px solid #1e1e1e',
              gap: 12
            }}
          >
            <div style={{ 
              width: 24, 
              height: 24, 
              border: '1px solid #2a2a2a', 
              display: 'flex', 
              alignItems: 'center', 
              justifyContent: 'center',
              flexShrink: 0
            }}>
              <div style={{ 
                width: 8, 
                height: 8, 
                borderRadius: '50%', 
                background: getSeverityColor(event.severity) 
              }} />
            </div>

            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ 
                fontSize: 9, 
                color: '#ccc', 
                fontWeight: 'bold', 
                letterSpacing: '0.5px', 
                textTransform: 'uppercase',
                whiteSpace: 'nowrap',
                overflow: 'hidden',
                textOverflow: 'ellipsis'
              }}>
                {event.title}
              </div>
              <div style={{ 
                fontSize: 8, 
                color: '#444', 
                fontWeight: 'bold',
                textTransform: 'uppercase',
                whiteSpace: 'nowrap',
                overflow: 'hidden',
                textOverflow: 'ellipsis'
              }}>
                {event.description}
              </div>
            </div>

            {event.amount && (
              <div style={{ 
                fontSize: 10, 
                fontWeight: 'bold', 
                color: event.direction === 'credit' ? '#4caf50' : '#f44336',
                fontFamily: 'Courier New, monospace'
              }}>
                {new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(event.amount)}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
