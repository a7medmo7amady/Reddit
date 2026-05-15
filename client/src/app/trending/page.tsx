"use client";
import { useEffect } from "react";
import { useRouter } from "next/navigation";

export default function Trending() {
  const router = useRouter();
  useEffect(() => { router.replace("/?tab=trending"); }, [router]);
  return null;
}
