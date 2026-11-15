const STORAGE_KEY = "chat-session";

export interface SessionChatMessage {
  text: string;
  role: "user" | "bot";
}

function isSessionChatMessage(value: unknown): value is SessionChatMessage {
  if (!value || typeof value !== "object") {
    return false;
  }

  const message = value as Record<string, unknown>;

  return (
    typeof message.text === "string" &&
    (message.role === "user" || message.role === "bot")
  );
}

export function loadSessionChat(): SessionChatMessage[] {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY);
    if (!raw) {
      return [];
    }

    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) {
      return [];
    }

    return parsed.filter(isSessionChatMessage);
  } catch {
    return [];
  }
}

export function saveSessionChat(
  messages: Array<{ text: string; role?: "user" | "bot" }>,
) {
  try {
    const stored = messages.filter(
      (message): message is SessionChatMessage =>
        (message.role === "user" || message.role === "bot") &&
        message.text.trim() !== "",
    );
    if (stored.length === 0) {
      sessionStorage.removeItem(STORAGE_KEY);
      return;
    }

    sessionStorage.setItem(STORAGE_KEY, JSON.stringify(stored));
  } catch {
    return;
  }
}

export function clearSessionChat() {
  try {
    sessionStorage.removeItem(STORAGE_KEY);
  } catch {
    return;
  }
}
