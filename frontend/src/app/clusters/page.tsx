"use client";

import { useState, useEffect } from "react";
import { Network, Plus, ShieldCheck, Zap, GitMerge, AlertCircle, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api";
import { Badge } from "@/components/ui/badge";

export default function ClustersPage() {
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [clusters, setClusters] = useState<any[]>([]);
  const [idleServers, setIdleServers] = useState<any[]>([]);
  const [formData, setFormData] = useState({
    name: "",
    manager_id: "",
    cni: "flannel",
    worker_ids: [] as number[],
  });

  const fetchClusters = async () => {
    try {
      const res = await api.clusters.list();
      setClusters(res.clusters || []);
    } catch (err) {
      console.error("Failed to fetch clusters", err);
    }
  };

  const fetchIdleServers = async () => {
    try {
      const res = await api.servers.idle();
      setIdleServers(res.servers || []);
    } catch (err) {
      console.error("Failed to fetch idle servers", err);
    }
  };

  useEffect(() => {
    fetchClusters();
    fetchIdleServers();
  }, []);

  useEffect(() => {
    if (isCreateOpen) {
      fetchIdleServers();
    }
  }, [isCreateOpen]);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    const managerId = parseInt(formData.manager_id);
    if (!managerId || !formData.name) {
      alert("Cluster name and manager server are required.");
      return;
    }
    setIsLoading(true);
    try {
      await api.clusters.design({
        name: formData.name,
        manager_server_id: managerId,
        worker_server_ids: formData.worker_ids,
        cni_plugin: formData.cni,
      });
      setIsCreateOpen(false);
      setFormData({ name: "", manager_id: "", cni: "flannel", worker_ids: [] });
      fetchClusters();
    } catch (err: any) {
      alert(err.message || "Failed to create cluster");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="space-y-8 animate-fade-in max-w-6xl mx-auto">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-white">Clusters</h1>
          <p className="text-zinc-400 mt-1">Design and oversee your Kubernetes topologies.</p>
        </div>
        <Button onClick={() => setIsCreateOpen(true)}>
          <Plus className="w-4 h-4 mr-2" />
          Create Cluster
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card className="bg-zinc-900/10 border-zinc-800 backdrop-blur-sm">
          <CardContent className="p-6 flex gap-4">
            <div className="p-2 rounded-md bg-zinc-100/5 text-zinc-100 h-fit ring-1 ring-white/10">
              <ShieldCheck size={20} />
            </div>
            <div>
              <h3 className="text-sm font-medium text-white">Secure by Default</h3>
              <p className="text-xs text-zinc-500 mt-1">mTLS encryption enabled between all nodes.</p>
            </div>
          </CardContent>
        </Card>
        <Card className="bg-zinc-900/10 border-zinc-800 backdrop-blur-sm">
          <CardContent className="p-6 flex gap-4">
            <div className="p-2 rounded-md bg-zinc-100/5 text-zinc-100 h-fit ring-1 ring-white/10">
              <Zap size={20} />
            </div>
            <div>
              <h3 className="text-sm font-medium text-white">Instant Provisioning</h3>
              <p className="text-xs text-zinc-500 mt-1">K3s installs in under 30 seconds.</p>
            </div>
          </CardContent>
        </Card>
        <Card className="bg-zinc-900/10 border-zinc-800 backdrop-blur-sm">
          <CardContent className="p-6 flex gap-4">
            <div className="p-2 rounded-md bg-zinc-100/5 text-zinc-100 h-fit ring-1 ring-white/10">
              <GitMerge size={20} />
            </div>
            <div>
              <h3 className="text-sm font-medium text-white">GitOps Ready</h3>
              <p className="text-xs text-zinc-500 mt-1">Push to Git to deploy.</p>
            </div>
          </CardContent>
        </Card>
      </div>

      {clusters.length === 0 ? (
        <Card className="border-dashed border border-zinc-800 bg-transparent shadow-none">
          <div className="h-96 flex flex-col items-center justify-center text-center p-8">
            <div className="w-16 h-16 rounded-full bg-zinc-900 flex items-center justify-center mb-6 ring-1 ring-zinc-800">
              <Network className="w-8 h-8 text-zinc-500" />
            </div>
            <h3 className="text-lg font-medium text-white">Topology Canvas</h3>
            <p className="text-zinc-500 max-w-sm mt-2 mb-8 text-sm">
              No active clusters. Design your first cluster by assigning roles to your registered servers.
            </p>
            <Button variant="secondary" onClick={() => setIsCreateOpen(true)}>
              Open Designer
            </Button>
          </div>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {clusters.map((cluster) => (
            <Card key={cluster.id} className="border-zinc-800 bg-zinc-900/30">
              <CardContent className="p-6">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="font-semibold text-white">{cluster.name}</h3>
                  <Badge variant="outline" className={
                    cluster.status === "active" ? "border-emerald-500/50 text-emerald-500" :
                    cluster.status === "provisioning" ? "border-blue-500/50 text-blue-500" :
                    "border-zinc-600 text-zinc-400"
                  }>
                    {cluster.status}
                  </Badge>
                </div>
                <div className="text-sm text-zinc-400 space-y-1">
                  <p>Manager: {cluster.manager_server?.hostname || cluster.manager_server?.ip || `#${cluster.manager_server_id}`}</p>
                  <p>Workers: {cluster.workers?.length || 0}</p>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {isCreateOpen && (
        <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
          <DialogContent className="sm:max-w-[500px] border-zinc-800 bg-zinc-950">
            <DialogHeader>
              <DialogTitle>Design New Cluster</DialogTitle>
              <DialogDescription>
                Define the control plane and network settings for your new Kubernetes cluster.
              </DialogDescription>
            </DialogHeader>

            <form onSubmit={handleCreate} className="space-y-4 mt-6">
              <div className="space-y-2">
                <label className="text-xs font-medium text-zinc-400">Cluster Name</label>
                <Input
                  placeholder="production-v1"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  required
                />
              </div>

              <div className="space-y-2">
                <label className="text-xs font-medium text-zinc-400">Manager Node (Ready servers only)</label>
                <select
                  className="w-full rounded-md border border-zinc-800 bg-zinc-900/50 px-3 py-2 text-sm text-zinc-200 focus:outline-none focus:ring-2 focus:ring-zinc-600"
                  value={formData.manager_id}
                  onChange={(e) => setFormData({ ...formData, manager_id: e.target.value })}
                  required
                >
                  <option value="">Select a server</option>
                  {idleServers.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.hostname || s.ip} ({s.ip})
                    </option>
                  ))}
                  {idleServers.length === 0 && (
                    <option value="" disabled>No idle servers available</option>
                  )}
                </select>
                <p className="text-[10px] text-zinc-500 flex items-center gap-1">
                  <AlertCircle size={10} />
                  Must be a registered server in "Ready" state.
                </p>
              </div>

              <div className="space-y-2">
                <label className="text-xs font-medium text-zinc-400">Worker Nodes (optional)</label>
                <div className="space-y-2 max-h-32 overflow-y-auto">
                  {idleServers.filter((s) => s.id !== parseInt(formData.manager_id)).map((s) => (
                    <label key={s.id} className="flex items-center gap-2 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={formData.worker_ids.includes(s.id)}
                        onChange={(e) => {
                          const ids = e.target.checked
                            ? [...formData.worker_ids, s.id]
                            : formData.worker_ids.filter((id) => id !== s.id);
                          setFormData({ ...formData, worker_ids: ids });
                        }}
                        className="rounded border-zinc-700"
                      />
                      <span className="text-sm text-zinc-300">{s.hostname || s.ip} ({s.ip})</span>
                    </label>
                  ))}
                  {idleServers.filter((s) => s.id !== parseInt(formData.manager_id)).length === 0 && (
                    <p className="text-xs text-zinc-500">No other idle servers</p>
                  )}
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-xs font-medium text-zinc-400">CNI Plugin</label>
                <div className="grid grid-cols-2 gap-2">
                  <div
                    className={`border rounded-md p-3 cursor-pointer transition-all ${formData.cni === "flannel" ? "border-indigo-500 bg-indigo-500/10" : "border-zinc-800 hover:border-zinc-700"}`}
                    onClick={() => setFormData({ ...formData, cni: "flannel" })}
                  >
                    <div className="text-sm font-medium text-white">Flannel</div>
                    <div className="text-xs text-zinc-500">Lightweight overlay</div>
                  </div>
                  <div
                    className={`border rounded-md p-3 cursor-pointer transition-all ${formData.cni === "calico" ? "border-indigo-500 bg-indigo-500/10" : "border-zinc-800 hover:border-zinc-700"}`}
                    onClick={() => setFormData({ ...formData, cni: "calico" })}
                  >
                    <div className="text-sm font-medium text-white">Calico</div>
                    <div className="text-xs text-zinc-500">Network policy support</div>
                  </div>
                </div>
              </div>

              <DialogFooter>
                <Button type="button" variant="ghost" onClick={() => setIsCreateOpen(false)}>
                  Cancel
                </Button>
                <Button type="submit" isLoading={isLoading}>
                  Create Cluster
                </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
}
