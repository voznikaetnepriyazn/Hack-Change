import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it } from "vitest";
import { AccountIndexRedirect } from "./AccountIndexRedirect";
import { useProfileStore } from "../../stores/profileStore";
import { internUser, leadUser } from "../../test/fixtures";

function renderRedirect() {
  return render(
    <MemoryRouter initialEntries={["/account"]}>
      <Routes>
        <Route path="/account" element={<AccountIndexRedirect />} />
        <Route path="/account/chat" element={<p>чат стажёра</p>} />
        <Route path="/account/intern" element={<p>страница стажёра</p>} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("AccountIndexRedirect", () => {
  beforeEach(() => {
    useProfileStore.setState({ user: null });
  });

  it("waits until the profile is known", () => {
    renderRedirect();

    expect(screen.getByText("Загрузка...")).toBeInTheDocument();
  });

  it("sends an intern to the chat", () => {
    useProfileStore.setState({ user: internUser });
    renderRedirect();

    expect(screen.getByText("чат стажёра")).toBeInTheDocument();
  });

  it("sends a lead to the intern page", () => {
    useProfileStore.setState({ user: leadUser });
    renderRedirect();

    expect(screen.getByText("страница стажёра")).toBeInTheDocument();
  });
});