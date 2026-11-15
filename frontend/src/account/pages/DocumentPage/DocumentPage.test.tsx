import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { DocumentPage } from "./index";
import { useDocumentsStore } from "../../../stores/documentsStore";

vi.mock("../../api/document.api", () => ({
  getDocumentContent: vi.fn(),
}));

import { getDocumentContent } from "../../api/document.api";

const loadDocument = vi.mocked(getDocumentContent);

function renderDocument(documentId: string) {
  return render(
    <MemoryRouter initialEntries={[`/account/documents/${documentId}`]}>
      <Routes>
        <Route path="/account/documents/:documentId" element={<DocumentPage />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("DocumentPage", () => {
  beforeEach(() => {
    loadDocument.mockReset();
  });

  it("shows the loaded markdown", async () => {
    const document = useDocumentsStore.getState().documents[0];
    loadDocument.mockResolvedValue("# Правила\n\nТекст документа");

    renderDocument(String(document.id));

    expect(await screen.findByRole("heading", { name: "Правила" })).toBeInTheDocument();
    expect(screen.getByText("Текст документа")).toBeInTheDocument();
    expect(loadDocument).toHaveBeenCalledWith(document.link);
  });

  it("reports an unknown document", async () => {
    renderDocument("999");

    expect(await screen.findByText("Error: Документ не найден")).toBeInTheDocument();
    expect(loadDocument).not.toHaveBeenCalled();
  });

  it("reports an invalid document id", async () => {
    renderDocument("0");

    expect(await screen.findByText("Error: Некорректный ID документа")).toBeInTheDocument();
    expect(loadDocument).not.toHaveBeenCalled();
  });

  it("shows the loader error from a failed request", async () => {
    const document = useDocumentsStore.getState().documents[0];
    loadDocument.mockRejectedValue(new Error("Сеть недоступна"));

    renderDocument(String(document.id));

    expect(await screen.findByText("Error: Сеть недоступна")).toBeInTheDocument();
  });

  it("uses a generic message when the rejection is not an error", async () => {
    const document = useDocumentsStore.getState().documents[0];
    loadDocument.mockRejectedValue("boom");

    renderDocument(String(document.id));

    expect(await screen.findByText("Error: Неизвестная ошибка")).toBeInTheDocument();
  });
});
