import { describe, expect, it } from "vitest";
import { isCurrentUserResponse } from "./account.validation";

describe("isCurrentUserResponse", () => {
  it("accepts the authenticated user identity", () => {
    expect(
      isCurrentUserResponse({
        user: { id: "1", full_name: "Ева Смирнова", login: "e.smirnova", role: "intern" },
      }),
    ).toBe(true);
  });

  it("rejects malformed payloads", () => {
    expect(isCurrentUserResponse(null)).toBe(false);
    expect(isCurrentUserResponse(1)).toBe(false);
    expect(isCurrentUserResponse({})).toBe(false);
    expect(isCurrentUserResponse({ user: null })).toBe(false);
    expect(
      isCurrentUserResponse({
        user: { id: 1, full_name: "A", login: "a", role: "intern" },
      }),
    ).toBe(false);
    expect(
      isCurrentUserResponse({
        user: { id: "1", full_name: "A", login: 1, role: "intern" },
      }),
    ).toBe(false);
    expect(
      isCurrentUserResponse({
        user: { id: "1", full_name: "A", login: "a", role: "unknown" },
      }),
    ).toBe(false);
  });
});
