import { create } from "zustand";
import { User } from "../account/types/account.types";


interface ProfileState {
  title: string;
  description: string;
  user: User;
}

const intern = {
  id: 1,
  name: "Имя Фамилия",
  avatar: null,
  position: "Стажёр",
  department: "Отдел разработки",
  phone: "+7 (999) 999-99-99",
  login: "@test",
  role: "intern" as const,
  vacationStart: new Date(),
  vacationEnd: new Date(new Date().setDate(new Date().getDate() + 10)),
};
const intern1 = {
  id: 2,
  avatar: "https://thumb.wikimedia.org/wikipedia/ru/thumb/0/0f/Tom_Holland_as_Spider-Man.jpg/330px-Tom_Holland_as_Spider-Man.jpg?utm_source=ru.wikipedia.org&utm_campaign=parser&utm_content=thumbnail",
  name: "Иван Иванов",
  position: "Стажёр",
  department: "Отдел разработки",
  phone: "+7 (999) 999-99-98",
  login: "@ivan",
  role: "intern" as const,
  vacationStart: new Date(new Date().setDate(new Date().getDate() - 4)),
  vacationEnd: new Date(new Date().setDate(new Date().getDate() + 10)),
};
const leader = {
  id: 3,
  avatar: "https://thumb.wikimedia.org/wikipedia/ru/thumb/0/0f/Tom_Holland_as_Spider-Man.jpg/330px-Tom_Holland_as_Spider-Man.jpg?utm_source=ru.wikipedia.org&utm_campaign=parser&utm_content=thumbnail",
  name: "Иван Иванов",
  position: "Руководитель",
  department: "Отдел разработки",
  phone: "+7 (999) 999-99-98",
  login: "@ivan",
  role: "leader" as const,
  vacationStart: new Date(new Date().setDate(new Date().getDate() - 4)),
  vacationEnd: new Date(new Date().setDate(new Date().getDate() + 10)),
};


export const useProfileStore = create<ProfileState>((set) => ({
  title: "Аккаунт",
  description: "Личная информация и настройки профиля стажёра.",
  user: intern,
}));
