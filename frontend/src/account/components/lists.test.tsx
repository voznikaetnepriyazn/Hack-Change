import type { ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { QuestionsList } from "./QuestionsList";
import { DocumentsList } from "./DocumentsList";
import { TasksList } from "./TasksList";
import { Tabs } from "./Tabs";
import { QuestionBlock } from "./QuestionBlock";
import { ProfileItem } from "./ProfileItem";
import { CommandItem } from "./CommandItem";
import { useProfileStore } from "../../stores/profileStore";
import { internUser, leadUser } from "../../test/fixtures";
import type { Question, Task } from "../types/account.types";

const questions: Question[] = [
  { id: 1, title: "Почта", description: "доступ", answer: "смотри письмо" },
  { id: 2, title: "Wi-Fi", description: "сеть", answer: "Avito_Corp" },
];

const tasks: Task[] = [
  {
    id: 1,
    code: "ALPHA-1",
    title: "Сегодня",
    description: "сделать сегодня",
    status: "active",
    deadline: new Date(),
  },
  {
    id: 2,
    code: "ALPHA-2",
    title: "Вчера",
    description: "уже поздно",
    status: "active",
    deadline: daysFromToday(-3),
  },
  {
    id: 3,
    code: "ALPHA-3",
    title: "Потом",
    description: "ещё рано",
    status: "completed",
    deadline: daysFromToday(3),
  },
];

function daysFromToday(days: number) {
  const date = new Date();
  date.setDate(date.getDate() + days);
  return date;
}

function renderWithRouter(ui: ReactNode) {
  return render(
    <MemoryRouter initialEntries={["/account/questions"]}>
      <Routes>
        <Route path="/account/questions" element={ui} />
        <Route path="/account/chat" element={<p>чат</p>} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("lists and profile cards", () => {
  it("filters, expands and collapses questions", () => {
    renderWithRouter(<QuestionsList questions={questions} items={[2]} search="wi" />);

    expect(screen.queryByText("Почта")).not.toBeInTheDocument();
    expect(screen.getByText("Wi-Fi")).toBeInTheDocument();

    fireEvent.click(screen.getByText("Wi-Fi"));
    expect(screen.getByText("Avito_Corp")).toBeInTheDocument();
    fireEvent.click(screen.getByText("Wi-Fi"));
  });

  it("shows every question when the tab has no ids", () => {
    renderWithRouter(<QuestionsList questions={questions} items={[]} search="" />);
    expect(screen.getByText("Почта")).toBeInTheDocument();
    expect(screen.getByText("Wi-Fi")).toBeInTheDocument();
  });

  it("filters documents by tab and search", () => {
    renderWithRouter(
      <DocumentsList
        documents={[
          { id: 1, title: "Памятка", description: "старт", link: "https://example.com/a" },
          { id: 2, title: "VPN", description: "сеть", link: "https://example.com/b" },
        ]}
        items={[1]}
        search="пам"
      />,
    );

    expect(screen.getByText("Памятка")).toBeInTheDocument();
    expect(screen.queryByText("VPN")).not.toBeInTheDocument();
  });

  it("opens a task, follows the assistant link and toggles the deadline states", () => {
    renderWithRouter(<TasksList tasks={tasks} search="" />);

    expect(screen.getByText("Дедлайн: сегодня")).toBeInTheDocument();
    expect(screen.getByText("Дедлайн: просрочено")).toBeInTheDocument();
    expect(screen.getAllByText(/Дедлайн:/)).toHaveLength(3);

    const overdue = screen.getByRole("heading", { name: "Вчера" }).parentElement?.parentElement;
    fireEvent.click(overdue?.querySelector("svg") as Element);
    fireEvent.click(screen.getAllByText("Открыть задачу")[0]);
    expect(screen.getByText("сделать сегодня")).toBeInTheDocument();
    fireEvent.click(screen.getByText("Попросите помочь AI-Buddy"));
    expect(screen.getByText("чат")).toBeInTheDocument();
  });

  it("hides a task that does not match the search", () => {
    renderWithRouter(<TasksList tasks={tasks} search="потом" />);
    expect(screen.getByText("Потом")).toBeInTheDocument();
    expect(screen.queryByText("Сегодня")).not.toBeInTheDocument();
  });

  it("switches tabs and opens the assistant from a question block", () => {
    const setActiveTab = vi.fn();
    renderWithRouter(
      <>
        <Tabs
          activeTab="all"
          setActiveTab={setActiveTab}
          tags={[
            { id: 0, name: "Все", tab: "all", items: [] },
            { id: 1, name: "Доступы", tab: "access", items: [1] },
          ]}
        />
        <QuestionBlock content={{ title: "Не нашли?", description: "Спросите" }} />
      </>,
    );

    fireEvent.click(screen.getByText("Доступы"));
    expect(setActiveTab).toHaveBeenCalledWith("access");
    fireEvent.click(screen.getByRole("button", { name: "Задать вопрос AI" }));
    expect(screen.getByText("чат")).toBeInTheDocument();
  });

  it("renders profile initials, an avatar and the intern post for a lead", () => {
    useProfileStore.setState({ user: leadUser });
    const { rerender } = renderWithRouter(
      <ProfileItem user={internUser} size="large" className="card" />,
    );
    expect(screen.getByText("Е")).toBeInTheDocument();
    expect(screen.getByText("Разработчик")).toBeInTheDocument();

    useProfileStore.setState({ user: internUser });
    rerender(
      <MemoryRouter>
        <ProfileItem user={internUser} size="large" />
      </MemoryRouter>,
    );
    expect(screen.queryByText("Разработчик")).not.toBeInTheDocument();

    rerender(
      <MemoryRouter>
        <ProfileItem user={leadUser} size="small" isShort />
      </MemoryRouter>,
    );
    expect(screen.getByAltText("avatar")).toHaveAttribute("src", leadUser.avatar);

    rerender(
      <MemoryRouter>
        <ProfileItem user={null} />
      </MemoryRouter>,
    );
  });

  it("renders an employee card from the passed profile", () => {
    const { rerender } = render(
      <CommandItem
        employee={{
          full_name: "Ева Смирнова",
          position: "Стажёр",
          email: "e.smirnova@avito.ru",
          responsibilities: "Онбординг и первые задачи",
        }}
      />,
    );

    expect(screen.getByRole("heading", { name: "Ева Смирнова" })).toBeInTheDocument();
    expect(screen.getByText("Е")).toBeInTheDocument();
    expect(screen.getByText("Стажёр")).toBeInTheDocument();
    expect(screen.getByText("e.smirnova@avito.ru")).toBeInTheDocument();
    expect(screen.getByText("Онбординг и первые задачи")).toBeInTheDocument();

    rerender(
      <CommandItem
        employee={{
          avatar: "https://example.com/ivan.jpg",
          position: "Разработчик",
        }}
      />,
    );

    expect(screen.getByAltText("avatar")).toHaveAttribute(
      "src",
      "https://example.com/ivan.jpg",
    );
    expect(screen.queryByRole("heading")).not.toBeInTheDocument();
    expect(screen.queryByText("e.smirnova@avito.ru")).not.toBeInTheDocument();
    expect(screen.queryByText("Онбординг и первые задачи")).not.toBeInTheDocument();
  });
});
