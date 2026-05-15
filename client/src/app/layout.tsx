import type { Metadata } from "next";
import { Suspense } from "react";
import "./globals.css";

export const metadata: Metadata = {
  title: "Reddit | Log In or Sign Up",
  description: "Log in or create a Reddit account.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <Suspense fallback={<div />}>{children}</Suspense>
      </body>
    </html>
  );
}
