import axios from "axios";

// Garante que o baseURL termine com uma barra para que chamadas relativas (sem barra inicial) funcionem corretamente
const rawBaseUrl = process.env.NEXT_PUBLIC_API_URL || "";
const apiBaseUrl = rawBaseUrl.endsWith('/') ? rawBaseUrl : `${rawBaseUrl}/`;

const apiServer = axios.create({
  baseURL: apiBaseUrl,
});

export { apiServer };
export default apiServer;
