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
  otherParticipantIds?: string[];
  lastMessage?: InboxMessage;
  unreadCount: number;
  muted: boolean;
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

export interface ChatUser {
  id: number;
  username: string;
  displayName?: string | null;
  avatar?: string | null;
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

export async function getUserById(userId: string): Promise<ChatUser> {
  const response = await fetchWithAuth(`/users/id/${encodeURIComponent(userId)}`);
  return parseJson<ChatUser>(response);
}

export async function getUserByUsername(username: string): Promise<ChatUser> {
  const response = await fetchWithAuth(`/users/${encodeURIComponent(username)}`);
  return parseJson<ChatUser>(response);
}

export async function getInbox(): Promise<InboxItem[]> {
  const response = await fetchWithAuth("/chat/inbox");
  const inbox = await parseJson<InboxItem[] | null>(response);
  return Array.isArray(inbox) ? inbox : [];
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

export async function getOrCreateCommunityRoom(
  communityId: string
): Promise<Conversation> {
  const response = await fetchWithAuth(
    `/chat/communities/${encodeURIComponent(communityId)}/room`
  );
  return parseJson<Conversation>(response);
}

export async function getConversationMessages(
  conversationId: string
): Promise<ChatMessage[]> {
  const response = await fetchWithAuth(
    `/chat/conversations/${encodeURIComponent(conversationId)}/messages`
  );
  const messages = await parseJson<ChatMessage[] | null>(response);
  return Array.isArray(messages) ? messages : [];
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

export async function hideConversation(conversationId: string): Promise<void> {
  const response = await fetchWithAuth(
    `/chat/conversations/${encodeURIComponent(conversationId)}/hide`,
    { method: "POST" }
  );
  if (!response.ok) {
    throw new Error("Failed to delete chat");
  }
}

export async function setConversationMuted(
  conversationId: string,
  muted: boolean
): Promise<void> {
  const response = await fetchWithAuth(
    `/chat/conversations/${encodeURIComponent(conversationId)}/muted`,
    {
      method: "POST",
      body: JSON.stringify({ muted }),
    }
  );
  if (!response.ok) {
    throw new Error(`Failed to ${muted ? "mute" : "unmute"} chat`);
  }
}
