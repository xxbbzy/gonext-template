/**
 * @deprecated Use `client` from `@/lib/api-client.gen` instead.
 *
 * This file is kept for backward compatibility during migration.
 * Type re-exports remain available.
 */
import type { components } from "@/types/api";

// Re-export types for backward compatibility
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

// Re-export the new client as default for gradual migration
export { client as default } from "./api-client.gen";
