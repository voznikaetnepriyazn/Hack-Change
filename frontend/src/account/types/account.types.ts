import { AuthUser, UserRole } from "../../auth/types/auth.types";

interface GitHubProfile {
  login: string;
  username: string;
}

interface Requests {
  id: string;
  name: string;
  description: string;
  status: "pending" | "approved" | "rejected";
}

export interface User {
  id: string;
  name: string;
  avatar: string | null;
  position: string;
  phone?: string;
  login: string;
  command: string;
  post: string;
  formatWork: string;
  office: string;
  statusVacation?: string;
  githubProfile?: GitHubProfile | null;
  requests?: Requests[];
  role: UserRole;
  vacationStart?: Date | null;
  vacationEnd?: Date | null;
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
