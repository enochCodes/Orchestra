"use client";

import { useState, useEffect } from "react";
import { Server, Plus, RefreshCw, Search, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { api } from "@/lib/api";

const statusColors: Record<string, string> = {
  pending: "bg-amber-500/10 text-amber-500 border-amber-500/30",
  preflight: "bg-blue-500/10 text-blue-500 border-blue-500/30",
  ready: "bg-emerald-500/10 text-emerald-500 border-emerald-500/30",
  error: "bg-red-500/10 text-red-500 border-red-500/30",
};

const roleLabels: Record<string, string> = {
  none: "Idle",
  manager: "Manager",
  worker: "Worker",
};

export default function ServersPage() {
  const [isRegisterOpen, setIsRegisterOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [servers, setServers] = useState<any[]>([]);
  const [teams, setTeams] = useState<any[]>([]);
  const [search, setSearch] = useState("");
  const [formData, setFormData] = useState({
    hostname: "",
    ip: "",
    ssh_user: "root",
    ssh_port: "22",
    ssh_key: "",
  });

  const fetchServers = async () => {
    try {
      const res = await api.servers.list();
      setServers(res.servers || []);
    } catch (err) {
      console.error("Failed to fetch servers", err);
    }
  };

  const fetchTeams = async () => {
    try {
      const res = await api.servers.teams.list();
      setTeams(res.teams || []);
    } catch (err) {
      console.error("Failed to fetch teams", err);
    }
  };

  useEffect(() => {
    fetchServers();
    fetchTeams();
  }, []);

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    try {
      await api.servers.register({
        ...formData,
        ssh_port: parseInt(formData.ssh_port) || 22,
      });
      setIsRegisterOpen(false);
      setFormData({ hostname: "", ip: "", ssh_user: "root", ssh_port: "22", ssh_key: "" });
      fetchServers();
    } catch (err: any) {
      alert(err.message || "Failed to register server");
    } finally {
      setIsLoading(false);
    }
  };

  const filtered = servers.filter(
    (s) =>
      !search ||
      s.hostname?.toLowerCase().includes(search.toLowerCase()) ||
      s.ip?.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="space-y-8 animate-fade-in max-w-6xl mx-auto">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-white">Servers</h1>
          <p className="text-zinc-400 mt-1">Manage your physical inventory and provisioning status.</p>
        </div>
        <Button onClick={() => setIsRegisterOpen(true)}>
          <Plus className="w-4 h-4 mr-2" />
          Register Server
        </Button>
      </div>

      <div className="flex gap-4 items-center bg-zinc-900/50 p-1 rounded-lg border border-zinc-800">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-500" />
          <Input
            placeholder="Search by hostname, IP, or tags..."
            className="pl-9 bg-transparent border-0 focus-visible:ring-0 h-9"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <div className="h-5 w-px bg-zinc-800" />
        <Button variant="ghost" size="sm" className="text-zinc-400 h-8" onClick={fetchServers}>
          <RefreshCw className="w-3.5 h-3.5 mr-2" />
          Refresh
        </Button>
      </div>

      <div className="rounded-lg border border-zinc-800 overflow-hidden bg-black">
        <Table>
          <TableHeader className="bg-zinc-900/40">
            <TableRow className="hover:bg-transparent border-zinc-800">
              <TableHead className="w-[200px]">Hostname</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Specs</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Last Seen</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filtered.length === 0 ? (
              <TableRow className="hover:bg-transparent border-none">
                <TableCell colSpan={6} className="h-64 text-center">
                  <div className="flex flex-col items-center justify-center text-zinc-500">
                    <div className="w-12 h-12 bg-zinc-900 rounded-full flex items-center justify-center mb-4 border border-zinc-800">
                      <Server className="w-6 h-6 opacity-40" />
                    </div>
                    <h3 className="text-sm font-medium text-zinc-300">No servers found</h3>
                    <p className="max-w-xs mt-1 text-xs text-zinc-500">
                      Register your first physical machine to begin provisioning.
                    </p>
                    <Button variant="outline" size="sm" className="mt-4" onClick={() => setIsRegisterOpen(true)}>
                      Register Server
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ) : (
              filtered.map((server) => (
                <TableRow key={server.id} className="hover:bg-zinc-900/40 border-zinc-800">
                  <TableCell className="font-medium text-white">
                    {server.hostname || server.ip || `Server ${server.id}`}
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline" className={statusColors[server.status] || "bg-zinc-500/10"}>
                      {server.status}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-zinc-400 text-sm">
                    {server.cpu_cores ? `${server.cpu_cores} CPU` : "-"} {server.ram_bytes ? `• ${Math.round(server.ram_bytes / 1024 / 1024 / 1024)}GB` : ""}
                  </TableCell>
                  <TableCell className="text-zinc-400">{roleLabels[server.role] || server.role}</TableCell>
                  <TableCell className="text-zinc-500 text-sm">
                    {server.updated_at ? new Date(server.updated_at).toLocaleDateString() : "-"}
                  </TableCell>
                  <TableCell className="text-right">
                    <Button variant="ghost" size="sm">View</Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {isRegisterOpen && (
        <Dialog open={isRegisterOpen} onOpenChange={setIsRegisterOpen}>
          <DialogContent className="sm:max-w-[500px] border-zinc-800 bg-zinc-950">
            <DialogHeader>
              <DialogTitle>Register New Server</DialogTitle>
              <DialogDescription>
                Orchestra will SSH into this machine, run pre-flight checks, and install the K3s agent.
              </DialogDescription>
            </DialogHeader>

            <form onSubmit={handleRegister} className="space-y-4 mt-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-xs font-medium text-zinc-400">Hostname</label>
                <Input
                  placeholder="node-01"
                  value={formData.hostname}
                  onChange={(e) => setFormData({ ...formData, hostname: e.target.value })}
                />
              </div>
              <div className="space-y-2">
                <label className="text-xs font-medium text-zinc-400">IP Address</label>
                <Input
                  placeholder="192.168.1.10"
                  value={formData.ip}
                  onChange={(e) => setFormData({ ...formData, ip: e.target.value })}
                  required
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-xs font-medium text-zinc-400">SSH User</label>
                <Input
                  placeholder="root"
                  value={formData.ssh_user}
                  onChange={(e) => setFormData({ ...formData, ssh_user: e.target.value })}
                  required
                />
              </div>
              <div className="space-y-2">
                <label className="text-xs font-medium text-zinc-400">SSH Port</label>
                <Input
                  placeholder="22"
                  type="number"
                  value={formData.ssh_port}
                  onChange={(e) => setFormData({ ...formData, ssh_port: e.target.value })}
                  required
                />
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-xs font-medium text-zinc-400">Private Key (PEM)</label>
              <textarea
                className="w-full h-32 rounded-md border border-zinc-800 bg-zinc-900/50 px-3 py-2 text-sm text-zinc-200 placeholder:text-zinc-600 focus:outline-none focus:ring-2 focus:ring-zinc-600 resize-none font-mono"
                placeholder="-----BEGIN OPENSSH PRIVATE KEY-----"
                value={formData.ssh_key}
                onChange={(e) => setFormData({ ...formData, ssh_key: e.target.value })}
                required
              />
            </div>

            <DialogFooter>
              <Button type="button" variant="ghost" onClick={() => setIsRegisterOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" isLoading={isLoading}>
                Register Server
              </Button>
            </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
}
