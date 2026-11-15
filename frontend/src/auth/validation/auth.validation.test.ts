import { describe, expect, it } from "vitest";
import { validateLoginData } from "./auth.validation";

describe("validateLoginData", () => {
  it("strips @ and whitespace from the login", () => {
    expect(validateLoginData("  @@e.smirnova  ", "Password123!")).toEqual({
      login: "e.smirnova",
      password: "Password123!",
    });
  });

  it("rejects an empty login or password", () => {
    expect(() => validateLoginData("@@@", "Password123!")).toThrow(
      "Password and login are required",
    );
    expect(() => validateLoginData("e.smirnova", "")).toThrow(
      "Password and login are required",
    );
  });

  it("rejects a password outside 8..100 characters", () => {
    expect(() => validateLoginData("e.smirnova", "short")).toThrow(
      "Password must be between 8 and 100 characters",
    );
    expect(() => validateLoginData("e.smirnova", "p".repeat(101))).toThrow(
      "Password must be between 8 and 100 characters",
    );
  });
});
