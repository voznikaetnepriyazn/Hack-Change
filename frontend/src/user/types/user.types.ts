import { AuthUser } from "../../auth/types/auth.types";

export interface CurrentUserResponse {
  user: AuthUser;
}