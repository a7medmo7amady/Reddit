"use client";

import Image from "next/image";
import { useState, useEffect, useCallback } from "react";
import { useSearchParams } from "next/navigation";
import styles from "./page.module.css";
import Link from "next/link";
import { saveToken, getToken, logout } from "@/lib/auth";
import { getMyUsername } from "@/lib/jwt";
import { buildApiUrl } from "@/lib/config";
import AuthPopup from "@/components/AuthPopup";

const API_URL = process.env.NEXT_PUBLIC_API_GATEWAY_URL ?? "http://localhost:8088";

type TabMode = "home" | "trending" | "followed";

interface PostImage { thumbnail: string; preview: string; full: string; }

interface Post {
  id?: string | number;
  stringId?: string;
  title: string;
  body?: string;
  authorId?: string;
  author?: string;
  community: string;
  type?: string;
  upvotes?: number;
  downvotes?: number;
  score?: number;
  commentCount?: number;
  createdAt: string;
  images?: PostImage[];
  video?: { status: string; playbackUrl?: string };
}

function formatScore(n: number): string {
  if (Math.abs(n) >= 1000) return (n / 1000).toFixed(1).replace(/\.0$/, "") + "k";
  return String(n);
}

function timeAgo(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const m = Math.floor(diff / 60000);
  if (m < 1) return "just now";
  if (m < 60) return `${m}m ago`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h}h ago`;
  const d = Math.floor(h / 24);
  return `${d}d ago`;
}

const NAV_LINKS = [
  { label: "Home",    href: "/" },
  { label: "Trending", href: "/trending" },
  { label: "Following", href: "/following" },
];

const PAGE_TAB: Record<string, "home" | "trending" | "followed"> = {
  "/":         "home",
  "/trending": "trending",
  "/following": "followed",
};

export default function Home() {
  const searchParams = useSearchParams();
  const [isAuthed, setIsAuthed] = useState(false);
  const [showAuthPopup, setShowAuthPopup] = useState(false);
  const tabParam = searchParams.get("tab");
  const [activeTab, setActiveTab] = useState<TabMode>(
    tabParam === "followed" ? "followed" : tabParam === "trending" ? "trending" : "home"
  );
  const [posts, setPosts] = useState<Post[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [username, setUsername] = useState<string | null>(null);
  const [votedPosts, setVotedPosts] = useState<Record<string, number>>({});

  useEffect(() => {
    setUsername(getMyUsername());
  }, [isAuthed]);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const urlToken = params.get("accessToken");
    if (urlToken) {
      saveToken(urlToken);
      queueMicrotask(() => setIsAuthed(true));
      const clean = new URL(window.location.href);
      clean.searchParams.delete("accessToken");
      window.history.replaceState({}, "", clean.toString());
      return;
    }

    if (getToken()) {
      queueMicrotask(() => setIsAuthed(true));
    }
  }, []);

  const fetchPosts = useCallback(async () => {
    setIsLoading(true);
    try {
      const headers: HeadersInit = {};
      const token = getToken();
      if (token) headers["Authorization"] = `Bearer ${token}`;

      let fetchedPosts: Post[] = [];

      if (activeTab === "followed") {
        const followed = localStorage.getItem("followedCommunities");
        if (!followed) {
          setPosts([]);
          setIsLoading(false);
          return;
        }

        const res = await fetch(`${API_URL}/posts/feed?communities=${followed}`, token ? { headers } : undefined);
        if (res.ok) {
          const data = await res.json();
          fetchedPosts = data.posts || [];
        }
      } else if (activeTab === "trending") {
        const res = await fetch(`${API_URL}/posts/trending`, token ? { headers } : undefined);
        if (res.ok) {
          const data = await res.json();
          fetchedPosts = data.posts || [];
        }
      } else if (activeTab === "home") {
        const followed = localStorage.getItem("followedCommunities");
        if (isAuthed && followed) {
          const [trendingRes, followedRes] = await Promise.all([
            fetch(`${API_URL}/posts/trending`, token ? { headers } : undefined),
            fetch(`${API_URL}/posts/feed?communities=${followed}`, token ? { headers } : undefined)
          ]);
          const trendingData = trendingRes.ok ? await trendingRes.json() : { posts: [] };
          const followedData = followedRes.ok ? await followedRes.json() : { posts: [] };
          
          fetchedPosts = [...(trendingData.posts || []), ...(followedData.posts || [])];
          // Simple random shuffle for the mix
          fetchedPosts.sort(() => Math.random() - 0.5);
        } else {
          const res = await fetch(`${API_URL}/posts/trending`, token ? { headers } : undefined);
          if (res.ok) {
            const data = await res.json();
            fetchedPosts = data.posts || [];
          }
        }
      }

      // Deduplicate by ID
      const uniquePosts: Post[] = [];
      const seen = new Set<string>();
      for (const p of fetchedPosts) {
        const k = p.stringId || String(p.id);
        if (!seen.has(k)) {
          seen.add(k);
          uniquePosts.push(p);
        }
      }

      setPosts(uniquePosts);
    } catch (err) {
      console.error("Error fetching posts:", err);
    } finally {
      setIsLoading(false);
    }
  }, [activeTab, isAuthed]);

  useEffect(() => { fetchPosts(); }, [fetchPosts]);

  const handleLogout = async () => { await logout(); setIsAuthed(false); };

  const handleVote = async (postKey: string, direction: number) => {
    if (!isAuthed) { setShowAuthPopup(true); return; }
    const prev = votedPosts[postKey] ?? 0;
    const next = prev === direction ? 0 : direction;
    const delta = next - prev;
    setVotedPosts(v => ({ ...v, [postKey]: next }));
    setPosts(ps => ps.map(p => {
      const k = p.stringId || String(p.id);
      if (k !== postKey) return p;
      return { ...p, score: (p.score ?? 0) + delta };
    }));
    try {
      const res = await fetch(`${API_URL}/posts/${postKey}/vote`, {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${getToken()}` },
        body: JSON.stringify({ direction: next }),
      });
      if (res.ok) {
        const data = await res.json();
        setPosts(ps => ps.map(p => {
          const k = p.stringId || String(p.id);
          if (k !== postKey) return p;
          return { ...p, upvotes: data.upvotes, downvotes: data.downvotes, score: data.score };
        }));
      }
    } catch { /* optimistic — keep local delta */ }
  };

  const handleTabClick = (tab: TabMode) => {
    if (tab === "followed" && !isAuthed) { setShowAuthPopup(true); return; }
    setActiveTab(tab);
  };

  const handleCreatePostClick = (e: React.MouseEvent) => {
    if (!isAuthed) { e.preventDefault(); setShowAuthPopup(true); }
  };

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
            <input type="text" placeholder="Search Reddit" className={styles.searchInput} />
          </div>
        </div>

        <div className={styles.headerRight}>
          {isAuthed && username ? (
            <div className={styles.userMenu}>
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
        {/* Left Sidebar */}
        <nav className={styles.leftSidebar}>
          <ul className={styles.navList}>
            {NAV_LINKS.map(({ label, href }) => {
              const isActive =
                (label === "Home"    && activeTab === "home") ||
                (label === "Trending" && activeTab === "trending") ||
                (label === "Following" && activeTab === "followed");
              return (
                <li key={label}>
                  <Link href={href} className={`${styles.navItem} ${isActive ? styles.navActive : ""}`}>
                    <span>{label}</span>
                  </Link>
                </li>
              );
            })}
          </ul>

          <div className={styles.sidebarDivider} />

          {isAuthed && (
            <button className={styles.createCommunityBtn} onClick={() => setShowAuthPopup(false)}>
              + Create Community
            </button>
          )}
        </nav>

        {/* Feed */}
        <main className={styles.feed}>
          {/* Create post prompt */}
          <div className={styles.createPostCard}>
            <div className={styles.createPostAvatar}>
              {username ? username.charAt(0).toUpperCase() : "?"}
            </div>
            <Link href="/submit" className={styles.createPostInput} onClick={handleCreatePostClick}>
              Create Post
            </Link>
            <Link href="/submit" className={styles.createPostBtn} onClick={handleCreatePostClick}>
              Post
            </Link>
          </div>


          {isLoading ? (
            <div className={styles.placeholder}>
              {[...Array(5)].map((_, i) => <div key={i} className={styles.skeletonCard} />)}
            </div>
          ) : posts.length === 0 ? (
            <div className={styles.emptyState}>No posts yet — be the first to post!</div>
          ) : (
            <div className={styles.postList}>
              {posts.map(post => {
                const score = post.score ?? ((post.upvotes ?? 0) - (post.downvotes ?? 0));
                const authorName = post.author || post.authorId || "unknown";
                const key = post.stringId || String(post.id);
                const voted = votedPosts[key] ?? 0;
                return (
                  <article key={key} className={styles.postCard}>
                    <div className={styles.postBody}>
                      <div className={styles.postMeta}>
                        <div className={styles.postAuthorAvatar}>{authorName.charAt(0).toUpperCase()}</div>
                        <Link href={`/r/${post.community}`} className={styles.communityTag}>
                          r/{post.community}
                        </Link>
                        <span className={styles.metaDot}>•</span>
                        <span className={styles.metaText}>u/{authorName}</span>
                        <span className={styles.metaDot}>•</span>
                        <span className={styles.metaText}>{timeAgo(post.createdAt)}</span>
                      </div>
                      <Link href={`/posts/${key}`} className={styles.postTitleLink}>
                        <h2 className={styles.postTitle}>{post.title}</h2>
                      </Link>
                      {post.body && <p className={styles.postExcerpt}>{post.body}</p>}
                      {post.images && post.images.length > 0 && (
                        // eslint-disable-next-line @next/next/no-img-element
                        <img src={`${API_URL}${post.images[0].preview}`} alt="" className={styles.postMediaPreview} />
                      )}
                      {post.type === "video" && post.video?.playbackUrl && (
                        <video src={post.video.playbackUrl} className={styles.postMediaPreview} muted playsInline />
                      )}
                      {post.type === "video" && !post.video?.playbackUrl && (
                        <div className={styles.videoPlaceholder}>&#9654; Video processing...</div>
                      )}
                      <div className={styles.postActions}>
                        {/* Vote pill */}
                        <div className={styles.votePill}>
                          <button
                            className={`${styles.voteBtn} ${voted === 1 ? styles.upvoted : ""}`}
                            onClick={() => handleVote(key, 1)}
                            aria-label="Upvote"
                          >
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><path d="M12 4 L20 14 H4 Z"/></svg>
                          </button>
                          <span className={`${styles.score} ${voted === 1 ? styles.scoreUp : voted === -1 ? styles.scoreDown : ""}`}>
                            {formatScore(score)}
                          </span>
                          <button
                            className={`${styles.voteBtn} ${voted === -1 ? styles.downvoted : ""}`}
                            onClick={() => handleVote(key, -1)}
                            aria-label="Downvote"
                          >
                            <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><path d="M12 20 L4 10 H20 Z"/></svg>
                          </button>
                        </div>

                        {/* Comments */}
                        <Link href={`/posts/${key}`} className={styles.actionBtn}>
                          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
                          </svg>
                          {post.commentCount ?? 0} Comments
                        </Link>

                        {/* Share */}
                        <button className={styles.actionBtn}>
                          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <path d="M4 12v8a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-8"/>
                            <polyline points="16 6 12 2 8 6"/>
                            <line x1="12" y1="2" x2="12" y2="15"/>
                          </svg>
                          Share
                        </button>
                      </div>
                    </div>
                  </article>
                );
              })}
            </div>
          )}
        </main>

        {/* Right Sidebar */}
        <aside className={styles.rightSidebar}>
          <div className={styles.widget}>
            <div className={styles.widgetHeader}>Home</div>
            <div className={styles.widgetBody}>
              <p className={styles.widgetText}>
                Your personal Reddit frontpage. Come here to check in with your favourite communities.
              </p>
              {isAuthed ? (
                <Link href="/submit" className={styles.widgetBtn}>
                  Create Post
                </Link>
              ) : (
                <>
                  <button className={styles.widgetBtn} onClick={() => setShowAuthPopup(true)}>Sign Up</button>
                  <button className={styles.widgetBtnOutline} onClick={() => setShowAuthPopup(true)}>Log In</button>
                </>
              )}
            </div>
          </div>
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
