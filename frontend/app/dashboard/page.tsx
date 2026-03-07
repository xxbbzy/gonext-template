"use client";

import { useTranslations } from "next-intl";
import { useAuthStore } from "@/stores/auth";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Package, Upload, Users } from "lucide-react";

export default function DashboardPage() {
  const tDashboard = useTranslations("dashboard");
  const user = useAuthStore((s) => s.user);

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-gray-900">
          {tDashboard("welcome")}, {user?.username}
        </h2>
        <p className="mt-1 text-gray-500">{tDashboard("overview")}</p>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">
              {tDashboard("totalItems")}
            </CardTitle>
            <Package className="h-4 w-4 text-gray-400" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">—</div>
            <p className="text-xs text-gray-500 mt-1">
              {tDashboard("itemsHint")}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">
              {tDashboard("upload")}
            </CardTitle>
            <Upload className="h-4 w-4 text-gray-400" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">—</div>
            <p className="text-xs text-gray-500 mt-1">
              {tDashboard("uploadHint")}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">
              {tDashboard("role")}
            </CardTitle>
            <Users className="h-4 w-4 text-gray-400" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold capitalize">{user?.role}</div>
            <p className="text-xs text-gray-500 mt-1">
              {tDashboard("roleHint")}
            </p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{tDashboard("quickStart")}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3 text-sm text-gray-600">
          <p>{tDashboard("quickStartLineOne")}</p>
          <p>{tDashboard("quickStartLineTwo")}</p>
          <p>{tDashboard("quickStartLineThree")}</p>
          <p>{tDashboard("quickStartLineFour")}</p>
        </CardContent>
      </Card>
    </div>
  );
}
