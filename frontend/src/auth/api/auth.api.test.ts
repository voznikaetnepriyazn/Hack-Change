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
    login: "e.smirnova",
    role: "intern",
  },
};

describe("isLoginResponse", () => {
  it("accepts a user with id, login and role", () => {
    expect(isLoginResponse(loginPayload)).toBe(true);
  });

  it("rejects malformed payloads", () => {
    expect(isLoginResponse(null)).toBe(false);
    expect(isLoginResponse("user")).toBe(false);
    expect(isLoginResponse({})).toBe(false);
    expect(isLoginResponse({ user: null })).toBe(false);
    expect(isLoginResponse({ user: { id: 1, login: "a", role: "intern" } })).toBe(false);
    expect(isLoginResponse({ user: { id: "1", login: 1, role: "intern" } })).toBe(false);
    expect(isLoginResponse({ user: { id: "1", login: "a", role: 1 } })).toBe(false);
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

  it("returns the logout message", async () => {
    request.mockResolvedValue({ message: "bye" });

    await expect(logoutUser()).resolves.toEqual({ message: "bye" });
    expect(request).toHaveBeenCalledWith("auth/logout", { method: "POST" });
  });

  it("rejects a response that is not an object", async () => {
    request.mockResolvedValue(null);

    await expect(logoutUser()).rejects.toThrow("Invalid response format");
  });

  it("rejects a response without a message", async () => {
    request.mockResolvedValue({ ok: true });

    await expect(logoutUser()).rejects.toThrow("Invalid response format");
  });
});
