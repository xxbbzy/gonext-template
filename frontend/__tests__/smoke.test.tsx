import { describe, it, expect } from "vitest";

describe("Frontend smoke test", () => {
  it("basic assertion works", () => {
    expect(1 + 1).toBe(2);
  });

  it("can import React", async () => {
    const React = await import("react");
    expect(React.createElement).toBeDefined();
  });
});
