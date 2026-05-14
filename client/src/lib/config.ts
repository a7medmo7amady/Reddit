function trimTrailingSlash(value: string): string {
  return value.replace(/\/+$/, "");
}

export function getApiBaseUrl(): string {
  return trimTrailingSlash(process.env.NEXT_PUBLIC_API_GATEWAY_URL ?? "");
}

export function buildApiUrl(path: string): string {
  const baseUrl = getApiBaseUrl();
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  return `${baseUrl}${normalizedPath}`;
}

export function getChatWebSocketUrl(token: string): string {
  const configuredUrl = process.env.NEXT_PUBLIC_CHAT_WS_URL;
  const url = configuredUrl
    ? new URL(configuredUrl)
    : new URL(
        buildApiUrl("/chat/ws"),
        typeof window === "undefined" ? undefined : window.location.origin
      );

  if (url.protocol === "http:") url.protocol = "ws:";
  if (url.protocol === "https:") url.protocol = "wss:";
  url.searchParams.set("access_token", token);
  return url.toString();
}
