import { buildApiUrl } from "./config";

const TOKEN_COOKIE = "access_token";
const TOKEN_TTL_MS = 15 * 60 * 1000; // 15 minutes — matches JWT expiry

export function saveToken(token: string): void {
  const expires = new Date(Date.now() + TOKEN_TTL_MS);
  document.cookie = [
    `${TOKEN_COOKIE}=${encodeURIComponent(token)}`,
    `expires=${expires.toUTCString()}`,
    "path=/",
    "SameSite=Strict",
  ].join("; ");
}

export function getToken(): string | null {
  if (typeof document === "undefined") return null;
  const match = document.cookie.match(
    new RegExp(`(?:^|;\\s*)${TOKEN_COOKIE}=([^;]*)`)
  );
  return match ? decodeURIComponent(match[1]) : null;
}

export function clearToken(): void {
  document.cookie = `${TOKEN_COOKIE}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/; SameSite=Strict`;
}

export async function fetchWithAuth(
  path: string,
  options: RequestInit = {}
): Promise<Response> {
  const token = getToken();
  const headers = new Headers(options.headers);
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }
  headers.set("Content-Type", "application/json");

  return fetch(buildApiUrl(path), {
    ...options,
    credentials: "include",
    headers,
  });
}

export async function logout(): Promise<void> {
  await fetch(buildApiUrl("/auth/logout"), {
    method: "POST",
    credentials: "include",
  }).catch(() => {});
  clearToken();
}
