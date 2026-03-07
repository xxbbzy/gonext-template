"use client";

import * as React from "react";
import { createPortal } from "react-dom";
import { X } from "lucide-react";
import { cn } from "@/lib/utils";

interface DialogContextValue {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

const DialogContext = React.createContext<DialogContextValue | null>(null);

function useDialogContext() {
  const context = React.useContext(DialogContext);
  if (!context) {
    throw new Error("Dialog components must be used within Dialog");
  }

  return context;
}

interface DialogProps {
  children: React.ReactNode;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

function Dialog({ children, open, onOpenChange }: DialogProps) {
  React.useEffect(() => {
    if (!open) {
      return;
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onOpenChange(false);
      }
    };

    document.body.style.overflow = "hidden";
    window.addEventListener("keydown", handleKeyDown);

    return () => {
      document.body.style.overflow = "";
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, [open, onOpenChange]);

  return (
    <DialogContext.Provider value={{ open, onOpenChange }}>
      {children}
    </DialogContext.Provider>
  );
}

function DialogTrigger({
  children,
  asChild,
}: {
  children: React.ReactElement<React.HTMLAttributes<HTMLElement>>;
  asChild?: boolean;
}) {
  const { onOpenChange } = useDialogContext();

  if (asChild) {
    return React.cloneElement(children, {
      onClick: (event: React.MouseEvent<HTMLElement>) => {
        children.props.onClick?.(event);
        onOpenChange(true);
      },
    });
  }

  return React.cloneElement(children, {
    onClick: () => onOpenChange(true),
  });
}

function DialogPortal({ children }: { children: React.ReactNode }) {
  const [mounted, setMounted] = React.useState(false);

  React.useEffect(() => {
    setMounted(true);
    return () => setMounted(false);
  }, []);

  if (!mounted) {
    return null;
  }

  return createPortal(children, document.body);
}

function DialogOverlay({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  const { open } = useDialogContext();

  if (!open) {
    return null;
  }

  return (
    <DialogPortal>
      <div
        className={cn(
          "fixed inset-0 z-50 bg-slate-950/35 backdrop-blur-[1px]",
          className
        )}
        {...props}
      />
    </DialogPortal>
  );
}

function DialogContent({
  children,
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  const { open, onOpenChange } = useDialogContext();

  if (!open) {
    return null;
  }

  return (
    <DialogPortal>
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
        <button
          type="button"
          aria-label="Close dialog"
          className="absolute inset-0 cursor-default"
          onClick={() => onOpenChange(false)}
        />
        <div
          role="dialog"
          aria-modal="true"
          className={cn(
            "relative z-10 w-full max-w-2xl rounded-2xl border border-slate-200 bg-white p-6 shadow-2xl",
            className
          )}
          {...props}
        >
          <button
            type="button"
            aria-label="Close dialog"
            className="absolute right-4 top-4 rounded-full p-1 text-slate-500 transition hover:bg-slate-100 hover:text-slate-900"
            onClick={() => onOpenChange(false)}
          >
            <X className="h-4 w-4" />
          </button>
          {children}
        </div>
      </div>
    </DialogPortal>
  );
}

function DialogHeader({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div className={cn("mb-6 flex flex-col gap-1.5", className)} {...props} />
  );
}

function DialogTitle({
  className,
  ...props
}: React.HTMLAttributes<HTMLHeadingElement>) {
  return (
    <h2
      className={cn(
        "text-xl font-semibold tracking-tight text-slate-900",
        className
      )}
      {...props}
    />
  );
}

function DialogDescription({
  className,
  ...props
}: React.HTMLAttributes<HTMLParagraphElement>) {
  return <p className={cn("text-sm text-slate-500", className)} {...props} />;
}

function DialogFooter({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn(
        "mt-6 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end",
        className
      )}
      {...props}
    />
  );
}

export {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogOverlay,
  DialogPortal,
  DialogTitle,
  DialogTrigger,
};
