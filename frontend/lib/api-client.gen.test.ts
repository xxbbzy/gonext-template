import { describe, expect, it } from "vitest";

import {
  ApiClientError,
  getApiErrorMessage,
  getApiErrorMetadata,
  type ApiErrorResponse,
} from "@/lib/api-client.gen";

describe("api-client error helpers", () => {
  it("extracts unified metadata from ApiClientError", () => {
    const payload: ApiErrorResponse = {
      code: 429,
      message: "too many requests",
      request_id: "req-test-123",
      details: { retry_after_seconds: 60 },
    };
    const error = new ApiClientError(
      payload.message,
      429,
      payload.code,
      payload
    );

    expect(getApiErrorMessage(error, "fallback")).toBe("too many requests");
    expect(getApiErrorMetadata(error)).toEqual({
      status: 429,
      code: 429,
      requestId: "req-test-123",
      details: { retry_after_seconds: 60 },
    });
  });

  it("keeps message fallback compatibility for unknown errors", () => {
    expect(getApiErrorMessage(null, "fallback message")).toBe(
      "fallback message"
    );
    expect(getApiErrorMetadata(new Error("plain error"))).toBeNull();
  });
});
