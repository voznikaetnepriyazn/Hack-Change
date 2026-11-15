export type UserRole = "intern" | "lead" ;

export interface AuthUser {
  id: string;
  login: string;
  password: string;
  role: UserRole;
}

export interface LoginResponse {
  message: string;
  user: AuthUser;
}