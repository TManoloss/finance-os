import axios from "axios";
import { getSession, signOut } from "next-auth/react";

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL,
});

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    // Se for 401, verifica se é erro de autenticação real ou apenas configuração faltante
    if (error.response?.status === 401 && !originalRequest._retry) {
      const errorMsg = error.response.data?.error || "";
      
      // Se a mensagem contém "Pluggy" ou "configurações", não desloga, pois é erro de negócio
      if (errorMsg.toLowerCase().includes("pluggy") || errorMsg.toLowerCase().includes("configurações")) {
        return Promise.reject(error);
      }

      originalRequest._retry = true;
      if (typeof window !== 'undefined') {
        signOut();
      }
    }
    return Promise.reject(error);
  }
);

export default api;
