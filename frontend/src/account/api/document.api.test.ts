import { afterEach, describe, expect, it, vi } from "vitest";
import { getDocumentContent } from "./document.api";

describe("getDocumentContent", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("returns the document text", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({ ok: true, text: async () => "# Документ" }),
    );

    await expect(getDocumentContent("/docs/rules.md")).resolves.toBe("# Документ");
    expect(fetch).toHaveBeenCalledWith("/docs/rules.md");
  });

  it("rejects a failed response", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: false }));

    await expect(getDocumentContent("/docs/missing.md")).rejects.toThrow(
      "Не удалось загрузить документ",
    );
  });
});
