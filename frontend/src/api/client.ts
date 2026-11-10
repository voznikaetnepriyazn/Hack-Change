import { type DatabaseStatus, isDatabaseStatus } from './databaseStatus';
import { API_URL } from './config';
import ApiError from '../shared/errors/ApiError';

export type HealthResponse = {
  ok: boolean;
  service: string;
  database: DatabaseStatus;
}

export type RefreshAccessToken = () => Promise<string>;

const isHealthResponse = (data: unknown): data is HealthResponse => {
  return (
    typeof data === 'object' &&
    data !== null &&
    'ok' in data &&
    typeof data.ok === 'boolean' &&
    'service' in data &&
    typeof data.service === 'string' &&
    'database' in data &&
    isDatabaseStatus(data.database)
  );
}

function getErrorMessage(data: unknown): string | null {

  if (typeof data !== 'object' || data === null) {
    return null;
  }

  if (!('message' in data) || typeof data.message !== 'string') {
    return null;
  }

  return data.message;
}

export async function apiRequest(path: string, options?: RequestInit, accessToken?: string, refreshAccessToken?: RefreshAccessToken): Promise<unknown> {
  const url = `${API_URL.replace(/\/$/, '')}/${path.replace(/^\//, '')}`;
  const credentials = options?.credentials ?? 'include';

  const headers = new Headers(options?.headers);

  if (accessToken) {
    headers.set('Authorization', `Bearer ${accessToken}`);
  }

  const response = await fetch(url, { ...options, headers, credentials });
  let data: unknown = null;

  try {
    data = await response.json();
  } catch {
    throw new Error(`API returned an invalid JSON or empty data response`);
  }

  if (!response.ok) {
    const errorMessage = getErrorMessage(data);
    
    if (response.status === 401 && refreshAccessToken && accessToken) {
      const newAccessToken = await refreshAccessToken();
      return apiRequest(path, options, newAccessToken);
    }

    throw new ApiError(
      errorMessage ?? `API responded with ${response.status}`,
      response.status
    );
  }

  if (data === null) {
    throw new Error('API returned an empty or invalid JSON response');
  }

  return data;
}

export async function getHealth(): Promise<HealthResponse> {

  const data: unknown = await apiRequest('api/health');

  if (!isHealthResponse(data)) {
    throw new Error('Invalid health response');
  }

  return data;
}