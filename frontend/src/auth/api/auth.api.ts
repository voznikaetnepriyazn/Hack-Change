import { apiRequest } from '../../api/client';
import type { LoginResponse } from '../types/auth.types';
import { validateLoginData } from '../validation/auth.validation';

export const isLoginResponse = (
  data: unknown
): data is LoginResponse => {
  if (!data || typeof data !== 'object') {
    return false;
  }

  if (!('message' in data) || typeof data.message !== 'string') {
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
    'login' in user &&
    typeof user.login === 'string' &&
    'role' in user &&
    typeof user.role === 'string'
  );
};

export const loginUser = async (login: string, password: string): Promise<LoginResponse> => {
  const data: unknown = await apiRequest('auth/login', {
    method: 'POST',
    body: JSON.stringify(validateLoginData(login, password)),
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (!isLoginResponse(data)) {
    throw new Error('Invalid response format');
  }

  return data;
}

export const logoutUser = async () => {
  const data: unknown = await apiRequest('auth/logout', {
    method: 'POST'
  });

  if (!data || typeof data !== 'object') {
    throw new Error('Invalid response format');
  }

  if (!('message' in data) || typeof data.message !== 'string') {
    throw new Error('Invalid response format');
  }

  return {
    message: data.message,
  };
}