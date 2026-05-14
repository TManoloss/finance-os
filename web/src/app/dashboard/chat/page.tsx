"use client";

import { useState, useEffect, useRef } from "react";
import { useSession } from "next-auth/react";
import api from "@/lib/api";
import { Send, Bot, User, Terminal } from "lucide-react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export default function ChatPage() {
  const { data: session } = useSession() as any;
  const token = session?.accessToken;
  const [messages, setMessages] = useState([
    { role: "assistant", content: "SISTEMA_INICIALIZADO: Olá! Eu sou o Pierre. Como posso auxiliar na sua análise financeira hoje?" }
  ]);
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages]);

  const handleSend = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim() || loading) return;

    const userMessage = input.trim();
    setInput("");
    setMessages(prev => [...prev, { role: "user", content: userMessage }]);
    setLoading(true);

    try {
      const resp = await api.post("/chat", {
        message: userMessage,
        history: messages.slice(-5).map(m => ({ role: m.role, content: m.content }))
      }, {
        headers: { Authorization: `Bearer ${token}` }
      });

      const responseContent = resp.data?.data?.response || "SYSTEM_ERROR: O núcleo de IA não retornou uma resposta válida.";
      setMessages(prev => [...prev, { role: "assistant", content: responseContent }]);
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || "ERROR_REF_0xChat: Falha na comunicação com o núcleo de IA.";
      setMessages(prev => [...prev, { role: "assistant", content: errorMessage }]);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="grid-blueprint grid-cols-1 h-[calc(100vh-2px)] overflow-hidden">
      <header className="p-8 bg-elevated border-b-2 border-black flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <Terminal className="w-4 h-4" />
            <span className="text-[10px] font-bold tracking-widest uppercase">AI_COMMUNICATION_INTERFACE</span>
          </div>
          <h1 className="text-4xl font-black text-text-primary uppercase tracking-tighter">TERMINAL_PIERRE</h1>
          <p className="text-sm font-bold text-text-secondary uppercase">STATUS: ENCRYPTED_STREAM_ACTIVE</p>
        </div>
      </header>

      {/* Chat Area */}
      <div className="flex-1 overflow-hidden flex flex-col bg-background">
        <div 
          ref={scrollRef}
          className="flex-1 overflow-y-auto p-8 space-y-8 scroll-smooth"
        >
          {messages.map((msg, idx) => (
            <div 
              key={idx}
              className={cn(
                "flex flex-col gap-2 max-w-3xl",
                msg.role === "user" ? "ml-auto items-end" : "mr-auto items-start"
              )}
            >
              <div className="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-text-secondary">
                {msg.role === "user" ? (
                  <>USUARIO <User className="w-3 h-3" /></>
                ) : (
                  <><Bot className="w-3 h-3" /> PIERRE_AI</>
                )}
              </div>
              
              <div className={cn(
                "p-6 border-2 border-black text-sm leading-relaxed shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]",
                msg.role === "user" 
                  ? "bg-accent-primary text-white" 
                  : "bg-elevated text-text-primary"
              )}>
                {msg.role === "assistant" ? (
                  <div className="markdown-content">
                    <ReactMarkdown remarkPlugins={[remarkGfm]}>
                      {msg.content}
                    </ReactMarkdown>
                  </div>
                ) : (
                  msg.content
                )}
              </div>
            </div>
          ))}
          {loading && (
            <div className="flex flex-col gap-2 mr-auto items-start">
              <div className="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-text-secondary">
                <Bot className="w-3 h-3" /> PIERRE_AI
              </div>
              <div className="p-6 border-2 border-black bg-elevated text-text-secondary text-xs font-bold animate-pulse">
                &gt; PROCESSANDO_REQUISICAO...
              </div>
            </div>
          )}
        </div>

        {/* Input Area */}
        <form onSubmit={handleSend} className="p-8 border-t-2 border-black bg-elevated">
          <div className="relative">
            <div className="absolute left-6 top-1/2 -translate-y-1/2 text-text-primary font-black">&gt;</div>
            <input 
              type="text"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              disabled={loading}
              placeholder="DIGITE_SUA_COMANDOS_AQUI..."
              className="w-full bg-background border-2 border-black py-4 pl-12 pr-16 text-sm font-bold uppercase focus:outline-none focus:bg-background transition-all shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] focus:shadow-none focus:translate-x-[2px] focus:translate-y-[2px]"
            />
            <button 
              type="submit"
              disabled={loading || !input.trim()}
              className="absolute right-4 top-1/2 -translate-y-1/2 p-2 bg-black text-white hover:bg-text-secondary transition-all disabled:opacity-50"
            >
              <Send className="w-5 h-5" />
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
