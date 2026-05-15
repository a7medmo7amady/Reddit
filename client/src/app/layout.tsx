import type { Metadata } from "next";
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
      <body>{children}</body>
    </html>
  );
}
