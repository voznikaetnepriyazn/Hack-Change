import { create } from "zustand";
import { ContentBlock, Document, Tag } from "../account/types/account.types";

interface DocumentsState {
  documents: Document[];
  tags: Tag[];
  contentBlock: ContentBlock;
  title: string;
  description: string;
}

export const useDocumentsStore = create<DocumentsState>((set) => ({
  title: "Документы",
  description: "Инструкции, материалы и полезные ссылки для стажёров.",
  tags: [
    {
      id: 0,
      name: "Все документы",
      tab: "all",
      items: [],
    },
    {
      id: 1,
      name: "Инструкции",
      tab: "instructions",
      items: [1, 2],
    },
    {
      id: 2,
      name: "Обучение",
      tab: "training",
      items: [4, 5],
    },
  ],
  documents: [
    {
      id: 1,
      title: "Памятка стажёра",
      description: "Основная информация о начале стажировки",
      link: "https://www.google.com",
    },
    {
      id: 2,
      title: "Рабочие инструменты и доступы",
      description: "Инструкция по подключению сервисов",
      link: "https://www.google.com",
    },
    {
      id: 3,
      title: "Правила коммуникации",
      description: "Как устроено взаимодействие в команде",
      link: "https://www.google.com",
    },
    {
      id: 4,
      title: "Полезные обучающие материалы",
      description: "Подборка ссылок и внутренних ресурсов",
      link: "https://www.google.com",
    },
    {
      id: 5,
      title: "Документ 5",
      description: "Описание документа 5",
      link: "https://www.google.com",
    },
  ],
  contentBlock: {
    title: "Нужна помощь?",
    description:
      "Попросите AI Buddy кратко пересказать документ или объяснить непонятный термин.",
  },
}));
