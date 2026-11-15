import { fireEvent, render, screen } from "@testing-library/react";
import { RouterProvider } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { router } from "./router";
import { AuthProvider } from "../auth/context/AuthContext";
import { useProfileStore } from "../stores/profileStore";
import { internUser } from "../test/fixtures";

vi.mock("../account/api/account.api", () => ({
  getCurrentUser: vi.fn(),
  getIntern: vi.fn(),
  streamChat: vi.fn(() => () => {}),
}));

import { getCurrentUser } from "../account/api/account.api";

const currentUser = vi.mocked(getCurrentUser);
type Profile = Awaited<ReturnType<typeof getCurrentUser>>;

describe("router", () => {
  beforeEach(async () => {
    currentUser.mockReset();
    currentUser.mockResolvedValue(internUser as Profile);
    useProfileStore.setState({ user: null });
    await router.navigate("/");
  });

  it("opens the chat for a restored intern and redirects unknown paths", async () => {
    render(
      <AuthProvider>
        <RouterProvider router={router} />
      </AuthProvider>,
    );

    expect(await screen.findByText("Привет, Имя!")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("link", { name: "Задачи" }));
    expect(await screen.findByText("Список задач, их описание и дедлайны")).toBeInTheDocument();

    await router.navigate("/missing");
    expect(await screen.findByText("Привет, Имя!")).toBeInTheDocument();
  });
});
