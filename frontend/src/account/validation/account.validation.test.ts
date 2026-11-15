import { describe, expect, it } from "vitest";
import { isCurrentUserResponse } from "./account.validation";

describe("isCurrentUserResponse", () => {
  it("accepts a user with id, login and role", () => {
    expect(
      isCurrentUserResponse({
        user: { id: "1", login: "@test", role: "intern" },
      }),
    ).toBe(true);
  });

  it("rejects malformed payloads", () => {
    expect(isCurrentUserResponse(null)).toBe(false);
    expect(isCurrentUserResponse(1)).toBe(false);
    expect(isCurrentUserResponse({})).toBe(false);
    expect(isCurrentUserResponse({ user: null })).toBe(false);
    expect(isCurrentUserResponse({ user: { id: 1, login: "a", role: "intern" } })).toBe(
      false,
    );
    expect(isCurrentUserResponse({ user: { id: "1", login: 1, role: "intern" } })).toBe(
      false,
    );
    expect(isCurrentUserResponse({ user: { id: "1", login: "a", role: 1 } })).toBe(false);
  });
});
