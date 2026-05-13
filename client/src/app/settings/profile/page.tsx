"use client";

import Image from "next/image";
import Link from "next/link";
import { useEffect, useState } from "react";
import { fetchWithAuth, logout } from "@/lib/auth";
import { getMyUsername } from "@/lib/jwt";
import { useRouter } from "next/navigation";
import styles from "./page.module.css";

interface ProfileForm {
  displayName: string;
  bio: string;
  avatar: string;
  banner: string;
}

export default function SettingsProfilePage() {
  const router = useRouter();
  const myUsername = getMyUsername();

  const [form, setForm] = useState<ProfileForm>({
    displayName: "",
    bio: "",
    avatar: "",
    banner: "",
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!myUsername) {
      router.replace("/");
      return;
    }
    fetchWithAuth("/users/me")
      .then((r) => r.json())
      .then((data) => {
        setForm({
          displayName: data.displayName ?? "",
          bio: data.bio ?? "",
          avatar: data.avatar ?? "",
          banner: data.banner ?? "",
        });
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, [myUsername, router]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setForm((f) => ({ ...f, [e.target.name]: e.target.value }));
    setSuccess(false);
    setError(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setError(null);
    setSuccess(false);

    const body: Record<string, string> = {};
    if (form.displayName) body.displayName = form.displayName;
    if (form.bio) body.bio = form.bio;
    if (form.avatar) body.avatar = form.avatar;
    if (form.banner) body.banner = form.banner;

    const res = await fetchWithAuth("/users/me", {
      method: "PATCH",
      body: JSON.stringify(body),
    }).catch(() => null);

    setSaving(false);
    if (!res || !res.ok) {
      const data = await res?.json().catch(() => ({}));
      setError(data?.message ?? "Failed to save changes.");
    } else {
      setSuccess(true);
    }
  };

  const handleLogout = async () => {
    await logout();
    router.replace("/");
  };

  if (loading) {
    return (
      <div className={styles.page}>
        <nav className={styles.topbar}>
          <Link href="/">
            <Image src="/reddit-1.svg" alt="Reddit" width={100} height={32} />
          </Link>
        </nav>
        <div className={styles.loading}>Loading…</div>
      </div>
    );
  }

  return (
    <div className={styles.page}>
      <nav className={styles.topbar}>
        <Link href="/">
          <Image src="/reddit-1.svg" alt="Reddit" width={100} height={32} />
        </Link>
        <div className={styles.topbarRight}>
          {myUsername && (
            <Link href={`/u/${myUsername}`} className={styles.navUser}>
              u/{myUsername}
            </Link>
          )}
          <button className={styles.logoutBtn} onClick={handleLogout}>
            Log Out
          </button>
        </div>
      </nav>

      <div className={styles.layout}>
        <aside className={styles.sidebar}>
          <h2 className={styles.sidebarTitle}>Settings</h2>
          <nav className={styles.sidebarNav}>
            <span className={styles.sidebarLinkActive}>Profile</span>
          </nav>
        </aside>

        <main className={styles.main}>
          <h1 className={styles.pageTitle}>Customize profile</h1>

          <form className={styles.form} onSubmit={handleSubmit}>
            <section className={styles.section}>
              <h2 className={styles.sectionTitle}>Profile information</h2>

              <label className={styles.field}>
                <span className={styles.label}>Display name (optional)</span>
                <span className={styles.hint}>
                  Set a display name. This does not change your username.
                </span>
                <input
                  className={styles.input}
                  name="displayName"
                  type="text"
                  maxLength={30}
                  value={form.displayName}
                  onChange={handleChange}
                  placeholder="Display name"
                />
              </label>

              <label className={styles.field}>
                <span className={styles.label}>About (optional)</span>
                <span className={styles.hint}>A brief description of yourself.</span>
                <textarea
                  className={styles.textarea}
                  name="bio"
                  maxLength={200}
                  value={form.bio}
                  onChange={handleChange}
                  placeholder="About"
                  rows={4}
                />
              </label>
            </section>

            <section className={styles.section}>
              <h2 className={styles.sectionTitle}>Images</h2>

              <label className={styles.field}>
                <span className={styles.label}>Avatar URL</span>
                <input
                  className={styles.input}
                  name="avatar"
                  type="url"
                  value={form.avatar}
                  onChange={handleChange}
                  placeholder="https://…"
                />
              </label>

              <label className={styles.field}>
                <span className={styles.label}>Banner URL</span>
                <input
                  className={styles.input}
                  name="banner"
                  type="url"
                  value={form.banner}
                  onChange={handleChange}
                  placeholder="https://…"
                />
              </label>
            </section>

            {error && <p className={styles.error}>{error}</p>}
            {success && <p className={styles.successMsg}>Profile saved!</p>}

            <div className={styles.formActions}>
              <Link href={`/u/${myUsername}`} className={styles.cancelBtn}>
                Cancel
              </Link>
              <button className={styles.saveBtn} type="submit" disabled={saving}>
                {saving ? "Saving…" : "Save"}
              </button>
            </div>
          </form>
        </main>
      </div>
    </div>
  );
}
