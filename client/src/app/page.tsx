"use client";

import Image from "next/image";
import { useState, useEffect } from "react";
import styles from "./page.module.css";
import Link from "next/link";
import { saveToken, getToken, logout } from "@/lib/auth";
import { getMyUsername } from "@/lib/jwt";

type AuthMode = "login" | "signup";

const API_URL =
  process.env.NEXT_PUBLIC_API_GATEWAY_URL ?? "http://localhost:8088";

export default function Home() {
  const [mode, setMode] = useState<AuthMode>("login");
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isAuthed, setIsAuthed] = useState(false);

  const isSignup = mode === "signup";

  // Restore auth state from cookie, or pick up the token the OAuth redirect drops in the URL
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const urlToken = params.get("accessToken");

    if (urlToken) {
      saveToken(urlToken);
      setIsAuthed(true);
      const clean = new URL(window.location.href);
      clean.searchParams.delete("accessToken");
      window.history.replaceState({}, "", clean.toString());
      return;
    }

    if (getToken()) {
      setIsAuthed(true);
    }
  }, []);

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError(null);
    setIsLoading(true);

    const formData = new FormData(e.currentTarget);
    const email = formData.get("email") as string;
    const password = formData.get("password") as string;

    try {
      if (isSignup) {
        const confirmPassword = formData.get("confirmPassword") as string;
        if (password !== confirmPassword) {
          setError("Passwords do not match");
          return;
        }

        const res = await fetch(`${API_URL}/auth/signup`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
          body: JSON.stringify({
            username: formData.get("username"),
            email,
            password,
          }),
        });

        const data = await res.json().catch(() => ({}));
        if (!res.ok) {
          setError(data.message ?? "Signup failed. Please try again.");
          return;
        }
        saveToken(data.accessToken);
        setIsAuthed(true);
      } else {
        const res = await fetch(`${API_URL}/auth/login`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
          body: JSON.stringify({ identifier: email, password }),
        });

        const data = await res.json().catch(() => ({}));
        if (!res.ok) {
          setError(data.message ?? "Invalid email or password.");
          return;
        }
        saveToken(data.accessToken);
        setIsAuthed(true);
      }
    } catch {
      setError("Network error. Is the server running?");
    } finally {
      setIsLoading(false);
    }
  };

  const handleLogout = async () => {
    await logout();
    setIsAuthed(false);
  };

  if (isAuthed) {
    const username = getMyUsername();
    return (
      <main className={styles.page}>
        <header className={styles.header}>
          <div className={styles.brand}>
            <Image
              className={styles.redditLogo}
              src="/reddit-1.svg"
              alt="Reddit"
              width={124}
              height={40}
              priority
            />
          </div>
        </header>
        <section className={styles.authShell}>
          <div className={styles.authCard}>
            <h1>You&apos;re in!</h1>
            {username && (
              <Link
                href={`/u/${username}`}
                style={{ display: "block", marginBottom: 12, color: "#ff4500" }}
              >
                View your profile →
              </Link>
            )}
            <button className={styles.submitButton} onClick={handleLogout}>
              Log Out
            </button>
          </div>
        </section>
      </main>
    );
  }

  return (
    <main className={styles.page}>
      <header className={styles.header}>
        <div className={styles.brand}>
          <Image
            className={styles.redditLogo}
            src="/reddit-1.svg"
            alt="Reddit"
            width={124}
            height={40}
            priority
          />
        </div>
      </header>

      <section className={styles.authShell} aria-labelledby="auth-title">
        <form className={styles.authCard} onSubmit={handleSubmit}>
          <h1 id="auth-title">{isSignup ? "Sign Up" : "Log In"}</h1>

          <p className={styles.policyText}>
            By continuing, you agree to our{" "}
            <button type="button">User Agreement</button> and acknowledge that
            you understand the <button type="button">Privacy Policy</button>.
          </p>

          <div className={styles.oauthStack}>
            <a href={`${API_URL}/oauth2/authorization/google`}>
              <Image
                className={styles.googleLogo}
                src="/google-g-2015.svg"
                alt=""
                width={22}
                height={22}
                aria-hidden="true"
              />
              Continue with Google
            </a>
          </div>

          <div className={styles.divider}>
            <span>OR</span>
          </div>

          {error && (
            <p role="alert" className={styles.errorText}>
              {error}
            </p>
          )}

          <div className={styles.fields}>
            {isSignup && (
              <label>
                <input
                  autoComplete="username"
                  name="username"
                  placeholder=" "
                  type="text"
                  required
                />
                <span>Username *</span>
              </label>
            )}

            <label>
              <input
                autoComplete={isSignup ? "email" : "username"}
                name="email"
                placeholder=" "
                type={isSignup ? "email" : "text"}
                required
              />
              <span>{isSignup ? "Email *" : "Email or username *"}</span>
            </label>

            <label className={styles.passwordField}>
              <input
                autoComplete={isSignup ? "new-password" : "current-password"}
                name="password"
                placeholder=" "
                type={showPassword ? "text" : "password"}
                required
              />
              <span>Password *</span>
              <button
                aria-label={showPassword ? "Hide password" : "Show password"}
                type="button"
                onClick={() => setShowPassword((v) => !v)}
              >
                {showPassword ? "Hide" : "Show"}
              </button>
            </label>

            {isSignup && (
              <label className={styles.passwordField}>
                <input
                  autoComplete="new-password"
                  name="confirmPassword"
                  placeholder=" "
                  type={showConfirmPassword ? "text" : "password"}
                  required
                />
                <span>Confirm password *</span>
                <button
                  aria-label={
                    showConfirmPassword
                      ? "Hide confirm password"
                      : "Show confirm password"
                  }
                  type="button"
                  onClick={() => setShowConfirmPassword((v) => !v)}
                >
                  {showConfirmPassword ? "Hide" : "Show"}
                </button>
              </label>
            )}
          </div>

          <div className={styles.accountLinks}>
            {!isSignup && <button type="button">Forgot password?</button>}
            <p>
              {isSignup ? "Already a redditor?" : "New to Reddit?"}{" "}
              <button
                type="button"
                onClick={() => {
                  setMode(isSignup ? "login" : "signup");
                  setError(null);
                }}
              >
                {isSignup ? "Log In" : "Sign Up"}
              </button>
            </p>
          </div>

          <button
            className={styles.submitButton}
            type="submit"
            disabled={isLoading}
          >
            {isLoading ? "Please wait…" : isSignup ? "Sign Up" : "Log In"}
          </button>
        </form>
      </section>
    </main>
  );
}
