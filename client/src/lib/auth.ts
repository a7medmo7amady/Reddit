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

const API_URL = process.env.NEXT_PUBLIC_API_GATEWAY_URL ?? "http://localhost:8088";

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

  return fetch(`${API_URL}${path}`, {
    ...options,
    credentials: "include",
    headers,
  });
}

export async function logout(): Promise<void> {
  await fetch(`${API_URL}/auth/logout`, {
    method: "POST",
    credentials: "include",
  }).catch(() => {});
  clearToken();
}
