import { afterEach, describe, expect, it, vi } from "vitest";
import { apiRequest } from "./client";
import ApiError from "../shared/errors/ApiError";

function jsonResponse(body: unknown, status = 200) {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  } as Response;
}

describe("apiRequest", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("sends credentials and returns parsed JSON", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ ok: true }));
    vi.stubGlobal("fetch", fetchMock);

    await expect(apiRequest("/auth/login", { method: "POST" })).resolves.toEqual({
      ok: true,
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/auth/login",
      expect.objectContaining({
        method: "POST",
        credentials: "include",
      }),
    );
  });

  it("keeps an explicit credentials mode", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ ok: true }));
    vi.stubGlobal("fetch", fetchMock);

    await apiRequest("health", { credentials: "omit", headers: { "X-Test": "1" } });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/health",
      expect.objectContaining({ credentials: "omit" }),
    );
  });

  it("throws when the body is not JSON", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => {
          throw new Error("bad json");
        },
      }),
    );

    await expect(apiRequest("health")).rejects.toThrow(
      "API returned an invalid JSON or empty data response",
    );
  });

  it("accepts a 204 No Content response", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        status: 204,
        json: async () => {
          throw new Error("no content");
        },
      }),
    );

    await expect(apiRequest("auth/logout", { method: "POST" })).resolves.toBeNull();
  });

  it("throws ApiError with the server message", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(jsonResponse({ error: "Invalid login or password" }, 401)),
    );

    await expect(apiRequest("auth/login")).rejects.toMatchObject({
      message: "Invalid login or password",
      statusCode: 401,
    });
  });

  it("throws ApiError with the status when the error payload is unusable", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse({ error: 5 }, 500)));

    const error = await apiRequest("auth/login").catch((reason: unknown) => reason);

    expect(error).toBeInstanceOf(ApiError);
    expect((error as ApiError).message).toBe("API responded with 500");
  });

  it("throws ApiError when the error payload is not an object", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse(null, 502)));

    await expect(apiRequest("auth/login")).rejects.toThrow("API responded with 502");
  });

  it("throws when a successful response has an empty JSON body", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse(null, 200)));

    await expect(apiRequest("auth/logout")).rejects.toThrow(
      "API returned an empty or invalid JSON response",
    );
  });

  it("throws ApiError when the payload has no error field", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse({ reason: "nope" }, 400)));

    await expect(apiRequest("auth/login")).rejects.toThrow("API responded with 400");
  });
});
