export type UserRole = "intern" | "lead" | "manager";

export interface AuthUser {
  id: string;
  full_name: string;
  login: string;
  role: UserRole;
}

export interface LoginResponse {
  user: AuthUser;
}