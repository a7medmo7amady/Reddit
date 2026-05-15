import { getToken } from "./auth";

interface TokenPayload {
  sub: string;
  username: string;
  role: string;
  iat: number;
  exp: number;
}

export function decodeToken(): TokenPayload | null {
  const token = getToken();
  if (!token) return null;
  try {
    const payload = token.split(".")[1];
    const json = atob(payload.replace(/-/g, "+").replace(/_/g, "/"));
    return JSON.parse(json) as TokenPayload;
  } catch {
    return null;
  }
}

export function getMyUsername(): string | null {
  return decodeToken()?.username ?? null;
}

export function getMyUserId(): string | null {
  return decodeToken()?.sub ?? null;
}
