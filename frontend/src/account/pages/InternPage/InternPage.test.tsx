import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { InternPage } from "./index";
import { useInternStore } from "../../../stores/internStore";
import { internUser } from "../../../test/fixtures";

vi.mock("../../api/account.api", () => ({
  getIntern: vi.fn(),
}));

import { getIntern } from "../../api/account.api";

const loadIntern = vi.mocked(getIntern);

describe("InternPage", () => {
  beforeEach(() => {
    loadIntern.mockReset();
    useInternStore.setState({ user: null });
  });

  it("loads and shows the intern", async () => {
    loadIntern.mockResolvedValue(internUser);
    render(<InternPage />);

    expect(screen.getByText("Загрузка...")).toBeInTheDocument();
    expect(await screen.findByText("Ева Смирнова")).toBeInTheDocument();
    expect(screen.getByText("Запросов:")).toBeInTheDocument();
    expect(screen.getByText(/Доступы, задачи/)).toBeInTheDocument();
    await waitFor(() => expect(loadIntern).toHaveBeenCalledOnce());
  });

  it("does not request the intern again when the profile is already loaded", async () => {
    useInternStore.setState({ user: internUser });
    render(<InternPage />);

    expect(screen.getByText("Ева Смирнова")).toBeInTheDocument();
    await waitFor(() => expect(loadIntern).not.toHaveBeenCalled());
  });

  it("shows an error when the intern is missing or the request fails", async () => {
    loadIntern.mockResolvedValueOnce(null as unknown as typeof internUser);
    const { unmount } = render(<InternPage />);
    expect(await screen.findByText("Данные стажёра не найдены")).toBeInTheDocument();
    unmount();

    loadIntern.mockRejectedValueOnce(new Error("Сеть"));
    const failed = render(<InternPage />);
    expect(await screen.findByText("Сеть")).toBeInTheDocument();
    failed.unmount();

    loadIntern.mockRejectedValueOnce("offline");
    render(<InternPage />);
    expect(await screen.findByText("Не удалось загрузить данные стажёра")).toBeInTheDocument();
  });
});
