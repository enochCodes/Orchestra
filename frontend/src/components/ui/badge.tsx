import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const badgeVariants = cva(
    "inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-zinc-400 focus:ring-offset-2",
    {
        variants: {
            variant: {
                default:
                    "border-transparent bg-indigo-500/15 text-indigo-400 hover:bg-indigo-500/25",
                secondary:
                    "border-transparent bg-zinc-800 text-zinc-300 hover:bg-zinc-700",
                destructive:
                    "border-transparent bg-red-500/15 text-red-400 hover:bg-red-500/25",
                success:
                    "border-transparent bg-emerald-500/15 text-emerald-400 hover:bg-emerald-500/25",
                warning:
                    "border-transparent bg-amber-500/15 text-amber-400 hover:bg-amber-500/25",
                outline: "text-zinc-300 border-zinc-700",
            },
        },
        defaultVariants: {
            variant: "default",
        },
    }
);

export interface BadgeProps
    extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> { }

function Badge({ className, variant, ...props }: BadgeProps) {
    return (
        <div className={cn(badgeVariants({ variant }), className)} {...props} />
    );
}

export { Badge, badgeVariants };
