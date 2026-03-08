/**
 * Type-safe API client generated from OpenAPI spec.
 *
 * Uses openapi-fetch for end-to-end type safety.
 * Middleware handles auth token injection and 401 refresh with body-safe retry.
 */
import createClient from "openapi-fetch";
import type { paths } from "@/types/api";
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

const baseUrl = getApiBaseURL();

// --- Main client ---
export const client = createClient<paths>({ baseUrl });

// --- Body cache for safe 401 retries (POST/PUT body is consumed once) ---
const bodyCache = new WeakMap<Request, Blob>();

// --- Single-flight refresh ---
let refreshPromise: Promise<string> | null = null;

async function refreshAccessToken(): Promise<string> {
  const store = useAuthStore.getState();
  // Use raw fetch — NOT the openapi-fetch client (prevents recursion)
  const res = await fetch(`${baseUrl}/api/v1/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: store.refreshToken }),
  });
  if (!res.ok) {
    throw new Error("refresh failed");
  }
  const json = await res.json();
  const data = json.data;
  store.setTokens(data.access_token, data.refresh_token);
  return data.access_token as string;
}

// --- Middleware registration ---
client.use({
  async onRequest({ request }) {
    // 1. Inject auth token
    const token = useAuthStore.getState().accessToken;
    if (token) {
      request.headers.set("Authorization", `Bearer ${token}`);
    }

    // 2. Clone body for potential 401 retry (before it's consumed)
    if (request.method !== "GET" && request.method !== "HEAD") {
      const cloned = request.clone();
      bodyCache.set(request, await cloned.blob());
    }
  },

  async onResponse({ request, response }) {
    if (response.status !== 401) return response;

    // Don't retry refresh endpoint itself
    if (request.url.includes("/auth/refresh")) return response;

    // Single-flight: all concurrent 401s await the same promise
    if (!refreshPromise) {
      refreshPromise = refreshAccessToken().finally(() => {
        refreshPromise = null;
      });
    }

    let newToken: string;
    try {
      newToken = await refreshPromise;
    } catch {
      // Refresh failed — logout and redirect
      useAuthStore.getState().logout();
      if (typeof window !== "undefined") {
        window.location.href = `/login?redirect=${encodeURIComponent(window.location.pathname)}`;
      }
      return response;
    }

    // Rebuild request with cached body + new token
    const cachedBody = bodyCache.get(request) ?? null;
    const headers = new Headers(request.headers);
    headers.set("Authorization", `Bearer ${newToken}`);

    return fetch(request.url, {
      method: request.method,
      headers,
      body: cachedBody,
    });
  },
});

// --- Upload wrapper (typed, no double-cast in pages) ---
export async function uploadFile(file: File) {
  const formData = new FormData();
  formData.append("file", file);

  return client.POST("/api/v1/upload", {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    body: {} as any,
    bodySerializer: () => formData,
  });
}

// Re-export types for convenience
export type { paths };
