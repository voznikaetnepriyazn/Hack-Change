import { create } from "zustand";
import { User } from "../account/types/account.types";

interface Progress {
  day: string;
  step: string;
}

interface Metrics {
  requests: number;
  answersWithoutEscalation: number;
  escalation: number;
}

interface ThemesinQuestions {
  id: number;
  name: string;
}
interface InternState {
  title: string;
  description: string;
  user: User | null;
  progress: Progress;
  metrics: Metrics;
  themesinQuestions: ThemesinQuestions[];
  setInternUser: (user: User | null) => void;
}

export const useInternStore = create<InternState>((set) => ({
  title: "Мой стажёр",
  description: "Данные и статистска стажёра",
  user: null,
  progress: {
    day: " 14-й день",
    step: "2. Среда разработки настроена",
  },
  metrics: {
    requests: 20,
    answersWithoutEscalation: 8,
    escalation: 12,
  },
  themesinQuestions: [
    {
      id: 1,
      name: "Доступы",
    },
    {
      id: 2,
      name: "задачи",
    },
  ],

  setInternUser: (user: User | null) => set({ user }),
}));
