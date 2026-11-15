import { beforeEach, describe, expect, it, vi } from "vitest";
import { isLoginResponse, loginUser, logoutUser } from "./auth.api";

vi.mock("../../api/client", () => ({
  apiRequest: vi.fn(),
}));

import { apiRequest } from "../../api/client";

const request = vi.mocked(apiRequest);

const loginPayload = {
  user: {
    id: "usr_1001",
    full_name: "Ева Смирнова",
    login: "e.smirnova",
    role: "intern",
  },
};

describe("isLoginResponse", () => {
  it("accepts the backend user identity response", () => {
    expect(isLoginResponse(loginPayload)).toBe(true);
  });

  it("rejects malformed payloads", () => {
    expect(isLoginResponse(null)).toBe(false);
    expect(isLoginResponse("user")).toBe(false);
    expect(isLoginResponse({})).toBe(false);
    expect(isLoginResponse({ user: null })).toBe(false);
    expect(isLoginResponse({ user: { id: "1", login: "a", role: "intern" } })).toBe(false);
    expect(isLoginResponse({ user: { id: "1", full_name: "A", login: "a", role: "unknown" } })).toBe(false);
    expect(isLoginResponse({ user: { id: 1, full_name: "A", login: "a", role: "intern" } })).toBe(false);
    expect(isLoginResponse({ user: { id: "1", full_name: "A", login: 1, role: "intern" } })).toBe(false);
  });
});

describe("loginUser", () => {
  beforeEach(() => {
    request.mockReset();
  });

  it("posts validated credentials and returns the user", async () => {
    request.mockResolvedValue(loginPayload);

    await expect(loginUser("@e.smirnova", "Password123!")).resolves.toEqual(loginPayload);
    expect(request).toHaveBeenCalledWith("auth/login", {
      method: "POST",
      body: JSON.stringify({ login: "e.smirnova", password: "Password123!" }),
      headers: { "Content-Type": "application/json" },
    });
  });

  it("rejects an unexpected response shape", async () => {
    request.mockResolvedValue({ user: { id: "1" } });

    await expect(loginUser("e.smirnova", "Password123!")).rejects.toThrow(
      "Invalid response format",
    );
  });
});

describe("logoutUser", () => {
  beforeEach(() => {
    request.mockReset();
  });

  it("accepts the backend's empty 204 logout response", async () => {
    request.mockResolvedValue(null);

    await expect(logoutUser()).resolves.toBeUndefined();
    expect(request).toHaveBeenCalledWith("auth/logout", { method: "POST" });
  });
});
