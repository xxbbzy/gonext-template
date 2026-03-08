"use client";

import { Suspense, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import Link from "next/link";
import { useTranslations } from "next-intl";
import { client } from "@/lib/api-client.gen";
import { useAuthStore, User } from "@/stores/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

function LoginForm() {
  const tCommon = useTranslations("common");
  const tAuth = useTranslations("auth");
  const loginSchema = z.object({
    email: z.string().email(tAuth("emailInvalid")),
    password: z.string().min(1, tAuth("passwordRequired")),
  });

  type LoginForm = z.infer<typeof loginSchema>;
  const router = useRouter();
  const searchParams = useSearchParams();
  const login = useAuthStore((s) => s.login);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = async (data: LoginForm) => {
    setLoading(true);
    setError("");
    try {
      const { data: res, error: apiError } = await client.POST(
        "/api/v1/auth/login",
        { body: { email: data.email, password: data.password } }
      );
      if (apiError || !res?.data) {
        setError(
          (apiError as { message?: string })?.message || tAuth("loginFailed")
        );
        return;
      }
      const { access_token, refresh_token, user } = res.data as {
        access_token: string;
        refresh_token: string;
        user: User;
      };
      login(access_token, refresh_token, user);

      const redirect = searchParams.get("redirect") || "/dashboard";
      router.push(redirect);
    } catch {
      setError(tAuth("loginFailed"));
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card className="w-full max-w-md">
      <CardHeader className="text-center">
        <CardTitle className="text-2xl font-bold">{tCommon("login")}</CardTitle>
        <CardDescription>{tAuth("loginDesc")}</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          {error && (
            <div className="rounded-lg bg-red-50 border border-red-200 p-3 text-sm text-red-600">
              {error}
            </div>
          )}
          <div className="space-y-2">
            <label className="text-sm font-medium text-gray-700">
              {tAuth("email")}
            </label>
            <Input
              type="email"
              placeholder="your@email.com"
              error={errors.email?.message}
              {...register("email")}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium text-gray-700">
              {tAuth("password")}
            </label>
            <Input
              type="password"
              placeholder="••••••••"
              error={errors.password?.message}
              {...register("password")}
            />
          </div>
          <Button type="submit" className="w-full" loading={loading}>
            {tCommon("login")}
          </Button>
          <p className="text-center text-sm text-gray-500">
            {tAuth("noAccount")}{" "}
            <Link href="/register" className="text-blue-600 hover:underline">
              {tCommon("register")}
            </Link>
          </p>
        </form>
      </CardContent>
    </Card>
  );
}

export default function LoginPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 px-4">
      <Suspense
        fallback={
          <div className="flex items-center justify-center">
            <div className="animate-spin h-8 w-8 border-4 border-blue-600 border-t-transparent rounded-full" />
          </div>
        }
      >
        <LoginForm />
      </Suspense>
    </div>
  );
}
