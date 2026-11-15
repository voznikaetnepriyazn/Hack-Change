import { describe, expect, it } from "vitest";
import { API_URL } from "./config";

describe("API_URL", () => {
  it("falls back to /api", () => {
    expect(API_URL).toBe("/api");
  });
});
