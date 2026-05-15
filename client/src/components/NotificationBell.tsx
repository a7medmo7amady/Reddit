"use client";

import { useState, useEffect, useRef } from "react";
import { getToken } from "@/lib/auth";
import { getMyUserId } from "@/lib/jwt";

const API_URL = process.env.NEXT_PUBLIC_API_GATEWAY_URL ?? "http://localhost:8088";

interface NotificationItem {
  id: string;
  user_id: string;
  type: string;
  title: string;
  message: string;
  is_read: boolean;
  link?: string;
  created_at: string;
}

export default function NotificationBell() {
  const [open, setOpen] = useState(false);
  const [notifications, setNotifications] = useState<NotificationItem[]>([]);
  const [wsMessage, setWsMessage] = useState<NotificationItem | null>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const userId = getMyUserId();

  // Fetch recent notifications
  useEffect(() => {
    if (!userId) return;
    const token = getToken();
    if (!token) return;

    fetch(`${API_URL}/notifications/recent`, {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then((res) => (res.ok ? res.json() : []))
      .then((data: NotificationItem[]) => {
        setNotifications(data || []);
      })
      .catch(() => {});
  }, [userId]);

  // WebSocket for real-time notifications
  useEffect(() => {
    if (!userId) return;
    const token = getToken();
    const baseWs = API_URL.replace(/^http/, "ws");
    const wsUrl = `${baseWs}/notifications/ws?user_id=${userId}${token ? `&access_token=${token}` : ""}`;
    const ws = new WebSocket(wsUrl);

    ws.onmessage = (event) => {
      try {
        const msg: NotificationItem = JSON.parse(event.data);
        setWsMessage(msg);
        setNotifications((prev) => [msg, ...prev]);
      } catch {
        // ignore non-json
      }
    };

    ws.onerror = () => {};

    return () => {
      ws.close();
    };
  }, [userId]);

  // Close dropdown on outside click
  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  const unreadCount = notifications.filter((n) => !n.is_read).length;

  const markAllRead = async () => {
    const token = getToken();
    if (!token) return;
    await fetch(`${API_URL}/notifications/mark-read`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
    });
    setNotifications((prev) => prev.map((n) => ({ ...n, is_read: true })));
  };

  return (
    <div ref={dropdownRef} style={{ position: "relative", display: "inline-block" }}>
      <button
        onClick={() => setOpen((o) => !o)}
        style={{ background: "none", border: "none", cursor: "pointer", position: "relative" }}
        title="Notifications"
      >
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
          <path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
        </svg>
        {unreadCount > 0 && (
          <span
            style={{
              position: "absolute",
              top: -4,
              right: -4,
              background: "red",
              color: "white",
              borderRadius: "50%",
              width: 18,
              height: 18,
              fontSize: 11,
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
            }}
          >
            {unreadCount}
          </span>
        )}
      </button>

      {open && (
        <div
          style={{
            position: "absolute",
            right: 0,
            top: 36,
            width: 320,
            background: "white",
            border: "1px solid #ddd",
            borderRadius: 8,
            boxShadow: "0 4px 12px rgba(0,0,0,0.15)",
            zIndex: 1000,
            maxHeight: 400,
            overflowY: "auto",
          }}
        >
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              padding: "12px 16px",
              borderBottom: "1px solid #eee",
            }}
          >
            <strong>Notifications</strong>
            {notifications.length > 0 && (
              <button onClick={markAllRead} style={{ fontSize: 12, background: "none", border: "none", color: "#0079d3", cursor: "pointer" }}>
                Mark all read
              </button>
            )}
          </div>

          {notifications.length === 0 ? (
            <div style={{ padding: 16, color: "#888", textAlign: "center" }}>No notifications yet.</div>
          ) : (
            notifications.slice(0, 20).map((n) => (
              <div
                key={n.id || `${n.title}-${n.created_at}`}
                style={{
                  padding: "10px 16px",
                  borderBottom: "1px solid #f0f0f0",
                  background: n.is_read ? "white" : "#f6f7f8",
                  cursor: n.link ? "pointer" : "default",
                }}
                onClick={() => {
                  if (n.link) window.location.href = n.link;
                }}
              >
                <div style={{ fontWeight: 600, fontSize: 13 }}>{n.title}</div>
                <div style={{ fontSize: 12, color: "#555", marginTop: 2 }}>{n.message}</div>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
}
