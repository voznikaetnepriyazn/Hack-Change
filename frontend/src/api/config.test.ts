import { describe, expect, it } from "vitest";
import { API_URL } from "./config";

describe("API_URL", () => {
  it("points at the backend port without an /api prefix", () => {
    expect(API_URL).toBe("http://localhost:8080");
  });
});
