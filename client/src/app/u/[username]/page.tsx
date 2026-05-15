"use client";

import Image from "next/image";
import Link from "next/link";
import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { fetchWithAuth, logout } from "@/lib/auth";
import { buildApiUrl } from "@/lib/config";
import { getMyUsername } from "@/lib/jwt";
import styles from "./page.module.css";

type Tab = "overview" | "posts" | "comments" | "saved";

interface PublicProfile {
  username: string;
  displayName: string | null;
  bio: string | null;
  avatar: string | null;
  banner: string | null;
  karma: number;
  createdAt: string;
}

interface UserPost {
  id: string;
  title: string;
  body?: string;
  community: string;
  upvotes: number;
  downvotes: number;
  commentCount: number;
  createdAt: string;
}

interface UserComment {
  id: string;
  postId: string;
  authorId: string;
  author: string;
  body: string;
  upvotes: number;
  downvotes: number;
  createdAt: string;
}

function formatKarma(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + "M";
  if (n >= 1_000) return (n / 1_000).toFixed(1) + "K";
  return String(n);
}

function accountAge(iso: string): string {
  const months = Math.floor(
    (Date.now() - new Date(iso).getTime()) / (1000 * 60 * 60 * 24 * 30)
  );
  if (months < 1) return "less than a month";
  if (months < 12) return `${months} month${months !== 1 ? "s" : ""}`;
  const years = Math.floor(months / 12);
  return `${years} year${years !== 1 ? "s" : ""}`;
}

// API calls go through buildApiUrl() which reads NEXT_PUBLIC_API_GATEWAY_URL

function timeAgo(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const m = Math.floor(diff / 60000);
  if (m < 1) return "just now";
  if (m < 60) return `${m}m ago`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h}h ago`;
  return `${Math.floor(h / 24)}d ago`;
}

export default function UserProfilePage() {
  const { username } = useParams<{ username: string }>();
  const [profile, setProfile] = useState<PublicProfile | null>(null);
  const [notFound, setNotFound] = useState(false);
  const [fetchError, setFetchError] = useState<string | null>(null);
  const [tab, setTab] = useState<Tab>("overview");
  const [following, setFollowing] = useState(false);
  const [followLoading, setFollowLoading] = useState(false);
  const [userPosts, setUserPosts] = useState<UserPost[]>([]);
  const [postsLoading, setPostsLoading] = useState(false);
  const [userComments, setUserComments] = useState<UserComment[]>([]);
  const [commentsLoading, setCommentsLoading] = useState(false);

  const myUsername = getMyUsername();
  const isOwn = myUsername === username;

  const handleLogout = async () => {
    await logout();
    window.location.href = "/";
  };

  useEffect(() => {
    setNotFound(false);
    setFetchError(null);
    setProfile(null);
    fetch(buildApiUrl(`/users/${encodeURIComponent(username)}`))
      .then((r) => {
        if (r.status === 404) { setNotFound(true); return null; }
        if (!r.ok) { return r.json().then(d => { throw new Error(d?.error || `Server error ${r.status}`); }); }
        return r.json();
      })
      .then((data) => data && setProfile(data))
      .catch((err: Error) => {
        if (!notFound) setFetchError(err.message ?? "Could not load profile");
      });
  }, [username]); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (myUsername && myUsername !== username) {
      fetchWithAuth("/users/me/following")
        .then(res => res.ok ? res.json() : [])
        .then(followingList => {
          if (Array.isArray(followingList)) {
            setFollowing(followingList.some((user: any) => user.username === username));
          }
        })
        .catch(() => {});
    }
  }, [myUsername, username]);

  useEffect(() => {
    if (tab !== "posts" || !username) return;
    setPostsLoading(true);
    fetch(buildApiUrl(`/posts?author=${encodeURIComponent(username as string)}&limit=50`), { cache: 'no-store' })
      .then(r => r.json())
      .then(data => setUserPosts(data.posts || []))
      .catch(() => setUserPosts([]))
      .finally(() => setPostsLoading(false));
  }, [tab, username]);

  useEffect(() => {
    if (tab !== "comments" || !username) return;
    setCommentsLoading(true);
    fetch(buildApiUrl(`/comments?author=${encodeURIComponent(username as string)}&limit=50`), { cache: 'no-store' })
      .then(r => r.json())
      .then(data => setUserComments(data.comments || []))
      .catch(() => setUserComments([]))
      .finally(() => setCommentsLoading(false));
  }, [tab, username]);

  const toggleFollow = async () => {
    if (!myUsername) return;
    setFollowLoading(true);
    const endpoint = following
      ? `/users/unfollow/${username}`
      : `/users/follow/${username}`;
    await fetchWithAuth(endpoint, { method: "POST" }).catch(() => {});
    setFollowing((f) => !f);
    setFollowLoading(false);
  };

  if (notFound) {
    return (
      <div className={styles.page}>
        <nav className={styles.topbar}>
          <Link href="/">
            <Image src="/reddit-1.svg" alt="Reddit" width={100} height={32} />
          </Link>
        </nav>
        <div className={styles.notFound}>
          <h1>u/{username}</h1>
          <p>This account doesn&apos;t exist. Try searching for something else.</p>
          <Link href="/" className={styles.homeLink}>Go home</Link>
        </div>
      </div>
    );
  }

  if (fetchError) {
    return (
      <div className={styles.page}>
        <nav className={styles.topbar}>
          <Link href="/">
            <Image src="/reddit-1.svg" alt="Reddit" width={100} height={32} />
          </Link>
        </nav>
        <div className={styles.notFound}>
          <h1>u/{username}</h1>
          <p>Could not load profile: {fetchError}</p>
          <Link href="/" className={styles.homeLink}>Go home</Link>
        </div>
      </div>
    );
  }

  if (!profile) {
    return (
      <div className={styles.page}>
        <nav className={styles.topbar}>
          <Link href="/">
            <Image src="/reddit-1.svg" alt="Reddit" width={100} height={32} />
          </Link>
        </nav>
        <div className={styles.skeleton}>
          <div className={styles.skeletonBanner} />
          <div className={styles.skeletonAvatar} />
          <div className={styles.skeletonLine} style={{ width: 160 }} />
          <div className={styles.skeletonLine} style={{ width: 100 }} />
        </div>
      </div>
    );
  }

  const tabs: { id: Tab; label: string }[] = [
    { id: "overview", label: "Overview" },
    { id: "posts", label: "Posts" },
    { id: "comments", label: "Comments" },
    ...(isOwn ? [{ id: "saved" as Tab, label: "Saved" }] : []),
  ];

  return (
    <div className={styles.page}>
      {/* ── top nav ── */}
      <nav className={styles.topbar}>
        <Link href="/">
          <Image src="/reddit-1.svg" alt="Reddit" width={100} height={32} />
        </Link>
        {myUsername ? (
          <Link href={`/u/${myUsername}`} className={styles.navUser}>
            u/{myUsername}
          </Link>
        ) : (
          <Link href="/" className={styles.navUser}>Log in</Link>
        )}
      </nav>

      <div className={styles.banner}>
        {profile.banner ? (
          <Image
            src={profile.banner}
            alt=""
            fill
            style={{ objectFit: "cover" }}
            priority
          />
        ) : (
          <div className={styles.bannerDefault} />
        )}
      </div>

      <div className={styles.profileHeader}>
        <div className={styles.avatarWrap}>
          {profile.avatar ? (
            <Image
              src={profile.avatar}
              alt={profile.username}
              width={80}
              height={80}
              className={styles.avatar}
            />
          ) : (
            <div className={styles.avatarDefault}>
              {(profile.username?.[0] ?? "?").toUpperCase()}
            </div>
          )}
        </div>

        <div className={styles.headerInfo}>
          <div className={styles.nameRow}>
            <div>
              <h1 className={styles.displayName}>
                {profile.displayName ?? profile.username}
              </h1>
              <p className={styles.handle}>u/{profile.username ?? username}</p>
            </div>

            <div className={styles.actions}>
              {isOwn ? (
                <>
                  <Link href="/settings/profile" className={styles.editBtn}>
                    Edit Profile
                  </Link>
                  <button className={styles.logoutBtn} onClick={handleLogout}>
                    Log Out
                  </button>
                </>
              ) : myUsername ? (
                <button
                  className={following ? styles.unfollowBtn : styles.followBtn}
                  onClick={toggleFollow}
                  disabled={followLoading}
                >
                  {following ? "Unfollow" : "Follow"}
                </button>
              ) : null}
            </div>
          </div>

          <div className={styles.stats}>
            <div className={styles.stat}>
              <span className={styles.statValue}>{formatKarma(profile.karma)}</span>
              <span className={styles.statLabel}>Karma</span>
            </div>
            <div className={styles.statDivider} />
            <div className={styles.stat}>
              <span className={styles.statValue}>{accountAge(profile.createdAt)}</span>
              <span className={styles.statLabel}>on Reddit</span>
            </div>
          </div>

          {profile.bio && <p className={styles.bio}>{profile.bio}</p>}
        </div>
      </div>

      {/* ── tabs ── */}
      <div className={styles.tabs}>
        {tabs.map((t) => (
          <button
            key={t.id}
            className={`${styles.tab} ${tab === t.id ? styles.tabActive : ""}`}
            onClick={() => setTab(t.id)}
          >
            {t.label}
          </button>
        ))}
      </div>

      {/* ── content ── */}
      <div className={styles.content}>
        <div className={styles.feed}>
          {tab === "posts" ? (
            postsLoading ? (
              <div className={styles.empty}><p>Loading posts...</p></div>
            ) : userPosts.length === 0 ? (
              <div className={styles.empty}><p>No posts yet</p></div>
            ) : (
              <div className={styles.postList}>
                {userPosts.map(post => {
                  const score = (post.upvotes ?? 0) - (post.downvotes ?? 0);
                  return (
                    <Link key={post.id} href={`/posts/${post.id}`} className={styles.postCard}>
                      <div className={styles.postCardMeta}>
                        <span className={styles.postCardCommunity}>r/{post.community}</span>
                        <span className={styles.postCardDot}>•</span>
                        <span className={styles.postCardTime}>{timeAgo(post.createdAt)}</span>
                      </div>
                      <h3 className={styles.postCardTitle}>{post.title}</h3>
                      {post.body && <p className={styles.postCardBody}>{post.body}</p>}
                      <div className={styles.postCardFooter}>
                        <span className={styles.postCardStat}>{score} points</span>
                        <span className={styles.postCardStat}>{post.commentCount ?? 0} comments</span>
                      </div>
                    </Link>
                  );
                })}
              </div>
            )
          ) : tab === "comments" ? (
            commentsLoading ? (
              <div className={styles.empty}><p>Loading comments...</p></div>
            ) : userComments.length === 0 ? (
              <div className={styles.empty}><p>No comments yet</p></div>
            ) : (
              <div className={styles.postList}>
                {userComments.map(comment => {
                  const score = (comment.upvotes ?? 0) - (comment.downvotes ?? 0);
                  return (
                    <Link key={comment.id} href={`/posts/${comment.postId}`} className={styles.postCard}>
                      <div className={styles.postCardMeta}>
                        <span className={styles.postCardTime}>{timeAgo(comment.createdAt)}</span>
                      </div>
                      <p className={styles.postCardBody} style={{ marginTop: '8px' }}>{comment.body}</p>
                      <div className={styles.postCardFooter}>
                        <span className={styles.postCardStat}>{score} points</span>
                      </div>
                    </Link>
                  );
                })}
              </div>
            )
          ) : (
            <div className={styles.empty}>
              <p>No {tab} yet</p>
            </div>
          )}
        </div>

        <aside className={styles.sidebar}>
          <div className={styles.card}>
            <h2 className={styles.cardTitle}>
              {profile.displayName ?? profile.username}
            </h2>
            {profile.bio && <p className={styles.cardBio}>{profile.bio}</p>}
            <div className={styles.cardStats}>
              <div>
                <p className={styles.cardStatValue}>{formatKarma(profile.karma)}</p>
                <p className={styles.cardStatLabel}>Karma</p>
              </div>
              <div>
                <p className={styles.cardStatValue}>{accountAge(profile.createdAt)}</p>
                <p className={styles.cardStatLabel}>Redditor for</p>
              </div>
            </div>
            {isOwn && (
              <Link href="/settings/profile" className={styles.editBtnFull}>
                Edit Profile
              </Link>
            )}
          </div>
        </aside>
      </div>
    </div>
  );
}
