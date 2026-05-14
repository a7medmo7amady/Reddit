import { useState } from "react";
import Image from "next/image";
import styles from "./AuthPopup.module.css";
import { saveToken } from "@/lib/auth";

type AuthMode = "login" | "signup";

const API_URL =
  process.env.NEXT_PUBLIC_API_GATEWAY_URL ?? "http://localhost:8088";

interface AuthPopupProps {
  onClose: () => void;
  onSuccess: () => void;
}

export default function AuthPopup({ onClose, onSuccess }: AuthPopupProps) {
  const [mode, setMode] = useState<AuthMode>("login");
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const isSignup = mode === "signup";

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
          setIsLoading(false);
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
          setIsLoading(false);
          return;
        }
        saveToken(data.accessToken);
        onSuccess();
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
          setIsLoading(false);
          return;
        }
        saveToken(data.accessToken);
        onSuccess();
      }
    } catch {
      setError("Network error. Is the server running?");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.popup} onClick={(e) => e.stopPropagation()}>
        <button className={styles.closeButton} onClick={onClose} aria-label="Close">
          &times;
        </button>
        <div className={styles.header}>
          <Image
            className={styles.redditLogo}
            src="/reddit-1.svg"
            alt="Reddit"
            width={124}
            height={40}
            priority
          />
        </div>

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
      </div>
    </div>
  );
}
