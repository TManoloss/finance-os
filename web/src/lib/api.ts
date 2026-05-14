import axios from "axios";
import { getSession, signOut } from "next-auth/react";

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL,
});

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      // No servidor não podemos dar signOut direto assim
      if (typeof window !== 'undefined') {
        signOut();
      }
    }
    return Promise.reject(error);
  }
);

export default api;
