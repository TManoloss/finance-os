"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { Filter } from "lucide-react";

interface Category {
  id: string;
  name: string;
}

export default function CategoryFilter({ categories }: { categories: Category[] }) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const currentCategoryId = searchParams.get("category_id") || "all";

  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const value = e.target.value;
    const params = new URLSearchParams(searchParams.toString());
    
    if (value === "all") {
      params.delete("category_id");
    } else {
      params.set("category_id", value);
    }
    // Ao mudar o filtro, volta para a página 1
    params.delete("page");
    
    router.push(`/dashboard/transactions?${params.toString()}`);
  };

  return (
    <div className="flex-1 sm:flex-none flex items-center gap-2 bg-background border-2 border-black px-3 py-2 text-[10px] md:text-xs font-black uppercase shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]">
      <Filter className="w-3 h-3 md:w-4 h-4 shrink-0" />
      <select 
        value={currentCategoryId}
        onChange={handleChange}
        className="bg-transparent border-none focus:outline-none w-full cursor-pointer appearance-none pr-4"
      >
        <option value="all">TODAS_CATEGORIAS</option>
        {categories.map((cat) => (
          <option key={cat.id} value={cat.id}>
            {cat.name.toUpperCase()}
          </option>
        ))}
      </select>
    </div>
  );
}
