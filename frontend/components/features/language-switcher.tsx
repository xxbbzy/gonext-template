"use client";

import { useTransition } from "react";
import { useRouter } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import {
  defaultLocale,
  localeCookieName,
  locales,
  type Locale,
} from "@/lib/i18n";

function persistLocale(nextLocale: Locale) {
  globalThis.document.cookie = `${localeCookieName}=${nextLocale}; path=/; max-age=31536000; SameSite=Lax`;
}

export function LanguageSwitcher({ className }: { className?: string }) {
  const t = useTranslations("common");
  const locale = useLocale() as Locale;
  const router = useRouter();
  const [isPending, startTransition] = useTransition();

  const switchLocale = (nextLocale: Locale) => {
    if (nextLocale === locale) {
      return;
    }

    persistLocale(nextLocale);
    startTransition(() => {
      router.refresh();
    });
  };

  return (
    <div
      className={cn(
        "inline-flex items-center gap-1 rounded-full border border-gray-200 bg-white/90 p-1 shadow-sm backdrop-blur",
        className
      )}
    >
      {locales.map((item) => (
        <Button
          key={item}
          type="button"
          size="sm"
          variant={item === locale ? "default" : "ghost"}
          loading={isPending && item !== locale}
          onClick={() => switchLocale(item)}
        >
          {item === defaultLocale ? t("localeZh") : t("localeEn")}
        </Button>
      ))}
    </div>
  );
}
