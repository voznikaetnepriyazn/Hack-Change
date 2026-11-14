import { create } from "zustand";
import { User } from "../account/types/account.types";

interface Description {
  intern: string;
  lead: string;
}

interface ProfileState {
  title: string;
  description: Description;
  user: User | null;
  setProfileUser: (user: User | null) => void;
}

export const useProfileStore = create<ProfileState>((set) => ({
  title: "Аккаунт",
  description: {
    intern: "Личная информация и настройки профиля стажёра.",
    lead: "Личная информация руководителя.",
  },
  user: null,

  setProfileUser: (user: User | null) => set({ user }),
}));
