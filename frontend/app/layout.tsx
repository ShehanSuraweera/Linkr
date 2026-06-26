import type { Metadata } from "next";
import { Geist } from "next/font/google";
import Providers from "./providers";
import "./globals.css";

const geist = Geist({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "Linkr",
  description: "URL shortener with analytics",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className={`${geist.className} bg-background text-foreground antialiased min-h-screen`}>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
