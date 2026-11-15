import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ChatPage } from "./index";
import { useProfileStore } from "../../../stores/profileStore";
import { internUser, leadUser } from "../../../test/fixtures";
import ApiError from "../../../shared/errors/ApiError";
import type { ChatStreamHandlers } from "../../api/account.api";

const { streamChat } = vi.hoisted(() => ({ streamChat: vi.fn() }));

vi.mock("../../api/account.api", () => ({
  streamChat,
}));

describe("ChatPage", () => {
  beforeEach(() => {
    streamChat.mockReset();
    sessionStorage.clear();
    useProfileStore.setState({ user: internUser });
  });

  it("restores the intern chat and sends a question", () => {
    sessionStorage.setItem(
      "chat-session",
      JSON.stringify([{ text: "сохранённый", role: "user" }]),
    );
    const stop = vi.fn();
    streamChat.mockImplementation((_message: string, _role: string, handlers: ChatStreamHandlers) => {
      handlers.onDelta("кусок");
      handlers.onDone();
      return stop;
    });

    render(<ChatPage />);
    expect(screen.getByText("сохранённый")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /Расскажи о моей стажировке/ }));
    const input = screen.getByPlaceholderText("Напиши свой вопрос");
    expect(input).toHaveValue("Расскажи о моей стажировке");

    fireEvent.keyDown(input, { key: "Escape" });
    expect(streamChat).not.toHaveBeenCalled();

    fireEvent.keyDown(input, { key: "Enter" });
    expect(streamChat).toHaveBeenCalledWith(
      "Расскажи о моей стажировке",
      "intern",
      expect.objectContaining({
        onDelta: expect.any(Function),
        onEmployeeCards: expect.any(Function),
        onDone: expect.any(Function),
        onError: expect.any(Function),
      }),
    );
    expect(screen.getByRole("heading", { name: "Привет, Ева Смирнова!" })).toBeInTheDocument();
    expect(screen.getByText("кусок")).toBeInTheDocument();
    expect(stop).toBeTypeOf("function");
  });

  it("renders employee cards received from the stream", () => {
    streamChat.mockImplementation((_message: string, _role: string, handlers: ChatStreamHandlers) => {
      handlers.onEmployeeCards?.([
        {
          full_name: "Иванов Иван Иванович",
          position: "Backend Developer",
          email: "i.ivanov@avito.ru",
          responsibilities: "Ревью кода",
          office: "Москва",
        },
        {
          avatar: "https://example.com/eva.jpg",
          first_name: "Ева",
          role: "Стажёр",
        },
      ]);
      handlers.onDone();
      return vi.fn();
    });

    render(<ChatPage />);
    ask("Кто в моей команде?");

    expect(screen.getByRole("heading", { name: "Иванов Иван Иванович" })).toBeInTheDocument();
    expect(screen.getByText("И")).toBeInTheDocument();
    expect(screen.getByText("Backend Developer")).toBeInTheDocument();
    expect(screen.getByText("i.ivanov@avito.ru")).toBeInTheDocument();
    expect(screen.getByText("Ревью кода")).toBeInTheDocument();
    expect(screen.getAllByText("С чем можно обратиться")).toHaveLength(2);
    expect(screen.getByAltText("avatar")).toHaveAttribute("src", "https://example.com/eva.jpg");
    expect(screen.queryByText("Москва")).not.toBeInTheDocument();
    expect(screen.queryByText("Ева")).not.toBeInTheDocument();
    expect(screen.queryByText("Стажёр")).not.toBeInTheDocument();
  });

  it("does not send an empty question and blocks a second send while streaming", () => {
    streamChat.mockReturnValue(vi.fn());
    render(<ChatPage />);

    fireEvent.click(screen.getByRole("button", { name: "" }));
    expect(streamChat).not.toHaveBeenCalled();

    const input = screen.getByPlaceholderText("Напиши свой вопрос");
    fireEvent.change(input, { target: { value: "   " } });
    fireEvent.keyDown(input, { key: "Enter" });
    expect(streamChat).not.toHaveBeenCalled();

    fireEvent.change(input, { target: { value: "вопрос" } });
    fireEvent.click(screen.getAllByRole("button").at(-1)!);
    fireEvent.click(screen.getAllByRole("button").at(-1)!);
    expect(streamChat).toHaveBeenCalledOnce();
  });

  it("tells a lead that the chat is intern-only", () => {
    useProfileStore.setState({ user: leadUser });
    render(<ChatPage />);

    fireEvent.change(screen.getByPlaceholderText("Напиши свой вопрос"), {
      target: { value: "вопрос" },
    });
    fireEvent.keyDown(screen.getByPlaceholderText("Напиши свой вопрос"), { key: "Enter" });

    expect(screen.getByText("Чат доступен только стажёру")).toBeInTheDocument();
    expect(streamChat).not.toHaveBeenCalled();
  });

  function ask(text: string) {
    fireEvent.change(screen.getByPlaceholderText("Напиши свой вопрос"), {
      target: { value: text },
    });
    fireEvent.keyDown(screen.getByPlaceholderText("Напиши свой вопрос"), { key: "Enter" });
  }

  it("keeps a partial answer when the stream fails afterwards", () => {
    streamChat.mockImplementation((_message: string, _role: string, handlers: ChatStreamHandlers) => {
      handlers.onDelta("частично");
      handlers.onError(new Error(""));
      return vi.fn();
    });
    render(<ChatPage />);
    ask("вопрос");

    expect(screen.getByText("частично")).toBeInTheDocument();
  });

  it("replaces an empty answer with the stream error", () => {
    streamChat.mockImplementation((_message: string, _role: string, handlers: ChatStreamHandlers) => {
      handlers.onError(new Error("Сеть недоступна"));
      return vi.fn();
    });
    render(<ChatPage />);
    ask("ещё");

    expect(screen.getByText("Сеть недоступна")).toBeInTheDocument();
  });

  it("signs the user out when the stream returns 401", () => {
    streamChat.mockImplementation((_message: string, _role: string, handlers: ChatStreamHandlers) => {
      handlers.onError(new ApiError("сессия истекла", 401));
      return vi.fn();
    });
    render(<ChatPage />);
    ask("снова");

    expect(useProfileStore.getState().user).toBeNull();
    expect(sessionStorage.getItem("chat-session")).toBeNull();
  });

  it("stops the stream when the page unmounts during a request", () => {
    const stop = vi.fn();
    streamChat.mockReturnValue(stop);
    useProfileStore.setState({ user: internUser });
    const { unmount } = render(<ChatPage />);

    fireEvent.change(screen.getByPlaceholderText("Напиши свой вопрос"), {
      target: { value: "вопрос" },
    });
    fireEvent.keyDown(screen.getByPlaceholderText("Напиши свой вопрос"), { key: "Enter" });
    unmount();

    expect(stop).toHaveBeenCalledOnce();
  });
});
