"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";

interface NavItem {
    label: string;
    href: string;
}

interface SubNavProps {
    items: NavItem[];
}

export function SubNav({ items }: SubNavProps) {
    const pathname = usePathname();

    return (
        <div className="border-b border-zinc-800 mb-8 overflow-x-auto no-scrollbar">
            <nav className="flex items-center gap-6">
                {items.map((item) => {
                    const isActive = pathname === item.href;
                    return (
                        <Link
                            key={item.href}
                            href={item.href}
                            className={cn(
                                "pb-3 text-sm font-medium transition-colors border-b-2 relative -mb-[1px]",
                                isActive
                                    ? "text-white border-white"
                                    : "text-zinc-500 border-transparent hover:text-zinc-300"
                            )}
                        >
                            {item.label}
                        </Link>
                    )
                })}
            </nav>
        </div>
    );
}
