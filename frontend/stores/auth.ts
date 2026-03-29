import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";

import type { components } from "@/types/api";

type ApiUser = components["schemas"]["UserResponse"];

export type User = ApiUser;

export function toStoredUser(user: ApiUser): User {
  return {
    id: user.id,
    username: user.username,
    email: user.email,
    role: user.role,
  };
}

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  user: User | null;
  isAuthenticated: boolean;

  setTokens: (accessToken: string, refreshToken: string) => void;
  setUser: (user: User) => void;
  login: (accessToken: string, refreshToken: string, user: User) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: null,
      refreshToken: null,
      user: null,
      isAuthenticated: false,

      setTokens: (accessToken: string, refreshToken: string) =>
        set({ accessToken, refreshToken, isAuthenticated: true }),

      setUser: (user: User) => set({ user }),

      login: (accessToken: string, refreshToken: string, user: User) =>
        set({
          accessToken,
          refreshToken,
          user,
          isAuthenticated: true,
        }),

      logout: () =>
        set({
          accessToken: null,
          refreshToken: null,
          user: null,
          isAuthenticated: false,
        }),
    }),
    {
      name: "auth-storage",
      storage: createJSONStorage(() =>
        typeof window !== "undefined"
          ? localStorage
          : {
              getItem: () => null,
              setItem: () => {},
              removeItem: () => {},
            }
      ),
    }
  )
);
