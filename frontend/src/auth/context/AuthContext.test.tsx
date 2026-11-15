import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { AuthProvider, useAuth } from "./AuthContext";
import type { AuthContextType } from "./AuthContext";
import { useProfileStore } from "../../stores/profileStore";
import { internUser, leadUser } from "../../test/fixtures";

vi.mock("../api/auth.api", () => ({
  loginUser: vi.fn(),
  logoutUser: vi.fn(),
}));

vi.mock("../../account/api/account.api", () => ({
  getCurrentUser: vi.fn(),
}));

import { loginUser, logoutUser } from "../api/auth.api";
import { getCurrentUser } from "../../account/api/account.api";

const loginRequest = vi.mocked(loginUser);
const logoutRequest = vi.mocked(logoutUser);
const currentUser = vi.mocked(getCurrentUser);
type Profile = Awaited<ReturnType<typeof getCurrentUser>>;

function SessionProbe() {
  const auth = useAuth();
  return (
    <button type="button" onClick={() => void auth.logout()}>
      {auth.isAuthLoading ? "loading" : "ready"}
    </button>
  );
}

describe("AuthProvider", () => {
  beforeEach(() => {
    loginRequest.mockReset();
    logoutRequest.mockReset();
    currentUser.mockReset();
    useProfileStore.setState({ user: null });
    sessionStorage.setItem("chat-session", JSON.stringify([{ text: "старый", role: "user" }]));
  });

  it("restores the profile and keeps the saved chat", async () => {
    useProfileStore.setState({ user: internUser });
    currentUser.mockResolvedValue(internUser as Profile);
    render(
      <AuthProvider>
        <SessionProbe />
      </AuthProvider>,
    );

    expect(screen.getByRole("button")).toHaveTextContent("loading");
    await waitFor(() => expect(screen.getByRole("button")).toHaveTextContent("ready"));
    expect(currentUser).toHaveBeenCalledWith();
    expect(useProfileStore.getState().user?.id).toBe(internUser.id);
    expect(sessionStorage.getItem("chat-session")).toContain("старый");
  });

  it("does not request a profile or clear the chat when nobody is signed in", async () => {
    render(
      <AuthProvider>
        <SessionProbe />
      </AuthProvider>,
    );

    await waitFor(() => expect(screen.getByRole("button")).toHaveTextContent("ready"));
    expect(currentUser).not.toHaveBeenCalled();
    expect(useProfileStore.getState().user).toBeNull();
    expect(sessionStorage.getItem("chat-session")).toContain("старый");
  });

  it("clears the chat when the saved profile cannot be restored", async () => {
    useProfileStore.setState({ user: internUser });
    currentUser.mockRejectedValue(new Error("unauthorized"));
    render(
      <AuthProvider>
        <SessionProbe />
      </AuthProvider>,
    );

    await waitFor(() => expect(screen.getByRole("button")).toHaveTextContent("ready"));
    expect(currentUser).toHaveBeenCalledOnce();
    expect(useProfileStore.getState().user).toBeNull();
    expect(sessionStorage.getItem("chat-session")).toBeNull();
  });

  it("starts a new session on login and clears it on logout", async () => {
    currentUser.mockResolvedValue(internUser as Profile);
    let auth: AuthContextType | null = null;

    function Grabber() {
      auth = useAuth();
      return <p>{auth.isAuthLoading ? "loading" : "ready"}</p>;
    }

    render(
      <AuthProvider>
        <Grabber />
      </AuthProvider>,
    );
    await screen.findByText("ready");

    loginRequest.mockResolvedValue({
      user: {
        id: leadUser.id,
        full_name: leadUser.full_name,
        login: leadUser.login,
        role: "lead",
      },
    });
    currentUser.mockResolvedValue(leadUser as Profile);
    await auth!.login("@ivan", "MentorPass123!");

    expect(loginRequest).toHaveBeenCalledWith("@ivan", "MentorPass123!");
    expect(currentUser).toHaveBeenCalledWith("lead");
    expect(useProfileStore.getState().user?.role).toBe("lead");
    expect(sessionStorage.getItem("chat-session")).toBeNull();

    logoutRequest.mockResolvedValue(undefined);
    sessionStorage.setItem("chat-session", JSON.stringify([{ text: "ещё", role: "user" }]));
    await auth!.logout();

    expect(logoutRequest).toHaveBeenCalledWith();
    expect(useProfileStore.getState().user).toBeNull();
    expect(sessionStorage.getItem("chat-session")).toBeNull();
  });

  it("does not replace the profile when login returns no user", async () => {
    useProfileStore.setState({ user: internUser });
    currentUser.mockResolvedValue(internUser as Profile);
    let auth: AuthContextType | null = null;

    function Grabber() {
      auth = useAuth();
      return <p>ready-marker</p>;
    }

    render(
      <AuthProvider>
        <Grabber />
      </AuthProvider>,
    );
    await waitFor(() => expect(useProfileStore.getState().user).not.toBeNull());

    loginRequest.mockResolvedValue(
      undefined as unknown as Awaited<ReturnType<typeof loginUser>>,
    );
    await auth!.login("e.smirnova", "Password123!");
    expect(useProfileStore.getState().user?.role).toBe("intern");

    currentUser.mockResolvedValue(null as unknown as Profile);
    loginRequest.mockResolvedValue({
      user: {
        id: "1",
        full_name: "Ева Смирнова",
        login: "e.smirnova",
        role: "intern",
      },
    });
    await auth!.login("e.smirnova", "Password123!");
    expect(useProfileStore.getState().user?.role).toBe("intern");
  });

  it("rejects a failed login and a failed logout", async () => {
    currentUser.mockResolvedValue(internUser as Profile);
    let auth: AuthContextType | null = null;

    function Grabber() {
      auth = useAuth();
      return null;
    }

    render(
      <AuthProvider>
        <Grabber />
      </AuthProvider>,
    );
    await waitFor(() => expect(auth).not.toBeNull());

    loginRequest.mockRejectedValue(new Error("no"));
    await expect(auth!.login("e.smirnova", "Password123!")).rejects.toThrow("Failed to login");
    expect(sessionStorage.getItem("chat-session")).toContain("старый");

    logoutRequest.mockRejectedValue(new Error("no"));
    await expect(auth!.logout()).rejects.toThrow("no");
    expect(sessionStorage.getItem("chat-session")).toContain("старый");
  });
});

describe("useAuth", () => {
  it("throws outside of the provider", () => {
    function Orphan() {
      useAuth();
      return null;
    }

    expect(() => render(<Orphan />)).toThrow("useAuth must be used within an AuthProvider");
  });
});
