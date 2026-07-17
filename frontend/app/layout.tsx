import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Free Resume Generator",
  description: "Build, manage, and export polished resumes",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
