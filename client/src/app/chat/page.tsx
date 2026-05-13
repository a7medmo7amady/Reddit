"use client";

import Image from "next/image";
import Link from "next/link";
import { FormEvent, useEffect, useMemo, useRef, useState } from "react";
import {
  ChatMessage,
  Conversation,
  InboxItem,
  createDirectConversation,
  getConversationMessages,
  getInbox,
  markConversationRead,
  sendChatMessage,
} from "@/lib/chat";
import { getChatWebSocketUrl } from "@/lib/config";
import { getToken, logout } from "@/lib/auth";
import { getMyUserId, getMyUsername } from "@/lib/jwt";
import styles from "./page.module.css";

type ConnectionState = "connecting" | "online" | "offline";

interface RealtimeEvent {
  type: string;
  conversationId?: string;
  userId?: string;
  expiresAt?: string;
  message?: ChatMessage;
}

interface TypingUser {
  userId: string;
  expiresAt: number;
}

function shortId(value: string): string {
  if (value.length <= 10) return value;
  return `${value.slice(0, 6)}...${value.slice(-4)}`;
}

function formatTime(value?: string): string {
  if (!value) return "";
  return new Intl.DateTimeFormat(undefined, {
    hour: "numeric",
    minute: "2-digit",
  }).format(new Date(value));
}

function formatDate(value?: string): string {
  if (!value) return "No activity yet";
  const date = new Date(value);
  const today = new Date();
  if (date.toDateString() === today.toDateString()) return formatTime(value);
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
  }).format(date);
}

function upsertMessage(messages: ChatMessage[], message: ChatMessage) {
  if (messages.some((item) => item.id === message.id)) return messages;
  return [...messages, message].sort(
    (a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
  );
}

export default function ChatPage() {
  const [inbox, setInbox] = useState<InboxItem[]>([]);
  const [activeConversationId, setActiveConversationId] = useState<string | null>(
    null
  );
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [draft, setDraft] = useState("");
  const [newParticipant, setNewParticipant] = useState("");
  const [isLoadingInbox, setIsLoadingInbox] = useState(true);
  const [isLoadingMessages, setIsLoadingMessages] = useState(false);
  const [isSending, setIsSending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [connectionState, setConnectionState] =
    useState<ConnectionState>("connecting");
  const [typingUsers, setTypingUsers] = useState<Record<string, TypingUser[]>>(
    {}
  );

  const socketRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const lastTypingSentRef = useRef(0);
  const activeConversationRef = useRef<string | null>(null);
  const realtimeHandlerRef = useRef<(payload: RealtimeEvent) => void>(() => {});
  const bottomRef = useRef<HTMLDivElement | null>(null);

  const myUserId = getMyUserId();
  const myUsername = getMyUsername();
  const activeInboxItem = inbox.find(
    (item) => item.conversationId === activeConversationId
  );

  useEffect(() => {
    activeConversationRef.current = activeConversationId;
  }, [activeConversationId]);

  useEffect(() => {
    realtimeHandlerRef.current = handleRealtimeEvent;
  });

  const activeTypingUsers = useMemo(() => {
    if (!activeConversationId) return [];
    const now = Date.now();
    return (typingUsers[activeConversationId] ?? []).filter(
      (item) => item.expiresAt > now && item.userId !== myUserId
    );
  }, [activeConversationId, myUserId, typingUsers]);

  useEffect(() => {
    if (!getToken()) {
      window.location.href = "/";
      return;
    }

    getInbox()
      .then((items) => {
        setInbox(items);
        if (items[0]) setActiveConversationId(items[0].conversationId);
      })
      .catch((err: Error) => setError(err.message))
      .finally(() => setIsLoadingInbox(false));
  }, []);

  useEffect(() => {
    if (!activeConversationId) {
      return;
    }

    let ignore = false;
    queueMicrotask(() => {
      setIsLoadingMessages(true);
      setError(null);
      getConversationMessages(activeConversationId)
        .then((items) => {
          if (!ignore) setMessages(items);
        })
        .then(() => markConversationRead(activeConversationId).catch(() => {}))
        .catch((err: Error) => {
          if (!ignore) setError(err.message);
        })
        .finally(() => {
          if (!ignore) setIsLoadingMessages(false);
        });
    });

    return () => {
      ignore = true;
    };
  }, [activeConversationId]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ block: "end" });
  }, [messages, activeTypingUsers.length]);

  useEffect(() => {
    const interval = window.setInterval(() => {
      const now = Date.now();
      setTypingUsers((current) => {
        const next: Record<string, TypingUser[]> = {};
        for (const [conversationId, users] of Object.entries(current)) {
          const active = users.filter((user) => user.expiresAt > now);
          if (active.length > 0) next[conversationId] = active;
        }
        return next;
      });
    }, 1000);

    return () => window.clearInterval(interval);
  }, []);

  useEffect(() => {
    let closedByPage = false;

    function connect() {
      const token = getToken();
      if (!token) return;

      setConnectionState("connecting");
      const socket = new WebSocket(getChatWebSocketUrl(token));
      socketRef.current = socket;

      socket.onopen = () => setConnectionState("online");
      socket.onclose = () => {
        if (socketRef.current === socket) socketRef.current = null;
        setConnectionState("offline");
        if (!closedByPage) {
          reconnectTimerRef.current = setTimeout(connect, 1800);
        }
      };
      socket.onerror = () => setConnectionState("offline");
      socket.onmessage = (event) => {
        try {
          const payload = JSON.parse(event.data) as RealtimeEvent;
          realtimeHandlerRef.current(payload);
        } catch {
          return;
        }
      };
    }

    connect();

    return () => {
      closedByPage = true;
      if (reconnectTimerRef.current) clearTimeout(reconnectTimerRef.current);
      socketRef.current?.close();
    };
  }, []);

  function handleRealtimeEvent(payload: RealtimeEvent) {
    if (payload.type === "chat.message" && payload.message) {
      setMessages((current) => {
        if (payload.conversationId !== activeConversationRef.current) {
          return current;
        }
        return upsertMessage(current, payload.message as ChatMessage);
      });
      refreshInbox();
      return;
    }

    if (
      payload.type === "chat.typing" &&
      payload.conversationId &&
      payload.userId
    ) {
      const expiresAt = payload.expiresAt
        ? new Date(payload.expiresAt).getTime()
        : Date.now() + 5000;
      setTypingUsers((current) => {
        const users = current[payload.conversationId as string] ?? [];
        const withoutUser = users.filter((user) => user.userId !== payload.userId);
        return {
          ...current,
          [payload.conversationId as string]: [
            ...withoutUser,
            { userId: payload.userId as string, expiresAt },
          ],
        };
      });
    }
  }

  function refreshInbox() {
    getInbox()
      .then((items) => setInbox(items))
      .catch(() => {});
  }

  async function handleCreateConversation(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const participantId = newParticipant.trim();
    if (!participantId) return;

    setError(null);
    try {
      const conversation: Conversation =
        await createDirectConversation(participantId);
      await refreshInbox();
      setActiveConversationId(conversation.id);
      setNewParticipant("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not start chat");
    }
  }

  async function handleSendMessage(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const content = draft.trim();
    if (!content || !activeConversationId || isSending) return;

    setIsSending(true);
    setDraft("");
    setError(null);
    try {
      const message = await sendChatMessage(activeConversationId, content);
      setMessages((current) => upsertMessage(current, message));
      refreshInbox();
    } catch (err) {
      setDraft(content);
      setError(err instanceof Error ? err.message : "Message failed to send");
    } finally {
      setIsSending(false);
    }
  }

  function sendTypingSignal(value: string) {
    setDraft(value);
    if (!activeConversationId || value.trim().length === 0) return;
    const now = Date.now();
    if (now - lastTypingSentRef.current < 1500) return;
    lastTypingSentRef.current = now;

    const socket = socketRef.current;
    if (!socket || socket.readyState !== WebSocket.OPEN) return;
    socket.send(
      JSON.stringify({
        type: "chat.typing",
        conversationId: activeConversationId,
      })
    );
  }

  async function handleLogout() {
    await logout();
    window.location.href = "/";
  }

  return (
    <main className={styles.page}>
      <nav className={styles.topbar}>
        <Link href="/" className={styles.brand}>
          <Image src="/reddit-1.svg" alt="Reddit" width={102} height={32} />
        </Link>
        <div className={styles.searchBox}>Search Reddit</div>
        <div className={styles.navActions}>
          {myUsername && (
            <Link href={`/u/${myUsername}`} className={styles.userLink}>
              u/{myUsername}
            </Link>
          )}
          <button className={styles.iconButton} onClick={handleLogout}>
            Log out
          </button>
        </div>
      </nav>

      <div className={styles.shell}>
        <aside className={styles.leftRail}>
          <Link href="/" className={styles.railItem}>
            Home
          </Link>
          <Link href="/chat" className={`${styles.railItem} ${styles.activeRail}`}>
            Chat
          </Link>
          {myUsername && (
            <Link href={`/u/${myUsername}`} className={styles.railItem}>
              Profile
            </Link>
          )}
        </aside>

        <section className={styles.inboxPanel} aria-label="Chats">
          <div className={styles.panelHeader}>
            <div>
              <h1>Chats</h1>
              <p>{connectionState}</p>
            </div>
          </div>

          <form className={styles.newChatForm} onSubmit={handleCreateConversation}>
            <input
              value={newParticipant}
              onChange={(event) => setNewParticipant(event.target.value)}
              placeholder="User ID"
              aria-label="User ID"
            />
            <button type="submit">Start</button>
          </form>

          {error && <p className={styles.errorText}>{error}</p>}

          <div className={styles.inboxList}>
            {isLoadingInbox ? (
              <p className={styles.muted}>Loading chats...</p>
            ) : inbox.length === 0 ? (
              <p className={styles.emptyText}>No chats yet. Start a DM above.</p>
            ) : (
              inbox.map((item) => (
                <button
                  key={item.conversationId}
                  className={`${styles.inboxItem} ${
                    item.conversationId === activeConversationId
                      ? styles.inboxItemActive
                      : ""
                  }`}
                  onClick={() => setActiveConversationId(item.conversationId)}
                >
                  <span className={styles.avatar}>
                    {item.type === "community" ? "r/" : "u/"}
                  </span>
                  <span className={styles.inboxText}>
                    <span className={styles.inboxTitle}>
                      {item.type === "community"
                        ? item.communityId ?? shortId(item.conversationId)
                        : shortId(item.conversationId)}
                    </span>
                    <span className={styles.inboxPreview}>
                      {item.lastMessage?.content ?? "No messages yet"}
                    </span>
                  </span>
                  <span className={styles.inboxMeta}>
                    {formatDate(item.updatedAt)}
                    {item.unreadCount > 0 && (
                      <span className={styles.unread}>{item.unreadCount}</span>
                    )}
                  </span>
                </button>
              ))
            )}
          </div>
        </section>

        <section className={styles.chatPanel} aria-label="Conversation">
          {activeConversationId ? (
            <>
              <header className={styles.chatHeader}>
                <div className={styles.avatarLarge}>u/</div>
                <div>
                  <h2>
                    {activeInboxItem?.type === "community"
                      ? activeInboxItem.communityId
                      : `Chat ${shortId(activeConversationId)}`}
                  </h2>
                  <p>{messages.length} messages</p>
                </div>
              </header>

              <div className={styles.messages}>
                {isLoadingMessages ? (
                  <p className={styles.muted}>Loading messages...</p>
                ) : messages.length === 0 ? (
                  <div className={styles.emptyState}>
                    <h3>Start the conversation</h3>
                    <p>Send a message and it will appear here in real time.</p>
                  </div>
                ) : (
                  messages.map((message) => {
                    const mine = message.senderId === myUserId;
                    return (
                      <article
                        key={message.id}
                        className={`${styles.messageRow} ${
                          mine ? styles.messageMine : ""
                        }`}
                      >
                        <div className={styles.messageAvatar}>
                          {mine ? "me" : "u/"}
                        </div>
                        <div className={styles.messageBubble}>
                          <div className={styles.messageMeta}>
                            <Link href={`/u/${message.senderId}`}>
                              {mine ? "You" : `u/${message.senderId}`}
                            </Link>
                            <span>{formatTime(message.createdAt)}</span>
                          </div>
                          <p>{message.content}</p>
                        </div>
                      </article>
                    );
                  })
                )}
                {activeTypingUsers.length > 0 && (
                  <div className={styles.typingLine}>
                    {activeTypingUsers.map((user) => `u/${user.userId}`).join(", ")}{" "}
                    typing...
                  </div>
                )}
                <div ref={bottomRef} />
              </div>

              <form className={styles.composer} onSubmit={handleSendMessage}>
                <textarea
                  value={draft}
                  onChange={(event) => sendTypingSignal(event.target.value)}
                  placeholder="Message"
                  rows={1}
                  maxLength={2000}
                />
                <button type="submit" disabled={isSending || !draft.trim()}>
                  Send
                </button>
              </form>
            </>
          ) : (
            <div className={styles.noConversation}>
              <h2>Select a chat</h2>
              <p>Your DMs and community rooms will show up here.</p>
            </div>
          )}
        </section>

        <aside className={styles.contextPanel}>
          <h2>About chat</h2>
          <p>
            Realtime DMs, typing indicators, read tracking, and missed-message
            recovery through the chat service.
          </p>
          <div className={styles.contextStat}>
            <span>{inbox.length}</span>
            <p>conversations</p>
          </div>
          <div className={styles.contextStat}>
            <span>{connectionState === "online" ? "Live" : "Retrying"}</span>
            <p>websocket</p>
          </div>
        </aside>
      </div>
    </main>
  );
}
