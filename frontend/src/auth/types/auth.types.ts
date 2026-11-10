export interface AuthUser {
  id: number;
  email: string;
  createdAt: string;
  updatedAt: string;
}

export interface RegisterResponse {
  message: string;
  user: AuthUser;
}

export interface LoginResponse {
  message: string;
  user: AuthUser;
  accessToken: string;
}