import NextAuth from "next-auth";
import CredentialsProvider from "next-auth/providers/credentials";
import axios from "axios";

export const { handlers, auth, signIn, signOut } = NextAuth({
  providers: [
    CredentialsProvider({
      name: "Credentials",
      credentials: {
        email: { label: "Email", type: "email" },
        password: { label: "Password", type: "password" },
      },
      async authorize(credentials) {
        const loginUrl = process.env.NEXT_PUBLIC_API_URL ? `${process.env.NEXT_PUBLIC_API_URL}/api/v1/auth/login` : null;
        
        if (!loginUrl) {
           console.error("[Auth DEBUG] API URL não configurada!");
           return null;
        }

        try {
          const resp = await axios.post(loginUrl, {
            email: credentials?.email,
            password: credentials?.password,
          });

          console.log("[Auth DEBUG] Status API:", resp.status);
          
          if (resp.data && resp.data.success) {
            const userData = resp.data.data;
            return {
              id: userData.user?.id || 'unknown',
              name: userData.user?.name || 'User',
              email: userData.user?.email || credentials?.email,
              pluggyClientId: userData.user?.pluggy_client_id,
              accessToken: userData.access_token,
              refreshToken: userData.refresh_token,
            };
          }
          return null;
        } catch (error: any) {
          console.error("[Auth DEBUG] Erro na requisição:", error.message);
          return null;
        }
      },
    }),
  ],
  callbacks: {
    async jwt({ token, user, trigger, session }) {
      if (user) {
        token.accessToken = (user as any).accessToken;
        token.refreshToken = (user as any).refreshToken;
        token.pluggyClientId = (user as any).pluggyClientId;
        token.id = user.id;
      }
      if (trigger === "update" && session?.pluggyClientId) {
        token.pluggyClientId = session.pluggyClientId;
      }
      return token;
    },
    async session({ session, token }) {
      (session as any).accessToken = token.accessToken;
      (session as any).refreshToken = token.refreshToken;
      (session as any).user.id = token.id;
      (session as any).user.pluggyClientId = token.pluggyClientId;
      return session;
    },
  },
  pages: {
    signIn: "/login",
  },
});
