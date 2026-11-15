import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { ProtectedRoute } from "./ProtectedRoute";
import { AuthContext } from "../context/AuthContext";
import { useProfileStore } from "../../stores/profileStore";
import { internUser } from "../../test/fixtures";

const auth = {
  isAuthLoading: false,
  login: async () => {},
  logout: async () => {},
};

function renderRoute(isAuthLoading: boolean) {
  return render(
    <MemoryRouter initialEntries={["/account"]}>
      <AuthContext.Provider value={{ ...auth, isAuthLoading }}>
        <Routes>
          <Route
            path="/account"
            element={
              <ProtectedRoute>
                <p>секрет</p>
              </ProtectedRoute>
            }
          />
          <Route path="/auth/login" element={<p>страница входа</p>} />
        </Routes>
      </AuthContext.Provider>
    </MemoryRouter>,
  );
}

describe("ProtectedRoute", () => {
  it("shows the loader while auth is restoring", () => {
    useProfileStore.setState({ user: null });
    renderRoute(true);

    expect(screen.getByText("Загрузка...")).toBeInTheDocument();
  });

  it("sends an anonymous visitor to login", () => {
    useProfileStore.setState({ user: null });
    renderRoute(false);

    expect(screen.getByText("страница входа")).toBeInTheDocument();
  });

  it("renders the page for a signed-in user", () => {
    useProfileStore.setState({ user: internUser });
    renderRoute(false);

    expect(screen.getByText("секрет")).toBeInTheDocument();
  });
});
