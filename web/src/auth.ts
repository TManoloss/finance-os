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
        try {
          const loginUrl = `${process.env.NEXT_PUBLIC_API_URL}/api/v1/auth/login`;
          
          const resp = await axios.post(loginUrl, {
            email: credentials?.email,
            password: credentials?.password,
          });

          console.log("[Auth DEBUG] Resposta completa da API:", JSON.stringify(resp.data, null, 2));

          if (resp.data.success) {
            return {
              id: resp.data.data.user.id,
              name: resp.data.data.user.name,
              email: resp.data.data.user.email,
              pluggyClientId: resp.data.data.user.pluggy_client_id,
              accessToken: resp.data.data.access_token,
              refreshToken: resp.data.data.refresh_token,
            };
          }
          return null;
        } catch (error: any) {
          console.error("[Auth DEBUG] Erro completo:", error);
          if (error.response) {
            console.error("[Auth DEBUG] Erro resposta:", JSON.stringify(error.response.data, null, 2));
          }
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
