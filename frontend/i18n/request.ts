import { cookies } from "next/headers";
import { getRequestConfig } from "next-intl/server";
import { getLocale, getMessages, localeCookieName } from "@/lib/i18n";

export default getRequestConfig(async () => {
  const cookieStore = await cookies();
  const locale = getLocale(cookieStore.get(localeCookieName)?.value);

  return {
    locale,
    messages: await getMessages(locale),
  };
});
