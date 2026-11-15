import { afterEach, describe, expect, it, vi } from "vitest";
import { getCurrentUser, getIntern, streamChat } from "./account.api";
import ApiError from "../../shared/errors/ApiError";
import type { EmployeeCard } from "../types/account.types";

class FakeEventSource {
  static latest: FakeEventSource | null = null;
  onmessage: ((event: MessageEvent) => void) | null = null;
  onerror: (() => void) | null = null;
  closed = false;
  private listeners = new Map<string, Array<(event: MessageEvent) => void>>();

  constructor(
    public url: string,
    public init?: EventSourceInit,
  ) {
    FakeEventSource.latest = this;
  }

  addEventListener(type: string, listener: (event: MessageEvent) => void) {
    const list = this.listeners.get(type) ?? [];
    list.push(listener);
    this.listeners.set(type, list);
  }

  close() {
    this.closed = true;
  }

  emitMessage(data: string) {
    this.onmessage?.({ data } as MessageEvent);
  }

  emitDone() {
    this.listeners.get("done")?.forEach((listener) => listener({ data: "" } as MessageEvent));
  }

  emitEmployeeCards(data: string) {
    this.listeners.get("employee_cards")?.forEach((listener) =>
      listener({ data } as MessageEvent),
    );
  }

  emitError() {
    this.onerror?.();
  }
}

describe("account profile api", () => {
  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it("restores the authenticated profile from the backend", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({
        user: {
          id: "usr-1",
          full_name: "Ева Смирнова",
          login: "e.smirnova",
          role: "intern",
        },
      }),
    });
    vi.stubGlobal("fetch", fetchMock);

    await expect(getCurrentUser()).resolves.toMatchObject({
      id: "usr-1",
      name: "Ева Смирнова",
      login: "e.smirnova",
      role: "intern",
    });
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/auth/me",
      expect.objectContaining({ credentials: "include" }),
    );
  });

  it("returns the intern card", async () => {
    vi.useFakeTimers();
    const pending = getIntern();
    await vi.advanceTimersByTimeAsync(1000);

    await expect(pending).resolves.toMatchObject({ role: "intern", name: "Имя Фамилия" });
  });
});

describe("streamChat", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    FakeEventSource.latest = null;
  });

  it("rejects a non-intern without opening a stream", () => {
    const onError = vi.fn();
    const stop = streamChat("привет", "lead", {
      onDelta: vi.fn(),
      onDone: vi.fn(),
      onError,
    });

    expect(onError).toHaveBeenCalledWith(expect.objectContaining({
      message: "Chat access is restricted",
    }));
    expect(stop()).toBeUndefined();
  });

  it("streams deltas and finishes on the done event", async () => {
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 202 });
    vi.stubGlobal("fetch", fetchMock);
    vi.stubGlobal("EventSource", FakeEventSource);

    const onDelta = vi.fn();
    const onEmployeeCards = vi.fn();
    const onDone = vi.fn();
    const onError = vi.fn();
    streamChat("привет", "intern", { onDelta, onEmployeeCards, onDone, onError });

    await vi.waitFor(() => expect(FakeEventSource.latest).not.toBeNull());
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/chat",
      expect.objectContaining({ method: "POST", credentials: "include" }),
    );
    expect(FakeEventSource.latest?.url).toBe("/api/stream");
    expect(FakeEventSource.latest?.init).toEqual({ withCredentials: true });

    FakeEventSource.latest?.emitMessage("");
    FakeEventSource.latest?.emitMessage("Ответ");
    expect(onDelta).toHaveBeenCalledTimes(1);
    expect(onDelta).toHaveBeenCalledWith("Ответ");

    const cards: EmployeeCard[] = [
      { full_name: "Иванов Иван Иванович", position: "Backend Developer" },
    ];
    FakeEventSource.latest?.emitEmployeeCards(JSON.stringify(cards));
    expect(onEmployeeCards).toHaveBeenCalledWith(cards);

    FakeEventSource.latest?.emitDone();
    expect(onDone).toHaveBeenCalledOnce();
    expect(FakeEventSource.latest?.closed).toBe(true);

    FakeEventSource.latest?.emitError();
    expect(onError).not.toHaveBeenCalled();
  });

  it("reports invalid employee card payloads", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: true, status: 202 }));
    vi.stubGlobal("EventSource", FakeEventSource);
    const onError = vi.fn();

    streamChat("вопрос", "intern", {
      onDelta: vi.fn(),
      onDone: vi.fn(),
      onError,
    });
    await vi.waitFor(() => expect(FakeEventSource.latest).not.toBeNull());

    FakeEventSource.latest?.emitEmployeeCards(JSON.stringify([{ password: "private" }]));
    expect(onError).toHaveBeenCalledWith(
      expect.objectContaining({ message: "Invalid employee cards event" }),
    );
    expect(FakeEventSource.latest?.closed).toBe(true);
  });

  it("reports an API error from the chat request", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 403,
        json: async () => ({ error: "Chat access is restricted" }),
      }),
    );
    const onError = vi.fn();

    streamChat("привет", "intern", { onDelta: vi.fn(), onDone: vi.fn(), onError });

    await vi.waitFor(() => expect(onError).toHaveBeenCalledOnce());
    expect(onError.mock.calls[0][0]).toBeInstanceOf(ApiError);
    expect(onError.mock.calls[0][0].message).toBe("Chat access is restricted");
  });

  it("uses the status text when the error body is not JSON or has no message", async () => {
    const onError = vi.fn();
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
        json: async () => {
          throw new Error("bad");
        },
      }),
    );

    streamChat("привет", "intern", { onDelta: vi.fn(), onDone: vi.fn(), onError });
    await vi.waitFor(() => expect(onError).toHaveBeenCalledOnce());
    expect(onError.mock.calls[0][0].message).toBe("API responded with 500");

    onError.mockClear();
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 418,
        json: async () => ({ error: 1 }),
      }),
    );
    streamChat("привет", "intern", { onDelta: vi.fn(), onDone: vi.fn(), onError });
    await vi.waitFor(() => expect(onError).toHaveBeenCalledOnce());
    expect(onError.mock.calls[0][0].message).toBe("API responded with 418");
  });

  it("does not open the stream when the caller already stopped", async () => {
    let resolveFetch: (response: Response) => void = () => {};
    vi.stubGlobal(
      "fetch",
      vi.fn().mockImplementation(
        () =>
          new Promise<Response>((resolve) => {
            resolveFetch = resolve;
          }),
      ),
    );
    vi.stubGlobal("EventSource", FakeEventSource);

    const stop = streamChat("привет", "intern", {
      onDelta: vi.fn(),
      onDone: vi.fn(),
      onError: vi.fn(),
    });
    stop();
    resolveFetch({ ok: true, status: 202 } as Response);
    await Promise.resolve();
    await Promise.resolve();

    expect(FakeEventSource.latest).toBeNull();
  });

  it("ignores an abort and reports other failures", async () => {
    const onError = vi.fn();
    vi.stubGlobal(
      "fetch",
      vi.fn().mockRejectedValue(new DOMException("aborted", "AbortError")),
    );

    streamChat("привет", "intern", { onDelta: vi.fn(), onDone: vi.fn(), onError });
    await Promise.resolve();
    await Promise.resolve();
    expect(onError).not.toHaveBeenCalled();

    vi.stubGlobal("fetch", vi.fn().mockRejectedValue("offline"));
    streamChat("привет", "intern", { onDelta: vi.fn(), onDone: vi.fn(), onError });
    await vi.waitFor(() => expect(onError).toHaveBeenCalledOnce());
    expect(onError.mock.calls[0][0].message).toBe("Failed to start chat");

    onError.mockClear();
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("network")));
    streamChat("привет", "intern", { onDelta: vi.fn(), onDone: vi.fn(), onError });
    await vi.waitFor(() => expect(onError).toHaveBeenCalledOnce());
    expect(onError.mock.calls[0][0].message).toBe("network");
  });

  it("reports a stream error and closes the source when stopped", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: true, status: 202 }));
    vi.stubGlobal("EventSource", FakeEventSource);
    const onError = vi.fn();

    const stop = streamChat("привет", "intern", {
      onDelta: vi.fn(),
      onDone: vi.fn(),
      onError,
    });
    await vi.waitFor(() => expect(FakeEventSource.latest).not.toBeNull());
    FakeEventSource.latest?.emitError();
    expect(onError).toHaveBeenCalledWith(expect.objectContaining({ message: "Stream failed" }));
    expect(FakeEventSource.latest?.closed).toBe(true);

    stop();
    expect(FakeEventSource.latest?.closed).toBe(true);
  });
});
