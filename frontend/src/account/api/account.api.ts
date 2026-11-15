import { API_URL } from "../../api/config";
import { isCurrentUserResponse } from "../validation/account.validation";
import ApiError from "../../shared/errors/ApiError";
import { User } from "../types/account.types";

const intern: User = {
  id: "1",
  name: "Имя Фамилия",
  avatar: null,
  position: "Стажёр",
  phone: "+7 (999) 999-99-99",
  login: "@test",
  command: "Команда 1",
  post: "Должность",
  formatWork: "Гибрид, 20 ч/нед",
  office: "Москва",
  role: "intern" as const,
  statusVacation: "Онбординг",
  githubProfile: {
    login: "e.smirnova",
    username: "eva-smirnova-dev",
  },
  requests: [
    {
      id: "1",
      name: "Отгул на 15 ноября",
      description: "",
      status: "pending",
    },
    {
      id: "2",
      name: "Запрос доступа к Redash",
      description: "",
      status: "approved",
    },
    {
      id: "3",
      name: "Запрос на замену ноутбука",
      description: "Нет свободных устройств",
      status: "rejected",
    },
  ],
  vacationStart: new Date(),
  vacationEnd: new Date(new Date().setDate(new Date().getDate() + 10)),
};
const intern1: User = {
  id: "2",
  avatar:
    "https://thumb.wikimedia.org/wikipedia/ru/thumb/0/0f/Tom_Holland_as_Spider-Man.jpg/330px-Tom_Holland_as_Spider-Man.jpg?utm_source=ru.wikipedia.org&utm_campaign=parser&utm_content=thumbnail",
  name: "Иван Иванов",
  position: "Стажёр",
  phone: "+7 (999) 999-99-98",
  login: "@ivan",
  role: "intern" as const,
  vacationStart: null,
  vacationEnd: null,
  command: "Команда 1",
  post: "Должность",
  formatWork: "Гибрид, 20 ч/нед",
  office: "Москва",
  statusVacation: "Онбординг",
  githubProfile: {
    login: "i.ivanov",
    username: "ivan-ivanov-dev",
  },
  requests: [
    {
      id: "1",
      name: "Отгул на 20 ноября",
      description: "",
      status: "pending",
    },
  ],
};
const lead: User = {
  id: "3",
  avatar:
    "https://thumb.wikimedia.org/wikipedia/ru/thumb/0/0f/Tom_Holland_as_Spider-Man.jpg/330px-Tom_Holland_as_Spider-Man.jpg?utm_source=ru.wikipedia.org&utm_campaign=parser&utm_content=thumbnail",
  name: "Иван Иванов",
  position: "Руководитель",
  phone: "+7 (999) 999-99-98",
  login: "@ivan",
  role: "lead" as const,
  command: "Команда 1",
  post: "Должность",
  formatWork: "Гибрид, 40 ч/нед",
  office: "Москва",
};

export const getCurrentUser = async (role?: string) => {
  const data = {
    user: null as User | null,
  };
  await new Promise((resolve) => setTimeout(resolve, 1000));
  data.user = role === "lead" ? lead : intern;

  if (!isCurrentUserResponse(data)) {
    throw new ApiError("Invalid response format", 400);
  }

  return data.user;
};

export const getIntern = async () => {
  const data = {
    user: null as User | null,
  };
  await new Promise((resolve) => setTimeout(resolve, 1000));
  data.user = intern;

  return data.user;
};


export interface ChatStreamHandlers {
  onDelta: (text: string) => void;
  onDone: () => void;
  onError: (error: Error) => void;
}

function apiUrl(path: string) {
  return `${API_URL.replace(/\/$/, "")}/${path.replace(/^\//, "")}`;
}

async function readErrorMessage(response: Response) {
  try {
    const data: unknown = await response.json();
    if (
      data &&
      typeof data === "object" &&
      "error" in data &&
      typeof data.error === "string"
    ) {
      return data.error;
    }
  } catch {
    return `API responded with ${response.status}`;
  }

  return `API responded with ${response.status}`;
}

export function streamChat(
  message: string,
  role: string,
  handlers: ChatStreamHandlers,
) {
  if (role !== "intern") {
    handlers.onError(new Error("Chat access is restricted"));
    return () => {};
  }

  const controller = new AbortController();
  let source: EventSource | null = null;
  let stopped = false;
  let finished = false;

  const stop = () => {
    stopped = true;
    controller.abort();
    source?.close();
  };

  void (async () => {
    try {
      const response = await fetch(apiUrl("chat"), {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ message }),
        signal: controller.signal,
      });

      if (!response.ok) {
        throw new ApiError(await readErrorMessage(response), response.status);
      }

      if (stopped) {
        return;
      }

      source = new EventSource(apiUrl("stream"), { withCredentials: true });

      source.onmessage = (event) => {
        if (event.data) {
          handlers.onDelta(event.data);
        }
      };

      source.addEventListener("done", () => {
        finished = true;
        source?.close();
        handlers.onDone();
      });

      source.onerror = () => {
        source?.close();
        if (!stopped && !finished) {
          handlers.onError(new Error("Stream failed"));
        }
      };
    } catch (error) {
      if (stopped || (error instanceof DOMException && error.name === "AbortError")) {
        return;
      }

      handlers.onError(
        error instanceof Error ? error : new Error("Failed to start chat"),
      );
    }
  })();

  return stop;
}