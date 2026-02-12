"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";

const links = [
    { href: "/", label: "Overview" },
    { href: "/applications", label: "Applications" }, // Promoted as primary
    { href: "/deployments", label: "Deployments" },
    { href: "/servers", label: "Servers" },
    { href: "/clusters", label: "Clusters" },
    { href: "/monitoring", label: "Monitoring" },
    { href: "/settings", label: "Settings" },
];

export function Navigation() {
    const pathname = usePathname();

    return (
        <div className="border-b border-zinc-800 bg-black">
            <div className="flex h-12 items-center px-6 max-w-[1400px] mx-auto overflow-x-auto no-scrollbar">
                <nav className="flex items-center space-x-6 text-sm font-medium">
                    {links.map((link) => {
                        const isActive = pathname === link.href;
                        return (
                            <Link
                                key={link.href}
                                href={link.href}
                                className={cn(
                                    "transition-colors hover:text-white whitespace-nowrap pb-3.5 pt-3 border-b-2",
                                    isActive
                                        ? "border-white text-white"
                                        : "border-transparent text-zinc-400"
                                )}
                            >
                                {link.label}
                            </Link>
                        );
                    })}
                </nav>
            </div>
        </div>
    );
}
