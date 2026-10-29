import type { Metadata } from "next";
import "./globals.css";
import { I18nProvider } from "@/i18n/provider";
export const metadata: Metadata = { title: "Supay | Shopping, Stores & Wallet", description: "Supay connects customers, merchants, and wallet services in one everyday commerce platform." };
export default function RootLayout({children}: Readonly<{children:React.ReactNode}>) {
 return <html lang="vi"><body><I18nProvider>{children}</I18nProvider></body></html>;
}
