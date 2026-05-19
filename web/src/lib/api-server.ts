import axios from "axios";

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

const apiServer = axios.create({
  baseURL: apiBaseUrl,
});

export { apiServer };
export default apiServer;
