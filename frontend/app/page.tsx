import Link from "next/link";
import { Button } from "@/components/ui/button";

export default function Home() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-gradient-to-br from-blue-50 via-white to-indigo-50">
      <div className="text-center space-y-6 max-w-2xl px-4">
        <div className="inline-flex items-center rounded-full border border-blue-200 bg-blue-50 px-4 py-1.5 text-sm text-blue-700 mb-4">
          🚀 AI-Friendly Full-Stack Scaffold
        </div>
        <h1 className="text-5xl font-extrabold tracking-tight text-gray-900">
          Go<span className="text-blue-600">Next</span> Template
        </h1>
        <p className="text-lg text-gray-500 leading-relaxed">
          基于 Go + Next.js 的全栈项目脚手架，强调约定优于配置、端到端类型安全、一键启动。
          采用 Gin + GORM + Wire 后端架构和 Next.js 15 + TypeScript + shadcn/ui 前端技术栈。
        </p>
        <div className="flex gap-4 justify-center">
          <Link href="/login">
            <Button size="lg">开始使用</Button>
          </Link>
          <Link href="https://github.com" target="_blank">
            <Button variant="outline" size="lg">查看源码</Button>
          </Link>
        </div>
        <div className="grid grid-cols-3 gap-6 pt-8 text-sm">
          <div className="space-y-2">
            <div className="text-2xl">⚡</div>
            <p className="font-medium text-gray-900">一键启动</p>
            <p className="text-gray-500">make dev 即可启动全栈开发环境</p>
          </div>
          <div className="space-y-2">
            <div className="text-2xl">🔒</div>
            <p className="font-medium text-gray-900">端到端类型安全</p>
            <p className="text-gray-500">Go → OpenAPI → TypeScript 全链路类型覆盖</p>
          </div>
          <div className="space-y-2">
            <div className="text-2xl">🤖</div>
            <p className="font-medium text-gray-900">AI 友好</p>
            <p className="text-gray-500">强约定架构，AI 可准确理解和生成代码</p>
          </div>
        </div>
      </div>
    </div>
  );
}
