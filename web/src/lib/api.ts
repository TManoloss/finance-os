import axios from "axios";
import { getSession, signOut } from "next-auth/react";

// Normaliza a URL base para garantir que termine com /api/v1/
const getApiBaseUrl = () => {
  let url = process.env.NEXT_PUBLIC_API_URL || "";
  
  if (!url) return "";

  // Remove barra final se existir para padronizar
  if (url.endsWith('/')) {
    url = url.slice(0, -1);
  }

  // Se não termina com /api/v1, adiciona
  if (!url.endsWith('/api/v1')) {
    url = `${url}/api/v1`;
  }

  // Garante que termine com barra para chamadas relativas
  return `${url}/`;
};

const apiBaseUrl = getApiBaseUrl();

console.log("[API_DEBUG] BaseURL configurada:", apiBaseUrl);

const api = axios.create({
  baseURL: apiBaseUrl,
});

api.interceptors.request.use(async (config) => {
  if (typeof window !== "undefined") {
    const session: any = await getSession();
    if (session?.accessToken) {
      config.headers.Authorization = `Bearer ${session.accessToken}`;
    }
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    const status = error.response?.status;
    const errorMsg = error.response?.data?.error || "";

    console.log(`[API_DEBUG] Erro detectado: Status ${status} | URL: ${originalRequest.url} | Msg: ${errorMsg}`);

    // Se for 401, verifica se é erro de autenticação real ou apenas configuração faltante
    if (status === 401 && !originalRequest._retry) {
      // Se a mensagem contém "Pluggy" ou "configurações", não desloga, pois é erro de negócio
      if (errorMsg.toLowerCase().includes("pluggy") || errorMsg.toLowerCase().includes("configurações")) {
        console.warn("[API_DEBUG] Erro de configuração Pluggy detectado (401), mantendo sessão ativa.");
        return Promise.reject(error);
      }

      console.error("[API_DEBUG] Sessão expirada ou inválida (401). Realizando logout.");
      originalRequest._retry = true;
      if (typeof window !== 'undefined') {
        signOut();
      }
    }
    return Promise.reject(error);
  }
);

export { api };
export default api;
