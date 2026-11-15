import type { User } from "../account/types/account.types";

export const internUser: User = {
  id: "1",
  name: "Ева Смирнова",
  avatar: null,
  position: "Стажёр",
  phone: "+7 (999) 999-99-99",
  login: "@test",
  command: "Команда 1",
  post: "Разработчик",
  formatWork: "Гибрид",
  office: "Москва",
  role: "intern",
  statusVacation: "Онбординг",
  githubProfile: {
    login: "e.smirnova",
    username: "eva-smirnova-dev",
  },
  requests: [
    { id: "1", name: "Отгул", description: "", status: "pending" },
    { id: "2", name: "Доступ", description: "ок", status: "approved" },
    { id: "3", name: "Ноутбук", description: "нет", status: "rejected" },
    {
      id: "4",
      name: "Другое",
      description: "?",
      status: "unknown" as "pending",
    },
  ],
  vacationStart: new Date("2026-11-01"),
  vacationEnd: new Date("2026-11-10"),
};

export const leadUser: User = {
  id: "3",
  name: "Дмитрий Волков",
  avatar: "https://example.com/avatar.jpg",
  position: "Руководитель",
  login: "@ivan",
  command: "Команда 1",
  post: "Наставник",
  formatWork: "Офис",
  office: "Москва",
  role: "lead",
};
