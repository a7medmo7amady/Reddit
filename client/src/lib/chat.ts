import { fetchWithAuth } from "./auth";

export interface InboxMessage {
  id: string;
  senderId: string;
  content: string;
  createdAt: string;
}

export interface InboxItem {
  conversationId: string;
  type: "direct" | "community" | string;
  communityId?: string;
  lastMessage?: InboxMessage;
  unreadCount: number;
  updatedAt: string;
}

export interface Conversation {
  id: string;
  type: "direct" | "community" | string;
  communityId?: string;
  createdAt: string;
  updatedAt: string;
}

export interface ChatMessage {
  id: string;
  conversationId: string;
  senderId: string;
  content: string;
  type: "text" | string;
  createdAt: string;
  editedAt?: string;
  deletedAt?: string;
  deletedBy?: string;
}

async function parseJson<T>(response: Response): Promise<T> {
  const data = await response.json().catch(() => null);
  if (!response.ok) {
    const message =
      data && typeof data === "object" && "error" in data
        ? String(data.error)
        : "Chat request failed";
    throw new Error(message);
  }
  return data as T;
}

export async function getInbox(): Promise<InboxItem[]> {
  const response = await fetchWithAuth("/chat/inbox");
  return parseJson<InboxItem[]>(response);
}

export async function createDirectConversation(
  participantId: string
): Promise<Conversation> {
  const response = await fetchWithAuth("/chat/conversations", {
    method: "POST",
    body: JSON.stringify({ participantIds: [participantId] }),
  });
  return parseJson<Conversation>(response);
}

export async function getConversationMessages(
  conversationId: string
): Promise<ChatMessage[]> {
  const response = await fetchWithAuth(
    `/chat/conversations/${encodeURIComponent(conversationId)}/messages`
  );
  return parseJson<ChatMessage[]>(response);
}

export async function sendChatMessage(
  conversationId: string,
  content: string
): Promise<ChatMessage> {
  const response = await fetchWithAuth("/chat/messages", {
    method: "POST",
    body: JSON.stringify({ conversationId, content, type: "text" }),
  });
  return parseJson<ChatMessage>(response);
}

export async function markConversationRead(conversationId: string): Promise<void> {
  const response = await fetchWithAuth(
    `/chat/conversations/${encodeURIComponent(conversationId)}/read`,
    { method: "POST" }
  );
  if (!response.ok) {
    throw new Error("Failed to mark conversation read");
  }
}
