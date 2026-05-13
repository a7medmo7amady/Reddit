"use client";

import Image from "next/image";
import Link from "next/link";
import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { fetchWithAuth, logout } from "@/lib/auth";
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

export default function UserProfilePage() {
  const { username } = useParams<{ username: string }>();
  const [profile, setProfile] = useState<PublicProfile | null>(null);
  const [notFound, setNotFound] = useState(false);
  const [tab, setTab] = useState<Tab>("overview");
  const [following, setFollowing] = useState(false);
  const [followLoading, setFollowLoading] = useState(false);

  const myUsername = getMyUsername();
  const isOwn = myUsername === username;

  const handleLogout = async () => {
    await logout();
    window.location.href = "/";
  };

  useEffect(() => {
    fetch(
      `${process.env.NEXT_PUBLIC_API_GATEWAY_URL ?? "http://localhost:8088"}/users/${username}`
    )
      .then((r) => {
        if (r.status === 404) { setNotFound(true); return null; }
        return r.json();
      })
      .then((data) => data && setProfile(data))
      .catch(() => setNotFound(true));
  }, [username]);

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

      {/* ── banner ── */}
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

      {/* ── profile header ── */}
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
          <div className={styles.empty}>
            <p>No {tab} yet</p>
          </div>
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
