"use client";

import Link from "next/link";
import styles from "@/app/page.module.css";
import { getToken } from "@/lib/auth";
import { useState } from "react";

const API_URL = process.env.NEXT_PUBLIC_API_GATEWAY_URL ?? "http://localhost:8088";

export interface PostImage { thumbnail: string; preview: string; full: string; }

export interface Post {
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

interface PostCardProps {
  post: Post;
  isAuthed: boolean;
  onAuthRequired: () => void;
  /** If true, hides the community tag (useful inside a community page) */
  hideCommunity?: boolean;
}

export default function PostCard({ post, isAuthed, onAuthRequired, hideCommunity }: PostCardProps) {
  const key = post.stringId || String(post.id);
  const authorName = post.author || post.authorId || "unknown";
  const initialScore = post.score ?? ((post.upvotes ?? 0) - (post.downvotes ?? 0));
  const [voted, setVoted] = useState(0);
  const [score, setScore] = useState(initialScore);

  const handleVote = async (direction: number) => {
    if (!isAuthed) { onAuthRequired(); return; }
    const next = voted === direction ? 0 : direction;
    const delta = next - voted;
    setVoted(next);
    setScore(s => s + delta);
    try {
      const res = await fetch(`${API_URL}/posts/${key}/vote`, {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${getToken()}` },
        body: JSON.stringify({ direction: next }),
      });
      if (res.ok) {
        const data = await res.json();
        setScore(data.score ?? (data.upvotes - data.downvotes));
      }
    } catch { /* optimistic */ }
  };

  return (
    <article className={styles.postCard}>
      <div className={styles.postBody}>
        <div className={styles.postMeta}>
          <div className={styles.postAuthorAvatar}>{authorName.charAt(0).toUpperCase()}</div>
          {!hideCommunity && (
            <>
              <Link href={`/r/${post.community}`} className={styles.communityTag}>
                r/{post.community}
              </Link>
              <span className={styles.metaDot}>•</span>
            </>
          )}
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
          <div className={styles.votePill}>
            <button
              className={`${styles.voteBtn} ${voted === 1 ? styles.upvoted : ""}`}
              onClick={() => handleVote(1)}
              aria-label="Upvote"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><path d="M12 4 L20 14 H4 Z"/></svg>
            </button>
            <span className={`${styles.score} ${voted === 1 ? styles.scoreUp : voted === -1 ? styles.scoreDown : ""}`}>
              {formatScore(score)}
            </span>
            <button
              className={`${styles.voteBtn} ${voted === -1 ? styles.downvoted : ""}`}
              onClick={() => handleVote(-1)}
              aria-label="Downvote"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><path d="M12 20 L4 10 H20 Z"/></svg>
            </button>
          </div>

          <Link href={`/posts/${key}`} className={styles.actionBtn}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
            </svg>
            {post.commentCount ?? 0} Comments
          </Link>

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
}
