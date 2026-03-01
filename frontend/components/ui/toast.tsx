"use client";

import * as React from "react";
import { cn } from "@/lib/utils";
import { X } from "lucide-react";

interface Toast {
    id: string;
    title: string;
    description?: string;
    variant?: "default" | "success" | "error";
}

interface ToastContextType {
    toasts: Toast[];
    addToast: (toast: Omit<Toast, "id">) => void;
    removeToast: (id: string) => void;
}

const ToastContext = React.createContext<ToastContextType | undefined>(undefined);

export function useToast() {
    const context = React.useContext(ToastContext);
    if (!context) throw new Error("useToast must be used within ToastProvider");
    return context;
}

export function ToastProvider({ children }: { children: React.ReactNode }) {
    const [toasts, setToasts] = React.useState<Toast[]>([]);

    const addToast = React.useCallback((toast: Omit<Toast, "id">) => {
        const id = Math.random().toString(36).slice(2);
        setToasts((prev) => [...prev, { ...toast, id }]);
        setTimeout(() => setToasts((prev) => prev.filter((t) => t.id !== id)), 5000);
    }, []);

    const removeToast = React.useCallback((id: string) => {
        setToasts((prev) => prev.filter((t) => t.id !== id));
    }, []);

    return (
        <ToastContext.Provider value={{ toasts, addToast, removeToast }}>
            {children}
            <div className="fixed bottom-4 right-4 z-50 flex flex-col gap-2">
                {toasts.map((toast) => (
                    <div
                        key={toast.id}
                        className={cn(
                            "min-w-[320px] rounded-lg border p-4 shadow-lg transition-all animate-in slide-in-from-right",
                            toast.variant === "error" ? "border-red-200 bg-red-50 text-red-900" :
                                toast.variant === "success" ? "border-green-200 bg-green-50 text-green-900" :
                                    "border-gray-200 bg-white text-gray-900"
                        )}
                    >
                        <div className="flex items-start justify-between">
                            <div>
                                <p className="font-medium">{toast.title}</p>
                                {toast.description && <p className="mt-1 text-sm opacity-80">{toast.description}</p>}
                            </div>
                            <button onClick={() => removeToast(toast.id)} className="ml-4 opacity-50 hover:opacity-100">
                                <X className="h-4 w-4" />
                            </button>
                        </div>
                    </div>
                ))}
            </div>
        </ToastContext.Provider>
    );
}
