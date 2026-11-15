import { afterEach, describe, expect, it, vi } from "vitest";
import { clearSessionChat, loadSessionChat, saveSessionChat } from "./sessionChat";

describe("session chat storage", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    sessionStorage.clear();
  });

  it("returns an empty list when nothing is stored", () => {
    expect(loadSessionChat()).toEqual([]);
  });

  it("returns stored messages and drops invalid entries", () => {
    sessionStorage.setItem(
      "chat-session",
      JSON.stringify([
        { text: "Привет", role: "user" },
        { text: "Ответ", role: "bot" },
        { text: "нет роли" },
        { role: "user" },
        null,
      ]),
    );

    expect(loadSessionChat()).toEqual([
      { text: "Привет", role: "user" },
      { text: "Ответ", role: "bot" },
    ]);
  });

  it("returns an empty list for invalid JSON and non-arrays", () => {
    sessionStorage.setItem("chat-session", "{");
    expect(loadSessionChat()).toEqual([]);

    sessionStorage.setItem("chat-session", JSON.stringify({ text: "нет" }));
    expect(loadSessionChat()).toEqual([]);
  });

  it("saves only non-empty messages with a role", () => {
    saveSessionChat([
      { text: "  вопрос  ", role: "user" },
      { text: "   ", role: "bot" },
      { text: "без роли" },
    ]);

    expect(loadSessionChat()).toEqual([{ text: "  вопрос  ", role: "user" }]);
  });

  it("removes the key when there is nothing to save", () => {
    sessionStorage.setItem("chat-session", "[]");
    saveSessionChat([{ text: " ", role: "bot" }]);

    expect(sessionStorage.getItem("chat-session")).toBeNull();
  });

  it("clears the stored chat", () => {
    saveSessionChat([{ text: "вопрос", role: "user" }]);
    clearSessionChat();

    expect(loadSessionChat()).toEqual([]);
  });

  it("ignores storage failures", () => {
    vi.spyOn(Storage.prototype, "getItem").mockImplementation(() => {
      throw new Error("blocked");
    });
    expect(loadSessionChat()).toEqual([]);

    vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
      throw new Error("blocked");
    });
    expect(() => saveSessionChat([{ text: "вопрос", role: "user" }])).not.toThrow();

    vi.spyOn(Storage.prototype, "removeItem").mockImplementation(() => {
      throw new Error("blocked");
    });
    expect(() => clearSessionChat()).not.toThrow();
    expect(() => saveSessionChat([])).not.toThrow();
  });
});
