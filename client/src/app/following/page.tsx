"use client";
import { useEffect } from "react";
import { useRouter } from "next/navigation";

export default function Following() {
  const router = useRouter();
  useEffect(() => { router.replace("/?tab=followed"); }, [router]);
  return null;
}
