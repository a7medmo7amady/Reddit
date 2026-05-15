"use client";

import { useState, useEffect, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import Image from "next/image";
import Link from "next/link";
import { getToken, saveToken, logout } from "@/lib/auth";
import { getMyUsername } from "@/lib/jwt";
import AuthPopup from "@/components/AuthPopup";
import PostCard, { Post } from "@/components/PostCard";
import NotificationBell from "@/components/NotificationBell";
import styles from "@/app/page.module.css";
import communityStyles from "./community.module.css";

const API_URL = process.env.NEXT_PUBLIC_API_GATEWAY_URL ?? "http://localhost:8088";

interface CommunityInfo {
  id: number;
  name: string;
  description: string;
  memberCount: number;
}

export default function CommunityPage() {
  const params = useParams();
  const name = params?.name as string;

  const router = useRouter();
  const [searchQuery, setSearchQuery] = useState("");
  const [isAuthed, setIsAuthed] = useState(false);
  const [showAuthPopup, setShowAuthPopup] = useState(false);
  const [username, setUsername] = useState<string | null>(null);
  const [community, setCommunity] = useState<CommunityInfo | null>(null);
  const [posts, setPosts] = useState<Post[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [communityError, setCommunityError] = useState<string | null>(null);
  const [isMember, setIsMember] = useState(false);
  const [isBanned, setIsBanned] = useState(false);
  const [memberLoading, setMemberLoading] = useState(false);
  const [followedCommunities, setFollowedCommunities] = useState<{ id: number; name: string }[]>(() => {
    if (typeof window === "undefined") return [];
    const stored = localStorage.getItem("followedCommunities");
    if (!stored) return [];
    return stored.split(",").filter(Boolean).map((name, i) => ({ id: -(i + 1), name }));
  });

  // ── Auth init ────────────────────────────────────────────────────────────────
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const urlToken = params.get("accessToken");
    if (urlToken) {
      saveToken(urlToken);
      setIsAuthed(true);
      const clean = new URL(window.location.href);
      clean.searchParams.delete("accessToken");
      window.history.replaceState({}, "", clean.toString());
      return;
    }
    if (getToken()) setIsAuthed(true);
  }, []);

  useEffect(() => {
    setUsername(getMyUsername());
  }, [isAuthed]);

  // ── Fetch community details ───────────────────────────────────────────────────
  useEffect(() => {
    if (!name) return;
    fetch(`${API_URL}/communities/${name}`)
      .then(r => {
        if (!r.ok) throw new Error("Community not found");
        return r.json();
      })
      .then((data: CommunityInfo) => setCommunity(data))
      .catch(() => setCommunityError("Community not found or unavailable."));
  }, [name]);

  // ── Fetch membership & ban status ─────────────────────────────────────────────
  useEffect(() => {
    if (!isAuthed || !name) return;
    const token = getToken();
    if (!token) return;

    // Check membership
    fetch(`${API_URL}/communities/${name}/membership`, {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then(r => r.ok ? r.json() : { member: false })
      .then(data => setIsMember(data.member ?? false))
      .catch(() => {});

    // Check ban — hits feed-service via /posts/community/:name and uses 403 as signal
    fetch(`${API_URL}/posts/community/${name}?limit=1`, {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then(r => { if (r.status === 403) setIsBanned(true); })
      .catch(() => {});
  }, [isAuthed, name]);

  // ── Fetch followed communities for sidebar ─────────────────────────────────────
  const fetchCommunities = useCallback(() => {
    if (!isAuthed) return;
    const token = getToken();
    if (!token) return;
    fetch(`${API_URL}/communities/me`, { headers: { Authorization: `Bearer ${token}` } })
      .then(r => r.ok ? r.json() : null)
      .then((data: { id: number; name: string }[] | null) => {
        if (!data) return;
        setFollowedCommunities(data);
        localStorage.setItem("followedCommunities", data.map(c => c.name).join(","));
      })
      .catch(() => {});
  }, [isAuthed]);

  useEffect(() => { fetchCommunities(); }, [fetchCommunities]);

  // ── Fetch posts for this community ────────────────────────────────────────────
  useEffect(() => {
    if (!name) return;
    setIsLoading(true);
    const headers: HeadersInit = {};
    const token = getToken();
    if (token) headers["Authorization"] = `Bearer ${token}`;

    fetch(`${API_URL}/posts/community/${name}?limit=50`, { headers })
      .then(r => {
        if (r.status === 403) { setIsBanned(true); return { posts: [] }; }
        return r.ok ? r.json() : { posts: [] };
      })
      .then(data => setPosts(data.posts || []))
      .catch(() => setPosts([]))
      .finally(() => setIsLoading(false));
  }, [name, isAuthed]);

  // ── Join / Leave ──────────────────────────────────────────────────────────────
  const handleToggleMembership = async () => {
    if (!isAuthed) { setShowAuthPopup(true); return; }
    const token = getToken();
    if (!token) return;
    setMemberLoading(true);
    const endpoint = isMember ? "leave" : "join";
    try {
      const res = await fetch(`${API_URL}/communities/${name}/${endpoint}`, {
        method: "POST",
        headers: { Authorization: `Bearer ${token}` },
      });
      if (res.ok) {
        setIsMember(m => !m);
        setCommunity(c => c
          ? { ...c, memberCount: c.memberCount + (isMember ? -1 : 1) }
          : c
        );
        fetchCommunities();
      }
    } catch {}
    setMemberLoading(false);
  };

  const handleLogout = async () => {
    await logout();
    setIsAuthed(false);
    setFollowedCommunities([]);
    localStorage.removeItem("followedCommunities");
  };

  const NAV_LINKS = [
    { label: "Home", href: "/" },
    { label: "Trending", href: "/trending" },
    { label: "Following", href: "/following" },
  ];

  return (
    <div className={styles.app}>
      {/* Header */}
      <header className={styles.header}>
        <div className={styles.headerLeft}>
          <Link href="/" className={styles.brand}>
            <Image src="/reddit-1.svg" alt="Reddit" width={108} height={36} priority />
          </Link>
        </div>
        <div className={styles.headerCenter}>
          <div className={styles.searchBar}>
            <svg className={styles.searchIcon} width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="11" cy="11" r="8"></circle>
              <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
            </svg>
            <input
              type="text"
              placeholder="Search Reddit"
              className={styles.searchInput}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && searchQuery.trim()) {
                  router.push(`/search?q=${encodeURIComponent(searchQuery.trim())}`);
                }
              }}
            />
          </div>
        </div>
        <div className={styles.headerRight}>
          {isAuthed && username ? (
            <div className={styles.userMenu}>
              <NotificationBell />
              <Link href="/chat" className={styles.chatIcon} title="Chat">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
                </svg>
              </Link>
              <Link href={`/u/${username}`} className={styles.avatar} title={username}>
                {username.charAt(0).toUpperCase()}
              </Link>
              <span className={styles.usernameLabel}>u/{username}</span>
              <button className={styles.logoutBtn} onClick={handleLogout}>Log Out</button>
            </div>
          ) : (
            <div className={styles.authButtons}>
              <button className={styles.loginBtn} onClick={() => setShowAuthPopup(true)}>Log In</button>
              <button className={styles.signupBtn} onClick={() => setShowAuthPopup(true)}>Sign Up</button>
            </div>
          )}
        </div>
      </header>

      <div className={styles.body}>
        {/* Left sidebar */}
        <nav className={styles.leftSidebar}>
          <ul className={styles.navList}>
            {NAV_LINKS.map(({ label, href }) => (
              <li key={label}>
                <Link href={href} className={styles.navItem}><span>{label}</span></Link>
              </li>
            ))}
          </ul>
          <div className={styles.sidebarDivider} />
          {isAuthed && followedCommunities.length > 0 && (
            <div className={styles.sidebarSection}>
              <div className={styles.sidebarSectionTitle}>COMMUNITIES</div>
              <ul className={styles.navList}>
                {followedCommunities.map(c => (
                  <li key={c.id}>
                    <Link href={`/r/${c.name}`} className={`${styles.navItem} ${c.name === name ? styles.navActive : ""}`}>
                      <span className={styles.communityDot} />
                      <span>r/{c.name}</span>
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </nav>

        {/* Main feed */}
        <main className={styles.feed}>
          {communityError ? (
            <div className={styles.emptyState}>{communityError}</div>
          ) : (
            <>
              {/* Community header */}
              <div className={communityStyles.communityHeader}>
                <div className={communityStyles.communityInfo}>
                  <h1 className={communityStyles.communityName}>r/{name}</h1>
                  {community && (
                    <>
                      {community.description && (
                        <p className={communityStyles.communityDesc}>{community.description}</p>
                      )}
                      <span className={communityStyles.memberCount}>
                        {community.memberCount.toLocaleString()} {community.memberCount === 1 ? "member" : "members"}
                      </span>
                    </>
                  )}
                </div>
                {isAuthed && !isBanned && (
                  <button
                    className={isMember ? communityStyles.leaveBtn : communityStyles.joinBtn}
                    onClick={handleToggleMembership}
                    disabled={memberLoading}
                  >
                    {memberLoading ? "..." : isMember ? "Leave" : "Join"}
                  </button>
                )}
                {isBanned && (
                  <span className={communityStyles.bannedBadge}>You are banned from this community</span>
                )}
              </div>

              {/* Create post bar */}
              {isAuthed && !isBanned && (
                <div className={styles.createPostCard}>
                  <div className={styles.createPostAvatar}>
                    {username ? username.charAt(0).toUpperCase() : "?"}
                  </div>
                  <Link
                    href={`/submit?community=${name}`}
                    className={styles.createPostInput}
                  >
                    Create Post
                  </Link>
                  <Link
                    href={`/submit?community=${name}`}
                    className={styles.createPostBtn}
                  >
                    Post
                  </Link>
                </div>
              )}

              {/* Posts */}
              {isBanned ? (
                <div className={styles.emptyState}>
                  You have been banned from r/{name} and cannot view or post here.
                </div>
              ) : isLoading ? (
                <div className={styles.placeholder}>
                  {[...Array(5)].map((_, i) => <div key={i} className={styles.skeletonCard} />)}
                </div>
              ) : posts.length === 0 ? (
                <div className={styles.emptyState}>
                  No posts yet in r/{name} — be the first to post!
                </div>
              ) : (
                <div className={styles.postList}>
                  {posts.map(post => {
                    const key = post.stringId || String(post.id);
                    return (
                      <PostCard
                        key={key}
                        post={post}
                        isAuthed={isAuthed}
                        onAuthRequired={() => setShowAuthPopup(true)}
                        hideCommunity
                      />
                    );
                  })}
                </div>
              )}
            </>
          )}
        </main>

        {/* Right sidebar */}
        <aside className={styles.rightSidebar}>
          {community && (
            <div className={styles.widget}>
              <div className={styles.widgetHeader}>About r/{community.name}</div>
              <div className={styles.widgetBody}>
                {community.description && (
                  <p className={styles.widgetText}>{community.description}</p>
                )}
                <p className={styles.widgetText}>
                  <strong>{community.memberCount.toLocaleString()}</strong> members
                </p>
                {isAuthed && !isBanned && (
                  <Link href={`/submit?community=${name}`} className={styles.widgetBtn}>
                    Create Post
                  </Link>
                )}
                {!isAuthed && (
                  <>
                    <button className={styles.widgetBtn} onClick={() => setShowAuthPopup(true)}>Sign Up</button>
                    <button className={styles.widgetBtnOutline} onClick={() => setShowAuthPopup(true)}>Log In</button>
                  </>
                )}
              </div>
            </div>
          )}
        </aside>
      </div>

      {showAuthPopup && (
        <AuthPopup
          onClose={() => setShowAuthPopup(false)}
          onSuccess={() => { setShowAuthPopup(false); setIsAuthed(true); }}
        />
      )}
    </div>
  );
}
