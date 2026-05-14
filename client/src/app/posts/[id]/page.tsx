"use client";

import { useState, useEffect, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import styles from "./page.module.css";
import { getToken } from "@/lib/auth";
import { getMyUsername } from "@/lib/jwt";
import AuthPopup from "@/components/AuthPopup";

const API_URL = process.env.NEXT_PUBLIC_API_GATEWAY_URL ?? "http://localhost:8088";

interface Post {
  id: string;
  title: string;
  body?: string;
  community: string;
  authorId?: string;
  author?: string;
  upvotes?: number;
  downvotes?: number;
  score?: number;
  commentCount?: number;
  createdAt: string;
  images?: { url: string }[];
  video?: { playbackUrl?: string; status: string };
}

interface Comment {
  id: string;
  authorId: string;
  body: string;
  parentId?: string | null;
  upvotes: number;
  downvotes: number;
  createdAt: string;
}

function timeAgo(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const m = Math.floor(diff / 60000);
  if (m < 1) return "just now";
  if (m < 60) return `${m}m ago`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h}h ago`;
  return `${Math.floor(h / 24)}d ago`;
}

function formatScore(n: number): string {
  if (Math.abs(n) >= 1000) return (n / 1000).toFixed(1).replace(/\.0$/, "") + "k";
  return String(n);
}

export default function PostPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();

  const [post, setPost] = useState<Post | null>(null);
  const [postLoading, setPostLoading] = useState(true);
  const [comments, setComments] = useState<Comment[]>([]);
  const [commentsLoading, setCommentsLoading] = useState(true);
  const [commentBody, setCommentBody] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [commentError, setCommentError] = useState<string | null>(null);
  const [postVote, setPostVote] = useState(0);
  const [liveScore, setLiveScore] = useState<number | null>(null);
  const [commentVotes, setCommentVotes] = useState<Record<string, number>>({});
  const [showAuthPopup, setShowAuthPopup] = useState(false);
  const [isAuthed, setIsAuthed] = useState(false);
  const [username, setUsername] = useState<string | null>(null);

  useEffect(() => {
    if (getToken()) {
      setIsAuthed(true);
      setUsername(getMyUsername());
    }
  }, []);

  const fetchPost = useCallback(async () => {
    try {
      const headers: HeadersInit = {};
      const token = getToken();
      if (token) headers["Authorization"] = `Bearer ${token}`;
      const res = await fetch(`${API_URL}/posts/${id}`, { headers });
      if (!res.ok) throw new Error("Post not found");
      setPost(await res.json());
    } catch {
      router.replace("/");
    } finally {
      setPostLoading(false);
    }
  }, [id, router]);

  const fetchComments = useCallback(async () => {
    setCommentsLoading(true);
    try {
      const res = await fetch(`${API_URL}/posts/${id}/comments?limit=50`);
      if (!res.ok) throw new Error();
      const data = await res.json();
      setComments(data.comments || []);
    } catch {
      /* ignore — show empty state */
    } finally {
      setCommentsLoading(false);
    }
  }, [id]);

  useEffect(() => {
    fetchPost();
    fetchComments();
  }, [fetchPost, fetchComments]);

  const handlePostVote = async (direction: number) => {
    if (!isAuthed) { setShowAuthPopup(true); return; }
    const next = postVote === direction ? 0 : direction;
    setPostVote(next);
    try {
      const res = await fetch(`${API_URL}/posts/${id}/vote`, {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${getToken()}` },
        body: JSON.stringify({ direction: next }),
      });
      if (res.ok) {
        const data = await res.json();
        setLiveScore(data.score);
        setPostVote(data.userVote ?? next);
      }
    } catch { /* keep optimistic value */ }
  };

  const handleCommentVote = async (commentId: string, direction: number) => {
    if (!isAuthed) { setShowAuthPopup(true); return; }
    const prev = commentVotes[commentId] ?? 0;
    const next = prev === direction ? 0 : direction;
    setCommentVotes(v => ({ ...v, [commentId]: next }));
    try {
      await fetch(`${API_URL}/comments/${commentId}/vote`, {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${getToken()}` },
        body: JSON.stringify({ direction: next }),
      });
    } catch { /* optimistic */ }
  };

  const handleCommentSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!isAuthed) { setShowAuthPopup(true); return; }
    if (!commentBody.trim()) return;

    setSubmitting(true);
    setCommentError(null);
    try {
      const res = await fetch(`${API_URL}/posts/${id}/comments`, {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${getToken()}` },
        body: JSON.stringify({ body: commentBody }),
      });
      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.error || "Failed to post comment");
      }
      const newComment = await res.json();
      setComments(prev => [newComment, ...prev]);
      setCommentBody("");
      setPost(p => p ? { ...p, commentCount: (p.commentCount ?? 0) + 1 } : p);
    } catch (err: unknown) {
      setCommentError(err instanceof Error ? err.message : "An error occurred");
    } finally {
      setSubmitting(false);
    }
  };

  if (postLoading) {
    return (
      <div className={styles.page}>
        <header className={styles.header}>
          <Link href="/" className={styles.backBtn}>← Back</Link>
        </header>
        <div className={styles.skeleton} />
      </div>
    );
  }

  if (!post) return null;

  const baseScore = post.score ?? ((post.upvotes ?? 0) - (post.downvotes ?? 0));
  const score = liveScore !== null ? liveScore : baseScore + postVote;
  const authorName = post.author || post.authorId || "unknown";

  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <Link href="/" className={styles.backBtn}>← Back</Link>
        <span className={styles.headerCommunity}>r/{post.community}</span>
      </header>

      <div className={styles.content}>
        {/* Post card */}
        <article className={styles.postCard}>
          <div className={styles.postMeta}>
            <Link href={`/r/${post.community}`} className={styles.communityLink}>r/{post.community}</Link>
            <span className={styles.dot}>•</span>
            <span className={styles.metaText}>Posted by u/{authorName}</span>
            <span className={styles.dot}>•</span>
            <span className={styles.metaText}>{timeAgo(post.createdAt)}</span>
          </div>

          <h1 className={styles.postTitle}>{post.title}</h1>

          {post.body && <p className={styles.postBody}>{post.body}</p>}

          {post.images && post.images.length > 0 && (
            <div className={styles.imageContainer}>
              {post.images.map((img, i) => (
                // eslint-disable-next-line @next/next/no-img-element
                <img key={i} src={img.url} alt="" className={styles.postImage} />
              ))}
            </div>
          )}

          {post.video?.playbackUrl && (
            <div className={styles.videoContainer}>
              <video src={post.video.playbackUrl} controls className={styles.postVideo} />
            </div>
          )}

          <div className={styles.postActions}>
            <div className={styles.voteRow}>
              <button
                className={`${styles.voteBtn} ${postVote === 1 ? styles.upvoted : ""}`}
                onClick={() => handlePostVote(1)}
                aria-label="Upvote"
              >▲</button>
              <span className={`${styles.score} ${postVote === 1 ? styles.scoreUp : postVote === -1 ? styles.scoreDown : ""}`}>
                {formatScore(score)}
              </span>
              <button
                className={`${styles.voteBtn} ${postVote === -1 ? styles.downvoted : ""}`}
                onClick={() => handlePostVote(-1)}
                aria-label="Downvote"
              >▼</button>
            </div>
            <span className={styles.commentCount}>{post.commentCount ?? comments.length} Comments</span>
            <button className={styles.shareBtn}>Share</button>
          </div>
        </article>

        {/* Comment form */}
        <div className={styles.commentFormCard}>
          {isAuthed ? (
            <form onSubmit={handleCommentSubmit}>
              <div className={styles.formLabel}>
                Comment as <span className={styles.usernameHighlight}>u/{username}</span>
              </div>
              <textarea
                className={styles.commentInput}
                placeholder="What are your thoughts?"
                value={commentBody}
                onChange={e => setCommentBody(e.target.value)}
                rows={4}
              />
              {commentError && <div className={styles.commentError}>{commentError}</div>}
              <div className={styles.formFooter}>
                <button
                  type="submit"
                  className={styles.submitCommentBtn}
                  disabled={submitting || !commentBody.trim()}
                >
                  {submitting ? "Posting..." : "Comment"}
                </button>
              </div>
            </form>
          ) : (
            <div className={styles.authPrompt}>
              <span>Log in or sign up to leave a comment</span>
              <button className={styles.authPromptBtn} onClick={() => setShowAuthPopup(true)}>Log In</button>
            </div>
          )}
        </div>

        {/* Comments list */}
        <div className={styles.commentsList}>
          {commentsLoading ? (
            <>
              <div className={styles.commentSkeleton} />
              <div className={styles.commentSkeleton} />
              <div className={styles.commentSkeleton} />
            </>
          ) : comments.length === 0 ? (
            <div className={styles.emptyComments}>No comments yet — be the first!</div>
          ) : (
            comments.map(comment => {
              const cvote = commentVotes[comment.id] ?? 0;
              const cscore = (comment.upvotes - comment.downvotes) + cvote;
              return (
                <div key={comment.id} className={styles.commentCard}>
                  <div className={styles.commentMeta}>
                    <span className={styles.commentAuthor}>u/{comment.authorId}</span>
                    <span className={styles.dot}>•</span>
                    <span className={styles.metaText}>{timeAgo(comment.createdAt)}</span>
                  </div>
                  <p className={styles.commentBody}>{comment.body}</p>
                  <div className={styles.commentActions}>
                    <button
                      className={`${styles.commentVoteBtn} ${cvote === 1 ? styles.upvoted : ""}`}
                      onClick={() => handleCommentVote(comment.id, 1)}
                    >▲</button>
                    <span className={`${styles.commentScore} ${cvote === 1 ? styles.scoreUp : cvote === -1 ? styles.scoreDown : ""}`}>
                      {formatScore(cscore)}
                    </span>
                    <button
                      className={`${styles.commentVoteBtn} ${cvote === -1 ? styles.downvoted : ""}`}
                      onClick={() => handleCommentVote(comment.id, -1)}
                    >▼</button>
                  </div>
                </div>
              );
            })
          )}
        </div>
      </div>

      {showAuthPopup && (
        <AuthPopup
          onClose={() => setShowAuthPopup(false)}
          onSuccess={() => { setShowAuthPopup(false); setIsAuthed(true); setUsername(getMyUsername()); }}
        />
      )}
    </div>
  );
}
