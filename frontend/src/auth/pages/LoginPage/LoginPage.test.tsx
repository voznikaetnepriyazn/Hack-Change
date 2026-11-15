import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { LoginPage } from "./index";
import { AuthContext } from "../../context/AuthContext";

function renderLogin(login = vi.fn()) {
  render(
    <MemoryRouter initialEntries={["/auth/login"]}>
      <AuthContext.Provider
        value={{ isAuthLoading: false, login, logout: async () => {} }}
      >
        <Routes>
          <Route path="/auth/login" element={<LoginPage />} />
          <Route path="/account" element={<p>кабинет</p>} />
        </Routes>
      </AuthContext.Provider>
    </MemoryRouter>,
  );
  return login;
}

async function submit() {
  const form = screen.getByRole("button", { name: "Продолжить" }).closest("form");
  fireEvent.submit(form as HTMLFormElement);
}

describe("LoginPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("rejects an empty login", async () => {
    renderLogin();
    await submit();

    expect(
      screen.getByText("Логин должен начинаться с @ и быть от 3 до 100 символов"),
    ).toBeInTheDocument();
  });

  it("prefixes the login with @ and rejects a short password", async () => {
    renderLogin();
    fireEvent.focus(screen.getByPlaceholderText("@username"));
    expect(screen.getByPlaceholderText("@username")).toHaveValue("@");

    fireEvent.change(screen.getByPlaceholderText("@username"), {
      target: { value: "eva" },
    });
    expect(screen.getByPlaceholderText("@username")).toHaveValue("@eva");

    fireEvent.change(screen.getByPlaceholderText("********"), {
      target: { value: "short" },
    });
    await submit();

    expect(screen.getByText("Пароль должен быть от 8 до 100 символов")).toBeInTheDocument();
  });

  it("rejects a password longer than 100 characters and an empty role", async () => {
    renderLogin();
    fireEvent.change(screen.getByPlaceholderText("@username"), {
      target: { value: "@e.smirnova" },
    });
    fireEvent.change(screen.getByPlaceholderText("********"), {
      target: { value: "p".repeat(101) },
    });
    await submit();
    expect(screen.getByText("Пароль должен быть от 8 до 100 символов")).toBeInTheDocument();

    fireEvent.change(screen.getByPlaceholderText("********"), {
      target: { value: "Password123!" },
    });
    await submit();
    expect(screen.getByText("Должность не может быть пустой")).toBeInTheDocument();
  });

  it("rejects a login that is too long", async () => {
    renderLogin();
    fireEvent.change(screen.getByPlaceholderText("@username"), {
      target: { value: `@${"a".repeat(100)}` },
    });
    await submit();

    expect(
      screen.getByText("Логин должен начинаться с @ и быть от 3 до 100 символов"),
    ).toBeInTheDocument();
  });

  it("opens the account after a successful login", async () => {
    const login = renderLogin(vi.fn().mockResolvedValue(undefined));
    fireEvent.focus(screen.getByPlaceholderText("@username"));
    fireEvent.change(screen.getByPlaceholderText("@username"), {
      target: { value: "@e.smirnova" },
    });
    fireEvent.change(screen.getByPlaceholderText("********"), {
      target: { value: "Password123!" },
    });
    fireEvent.change(screen.getByRole("combobox"), { target: { value: "intern" } });
    await submit();

    expect(login).toHaveBeenCalledWith("@e.smirnova", "Password123!");
    expect(await screen.findByText("кабинет")).toBeInTheDocument();
  });

  it("shows the loader and then the server error", async () => {
    let rejectLogin: (error: Error) => void = () => {};
    const login = vi.fn().mockImplementation(
      () =>
        new Promise<void>((_, reject) => {
          rejectLogin = reject;
        }),
    );
    renderLogin(login);
    fireEvent.change(screen.getByPlaceholderText("@username"), {
      target: { value: "@e.smirnova" },
    });
    fireEvent.change(screen.getByPlaceholderText("********"), {
      target: { value: "Password123!" },
    });
    fireEvent.change(screen.getByRole("combobox"), { target: { value: "lead" } });
    await submit();

    expect(screen.getByText("Загрузка...")).toBeInTheDocument();
    rejectLogin(new Error("no"));
    expect(await screen.findByText("Введены неверные данные")).toBeInTheDocument();
  });
});
