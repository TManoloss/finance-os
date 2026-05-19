import { auth } from "@/auth"
import { NextResponse } from "next/server"

export default auth((req) => {
  const isLogged = !!req.auth
  const isAuthPage = req.nextUrl.pathname.startsWith("/login") || req.nextUrl.pathname.startsWith("/register")
  const isOnboardingPage = req.nextUrl.pathname.startsWith("/onboarding/pluggy")
  
  const pluggyId = (req.auth?.user as any)?.pluggyClientId
  const hasPluggyKeys = pluggyId && pluggyId !== "NULL" && pluggyId.trim() !== ""

  if (!isLogged && !isAuthPage) {
    // return NextResponse.redirect(new URL("/login", req.url))
  }

  if (isLogged && isAuthPage) {
    return NextResponse.redirect(new URL("/dashboard", req.url))
  }

  // Se logado mas sem chaves, redireciona para onboarding (exceto se já estiver lá ou for rota de API/Next)
  if (isLogged && !hasPluggyKeys && !isOnboardingPage && !req.nextUrl.pathname.startsWith("/api")) {
    return NextResponse.redirect(new URL("/onboarding/pluggy", req.url))
  }

  // Se já tem as chaves, não deve conseguir acessar o onboarding
  if (isLogged && hasPluggyKeys && isOnboardingPage) {
    return NextResponse.redirect(new URL("/dashboard", req.url))
  }

  return NextResponse.next()
})

export const config = {
  matcher: ["/((?!api|_next/static|_next/image|favicon.ico).*)"],
}
