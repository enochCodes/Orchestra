"use client";

import * as React from "react";
import { createPortal } from "react-dom";
import { X } from "lucide-react";
import { cn } from "@/lib/utils";

const DialogContext = React.createContext<{
    open: boolean;
    onOpenChange: (open: boolean) => void;
}>({
    open: false,
    onOpenChange: () => { },
});

interface DialogProps {
    open?: boolean;
    onOpenChange?: (open: boolean) => void;
    children: React.ReactNode;
}

export function Dialog({ open = false, onOpenChange = () => { }, children }: DialogProps) {
    return (
        <DialogContext.Provider value={{ open, onOpenChange }}>
            {open && <DialogPortal>{children}</DialogPortal>}
        </DialogContext.Provider>
    );
}

/** Renders dialog children into document.body via a portal so they escape any parent overflow/transform. */
function DialogPortal({ children }: { children: React.ReactNode }) {
    const [mounted, setMounted] = React.useState(false);
    React.useEffect(() => setMounted(true), []);
    if (!mounted) return null;
    return createPortal(children, document.body);
}

export function DialogContent({ className, children }: React.HTMLAttributes<HTMLDivElement>) {
    const { onOpenChange } = React.useContext(DialogContext);

    // Close on Escape key
    React.useEffect(() => {
        const handler = (e: KeyboardEvent) => {
            if (e.key === "Escape") onOpenChange(false);
        };
        document.addEventListener("keydown", handler);
        return () => document.removeEventListener("keydown", handler);
    }, [onOpenChange]);

    // Prevent body scroll while dialog is open
    React.useEffect(() => {
        document.body.style.overflow = "hidden";
        return () => { document.body.style.overflow = ""; };
    }, []);

    return (
        <div className="fixed inset-0 z-[9999] flex items-center justify-center p-4">
            {/* Backdrop */}
            <div
                className="fixed inset-0 bg-black/80 backdrop-blur-sm transition-opacity animate-in fade-in duration-200"
                onClick={() => onOpenChange(false)}
            />
            {/* Content — centered, scrollable when tall */}
            <div
                className={cn(
                    "relative z-[10000] w-full max-w-lg max-h-[85vh] overflow-y-auto",
                    "bg-black border border-zinc-800 rounded-xl shadow-2xl",
                    "animate-in zoom-in-95 duration-200 p-6",
                    className
                )}
            >
                {children}
                <button
                    onClick={() => onOpenChange(false)}
                    className="absolute right-4 top-4 rounded-sm opacity-70 transition-opacity hover:opacity-100 focus:outline-none"
                >
                    <X className="h-4 w-4 text-zinc-400" />
                    <span className="sr-only">Close</span>
                </button>
            </div>
        </div>
    );
}

export function DialogHeader({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
    return (
        <div className={cn("flex flex-col space-y-1.5 text-center sm:text-left mb-6", className)} {...props} />
    );
}

export function DialogTitle({ className, ...props }: React.HTMLAttributes<HTMLHeadingElement>) {
    return (
        <h3 className={cn("text-lg font-semibold leading-none tracking-tight text-white", className)} {...props} />
    );
}

export function DialogDescription({ className, ...props }: React.HTMLAttributes<HTMLParagraphElement>) {
    return (
        <p className={cn("text-sm text-zinc-400 mt-2", className)} {...props} />
    );
}

export function DialogFooter({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
    return (
        <div className={cn("flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2 mt-6", className)} {...props} />
    );
}
