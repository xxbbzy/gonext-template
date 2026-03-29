import createClient from "openapi-fetch";

import { useAuthStore } from "@/stores/auth";
import type { components, paths } from "@/types/api";

const DEFAULT_API_PORT = "8080";

export type ApiPaths = paths;
export type ApiSchemas = components["schemas"];
export type ApiErrorResponse = components["schemas"]["ErrorResponse"];
export type AuthResponse = components["schemas"]["AuthResponse"];
export type UserResponse = components["schemas"]["UserResponse"];
export type CreateItemRequest = components["schemas"]["CreateItemRequest"];
export type UpdateItemRequest = components["schemas"]["UpdateItemRequest"];
export type ItemResponse = components["schemas"]["ItemResponse"];
export type PagedItemsResponse = components["schemas"]["PagedItemsResponse"];
export type UploadResponse = components["schemas"]["UploadResponse"];
export type ListItemsQuery = NonNullable<
  paths["/api/v1/items"]["get"]["parameters"]["query"]
>;

type SuccessEnvelope<T> = {
  code: 0;
  data: T;
  message: "success";
};

type NullableSuccessEnvelope = {
  code: 0;
  data: unknown;
  message: "success";
};

function getApiBaseURL() {
  if (process.env.NEXT_PUBLIC_API_URL) {
    return process.env.NEXT_PUBLIC_API_URL;
  }

  if (typeof window !== "undefined") {
    return `${window.location.protocol}//${window.location.hostname}:${DEFAULT_API_PORT}`;
  }

  return `http://localhost:${DEFAULT_API_PORT}`;
}

function toErrorMessage(error: unknown, fallback: string) {
  if (
    error &&
    typeof error === "object" &&
    "message" in error &&
    typeof error.message === "string" &&
    error.message.trim() !== ""
  ) {
    return error.message;
  }

  return fallback;
}

function createRedirectPath() {
  if (typeof window === "undefined") {
    return "/login";
  }

  const currentPath = `${window.location.pathname}${window.location.search}`;
  return `/login?redirect=${encodeURIComponent(currentPath)}`;
}

const baseUrl = getApiBaseURL();

export const client = createClient<paths>({ baseUrl });

export class ApiClientError extends Error {
  readonly status: number;
  readonly code: number | null;
  readonly payload: ApiErrorResponse | null;

  constructor(
    message: string,
    status: number,
    code: number | null,
    payload: ApiErrorResponse | null = null
  ) {
    super(message);
    this.name = "ApiClientError";
    this.status = status;
    this.code = code;
    this.payload = payload;
  }
}

export function getApiErrorMessage(error: unknown, fallback: string) {
  if (error instanceof ApiClientError && error.message.trim() !== "") {
    return error.message;
  }

  if (error instanceof Error && error.message.trim() !== "") {
    return error.message;
  }

  return toErrorMessage(error, fallback);
}

function toApiClientError(
  error: ApiErrorResponse | undefined,
  response: Response,
  fallback: string
) {
  return new ApiClientError(
    toErrorMessage(error, fallback),
    response.status,
    error?.code ?? null,
    error ?? null
  );
}

function unwrapData<T>(
  result:
    | {
        data: SuccessEnvelope<T>;
        error?: never;
        response: Response;
      }
    | {
        data?: never;
        error: ApiErrorResponse;
        response: Response;
      },
  fallback: string
) {
  if ("error" in result && result.error) {
    throw toApiClientError(result.error, result.response, fallback);
  }

  return result.data.data;
}

function assertSuccess(
  result:
    | {
        data: NullableSuccessEnvelope;
        error?: never;
        response: Response;
      }
    | {
        data?: never;
        error: ApiErrorResponse;
        response: Response;
      },
  fallback: string
) {
  if ("error" in result && result.error) {
    throw toApiClientError(result.error, result.response, fallback);
  }
}

const bodyCache = new WeakMap<Request, Blob>();

let refreshPromise: Promise<string> | null = null;

async function refreshAccessToken() {
  const store = useAuthStore.getState();
  if (!store.refreshToken) {
    throw new ApiClientError("refresh failed", 401, 401);
  }

  const res = await fetch(`${baseUrl}/api/v1/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: store.refreshToken }),
  });

  const payload = (await res.json().catch(() => null)) as
    | components["schemas"]["AuthSuccessResponse"]
    | ApiErrorResponse
    | null;

  if (!res.ok) {
    throw new ApiClientError(
      toErrorMessage(payload, "refresh failed"),
      res.status,
      "code" in (payload ?? {}) && typeof payload?.code === "number"
        ? payload.code
        : null,
      payload && "message" in payload ? (payload as ApiErrorResponse) : null
    );
  }

  if (!payload || !("data" in payload)) {
    throw new ApiClientError("refresh failed", res.status, null);
  }

  const authPayload = payload as components["schemas"]["AuthSuccessResponse"];
  store.setTokens(
    authPayload.data.access_token,
    authPayload.data.refresh_token
  );
  return authPayload.data.access_token;
}

client.use({
  async onRequest({ request }) {
    const token = useAuthStore.getState().accessToken;
    if (token) {
      request.headers.set("Authorization", `Bearer ${token}`);
    }

    if (request.method !== "GET" && request.method !== "HEAD") {
      const cloned = request.clone();
      bodyCache.set(request, await cloned.blob());
    }
  },

  async onResponse({ request, response }) {
    if (response.status !== 401 || request.url.includes("/auth/refresh")) {
      return response;
    }

    if (!refreshPromise) {
      refreshPromise = refreshAccessToken().finally(() => {
        refreshPromise = null;
      });
    }

    let newToken: string;
    try {
      newToken = await refreshPromise;
    } catch {
      useAuthStore.getState().logout();
      if (typeof window !== "undefined") {
        window.location.href = createRedirectPath();
      }
      return response;
    }

    const cachedBody = bodyCache.get(request) ?? null;
    const headers = new Headers(request.headers);
    headers.set("Authorization", `Bearer ${newToken}`);

    return fetch(
      new Request(request, {
        headers,
        body: cachedBody,
      })
    );
  },
});

export async function registerUser(
  body: components["schemas"]["RegisterRequest"]
) {
  return unwrapData(
    await client.POST("/api/v1/auth/register", { body }),
    "Registration failed"
  );
}

export async function loginUser(body: components["schemas"]["LoginRequest"]) {
  return unwrapData(
    await client.POST("/api/v1/auth/login", { body }),
    "Login failed"
  );
}

export async function getProfile() {
  return unwrapData(
    await client.GET("/api/v1/auth/profile"),
    "Profile load failed"
  );
}

export async function listItems(query: Partial<ListItemsQuery> = {}) {
  return unwrapData(
    await client.GET("/api/v1/items", {
      params: { query },
    }),
    "Failed to load items"
  );
}

export async function createItem(body: CreateItemRequest) {
  return unwrapData(
    await client.POST("/api/v1/items", { body }),
    "Failed to create item"
  );
}

export async function updateItem(id: number, body: UpdateItemRequest) {
  return unwrapData(
    await client.PUT("/api/v1/items/{id}", {
      params: { path: { id } },
      body,
    }),
    "Failed to update item"
  );
}

export async function deleteItem(id: number) {
  assertSuccess(
    await client.DELETE("/api/v1/items/{id}", {
      params: { path: { id } },
    }),
    "Failed to delete item"
  );
}

export async function uploadFile(file: File) {
  const formData = new FormData();
  formData.append("file", file);

  return unwrapData(
    await client.POST("/api/v1/upload", {
      body: { file: file.name },
      bodySerializer: () => formData,
    }),
    "Failed to upload file"
  );
}
