import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "MiniBroker",
  description: "High-throughput distributed message broker with consumer groups",
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
