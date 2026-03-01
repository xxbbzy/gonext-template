import axios, { AxiosError, InternalAxiosRequestConfig } from "axios";
import { useAuthStore } from "@/stores/auth";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

const apiClient = axios.create({
    baseURL: API_BASE_URL,
    timeout: 10000,
    headers: {
        "Content-Type": "application/json",
    },
});

// Request interceptor: attach JWT token
apiClient.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
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
                    const res = await axios.post(`${API_BASE_URL}/api/v1/auth/refresh`, {
                        refresh_token: authStore.refreshToken,
                    });
                    const { access_token, refresh_token } = res.data.data;
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

// API response type
export interface ApiResponse<T = unknown> {
    code: number;
    data: T;
    message: string;
}

export interface PagedData<T> {
    items: T[];
    total: number;
    page: number;
    page_size: number;
    total_pages: number;
}
