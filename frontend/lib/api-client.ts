import axios, { AxiosError, InternalAxiosRequestConfig } from "axios";
import type { components } from "@/types/api";
import { useAuthStore } from "@/stores/auth";

const DEFAULT_API_PORT = "8080";

function getApiBaseURL() {
  if (process.env.NEXT_PUBLIC_API_URL) {
    return process.env.NEXT_PUBLIC_API_URL;
  }

  if (typeof window !== "undefined") {
    return `${window.location.protocol}//${window.location.hostname}:${DEFAULT_API_PORT}`;
  }

  return `http://localhost:${DEFAULT_API_PORT}`;
}

const apiClient = axios.create({
  timeout: 10000,
  headers: {
    "Content-Type": "application/json",
  },
});

// Request interceptor: attach JWT token
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    config.baseURL = config.baseURL || getApiBaseURL();
    const token = useAuthStore.getState().accessToken;
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor: handle 401 unauthorized
apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    if (error.response?.status === 401) {
      const authStore = useAuthStore.getState();
      // Try to refresh token
      if (authStore.refreshToken) {
        try {
          const res = await axios.post<ApiEnvelope<AuthResponse>>(
            `${getApiBaseURL()}/api/v1/auth/refresh`,
            {
              refresh_token: authStore.refreshToken,
            }
          );
          const { access_token, refresh_token } = res.data.data;
          if (!access_token || !refresh_token) {
            throw new Error("missing refreshed tokens");
          }
          authStore.setTokens(access_token, refresh_token);

          // Retry original request
          if (error.config) {
            error.config.headers.Authorization = `Bearer ${access_token}`;
            return apiClient(error.config);
          }
        } catch {
          authStore.logout();
          if (typeof window !== "undefined") {
            window.location.href = `/login?redirect=${encodeURIComponent(window.location.pathname)}`;
          }
        }
      } else {
        authStore.logout();
        if (typeof window !== "undefined") {
          window.location.href = `/login?redirect=${encodeURIComponent(window.location.pathname)}`;
        }
      }
    }
    return Promise.reject(error);
  }
);

export default apiClient;

export type ApiEnvelope<T = unknown> = Omit<
  components["schemas"]["Response"],
  "data"
> & {
  data: T;
};

export type ApiResponse<T = unknown> = ApiEnvelope<T>;
export type AuthResponse = components["schemas"]["AuthResponse"];
export type ItemResponse = components["schemas"]["ItemResponse"];
export type PagedItemsResponse = components["schemas"]["PagedItemsResponse"];
export type UploadResponse = components["schemas"]["UploadResponse"];
