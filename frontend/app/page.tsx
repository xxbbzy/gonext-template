"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";

export default function Home() {
  const t = useTranslations("home");

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-gradient-to-br from-blue-50 via-white to-indigo-50">
      <div className="text-center space-y-6 max-w-2xl px-4">
        <div className="inline-flex items-center rounded-full border border-blue-200 bg-blue-50 px-4 py-1.5 text-sm text-blue-700 mb-4">
          {t("badge")}
        </div>
        <h1 className="text-5xl font-extrabold tracking-tight text-gray-900">
          Go<span className="text-blue-600">Next</span> Template
        </h1>
        <p className="text-lg text-gray-500 leading-relaxed">
          {t("description")}
        </p>
        <div className="flex gap-4 justify-center">
          <Link href="/login">
            <Button size="lg">{t("getStarted")}</Button>
          </Link>
          <Link href="https://github.com" target="_blank">
            <Button variant="outline" size="lg">
              {t("viewSource")}
            </Button>
          </Link>
        </div>
        <div className="grid grid-cols-3 gap-6 pt-8 text-sm">
          <div className="space-y-2">
            <div className="text-2xl">⚡</div>
            <p className="font-medium text-gray-900">{t("featureOneTitle")}</p>
            <p className="text-gray-500">{t("featureOneDesc")}</p>
          </div>
          <div className="space-y-2">
            <div className="text-2xl">🔒</div>
            <p className="font-medium text-gray-900">{t("featureTwoTitle")}</p>
            <p className="text-gray-500">{t("featureTwoDesc")}</p>
          </div>
          <div className="space-y-2">
            <div className="text-2xl">🤖</div>
            <p className="font-medium text-gray-900">
              {t("featureThreeTitle")}
            </p>
            <p className="text-gray-500">{t("featureThreeDesc")}</p>
          </div>
        </div>
      </div>
    </div>
  );
}
