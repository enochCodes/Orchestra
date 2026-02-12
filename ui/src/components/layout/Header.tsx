"use client";

import { useState, useRef, useEffect } from "react";
import { Command, ChevronDown, Bell, HelpCircle, User, LogOut, Settings, MessageSquare, FileText } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/lib/auth-context";
import Link from "next/link";

function UserAvatar({ name }: { name: string }) {
  const initials = name
    .split(" ")
    .map((n) => n[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();
  return (
    <div className="h-8 w-8 rounded-full bg-gradient-to-tr from-indigo-500 to-purple-500 flex items-center justify-center text-white text-xs font-bold ring-2 ring-black">
      {initials || "U"}
    </div>
  );
}

export function Header() {
  const [profileOpen, setProfileOpen] = useState(false);
  const { user, logout } = useAuth();
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setProfileOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const displayName = user?.display_name || user?.email?.split("@")[0] || "User";

  return (
    <header className="border-b border-zinc-800 bg-black">
      <div className="flex h-16 items-center px-6 max-w-[1400px] mx-auto justify-between">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <span className="text-zinc-500 hover:text-white cursor-pointer transition-colors">enochCodes</span>
            <span className="text-zinc-800">/</span>
            <span className="font-semibold text-sm text-white">Orchestra</span>
          </div>
          <div className="h-4 w-px bg-zinc-800 mx-2" />
          <Button variant="ghost" size="sm" className="h-8 gap-2 bg-transparent border-none text-zinc-400 hover:text-white hover:bg-zinc-900 px-2">
            <span className="w-4 h-4 rounded-full bg-indigo-500/20 text-indigo-500 flex items-center justify-center text-[10px] font-bold">P</span>
            Personal
            <ChevronDown size={14} className="text-zinc-600" />
          </Button>
        </div>

        <div className="flex items-center gap-4">
          <Link href="https://github.com/enochcodes/orchestra/issues" target="_blank" rel="noopener noreferrer">
            <Button variant="ghost" size="sm" className="text-zinc-400 hover:text-white">
              <MessageSquare className="w-4 h-4 mr-1.5" />
              Feedback
            </Button>
          </Link>
          <Link href="https://github.com/enochcodes/orchestra/releases" target="_blank" rel="noopener noreferrer">
            <Button variant="ghost" size="sm" className="text-zinc-400 hover:text-white">
              <FileText className="w-4 h-4 mr-1.5" />
              Changelog
            </Button>
          </Link>
          <Link href="https://github.com/enochcodes/orchestra#readme" target="_blank" rel="noopener noreferrer">
            <Button variant="ghost" size="sm" className="text-zinc-400 hover:text-white">
              <HelpCircle className="w-4 h-4 mr-1.5" />
              Help
            </Button>
          </Link>

          <div className="h-6 w-px bg-zinc-800 mx-2" />

          <div className="relative" ref={dropdownRef}>
            <button
              onClick={() => setProfileOpen(!profileOpen)}
              className="flex items-center gap-2 p-1.5 rounded-lg hover:bg-zinc-900 transition-colors"
            >
              <UserAvatar name={displayName} />
              <ChevronDown size={14} className={`text-zinc-500 transition-transform ${profileOpen ? "rotate-180" : ""}`} />
            </button>

            {profileOpen && (
              <div className="absolute right-0 top-full mt-2 w-56 rounded-lg border border-zinc-800 bg-zinc-950 shadow-xl py-2 z-50">
                <div className="px-4 py-3 border-b border-zinc-800">
                  <p className="text-sm font-medium text-white truncate">{displayName}</p>
                  <p className="text-xs text-zinc-500 truncate">{user?.email}</p>
                  {user?.system_role && (
                    <span className="inline-block mt-1 text-[10px] px-2 py-0.5 rounded bg-indigo-500/20 text-indigo-400">
                      {user.system_role.replace("_", " ")}
                    </span>
                  )}
                </div>
                <Link href="/settings" onClick={() => setProfileOpen(false)}>
                  <div className="flex items-center gap-2 px-4 py-2 text-sm text-zinc-300 hover:bg-zinc-900 hover:text-white cursor-pointer">
                    <Settings size={14} />
                    Settings
                  </div>
                </Link>
                <button
                  onClick={() => {
                    setProfileOpen(false);
                    logout();
                  }}
                  className="w-full flex items-center gap-2 px-4 py-2 text-sm text-zinc-300 hover:bg-zinc-900 hover:text-red-400 cursor-pointer"
                >
                  <LogOut size={14} />
                  Sign out
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  );
}
