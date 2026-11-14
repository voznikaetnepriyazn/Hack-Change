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
    typeof user.id === 'number' &&
    'login' in user &&
    typeof user.login === 'string' &&
    'createdAt' in user &&
    typeof user.createdAt === 'string' &&
    'updatedAt' in user &&
    typeof user.updatedAt === 'string'
  );
};
