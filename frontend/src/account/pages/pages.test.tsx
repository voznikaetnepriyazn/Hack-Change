import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it } from "vitest";
import { QuestionsPage } from "./QuestionsPage";
import { DocumentsPage } from "./DocumentsPage";
import { TasksPage } from "./TasksPage";
import { ProfilePage } from "./ProfilePage";
import { AccountLayout } from "./AccountLayout";
import { AuthContext } from "../../auth/context/AuthContext";
import { useProfileStore } from "../../stores/profileStore";
import { internUser, leadUser } from "../../test/fixtures";
import type { User } from "../types/account.types";

const auth = {
  isAuthLoading: false,
  login: async () => {},
  logout: async () => {},
};

function renderPage(path = "/account/questions") {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <AuthContext.Provider value={auth}>
        <Routes>
          <Route path="/account" element={<AccountLayout />}>
            <Route path="questions" element={<QuestionsPage />} />
            <Route path="documents" element={<DocumentsPage />} />
            <Route path="tasks" element={<TasksPage />} />
            <Route path="profile" element={<ProfilePage />} />
            <Route path="chat" element={<p>чат</p>} />
          </Route>
        </Routes>
      </AuthContext.Provider>
    </MemoryRouter>,
  );
}

describe("account pages", () => {
  beforeEach(() => {
    useProfileStore.setState({ user: internUser });
  });

  it("searches questions and opens another tab", () => {
    renderPage();

    expect(screen.getByRole("heading", { name: "Частые вопросы" })).toBeInTheDocument();
    fireEvent.click(screen.getByText("Стажировка"));
    expect(screen.getByText(/обязательные встречи/)).toBeInTheDocument();

    fireEvent.change(screen.getByPlaceholderText("Найти ответ на вопрос..."), {
      target: { value: "vpn" },
    });
    expect(screen.queryByText(/обязательные встречи/)).not.toBeInTheDocument();

    fireEvent.click(screen.getByText("Все вопросы"));
    expect(screen.getAllByText(/Avito Connect/).length).toBeGreaterThan(0);
    expect(screen.queryByText(/корпоративной почте/)).not.toBeInTheDocument();
  });

  it("searches documents and keeps only the matching tab", () => {
    renderPage("/account/documents");

    expect(screen.getByRole("heading", { name: "Документы" })).toBeInTheDocument();
    fireEvent.click(screen.getByText("Обучение"));
    fireEvent.change(screen.getByPlaceholderText("Поиск по документам..."), {
      target: { value: "грейд" },
    });
    expect(screen.getByText("Грейды, карьера и развитие")).toBeInTheDocument();
    expect(screen.queryByText("Отпуска, отгулы и отсутствия")).not.toBeInTheDocument();
  });

  it("filters tasks by status and search", () => {
    renderPage("/account/tasks");

    expect(screen.getByRole("heading", { name: "Задачи" })).toBeInTheDocument();
    fireEvent.click(screen.getByText("Активные"));
    expect(screen.getByText("Задача 1")).toBeInTheDocument();
    expect(screen.queryByText("Задача 2")).not.toBeInTheDocument();

    fireEvent.change(screen.getByPlaceholderText("Поиск по задачам..."), {
      target: { value: "Задача 5" },
    });
    expect(screen.getByText("Задача 5")).toBeInTheDocument();
    expect(screen.queryByText("Задача 1")).not.toBeInTheDocument();

    fireEvent.click(screen.getByText("Завершённые"));
    expect(screen.queryByText("Задача 5")).not.toBeInTheDocument();
  });

  it("shows the intern profile, a lead profile and a missing user", () => {
    const internView = renderPage("/account/profile");
    expect(screen.getByText("Отгул")).toBeInTheDocument();
    expect(screen.getByText("На согласовании")).toBeInTheDocument();
    expect(screen.getByText("Одобрено")).toBeInTheDocument();
    expect(screen.getByText("Отклонено")).toBeInTheDocument();
    expect(screen.getByText("Неизвестно")).toBeInTheDocument();
    expect(screen.getByText("eva-smirnova-dev")).toBeInTheDocument();
    internView.unmount();

    useProfileStore.setState({ user: leadUser });
    const leadView = renderPage("/account/profile");
    expect(screen.getByText("Личная информация руководителя.")).toBeInTheDocument();
    expect(screen.queryByText("Мои заявки и статусы")).not.toBeInTheDocument();
    leadView.unmount();

    useProfileStore.setState({
      user: {
        ...internUser,
        phone: undefined,
        login: "",
        command: "",
        post: "",
        formatWork: "",
        office: "",
        statusVacation: undefined,
        githubProfile: { login: "", username: "only-name" },
        requests: [],
        vacationStart: null,
        vacationEnd: null,
      } as User,
    });
    const sparseView = renderPage("/account/profile");
    expect(screen.getByText("Нет заявок")).toBeInTheDocument();
    expect(screen.getByText(/Не выбрано/)).toBeInTheDocument();
    sparseView.unmount();

    useProfileStore.setState({ user: null });
    renderPage("/account/profile");
    expect(screen.getByText("Не удалось загрузить данные пользователя.")).toBeInTheDocument();
  });

  it("shows the loader while auth is still restoring", () => {
    render(
      <MemoryRouter initialEntries={["/account/profile"]}>
        <AuthContext.Provider value={{ ...auth, isAuthLoading: true }}>
          <Routes>
            <Route path="/account/profile" element={<ProfilePage />} />
          </Routes>
        </AuthContext.Provider>
      </MemoryRouter>,
    );

    expect(screen.getByText("Загрузка...")).toBeInTheDocument();
  });

  it("shows intern links and collapses the menu", () => {
    const { container } = renderPage();
    expect(screen.getByText("Задачи")).toBeInTheDocument();
    expect(screen.queryByText("Мой стажер")).not.toBeInTheDocument();

    const aside = container.querySelector("aside");
    const before = aside?.className;
    fireEvent.click(aside?.querySelector("svg") as Element);
    expect(aside?.className).not.toBe(before);
  });

  it("shows the lead menu", () => {
    useProfileStore.setState({ user: leadUser });
    renderPage();

    expect(screen.getByText("Мой стажер")).toBeInTheDocument();
    expect(screen.queryByText("Задачи")).not.toBeInTheDocument();
  });
});
