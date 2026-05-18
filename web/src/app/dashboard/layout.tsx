import Sidebar from "@/components/Sidebar";
import BottomNav from "@/components/BottomNav";
import MobileHeader from "@/components/MobileHeader";
import StressScoreBadge from "@/components/StressScoreBadge";
import SurvivalModeOverlay from "@/components/SurvivalModeOverlay";
import SurvivalModeWrapper from "@/components/SurvivalModeWrapper";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex h-screen bg-background overflow-hidden selection:bg-text-primary selection:text-background">
      <Sidebar />
      <SurvivalModeWrapper>
        <SurvivalModeOverlay />
        <MobileHeader />
        
        {/* Desktop Header */}
        <header className="hidden md:flex items-center justify-end p-4 bg-background border-b-2 border-border-subtle gap-4">
          <StressScoreBadge />
        </header>

        <main className="flex-1 overflow-y-auto pb-16 md:pb-0">
          <div className="min-h-full">
            {children}
          </div>
        </main>
      </SurvivalModeWrapper>
      <BottomNav />
    </div>
  );
}
