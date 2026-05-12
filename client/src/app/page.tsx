"use client";

import Image from "next/image";
import { useState } from "react";
import styles from "./page.module.css";

type AuthMode = "login" | "signup";

export default function Home() {
  const [mode, setMode] = useState<AuthMode>("login");
  const isSignup = mode === "signup";

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
        <form className={styles.authCard}>
          <h1 id="auth-title">{isSignup ? "Sign Up" : "Log In"}</h1>

          <p className={styles.policyText}>
            By continuing, you agree to our{" "}
            <button type="button">User Agreement</button> and acknowledge that
            you understand the <button type="button">Privacy Policy</button>.
          </p>

          <div className={styles.oauthStack}>
            <button type="button">
              <Image
                className={styles.googleLogo}
                src="/google-g-2015.svg"
                alt=""
                width={22}
                height={22}
                aria-hidden="true"
              />
              Continue with Google
            </button>
          </div>

          <div className={styles.divider}>
            <span>OR</span>
          </div>

          <div className={styles.fields}>
            {isSignup && (
              <label>
                <input
                  autoComplete="username"
                  name="username"
                  placeholder=" "
                  type="text"
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
              />
              <span>{isSignup ? "Email *" : "Email or username *"}</span>
            </label>

            <label className={styles.passwordField}>
              <input
                autoComplete={isSignup ? "new-password" : "current-password"}
                name="password"
                placeholder=" "
                type="password"
              />
              <span>Password *</span>
              <button aria-label="Show password" type="button">
                o
              </button>
            </label>

            {isSignup && (
              <label className={styles.passwordField}>
                <input
                  autoComplete="new-password"
                  name="confirmPassword"
                  placeholder=" "
                  type="password"
                />
                <span>Confirm password *</span>
                <button aria-label="Show confirm password" type="button">
                  o
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
                onClick={() => setMode(isSignup ? "login" : "signup")}
              >
                {isSignup ? "Log In" : "Sign Up"}
              </button>
            </p>
          </div>

          <button className={styles.submitButton} type="submit">
            {isSignup ? "Sign Up" : "Log In"}
          </button>
        </form>
      </section>
    </main>
  );
}
