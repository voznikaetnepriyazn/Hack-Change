import type { CurrentUserResponse } from "../types/account.types";

export const isCurrentUserResponse = (
  data: unknown
): data is CurrentUserResponse => {
  if (!data || typeof data !== 'object') {
    return false;
  }

  if (
    !('user' in data) ||
    typeof data.user !== 'object' ||
    data.user === null
  ) {
    return false;
  }

  const user = data.user;

  return (
    'id' in user &&
    typeof user.id === 'string' &&
    'full_name' in user &&
    typeof user.full_name === 'string' &&
    'login' in user &&
    typeof user.login === 'string' &&
    'role' in user &&
    (user.role === 'intern' || user.role === 'lead' || user.role === 'manager')
  );
};
