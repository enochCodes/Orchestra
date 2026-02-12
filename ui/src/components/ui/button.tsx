import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";
import { Loader2 } from "lucide-react";

const buttonVariants = cva(
    "inline-flex items-center justify-center rounded-md text-sm font-medium transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-zinc-500 disabled:pointer-events-none disabled:opacity-50 active:scale-95",
    {
        variants: {
            variant: {
                default: "bg-white text-black hover:bg-zinc-200 border border-transparent shadow-sm", // Vercel Primary
                destructive: "bg-red-500 text-white hover:bg-red-600 shadow-sm",
                outline: "border border-zinc-800 bg-black text-zinc-200 hover:text-white hover:border-zinc-700", // Vercel Secondary
                secondary: "bg-zinc-900 text-zinc-200 hover:bg-zinc-800 border border-zinc-800",
                ghost: "hover:bg-zinc-900 text-zinc-400 hover:text-white",
                link: "text-blue-500 underline-offset-4 hover:underline",
            },
            size: {
                default: "h-9 px-4 py-2", // Slightly shorter than standard 10
                sm: "h-8 rounded-md px-3 text-xs",
                lg: "h-10 rounded-md px-8",
                icon: "h-9 w-9",
            },
        },
        defaultVariants: {
            variant: "default",
            size: "default",
        },
    }
);

export interface ButtonProps
    extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
    asChild?: boolean;
    isLoading?: boolean;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
    ({ className, variant, size, asChild = false, isLoading, children, ...props }, ref) => {
        return (
            <button
                className={cn(buttonVariants({ variant, size, className }))}
                ref={ref}
                disabled={isLoading || props.disabled}
                {...props}
            >
                {isLoading && <Loader2 className="mr-2 h-3 w-3 animate-spin" />}
                {children}
            </button>
        );
    }
);
Button.displayName = "Button";

export { Button, buttonVariants };
