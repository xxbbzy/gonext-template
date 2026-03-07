import type { Metadata } from "next";
import { cookies } from "next/headers";
import { NextIntlClientProvider } from "next-intl";
import "./globals.css";
import { QueryProvider } from "@/lib/query-provider";
import { ToastProvider } from "@/components/ui/toast";
import { LanguageSwitcher } from "@/components/features/language-switcher";
import { getLocale, getMessages, localeCookieName } from "@/lib/i18n";

export const metadata: Metadata = {
  title: "GoNext Template",
  description: "AI-Friendly 全栈项目脚手架，基于 Go + Next.js",
};

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const cookieStore = await cookies();
  const locale = getLocale(cookieStore.get(localeCookieName)?.value);
  const messages = await getMessages(locale);

  return (
    <html lang={locale}>
      <body className="antialiased">
        <NextIntlClientProvider locale={locale} messages={messages}>
          <div className="fixed right-4 top-4 z-50">
            <LanguageSwitcher />
          </div>
          <QueryProvider>
            <ToastProvider>{children}</ToastProvider>
          </QueryProvider>
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
