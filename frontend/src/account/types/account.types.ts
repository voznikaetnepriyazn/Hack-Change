import { AuthUser } from "../../auth/types/auth.types";

export interface User {
  id: number;
  name: string;
  avatar: string | null;
  position: string;
  department: string;
  phone?: string;
  login: string;
  vacationStart: Date;
  vacationEnd: Date;
}

export interface CurrentUserResponse {
  user: AuthUser;
}

export interface Question {
  id: number;
  title: string;
  description: string;
  answer: string;
}

export interface Document {
  id: number;
  title: string;
  description: string;
  link: string;
}

export interface Task {
  id: number;
  title: string;
  description: string;
  status: "active" | "completed";
  deadline: Date;
}

export interface Tag {
  id: number;
  name: string;
  tab: string;
  items: number[];
}

export interface ContentBlock {
  title: string;
  description: string;
}
