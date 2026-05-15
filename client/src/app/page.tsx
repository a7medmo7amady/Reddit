"use client";

import Image from "next/image";
import { useState, useEffect, useCallback } from "react";
import { useSearchParams } from "next/navigation";
import styles from "./page.module.css";
import Link from "next/link";
import { saveToken, getToken, logout } from "@/lib/auth";
import { getMyUsername } from "@/lib/jwt";
import AuthPopup from "@/components/AuthPopup";
import CreateCommunityPopup from "@/components/CreateCommunityPopup";
import PostCard, { Post } from "@/components/PostCard";

const API_URL = process.env.NEXT_PUBLIC_API_GATEWAY_URL ?? "http://localhost:8088";

type TabMode = "home" | "trending" | "followed";

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
  const [showCreateCommunityPopup, setShowCreateCommunityPopup] = useState(false);
  const tabParam = searchParams.get("tab");
  const [activeTab, setActiveTab] = useState<TabMode>(
    tabParam === "followed" ? "followed" : tabParam === "trending" ? "trending" : "home"
  );
  const [posts, setPosts] = useState<Post[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [username, setUsername] = useState<string | null>(null);
  const [followedCommunities, setFollowedCommunities] = useState<{ id: number; name: string }[]>([]);

  useEffect(() => {
    setUsername(getMyUsername());
  }, [isAuthed]);

  const fetchCommunities = useCallback(() => {
    if (!isAuthed) { setFollowedCommunities([]); return; }
    const token = getToken();
    if (!token) return;
    fetch(`${API_URL}/communities/me`, { headers: { Authorization: `Bearer ${token}` } })
      .then(r => r.ok ? r.json() : [])
      .then((data: { id: number; name: string }[]) => {
        setFollowedCommunities(data);
        localStorage.setItem("followedCommunities", data.map(c => c.name).join(","));
      })
      .catch(() => {});
  }, [isAuthed]);

  useEffect(() => {
    fetchCommunities();
  }, [fetchCommunities]);

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

          {isAuthed && followedCommunities.length > 0 && (
            <div className={styles.sidebarSection}>
              <div className={styles.sidebarSectionTitle}>COMMUNITIES</div>
              <ul className={styles.navList}>
                {followedCommunities.map(c => (
                  <li key={c.id}>
                    <Link href={`/r/${c.name}`} className={styles.navItem}>
                      <span className={styles.communityDot} />
                      <span>r/{c.name}</span>
                    </Link>
                  </li>
                ))}
              </ul>
              <div className={styles.sidebarDivider} />
            </div>
          )}

          {isAuthed && (
            <button className={styles.createCommunityBtn} onClick={() => setShowCreateCommunityPopup(true)}>
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
                const key = post.stringId || String(post.id);
                return (
                  <PostCard
                    key={key}
                    post={post}
                    isAuthed={isAuthed}
                    onAuthRequired={() => setShowAuthPopup(true)}
                  />
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

      {showCreateCommunityPopup && (
        <CreateCommunityPopup
          onClose={() => setShowCreateCommunityPopup(false)}
          onSuccess={(community) => {
            setShowCreateCommunityPopup(false);
            setFollowedCommunities(prev => [...prev, { id: community.id, name: community.name }]);
            localStorage.setItem(
              "followedCommunities",
              [...followedCommunities.map(c => c.name), community.name].join(",")
            );
            fetchCommunities();
          }}
        />
      )}
    </div>
  );
}
