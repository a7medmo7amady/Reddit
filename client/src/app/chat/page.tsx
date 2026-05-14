"use client";

import Image from "next/image";
import Link from "next/link";
import { FormEvent, useEffect, useMemo, useRef, useState } from "react";
import {
  ChatMessage,
  ChatUser,
  Conversation,
  InboxItem,
  createDirectConversation,
  getConversationMessages,
  getInbox,
  getOrCreateCommunityRoom,
  getUserById,
  getUserByUsername,
  hideConversation,
  markConversationRead,
  sendChatMessage,
  setConversationMuted,
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

function dedupeInbox(items: InboxItem[]): InboxItem[] {
  const seen = new Set<string>();
  const deduped: InboxItem[] = [];

  for (const item of items) {
    const key =
      item.type === "direct"
        ? `direct:${[...(item.otherParticipantIds ?? [])].sort().join(",")}`
        : `${item.type}:${item.communityId ?? item.conversationId}`;

    if (seen.has(key)) continue;
    seen.add(key);
    deduped.push(item);
  }

  return deduped;
}

export default function ChatPage() {
  const [inbox, setInbox] = useState<InboxItem[]>([]);
  const [activeConversationId, setActiveConversationId] = useState<string | null>(
    null
  );
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [draft, setDraft] = useState("");
  const [newUsername, setNewUsername] = useState("");
  const [newCommunity, setNewCommunity] = useState("");
  const [isLoadingInbox, setIsLoadingInbox] = useState(true);
  const [isLoadingMessages, setIsLoadingMessages] = useState(false);
  const [isSending, setIsSending] = useState(false);
  const [isCreatingConversation, setIsCreatingConversation] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [, setConnectionState] = useState<ConnectionState>("connecting");
  const [typingUsers, setTypingUsers] = useState<Record<string, TypingUser[]>>(
    {}
  );
  const [currentUser, setCurrentUser] = useState<{
    id: string | null;
    username: string | null;
  }>({ id: null, username: null });
  const [profilesById, setProfilesById] = useState<Record<string, ChatUser>>({});

  const socketRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const lastTypingSentRef = useRef(0);
  const activeConversationRef = useRef<string | null>(null);
  const realtimeHandlerRef = useRef<(payload: RealtimeEvent) => void>(() => {});
  const bottomRef = useRef<HTMLDivElement | null>(null);

  const myUserId = currentUser.id;
  const myUsername = currentUser.username;
  const activeInboxItem = inbox.find(
    (item) => item.conversationId === activeConversationId
  );
  const visibleMessages = useMemo(
    () => (Array.isArray(messages) ? messages : []),
    [messages]
  );
  const activeConversationTitle = activeInboxItem
    ? conversationTitle(activeInboxItem)
    : activeConversationId
      ? `Chat ${shortId(activeConversationId)}`
      : "Chat";
  const activeOtherParticipantId =
    activeInboxItem?.type === "direct"
      ? activeInboxItem.otherParticipantIds?.[0]
      : null;
  const activeProfileHref = activeOtherParticipantId
    ? profileHrefForUserId(activeOtherParticipantId)
    : null;

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

  function conversationTitle(item: InboxItem): string {
    if (item.type === "community") {
      return item.communityId ? `r/${item.communityId}` : "Community chat";
    }

    const otherParticipantId = item.otherParticipantIds?.[0];
    if (!otherParticipantId) return "Direct message";

    const profile = profilesById[otherParticipantId];
    return profile?.displayName || profile?.username || `u/${otherParticipantId}`;
  }

  function participantLabel(userId: string): string {
    const profile = profilesById[userId];
    return profile?.displayName || profile?.username || `u/${userId}`;
  }

  function profileHrefForUserId(userId: string): string {
    const username =
      userId === myUserId ? myUsername : profilesById[userId]?.username;
    return `/u/${username ?? userId}`;
  }

  useEffect(() => {
    queueMicrotask(() => {
      setCurrentUser({
        id: getMyUserId(),
        username: getMyUsername(),
      });
    });

    if (!getToken()) {
      window.location.href = "/";
      return;
    }

    getInbox()
      .then((items) => {
        const deduped = dedupeInbox(items);
        setInbox(deduped);
        if (deduped[0]) setActiveConversationId(deduped[0].conversationId);
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
    const profileIds = [
      ...inbox.flatMap((item) => item.otherParticipantIds ?? []),
      ...visibleMessages.map((message) => message.senderId),
    ];
    const missingIds = Array.from(
      new Set(
        profileIds.filter(
          (userId) => userId && userId !== myUserId && !profilesById[userId]
        )
      )
    );
    if (missingIds.length === 0) return;

    let ignore = false;
    Promise.all(
      missingIds.map((userId) =>
        getUserById(userId)
          .then((profile) => [userId, profile] as const)
          .catch(() => null)
      )
    ).then((entries) => {
      if (ignore) return;
      setProfilesById((current) => {
        const next = { ...current };
        for (const entry of entries) {
          if (entry) next[entry[0]] = entry[1];
        }
        return next;
      });
    });

    return () => {
      ignore = true;
    };
  }, [inbox, myUserId, profilesById, visibleMessages]);

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
      const isActiveConversation =
        payload.conversationId === activeConversationRef.current;

      setMessages((current) => {
        if (!isActiveConversation) {
          return current;
        }
        return upsertMessage(current, payload.message as ChatMessage);
      });

      if (isActiveConversation && payload.conversationId) {
        markConversationRead(payload.conversationId)
          .catch(() => {})
          .finally(() => refreshInbox());
      } else {
        refreshInbox();
      }
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

  async function refreshInbox() {
    try {
      const items = await getInbox();
      const deduped = dedupeInbox(items);
      setInbox(deduped);
      return deduped;
    } catch {
      return [];
    }
  }

  async function handleCreateConversation(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const username = newUsername.trim().replace(/^u\//i, "");
    if (!username || isCreatingConversation) return;

    setError(null);
    setIsCreatingConversation(true);
    try {
      const participant = await getUserByUsername(username);
      const conversation: Conversation =
        await createDirectConversation(String(participant.id));
      setProfilesById((current) => ({
        ...current,
        [String(participant.id)]: participant,
      }));
      await refreshInbox();
      setActiveConversationId(conversation.id);
      setNewUsername("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not start chat");
    } finally {
      setIsCreatingConversation(false);
    }
  }

  async function handleCreateCommunityRoom(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const communityId = newCommunity.trim().replace(/^r\//i, "");
    if (!communityId || isCreatingConversation) return;

    setError(null);
    setIsCreatingConversation(true);
    try {
      const conversation = await getOrCreateCommunityRoom(communityId);
      await refreshInbox();
      setActiveConversationId(conversation.id);
      setNewCommunity("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not open community chat");
    } finally {
      setIsCreatingConversation(false);
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

  async function handleToggleMute() {
    if (!activeConversationId || !activeInboxItem) return;

    const muted = !activeInboxItem.muted;
    setInbox((current) =>
      current.map((item) =>
        item.conversationId === activeConversationId ? { ...item, muted } : item
      )
    );

    try {
      await setConversationMuted(activeConversationId, muted);
      refreshInbox();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not update chat");
      setInbox((current) =>
        current.map((item) =>
          item.conversationId === activeConversationId
            ? { ...item, muted: !muted }
            : item
        )
      );
    }
  }

  async function handleDeleteConversation() {
    if (!activeConversationId) return;

    const conversationId = activeConversationId;
    const remainingInbox = inbox.filter(
      (item) => item.conversationId !== conversationId
    );
    setInbox(remainingInbox);
    setActiveConversationId(remainingInbox[0]?.conversationId ?? null);
    setMessages([]);
    setError(null);

    try {
      await hideConversation(conversationId);
      refreshInbox();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not delete chat");
      refreshInbox();
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
            </div>
          </div>

          <form className={styles.newChatForm} onSubmit={handleCreateConversation}>
            <input
              value={newUsername}
              onChange={(event) => setNewUsername(event.target.value)}
              placeholder="Username"
              aria-label="Username"
            />
            <button type="submit" disabled={isCreatingConversation}>
              DM
            </button>
          </form>

          <form className={styles.newChatForm} onSubmit={handleCreateCommunityRoom}>
            <input
              value={newCommunity}
              onChange={(event) => setNewCommunity(event.target.value)}
              placeholder="Community"
              aria-label="Community"
            />
            <button type="submit" disabled={isCreatingConversation}>
              Room
            </button>
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
                      {conversationTitle(item)}
                    </span>
                    <span className={styles.inboxPreview}>
                      {item.muted ? "Muted · " : ""}
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
                {activeProfileHref ? (
                  <Link href={activeProfileHref} className={styles.avatarLarge}>
                    u/
                  </Link>
                ) : (
                  <div className={styles.avatarLarge}>
                    {activeInboxItem?.type === "community" ? "r/" : "u/"}
                  </div>
                )}
                <div className={styles.chatHeaderText}>
                  {activeProfileHref ? (
                    <Link href={activeProfileHref}>
                      <h2>{activeConversationTitle}</h2>
                    </Link>
                  ) : (
                    <h2>{activeConversationTitle}</h2>
                  )}
                  <p>
                    {visibleMessages.length} messages
                    {activeInboxItem?.muted ? " · muted" : ""}
                  </p>
                </div>
                <div className={styles.chatActions}>
                  <button type="button" onClick={handleToggleMute}>
                    {activeInboxItem?.muted ? "Unmute" : "Mute"}
                  </button>
                  <button
                    type="button"
                    className={styles.dangerButton}
                    onClick={handleDeleteConversation}
                  >
                    Delete
                  </button>
                </div>
              </header>

              <div className={styles.messages}>
                {isLoadingMessages ? (
                  <p className={styles.muted}>Loading messages...</p>
                ) : visibleMessages.length === 0 ? (
                  <div className={styles.emptyState}>
                    <h3>Start the conversation</h3>
                    <p>Send a message and it will appear here in real time.</p>
                  </div>
                ) : (
                  visibleMessages.map((message) => {
                    const mine = message.senderId === myUserId;
                    return (
                      <article
                        key={message.id}
                        className={`${styles.messageRow} ${
                          mine ? styles.messageMine : ""
                        }`}
                      >
                        <Link
                          href={profileHrefForUserId(message.senderId)}
                          className={styles.messageAvatar}
                        >
                          {mine ? "me" : "u/"}
                        </Link>
                        <div className={styles.messageBubble}>
                          <div className={styles.messageMeta}>
                            <Link
                              href={profileHrefForUserId(message.senderId)}
                            >
                              {mine ? "You" : participantLabel(message.senderId)}
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
                  <article className={styles.messageRow}>
                    <Link
                      href={profileHrefForUserId(activeTypingUsers[0].userId)}
                      className={styles.messageAvatar}
                    >
                      u/
                    </Link>
                    <div className={`${styles.messageBubble} ${styles.typingBubble}`}>
                      <div className={styles.messageMeta}>
                        <span>
                          {activeTypingUsers
                            .map((user) => participantLabel(user.userId))
                            .join(", ")}
                        </span>
                      </div>
                      <p>
                        <span className={styles.typingDot} />
                        <span className={styles.typingDot} />
                        <span className={styles.typingDot} />
                      </p>
                    </div>
                  </article>
                )}
                <div ref={bottomRef} />
              </div>

              <form className={styles.composer} onSubmit={handleSendMessage}>
                <textarea
                  value={draft}
                  onChange={(event) => sendTypingSignal(event.target.value)}
                  onKeyDown={(event) => {
                    if (event.key !== "Enter" || event.shiftKey) return;
                    event.preventDefault();
                    event.currentTarget.form?.requestSubmit();
                  }}
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
      </div>
    </main>
  );
}
