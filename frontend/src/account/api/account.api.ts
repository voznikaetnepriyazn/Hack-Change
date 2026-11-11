import { apiRequest } from "../../api/client";
import type { RefreshAccessToken } from "../../api/client";
import { isCurrentUserResponse } from "../validation/user.validation";
import ApiError from "../../shared/errors/ApiError";



export const getCurrentUser = async (accessToken: string, refreshAccessToken?: RefreshAccessToken) => {
  const data: unknown = await apiRequest('api/users/me', {
    method: 'GET'
  }, accessToken, refreshAccessToken);

  if (!isCurrentUserResponse(data)) {
    throw new ApiError('Invalid response format', 400);
  }

  return data.user;
}
