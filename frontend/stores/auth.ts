import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";

export interface User {
    id: number;
    username: string;
    email: string;
    role: string;
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
                typeof window !== "undefined" ? localStorage : {
                    getItem: () => null,
                    setItem: () => { },
                    removeItem: () => { },
                }
            ),
        }
    )
);
