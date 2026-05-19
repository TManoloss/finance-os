"use client";

import { useEffect, useState, use } from 'react';
import { useRouter } from 'next/navigation';
import { api } from '@/lib/api';
import { ChevronRight, ArrowLeft } from 'lucide-react';
import { Button } from '@/components/ui/button';

export default function ReplayPage({ params }: { params: Promise<{ month: string }> }) {
  const router = useRouter();
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [currentSlide, setCurrentSlide] = useState(0);
  
  // Unwrap the params Promise
  const resolvedParams = use(params);
  const month = resolvedParams.month;

  useEffect(() => {
    const fetchReplay = async () => {
      try {
        const res = await api.get(`/reports/monthly-replay?month=${month}`);
        setData(res.data.data);
      } catch (err) {
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    fetchReplay();
  }, [month]);

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center bg-[#0A0A0F]">
        <div className="text-[#7C6FFF] animate-pulse">Preparando seu Replay...</div>
      </div>
    );
  }

  if (!data || !data.narrative || !data.narrative.slides) {
    return (
      <div className="flex h-screen flex-col items-center justify-center bg-[#0A0A0F] text-white">
        <p>Replay não encontrado para este mês.</p>
        <Button variant="ghost" onClick={() => router.back()} className="mt-4">
          Voltar
        </Button>
      </div>
    );
  }

  const slides = data.narrative.slides;

  const nextSlide = () => {
    if (currentSlide < slides.length - 1) {
      setCurrentSlide(s => s + 1);
    } else {
      router.back();
    }
  };

  const prevSlide = () => {
    if (currentSlide > 0) {
      setCurrentSlide(s => s - 1);
    }
  };

  const slide = slides[currentSlide];

  return (
    <div className="flex h-screen w-full flex-col bg-gradient-to-br from-[#1A1A24] to-[#0A0A0F] text-[#F0F0F5] relative overflow-hidden">
      {/* Progress Bars */}
      <div className="absolute top-4 left-4 right-4 flex gap-2 z-10">
        {slides.map((_: any, idx: number) => (
          <div key={idx} className="h-1 flex-1 bg-[#2A2A3A] rounded-full overflow-hidden">
            <div 
              className="h-full bg-[#7C6FFF] transition-all duration-300"
              style={{ width: idx < currentSlide ? '100%' : idx === currentSlide ? '100%' : '0%' }}
            />
          </div>
        ))}
      </div>

      {/* Navigation areas (invisible) */}
      <div className="absolute inset-y-0 left-0 w-1/3 z-10" onClick={prevSlide} />
      <div className="absolute inset-y-0 right-0 w-2/3 z-10" onClick={nextSlide} />

      <div className="flex flex-1 flex-col items-center justify-center p-8 text-center z-0 animate-in fade-in zoom-in duration-500 key={currentSlide}">
        <h2 className="text-3xl font-bold mb-8 text-[#4ECDC4]">{slide.title}</h2>
        <p className="text-xl md:text-2xl text-[#F0F0F5] max-w-2xl leading-relaxed mb-12">
          {slide.text}
        </p>
        {slide.highlight_value && (
          <div className="text-5xl md:text-7xl font-extrabold text-[#7C6FFF] drop-shadow-lg">
            {slide.highlight_value}
          </div>
        )}
      </div>

      <div className="absolute bottom-8 w-full flex justify-center z-20">
        <Button variant="ghost" size="icon" onClick={() => router.back()} className="text-[#8888A0] hover:text-white">
          <ArrowLeft className="w-6 h-6" />
        </Button>
      </div>
    </div>
  );
}
