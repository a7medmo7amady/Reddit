"use client";

import { useSearchParams } from "next/navigation";
import { useState, useEffect } from "react";
import Link from "next/link";
import { getToken } from "@/lib/auth";

const API_URL = process.env.NEXT_PUBLIC_API_GATEWAY_URL ?? "http://localhost:8088";

export default function SearchPage() {
  const searchParams = useSearchParams();
  const q = searchParams.get("q") || "";
  const [results, setResults] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!q) return;
    setLoading(true);
    setError("");

    const headers: HeadersInit = {};
    const token = getToken();
    if (token) headers["Authorization"] = `Bearer ${token}`;

    fetch(`${API_URL}/search?q=${encodeURIComponent(q)}`, { headers })
      .then(async (res) => {
        if (!res.ok) throw new Error(await res.text());
        return res.json();
      })
      .then((data) => setResults(data))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [q]);

  return (
    <div style={{ maxWidth: 800, margin: "0 auto", padding: 24 }}>
      <h1>Search Results for &quot;{q}&quot;</h1>
      {loading && <p>Loading...</p>}
      {error && <p style={{ color: "red" }}>{error}</p>}

      {results && (
        <div>
          {results.type === "users" && (
            <div>
              <h2>Users (from Postgres)</h2>
              {(results.results || []).length === 0 ? (
                <p>No users found.</p>
              ) : (
                <ul>
                  {(results.results || []).map((user: any) => (
                    <li key={user.id || user.username}>
                      <Link href={`/u/${user.username}`}>u/{user.username}</Link>
                      {user.karma !== undefined && <span> — Karma: {user.karma}</span>}
                    </li>
                  ))}
                </ul>
              )}
            </div>
          )}

          {results.type === "posts" && (
            <div>
              <h2>Posts (from MongoDB)</h2>
              {(results.results?.posts || []).length === 0 ? (
                <p>No posts found.</p>
              ) : (
                <ul>
                  {(results.results?.posts || []).map((post: any) => (
                    <li key={post.id}>
                      <Link href={`/posts/${post.id}`}>{post.title}</Link>
                      <span> — r/{post.community} by u/{post.author}</span>
                    </li>
                  ))}
                </ul>
              )}

              <h2>Comments (from MongoDB)</h2>
              {(results.results?.comments || []).length === 0 ? (
                <p>No comments found.</p>
              ) : (
                <ul>
                  {(results.results?.comments || []).map((comment: any) => (
                    <li key={comment.id}>
                      <span>{comment.body?.slice(0, 120)}...</span>
                      <span> — by u/{comment.author}</span>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
