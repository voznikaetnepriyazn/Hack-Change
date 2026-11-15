import { API_URL } from "./config";
import ApiError from "../shared/errors/ApiError";

function getErrorMessage(data: unknown): string | null {
  if (typeof data !== "object" || data === null) {
    return null;
  }

  if (!("error" in data) || typeof data.error !== "string") {
    return null;
  }

  return data.error;
}

export async function apiRequest(
  path: string,
  options?: RequestInit,
): Promise<unknown> {
  const url = `${API_URL.replace(/\/$/, "")}/${path.replace(/^\//, "")}`;
  const credentials = options?.credentials ?? "include";

  const headers = new Headers(options?.headers);

  const response = await fetch(url, { ...options, headers, credentials });
  if (response.status === 204) {
    return null;
  }

  let data: unknown = null;

  try {
    data = await response.json();
  } catch {
    throw new Error(`API returned an invalid JSON or empty data response`);
  }

  if (!response.ok) {
    const errorMessage = getErrorMessage(data);

    throw new ApiError(
      errorMessage ?? `API responded with ${response.status}`,
      response.status,
    );
  }

  if (data === null) {
    throw new Error("API returned an empty or invalid JSON response");
  }

  return data;
}
