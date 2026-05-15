"use client";

import { useState, useRef, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import styles from "./page.module.css";
import { getToken } from "@/lib/auth";
import { getMyUsername } from "@/lib/jwt";

const API_URL = process.env.NEXT_PUBLIC_API_GATEWAY_URL ?? "http://localhost:8088";

export default function SubmitPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [community, setCommunity] = useState("");
  const [mediaFile, setMediaFile] = useState<File | null>(null);
  const [mediaPreview, setMediaPreview] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [username, setUsername] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<"text" | "media">("text");

  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (!getToken()) {
      router.replace("/");
      return;
    }
    setUsername(getMyUsername());
    // Pre-fill community from query param (e.g. from r/[name] page)
    const prefill = searchParams.get("community");
    if (prefill) setCommunity(prefill);
  }, [router, searchParams]);

  const handleMediaChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setMediaFile(file);
    setMediaPreview(URL.createObjectURL(file));
  };

  const removeMedia = () => {
    setMediaFile(null);
    setMediaPreview(null);
    if (fileInputRef.current) fileInputRef.current.value = "";
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!getToken()) { router.replace("/"); return; }
    if (!title.trim()) { setError("Title is required"); return; }
    if (!community.trim()) { setError("Community is required"); return; }

    setIsLoading(true);
    setError(null);

    const fd = new FormData();
    fd.append("title", title);
    fd.append("body", body);
    fd.append("community", community);
    if (mediaFile) {
      fd.append(mediaFile.type.startsWith("video/") ? "video" : "images", mediaFile);
    }

    try {
      const res = await fetch(`${API_URL}/posts`, {
        method: "POST",
        headers: { Authorization: `Bearer ${getToken()}` },
        body: fd,
      });
      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.error || "Failed to create post");
      }
      router.push("/");
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "An error occurred");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <Link href="/" className={styles.backBtn}>← Back</Link>
        <h1 className={styles.pageTitle}>Create post</h1>
      </header>

      <div className={styles.content}>
        <form className={styles.card} onSubmit={handleSubmit}>
          <div className={styles.communitySelector}>
            <div className={styles.communityAvatar}>
              {community ? community.charAt(0).toUpperCase() : username ? username.charAt(0).toUpperCase() : "u"}
            </div>
            <input
              className={styles.communityInput}
              placeholder="Choose a community"
              value={community}
              onChange={e => setCommunity(e.target.value)}
            />
          </div>

          <div className={styles.tabs}>
            <button
              type="button"
              className={`${styles.tab} ${activeTab === "text" ? styles.tabActive : ""}`}
              onClick={() => setActiveTab("text")}
            >
              Text
            </button>
            <button
              type="button"
              className={`${styles.tab} ${activeTab === "media" ? styles.tabActive : ""}`}
              onClick={() => setActiveTab("media")}
            >
              Images &amp; Video
            </button>
          </div>

          <input
            className={styles.titleInput}
            placeholder="Title*"
            value={title}
            onChange={e => setTitle(e.target.value)}
            maxLength={300}
            required
          />
          <div className={styles.titleCount}>{title.length}/300</div>

          {activeTab === "text" ? (
            <textarea
              className={styles.bodyInput}
              placeholder="Body text (optional)"
              value={body}
              onChange={e => setBody(e.target.value)}
              rows={6}
            />
          ) : (
            <div className={styles.mediaZone}>
              <input
                type="file"
                ref={fileInputRef}
                style={{ display: "none" }}
                accept="image/*,video/*"
                onChange={handleMediaChange}
              />
              {mediaPreview ? (
                <div className={styles.preview}>
                  <button type="button" className={styles.removeBtn} onClick={removeMedia}>×</button>
                  {mediaFile?.type.startsWith("video/")
                    ? <video src={mediaPreview} controls className={styles.previewMedia} />
                    : <img src={mediaPreview} alt="preview" className={styles.previewMedia} />}
                </div>
              ) : (
                <button type="button" className={styles.uploadBtn} onClick={() => fileInputRef.current?.click()}>
                  Upload Image or Video
                </button>
              )}
            </div>
          )}

          {error && <div className={styles.error}>{error}</div>}

          <div className={styles.footer}>
            <Link href="/" className={styles.cancelBtn}>Cancel</Link>
            <button type="submit" className={styles.submitBtn} disabled={isLoading || !title.trim()}>
              {isLoading ? "Posting..." : "Post"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
