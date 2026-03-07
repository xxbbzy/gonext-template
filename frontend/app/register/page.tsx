"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import Link from "next/link";
import { useTranslations } from "next-intl";
import apiClient, { ApiResponse } from "@/lib/api-client";
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

interface AuthData {
  access_token: string;
  refresh_token: string;
  user: User;
}

export default function RegisterPage() {
  const tCommon = useTranslations("common");
  const tAuth = useTranslations("auth");
  const registerSchema = z
    .object({
      username: z
        .string()
        .min(3, tAuth("usernameMin"))
        .max(50, tAuth("usernameMax")),
      email: z.string().email(tAuth("emailInvalid")),
      password: z
        .string()
        .min(6, tAuth("passwordMin"))
        .max(100, tAuth("passwordMax")),
      confirmPassword: z.string(),
    })
    .refine((data) => data.password === data.confirmPassword, {
      message: tAuth("passwordMismatch"),
      path: ["confirmPassword"],
    });

  type RegisterForm = z.infer<typeof registerSchema>;
  const router = useRouter();
  const login = useAuthStore((s) => s.login);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterForm>({
    resolver: zodResolver(registerSchema),
  });

  const onSubmit = async (data: RegisterForm) => {
    setLoading(true);
    setError("");
    try {
      const res = await apiClient.post<ApiResponse<AuthData>>(
        "/api/v1/auth/register",
        {
          username: data.username,
          email: data.email,
          password: data.password,
        }
      );
      const { access_token, refresh_token, user } = res.data.data;
      login(access_token, refresh_token, user);
      router.push("/dashboard");
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { message?: string } } };
      setError(axiosErr.response?.data?.message || tAuth("registerFailed"));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 px-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl font-bold">
            {tCommon("register")}
          </CardTitle>
          <CardDescription>{tAuth("registerDesc")}</CardDescription>
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
                {tAuth("username")}
              </label>
              <Input
                placeholder="your_username"
                error={errors.username?.message}
                {...register("username")}
              />
            </div>
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
            <div className="space-y-2">
              <label className="text-sm font-medium text-gray-700">
                {tAuth("confirmPassword")}
              </label>
              <Input
                type="password"
                placeholder="••••••••"
                error={errors.confirmPassword?.message}
                {...register("confirmPassword")}
              />
            </div>
            <Button type="submit" className="w-full" loading={loading}>
              {tCommon("register")}
            </Button>
            <p className="text-center text-sm text-gray-500">
              {tAuth("hasAccount")}{" "}
              <Link href="/login" className="text-blue-600 hover:underline">
                {tCommon("login")}
              </Link>
            </p>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
