import os from "os";
import path from "path";
import type { NextConfig } from "next";
import createNextIntlPlugin from "next-intl/plugin";

const withNextIntl = createNextIntlPlugin("./i18n/request.ts");

function getAllowedDevOrigins() {
  const port = process.env.PORT ?? "3000";
  const extraOrigins = (process.env.NEXT_ALLOWED_DEV_ORIGINS ?? "")
    .split(",")
    .map((origin) => origin.trim())
    .filter(Boolean);

  let interfaceOrigins: string[] = [];
  try {
    interfaceOrigins = Object.values(os.networkInterfaces())
      .flat()
      .filter((details): details is NonNullable<typeof details> =>
        Boolean(details)
      )
      .filter((details) => details.family === "IPv4" && !details.internal)
      .map((details) => `http://${details.address}:${port}`);
  } catch {
    interfaceOrigins = [];
  }

  return Array.from(
    new Set([
      `http://localhost:${port}`,
      `http://127.0.0.1:${port}`,
      ...interfaceOrigins,
      ...extraOrigins,
    ])
  );
}

const nextConfig: NextConfig = {
  allowedDevOrigins: getAllowedDevOrigins(),
  output: "standalone",
  turbopack: {
    root: path.resolve(__dirname),
  },
};

export default withNextIntl(nextConfig);
