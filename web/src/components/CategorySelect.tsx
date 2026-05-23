"use client";

import React, { useState, useRef, useEffect } from "react";
import { ChevronDown, Check, Loader2 } from "lucide-react";
import { useRouter } from "next/navigation";
import api from "@/lib/api";

interface Category {
  id: string;
  name: string;
  color?: string;
  icon?: string;
}

interface CategorySelectProps {
  transactionId: string;
  currentCategoryId?: string;
  currentCategoryName?: string;
  categories: Category[];
}

export default function CategorySelect({
  transactionId,
  currentCategoryId,
  currentCategoryName,
  categories,
}: CategorySelectProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [selectedId, setSelectedId] = useState(currentCategoryId || "");
  const [selectedName, setSelectedName] = useState(currentCategoryName || "SEM_CATEGORIA");
  const [loading, setLoading] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const router = useRouter();

  useEffect(() => {
    // Sync state if server data changes
    setSelectedId(currentCategoryId || "");
    setSelectedName(currentCategoryName || "SEM_CATEGORIA");
  }, [currentCategoryId, currentCategoryName]);

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleSelect = async (categoryId: string, categoryName: string) => {
    if (categoryId === selectedId) {
      setIsOpen(false);
      return;
    }

    setLoading(true);
    setIsOpen(false);

    try {
      await api.patch(`transactions/${transactionId}/category`, {
        category_id: categoryId,
      });

      setSelectedId(categoryId);
      setSelectedName(categoryName);
      
      // router.refresh() silently triggers a refetch of Server Component data
      // This will update the main state, lists, and recalculate reports!
      router.refresh();
    } catch (error) {
      console.error("[CATEGORY_SELECT_ERROR] Falha ao atualizar categoria:", error);
      alert("Erro ao atualizar categoria. Tente novamente.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="relative inline-block text-left" ref={dropdownRef}>
      {loading ? (
        <span className="inline-flex items-center gap-1.5 px-2 py-1 bg-black text-white text-[10px] font-black uppercase tracking-tighter border-2 border-black select-none">
          <Loader2 className="w-3 h-3 animate-spin text-accent-primary" />
          ATUALIZANDO...
        </span>
      ) : (
        <button
          onClick={() => setIsOpen(!isOpen)}
          className="inline-flex items-center gap-1.5 px-2 py-1 bg-black text-white text-[10px] font-black uppercase tracking-tighter border-2 border-black hover:bg-elevated hover:text-text-primary transition-all cursor-pointer shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] active:shadow-none active:translate-x-[1px] active:translate-y-[1px]"
        >
          {selectedName}
          <ChevronDown className="w-3 h-3 text-accent-primary" />
        </button>
      )}

      {isOpen && (
        <div className="absolute left-0 mt-1.5 w-56 bg-background border-2 border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] z-50 max-h-60 overflow-y-auto divide-y divide-black/10">
          <div className="py-1">
            {categories.map((cat) => {
              const isSelected = cat.id === selectedId;
              return (
                <button
                  key={cat.id}
                  onClick={() => handleSelect(cat.id, cat.name)}
                  className={`w-full flex items-center justify-between px-3 py-2 text-[10px] font-black uppercase tracking-wider text-left transition-all ${
                    isSelected
                      ? "bg-accent-primary text-white"
                      : "text-text-primary hover:bg-elevated hover:text-accent-primary"
                  }`}
                >
                  <span>{cat.name}</span>
                  {isSelected && <Check className="w-3 h-3 text-white" />}
                </button>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}
