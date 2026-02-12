"use client";

import { useState, useEffect } from "react";
import { User, Key, Shield } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth } from "@/lib/auth-context";
import { api } from "@/lib/api";

export default function SettingsPage() {
  const [activeTab, setActiveTab] = useState("general");
  const { user, refresh } = useAuth();
  const [displayName, setDisplayName] = useState("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    setDisplayName(user?.display_name || "");
  }, [user]);

  const tabs = [
    { id: "general", label: "General", icon: User },
    { id: "keys", label: "SSH Keys", icon: Key },
    { id: "security", label: "Security", icon: Shield },
  ];

  const handleSaveProfile = async () => {
    setSaving(true);
    try {
      await api.updateProfile({ display_name: displayName });
      await refresh();
    } catch (err: any) {
      alert(err.message || "Failed to save");
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="animate-fade-in max-w-6xl mx-auto flex gap-8">
      <div className="w-64 shrink-0 space-y-2">
        <h1 className="text-2xl font-semibold tracking-tight text-white mb-6">Settings</h1>
        {tabs.map(tab => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`w-full flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors ${activeTab === tab.id
              ? "bg-zinc-800 text-white"
              : "text-zinc-400 hover:text-white hover:bg-zinc-900"
              }`}
          >
            <tab.icon size={16} />
            {tab.label}
          </button>
        ))}
      </div>

      <div className="flex-1 space-y-6">
        {activeTab === "general" && (
          <Card className="bg-zinc-900/30 border-zinc-800">
            <CardHeader>
              <CardTitle>Profile Information</CardTitle>
              <CardDescription>Update your personal details.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-2">
                <label className="text-sm font-medium text-zinc-300">Display Name</label>
                <Input
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                  placeholder="Your name"
                />
              </div>
              <div className="grid gap-2">
                <label className="text-sm font-medium text-zinc-300">Email</label>
                <Input defaultValue={user?.email} disabled />
              </div>
              <Button onClick={handleSaveProfile} disabled={saving}>
                {saving ? "Saving..." : "Save Changes"}
              </Button>
            </CardContent>
          </Card>
        )}

        {activeTab === "keys" && (
          <Card className="bg-zinc-900/30 border-zinc-800">
            <CardHeader>
              <CardTitle>SSH Config</CardTitle>
              <CardDescription>Manage keys used for server provisioning.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="p-4 rounded-md border border-zinc-800 bg-zinc-950 flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <Key size={16} className="text-zinc-500" />
                  <div>
                    <p className="text-sm font-medium text-white">Default Cluster Key</p>
                    <p className="text-xs text-zinc-500">Used when registering servers</p>
                  </div>
                </div>
                <Button variant="outline" size="sm" disabled>Configure</Button>
              </div>
            </CardContent>
          </Card>
        )}

        {activeTab === "security" && (
          <div className="space-y-6">
            <Card className="bg-red-900/10 border-red-900/20">
              <CardHeader>
                <CardTitle className="text-red-500">Danger Zone</CardTitle>
                <CardDescription>Irreversible actions.</CardDescription>
              </CardHeader>
              <CardContent>
                <Button variant="destructive" disabled>Factory Reset Orchestra</Button>
              </CardContent>
            </Card>
          </div>
        )}
      </div>
    </div>
  );
}
