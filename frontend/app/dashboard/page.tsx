"use client";

import { useAuthStore } from "@/stores/auth";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Package, Upload, Users } from "lucide-react";

export default function DashboardPage() {
    const user = useAuthStore((s) => s.user);

    return (
        <div className="space-y-6">
            <div>
                <h2 className="text-2xl font-bold text-gray-900">
                    欢迎回来，{user?.username}
                </h2>
                <p className="mt-1 text-gray-500">这是你的项目仪表盘概览</p>
            </div>

            <div className="grid gap-4 md:grid-cols-3">
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-sm font-medium text-gray-500">项目总数</CardTitle>
                        <Package className="h-4 w-4 text-gray-400" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">—</div>
                        <p className="text-xs text-gray-500 mt-1">使用 API 获取真实数据</p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-sm font-medium text-gray-500">上传文件</CardTitle>
                        <Upload className="h-4 w-4 text-gray-400" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">—</div>
                        <p className="text-xs text-gray-500 mt-1">文件上传统计</p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                        <CardTitle className="text-sm font-medium text-gray-500">角色</CardTitle>
                        <Users className="h-4 w-4 text-gray-400" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold capitalize">{user?.role}</div>
                        <p className="text-xs text-gray-500 mt-1">当前用户角色</p>
                    </CardContent>
                </Card>
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>快速开始</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3 text-sm text-gray-600">
                    <p>🚀 这是 GoNext Template 脚手架的示例仪表盘页面。</p>
                    <p>📦 访问「项目管理」查看 CRUD 示例。</p>
                    <p>📤 访问「文件上传」体验文件上传功能。</p>
                    <p>🔧 修改 <code className="rounded bg-gray-100 px-1.5 py-0.5 font-mono text-xs">frontend/app/dashboard/page.tsx</code> 自定义此页面。</p>
                </CardContent>
            </Card>
        </div>
    );
}
