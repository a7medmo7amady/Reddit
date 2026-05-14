import { useState, useRef, useEffect } from "react";
import styles from "./PostForm.module.css";
import { getToken } from "@/lib/auth";
import { getMyUsername } from "@/lib/jwt";

const API_URL = process.env.NEXT_PUBLIC_API_GATEWAY_URL ?? "http://localhost:8088";

interface PostFormProps {
  onPostCreated?: () => void;
  requireAuth?: () => void;
}

export default function PostForm({ onPostCreated, requireAuth }: PostFormProps) {
  const [expanded, setExpanded] = useState(false);
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [community, setCommunity] = useState("");
  const [mediaFile, setMediaFile] = useState<File | null>(null);
  const [mediaPreview, setMediaPreview] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [username, setUsername] = useState<string | null>(null);

  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => { setUsername(getMyUsername()); }, []);

  const handleFocus = () => {
    if (!getToken()) { requireAuth?.(); return; }
    setExpanded(true);
  };

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
    if (!getToken()) { requireAuth?.(); return; }
    if (!title.trim()) { setError("Title is required"); return; }

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
      setTitle(""); setBody(""); setCommunity(""); removeMedia(); setExpanded(false);
      onPostCreated?.();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "An error occurred");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className={styles.container}>
      {!expanded ? (
        <div className={styles.collapsed}>
          <div className={styles.avatarSmall}>
            {username ? username.charAt(0).toUpperCase() : "?"}
          </div>
          <input
            id="post-title-input"
            className={styles.promptInput}
            placeholder="Create Post"
            onFocus={handleFocus}
            readOnly
          />
          <button className={styles.iconBtn} title="Image" onClick={handleFocus}>Img</button>
          <button className={styles.iconBtn} title="Link"  onClick={handleFocus}>Link</button>
        </div>
      ) : (
        <form className={styles.expanded} onSubmit={handleSubmit}>
          <div className={styles.tabs}>
            <button type="button" className={`${styles.tabBtn} ${styles.tabActive}`}>Text</button>
            <button type="button" className={styles.tabBtn} onClick={handleFocus}>Images &amp; Video</button>
          </div>

          <input
            className={styles.communityInput}
            placeholder="Community (e.g. programming)"
            value={community}
            onChange={e => setCommunity(e.target.value)}
          />

          <input
            id="post-title-input"
            className={styles.titleInput}
            placeholder="Title *"
            value={title}
            onChange={e => setTitle(e.target.value)}
            maxLength={300}
            required
          />
          <div className={styles.titleCount}>{title.length}/300</div>

          <textarea
            className={styles.bodyInput}
            placeholder="Text (optional)"
            value={body}
            onChange={e => setBody(e.target.value)}
            rows={4}
          />

          {mediaPreview && (
            <div className={styles.preview}>
              <button type="button" className={styles.removeBtn} onClick={removeMedia}>×</button>
              {mediaFile?.type.startsWith("video/")
                ? <video src={mediaPreview} controls className={styles.previewMedia} />
                : <img src={mediaPreview} alt="preview" className={styles.previewMedia} />}
            </div>
          )}

          {error && <div className={styles.error}>{error}</div>}

          <div className={styles.formFooter}>
            <div className={styles.mediaActions}>
              <input type="file" ref={fileInputRef} style={{ display: "none" }} accept="image/*,video/*" onChange={handleMediaChange} />
              <button type="button" className={styles.iconBtn} onClick={() => fileInputRef.current?.click()} title="Attach media">Media</button>
            </div>
            <div className={styles.formActions}>
              <button type="button" className={styles.cancelBtn} onClick={() => { setExpanded(false); setError(null); }}>Cancel</button>
              <button type="submit" className={styles.submitBtn} disabled={isLoading || !title.trim()}>
                {isLoading ? "Posting..." : "Post"}
              </button>
            </div>
          </div>
        </form>
      )}
    </div>
  );
}
