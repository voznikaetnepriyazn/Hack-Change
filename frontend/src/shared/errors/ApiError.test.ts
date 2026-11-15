import { describe, expect, it } from "vitest";
import ApiError from "./ApiError";

describe("ApiError", () => {
  it("keeps the message and status code", () => {
    const error = new ApiError("Invalid login or password", 401);

    expect(error).toBeInstanceOf(Error);
    expect(error.name).toBe("ApiError");
    expect(error.message).toBe("Invalid login or password");
    expect(error.statusCode).toBe(401);
  });
});
