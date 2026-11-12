import { AuthUser } from "../../auth/types/auth.types";

export interface CurrentUserResponse {
  user: AuthUser;
}

export interface Question {
  id: number;
  title: string;
  description: string;
  answer: string;
}

export interface Tag {
  id: number;
  name: string;
  items: number[];
}

export interface ContentBlock {
  title: string;
  description: string;
}