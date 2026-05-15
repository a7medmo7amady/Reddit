import { useState } from "react";
import { useRouter } from "next/navigation";
import styles from "./CreateCommunityPopup.module.css";
import { getToken } from "@/lib/auth";

const API_URL = process.env.NEXT_PUBLIC_API_GATEWAY_URL ?? "http://localhost:8088";

interface CommunityResult {
  id: number;
  name: string;
  description: string;
  memberCount: number;
}

interface CreateCommunityPopupProps {
  onClose: () => void;
  onSuccess: (community: CommunityResult) => void;
}

export default function CreateCommunityPopup({ onClose, onSuccess }: CreateCommunityPopupProps) {
  const router = useRouter();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;

    setIsLoading(true);
    setError(null);

    try {
      const token = getToken();
      if (!token) {
        setError("You must be logged in to create a community.");
        setIsLoading(false);
        return;
      }

      const res = await fetch(`${API_URL}/communities`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ name: name.trim(), description: description.trim() }),
      });

      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        setError(data.message ?? "Failed to create community. It might already exist or the name is invalid.");
        setIsLoading(false);
        return;
      }

      const community: CommunityResult = await res.json();
      onSuccess(community);
      router.push(`/r/${community.name}`);
    } catch (err) {
      setError("Network error. Please try again.");
      setIsLoading(false);
    }
  };

  const remainingChars = 21 - name.length;

  return (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.popup} onClick={(e) => e.stopPropagation()}>
        <button className={styles.closeButton} onClick={onClose} aria-label="Close">
          &times;
        </button>
        
        <div className={styles.header}>
          <h1>Create a community</h1>
        </div>

        <form className={styles.form} onSubmit={handleSubmit}>
          {error && <div className={styles.errorText}>{error}</div>}

          <div className={styles.fieldGroup}>
            <label htmlFor="community-name">Name</label>
            <div className={styles.subLabel}>
              Community names including capitalization cannot be changed.
            </div>
            <div className={styles.inputWrapper}>
              <span className={styles.prefix}>r/</span>
              <input
                id="community-name"
                type="text"
                value={name}
                onChange={(e) => {
                  const val = e.target.value.replace(/\s+/g, ''); // Remove spaces
                  if (val.length <= 21) setName(val);
                }}
                className={styles.inputWithPrefix}
                maxLength={21}
                required
                autoComplete="off"
              />
            </div>
            <div className={`${styles.charCount} ${remainingChars === 0 ? styles.textDanger : ""}`}>
              {remainingChars} Characters remaining
            </div>
          </div>

          <div className={styles.fieldGroup}>
            <label htmlFor="community-desc">Description <span className={styles.subLabel}>(optional)</span></label>
            <textarea
              id="community-desc"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="What is your community about?"
              maxLength={500}
            />
          </div>
          
          <div className={styles.footer}>
            <button type="button" className={styles.cancelBtn} onClick={onClose} disabled={isLoading}>
              Cancel
            </button>
            <button type="submit" className={styles.submitBtn} disabled={isLoading || name.length < 3}>
              {isLoading ? "Creating..." : "Create Community"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
