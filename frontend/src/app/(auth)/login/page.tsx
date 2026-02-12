"use client";

import { useState } from "react";
import { useAuth } from "@/lib/auth-context";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

export default function LoginPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const { login } = useAuth();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setIsLoading(true);
    try {
      await login(email, password);
    } catch (err: any) {
      setError(err.message || "Login failed");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 flex items-center justify-center bg-black">
      <div className="w-full max-w-sm space-y-8 p-8">
        <div className="text-center">
          <div className="inline-flex w-12 h-12 bg-white rounded-full items-center justify-center mb-4">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M12 2L2 22H22L12 2Z" fill="black" />
            </svg>
          </div>
          <h1 className="text-2xl font-bold text-white">Orchestra</h1>
          <p className="text-zinc-400 mt-1 text-sm">Sign in to your account</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="rounded-lg bg-red-500/10 border border-red-500/30 px-4 py-2 text-sm text-red-400">
              {error}
            </div>
          )}
          <div>
            <label className="text-xs font-medium text-zinc-400 block mb-2">Email</label>
            <Input
              type="email"
              placeholder="admin@orchestra.local"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="bg-zinc-900 border-zinc-800"
              required
            />
          </div>
          <div>
            <label className="text-xs font-medium text-zinc-400 block mb-2">Password</label>
            <Input
              type="password"
              placeholder="••••••••"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="bg-zinc-900 border-zinc-800"
              required
            />
          </div>
          <Button type="submit" className="w-full" isLoading={isLoading}>
            Sign In
          </Button>
        </form>

        <p className="text-center text-xs text-zinc-500">
          Default: admin@orchestra.local / admin123
        </p>
      </div>
    </div>
  );
}
