import { apiRequest } from '../../api/client';
import type { LoginResponse, RegisterResponse } from '../types/auth.types';
import { validateLoginData, validateRegisterData } from '../validation/auth.validation';

export const isRegisterResponse = (
  data: unknown
): data is RegisterResponse => {
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
    typeof user.id === 'number' &&
    'email' in user &&
    typeof user.email === 'string' &&
    'createdAt' in user &&
    typeof user.createdAt === 'string' &&
    'updatedAt' in user &&
    typeof user.updatedAt === 'string'
  );
};
export const isLoginResponse = (
  data: unknown
): data is LoginResponse => {
  if (!data || typeof data !== 'object') {
    return false;
  }

  if (!('message' in data) || typeof data.message !== 'string') {
    return false;
  }

  if (!('accessToken' in data) || typeof data.accessToken !== 'string') {
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
    'email' in user &&
    typeof user.email === 'string' &&
    'createdAt' in user &&
    typeof user.createdAt === 'string' &&
    'updatedAt' in user &&
    typeof user.updatedAt === 'string'
  );
};

export const registerUser = async (email: string, password: string): Promise<RegisterResponse> => {

  const data: unknown = await apiRequest('api/auth/register', {
    method: 'POST',
    body: JSON.stringify(validateRegisterData(email, password)),
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (!isRegisterResponse(data)) {
    throw new Error('Invalid response format');
  }

  return data;
}

export const loginUser = async (email: string, password: string): Promise<LoginResponse> => {
  const data: unknown = await apiRequest('api/auth/login', {
    method: 'POST',
    body: JSON.stringify(validateLoginData(email, password)),
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (!isLoginResponse(data)) {
    throw new Error('Invalid response format');
  }

  return data;
}

let refreshPromise: Promise<LoginResponse> | null = null;

export const refreshUserToken = async (): Promise<LoginResponse> => {

  if (refreshPromise) {
    return refreshPromise;
  }

  refreshPromise = (async () => {
    try {
      const data: unknown = await apiRequest('api/auth/refresh-token', {
        method: 'POST'
      });
  
      if (!isLoginResponse(data)) {
        throw new Error('Invalid response format');
      }
  
      return data;
      
    } finally {
      refreshPromise = null;
    }
  })();

  return refreshPromise;
}

export const logoutUser = async () => {
  const data: unknown = await apiRequest('api/auth/logout', {
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