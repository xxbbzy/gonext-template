export const locales = ["zh-CN", "en"] as const;

export type Locale = (typeof locales)[number];

export const defaultLocale: Locale = "zh-CN";
export const localeCookieName = "locale";

export function getLocale(value?: string | null): Locale {
  if (value && locales.includes(value as Locale)) {
    return value as Locale;
  }

  return defaultLocale;
}

export async function getMessages(locale: Locale) {
  switch (locale) {
    case "en":
      return (await import("@/messages/en.json")).default;
    default:
      return (await import("@/messages/zh-CN.json")).default;
  }
}
