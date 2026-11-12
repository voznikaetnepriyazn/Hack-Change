import { create } from "zustand";
import { ContentBlock, Question, Tag } from "../account/types/account.types";


interface QuestionsState {
  questions: Question[];
  tags: Tag[];
  contentBlock: ContentBlock;
  title: string;
  description: string;
}


export const useQuestionsStore = create<QuestionsState>((set) => ({
  title: "Частые вопросы",
  description: "Быстрые ответы на популярные вопросы о стажировке и работе в Авито.",
  questions: [
    {
      id: 1,
      title: "С чего начать стажировку?",
      description: "Первые шаги, знакомство с командой и план адаптации",
      answer:
        "ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ ответ",
    },
    {
      id: 2,
      title: "Вопрос 2",
      description: "Описание вопроса 2",
      answer: "Ответ на вопрос 2",
    },
    {
      id: 3,
      title: "Вопрос 3",
      description: "Описание вопроса 3",
      answer: "Ответ на вопрос 3",
    },
    {
      id: 4,
      title: "Вопрос 4",
      description: "Описание вопроса 4",
      answer: "Ответ на вопрос 4",
    },
    {
      id: 5,
      title: "Вопрос 5",
      description: "Описание вопроса 5",
      answer: "Ответ на вопрос 5",
    },
  ],
  tags: [
    {
      id: 1,
      name: "Стажировка",
      items: [1, 2],
    },
    {
      id: 2,
      name: "Доступы",
      items: [4, 5],
    },
    {
      id: 3,
      name: "Офис и быт",
      items: [3, 6],
    },
  ],
  contentBlock: {
    title: "Не нашли ответ?",
    description: "Спросите AI Buddy — он поможет разобраться и подскажет, где найти информацию.",
  },
}));
