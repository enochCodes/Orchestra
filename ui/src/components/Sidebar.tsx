"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
    LayoutDashboard,
    Server,
    Network,
    AppWindow,
    Activity,
    Settings,
    History,
    Key,
} from "lucide-react";
import { cn } from "@/lib/utils";

const links = [
    { href: "/", label: "Dashboard", icon: LayoutDashboard },
    { href: "/servers", label: "Servers", icon: Server },
    { href: "/clusters", label: "Clusters", icon: Network },
    { href: "/applications", label: "Applications", icon: AppWindow },
    { href: "/environments", label: "Environments", icon: Key },
    { href: "/deployments", label: "Deployments", icon: History },
    { href: "/monitoring", label: "Monitoring", icon: Activity },
    { href: "/settings", label: "Settings", icon: Settings },
];

export function Sidebar() {
    const pathname = usePathname();

    return (
        <div className="w-64 border-r border-zinc-800 bg-black flex flex-col h-screen sticky top-0">
            <div className="p-6">
                <div className="flex items-center gap-2 mb-8">
                    <div className="w-8 h-8 bg-white rounded-full flex items-center justify-center">
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                            <path d="M12 2L2 22H22L12 2Z" fill="black" />
                        </svg>
                    </div>
                    <span className="font-bold text-lg text-white tracking-tight">Orchestra</span>
                </div>

                <nav className="space-y-1">
                    {links.map((link) => {
                        const isActive = pathname === link.href;
                        return (
                            <Link
                                key={link.href}
                                href={link.href}
                                className={cn(
                                    "flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-all group",
                                    isActive
                                        ? "bg-zinc-900 text-white"
                                        : "text-zinc-400 hover:text-white hover:bg-zinc-900/50"
                                )}
                            >
                                <link.icon size={18} className={cn("transition-colors", isActive ? "text-indigo-400" : "text-zinc-500 group-hover:text-zinc-300")} />
                                {link.label}
                            </Link>
                        );
                    })}
                </nav>
            </div>

            <div className="mt-auto p-6 border-t border-zinc-800">
                <div className="flex items-center gap-3">
                    <div className="h-8 w-8 rounded-full bg-gradient-to-tr from-indigo-500 to-purple-500 flex items-center justify-center text-white text-xs font-bold ring-1 ring-zinc-700">
                        EC
                    </div>
                    <div className="flex flex-col">
                        <span className="text-sm font-medium text-white">enochCodes</span>
                        <span className="text-xs text-zinc-500">Pro Plan</span>
                    </div>
                </div>
            </div>
        </div>
    );
}
