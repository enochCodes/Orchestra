"use client";

import { useState, useEffect } from "react";
import { Server, Plus, RefreshCw, Search, Trash2, Eye, Globe, Cpu, HardDrive } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Card, CardContent } from "@/components/ui/card";
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
  const [selectedServer, setSelectedServer] = useState<any | null>(null);
  const [serverLogs, setServerLogs] = useState<string>("");
  const [formData, setFormData] = useState({
    hostname: "",
    ip: "",
    ssh_user: "root",
    ssh_port: "22",
    ssh_key: "",
  });

  // Nginx dialog state
  const [isNginxOpen, setIsNginxOpen] = useState(false);
  const [nginxConfigs, setNginxConfigs] = useState<any[]>([]);
  const [nginxForm, setNginxForm] = useState({
    domain: "",
    upstream_port: "",
    ssl_enabled: false,
    lets_encrypt: false,
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

  const handleViewServer = async (server: any) => {
    setSelectedServer(server);
    setServerLogs("");
    try {
      const logs = await api.servers.logs(server.id);
      setServerLogs(typeof logs === "string" ? logs : JSON.stringify(logs, null, 2));
    } catch {
      setServerLogs("No logs available.");
    }
    // Fetch nginx configs for this server
    try {
      const res = await api.nginx.list(server.id);
      setNginxConfigs(res.configs || []);
    } catch {
      setNginxConfigs([]);
    }
  };

  const handleDeleteServer = async (id: number) => {
    if (!confirm("Delete this server? This cannot be undone.")) return;
    try {
      await api.servers.delete(id);
      setSelectedServer(null);
      fetchServers();
    } catch (err: any) {
      alert(err.message);
    }
  };

  const handleAddNginx = async () => {
    if (!selectedServer || !nginxForm.domain || !nginxForm.upstream_port) return;
    setIsLoading(true);
    try {
      await api.nginx.create({
        server_id: selectedServer.id,
        domain: nginxForm.domain,
        upstream_port: parseInt(nginxForm.upstream_port),
        ssl_enabled: nginxForm.ssl_enabled,
        lets_encrypt: nginxForm.lets_encrypt,
      });
      setIsNginxOpen(false);
      setNginxForm({ domain: "", upstream_port: "", ssl_enabled: false, lets_encrypt: false });
      // Refresh nginx list
      const res = await api.nginx.list(selectedServer.id);
      setNginxConfigs(res.configs || []);
    } catch (err: any) {
      alert(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleDeleteNginx = async (id: number) => {
    if (!confirm("Remove this nginx config?")) return;
    try {
      await api.nginx.delete(id);
      if (selectedServer) {
        const res = await api.nginx.list(selectedServer.id);
        setNginxConfigs(res.configs || []);
      }
    } catch (err: any) {
      alert(err.message);
    }
  };

  const filtered = servers.filter(
    (s) =>
      !search ||
      s.hostname?.toLowerCase().includes(search.toLowerCase()) ||
      s.ip?.toLowerCase().includes(search.toLowerCase())
  );

  const formatRAM = (bytes: number) => {
    if (!bytes) return "—";
    const gb = bytes / 1024 / 1024 / 1024;
    return gb >= 1 ? `${gb.toFixed(1)} GB` : `${(bytes / 1024 / 1024).toFixed(0)} MB`;
  };

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
                    <div>
                      <span>{server.hostname || server.ip || `Server ${server.id}`}</span>
                      {server.hostname && (
                        <p className="text-[11px] text-zinc-500 font-mono">{server.ip}</p>
                      )}
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline" className={statusColors[server.status] || "bg-zinc-500/10"}>
                      {server.status}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-zinc-400 text-sm">
                    {server.cpu_cores ? `${server.cpu_cores} CPU` : "—"}{" "}
                    {server.ram_bytes ? `/ ${formatRAM(server.ram_bytes)}` : ""}
                  </TableCell>
                  <TableCell className="text-zinc-400">{roleLabels[server.role] || server.role}</TableCell>
                  <TableCell className="text-zinc-500 text-sm">
                    {server.updated_at ? new Date(server.updated_at).toLocaleDateString() : "—"}
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex items-center justify-end gap-1">
                      <Button variant="ghost" size="sm" onClick={() => handleViewServer(server)}>
                        <Eye size={14} className="mr-1" /> View
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="text-red-400 hover:text-red-300"
                        onClick={() => handleDeleteServer(server.id)}
                      >
                        <Trash2 size={14} />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* Server Detail Dialog */}
      {selectedServer && (
        <Dialog open={!!selectedServer} onOpenChange={() => setSelectedServer(null)}>
          <DialogContent className="sm:max-w-[650px] max-h-[85vh] overflow-y-auto border-zinc-800 bg-zinc-950">
            <DialogHeader>
              <DialogTitle>{selectedServer.hostname || selectedServer.ip}</DialogTitle>
              <DialogDescription>
                {selectedServer.ip} — {selectedServer.status}
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-5 mt-4">
              {/* Info Cards */}
              <div className="grid grid-cols-3 gap-3">
                <Card className="bg-zinc-900/50 border-zinc-800">
                  <CardContent className="p-4 flex items-center gap-3">
                    <Cpu size={18} className="text-zinc-500" />
                    <div>
                      <p className="text-xs text-zinc-500">CPU</p>
                      <p className="text-sm text-white font-medium">{selectedServer.cpu_cores || "—"} cores</p>
                    </div>
                  </CardContent>
                </Card>
                <Card className="bg-zinc-900/50 border-zinc-800">
                  <CardContent className="p-4 flex items-center gap-3">
                    <HardDrive size={18} className="text-zinc-500" />
                    <div>
                      <p className="text-xs text-zinc-500">RAM</p>
                      <p className="text-sm text-white font-medium">{formatRAM(selectedServer.ram_bytes)}</p>
                    </div>
                  </CardContent>
                </Card>
                <Card className="bg-zinc-900/50 border-zinc-800">
                  <CardContent className="p-4 flex items-center gap-3">
                    <Globe size={18} className="text-zinc-500" />
                    <div>
                      <p className="text-xs text-zinc-500">OS</p>
                      <p className="text-sm text-white font-medium">{selectedServer.os || "—"}</p>
                    </div>
                  </CardContent>
                </Card>
              </div>

              {/* Details */}
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <p className="text-zinc-500 text-xs mb-1">Role</p>
                  <p className="text-white">{roleLabels[selectedServer.role] || selectedServer.role}</p>
                </div>
                <div>
                  <p className="text-zinc-500 text-xs mb-1">SSH</p>
                  <p className="text-white font-mono text-xs">
                    {selectedServer.ssh_user}@{selectedServer.ip}:{selectedServer.ssh_port}
                  </p>
                </div>
                <div>
                  <p className="text-zinc-500 text-xs mb-1">Architecture</p>
                  <p className="text-white">{selectedServer.arch || "—"}</p>
                </div>
                <div>
                  <p className="text-zinc-500 text-xs mb-1">Cluster</p>
                  <p className="text-white">{selectedServer.cluster?.name || "None"}</p>
                </div>
              </div>

              {/* Error Message */}
              {selectedServer.error_message && (
                <div className="p-3 bg-red-500/10 border border-red-500/30 rounded text-red-400 text-xs font-mono">
                  {selectedServer.error_message}
                </div>
              )}

              {/* Nginx Configs */}
              <div>
                <div className="flex items-center justify-between mb-2">
                  <p className="text-xs font-medium text-zinc-400">Nginx Reverse Proxies</p>
                  <Button variant="outline" size="sm" onClick={() => setIsNginxOpen(true)}>
                    <Plus size={12} className="mr-1" /> Add
                  </Button>
                </div>
                {nginxConfigs.length === 0 ? (
                  <p className="text-xs text-zinc-500">No nginx configs for this server.</p>
                ) : (
                  <div className="space-y-2">
                    {nginxConfigs.map((nc) => (
                      <div key={nc.id} className="flex items-center justify-between p-3 bg-zinc-900/50 border border-zinc-800 rounded text-sm">
                        <div>
                          <span className="text-white">{nc.domain}</span>
                          <span className="text-zinc-500 ml-2">:{nc.upstream_port}</span>
                          {nc.ssl_enabled && <Badge variant="outline" className="ml-2 text-[10px]">SSL</Badge>}
                        </div>
                        <div className="flex items-center gap-2">
                          <Badge variant="outline" className={nc.status === "active" ? "border-emerald-500/50 text-emerald-500 text-[10px]" : "text-[10px]"}>
                            {nc.status}
                          </Badge>
                          <Button variant="ghost" size="sm" className="text-red-400" onClick={() => handleDeleteNginx(nc.id)}>
                            <Trash2 size={12} />
                          </Button>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Preflight Logs */}
              <div>
                <p className="text-xs font-medium text-zinc-400 mb-2">Preflight Report</p>
                <div className="bg-zinc-900/50 border border-zinc-800 rounded p-3 max-h-[200px] overflow-y-auto">
                  <pre className="text-xs text-zinc-300 font-mono whitespace-pre-wrap">{serverLogs || "Loading..."}</pre>
                </div>
              </div>

              {/* Actions */}
              <div className="flex gap-2 pt-2">
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-red-400 hover:text-red-300"
                  onClick={() => handleDeleteServer(selectedServer.id)}
                >
                  <Trash2 size={12} className="mr-1" /> Delete Server
                </Button>
              </div>
            </div>
          </DialogContent>
        </Dialog>
      )}

      {/* Add Nginx Dialog */}
      {isNginxOpen && (
        <Dialog open={isNginxOpen} onOpenChange={setIsNginxOpen}>
          <DialogContent className="sm:max-w-[450px] border-zinc-800 bg-zinc-950">
            <DialogHeader>
              <DialogTitle>Add Nginx Reverse Proxy</DialogTitle>
              <DialogDescription>
                Configure a reverse proxy for {selectedServer?.hostname || selectedServer?.ip}.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4 mt-4">
              <div className="space-y-2">
                <label className="text-xs font-medium text-zinc-400">Domain</label>
                <Input
                  placeholder="app.example.com"
                  value={nginxForm.domain}
                  onChange={(e) => setNginxForm({ ...nginxForm, domain: e.target.value })}
                />
              </div>
              <div className="space-y-2">
                <label className="text-xs font-medium text-zinc-400">Upstream Port</label>
                <Input
                  placeholder="3000"
                  type="number"
                  value={nginxForm.upstream_port}
                  onChange={(e) => setNginxForm({ ...nginxForm, upstream_port: e.target.value })}
                />
              </div>
              <div className="flex items-center gap-4">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={nginxForm.ssl_enabled}
                    onChange={(e) => setNginxForm({ ...nginxForm, ssl_enabled: e.target.checked })}
                    className="rounded border-zinc-700"
                  />
                  <span className="text-sm text-zinc-300">Enable SSL</span>
                </label>
                {nginxForm.ssl_enabled && (
                  <label className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={nginxForm.lets_encrypt}
                      onChange={(e) => setNginxForm({ ...nginxForm, lets_encrypt: e.target.checked })}
                      className="rounded border-zinc-700"
                    />
                    <span className="text-sm text-zinc-300">Let&apos;s Encrypt</span>
                  </label>
                )}
              </div>
            </div>
            <DialogFooter>
              <Button variant="ghost" onClick={() => setIsNginxOpen(false)}>Cancel</Button>
              <Button onClick={handleAddNginx} isLoading={isLoading}>Add Config</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}

      {/* Register Server Dialog */}
      {isRegisterOpen && (
        <Dialog open={isRegisterOpen} onOpenChange={setIsRegisterOpen}>
          <DialogContent className="sm:max-w-[500px] border-zinc-800 bg-zinc-950">
            <DialogHeader>
              <DialogTitle>Register New Server</DialogTitle>
              <DialogDescription>
                Orchestra will SSH into this machine, run pre-flight checks, and prepare it for provisioning.
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
