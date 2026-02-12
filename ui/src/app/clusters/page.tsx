"use client";

import { useState, useEffect } from "react";
import { Network, Plus, ShieldCheck, Zap, GitMerge, AlertCircle, Container, Server, Wrench } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api";
import { Badge } from "@/components/ui/badge";

const clusterTypes = [
  {
    id: "k8s",
    name: "Kubernetes (K3s)",
    description: "Full K8s orchestration with auto-scaling, rolling updates, and service mesh.",
    icon: Network,
  },
  {
    id: "docker_swarm",
    name: "Docker Swarm",
    description: "Lightweight container orchestration with built-in load balancing.",
    icon: Container,
  },
  {
    id: "manual",
    name: "Manual (Docker)",
    description: "Simple Docker containers — no orchestrator. Good for small deployments.",
    icon: Wrench,
  },
];

export default function ClustersPage() {
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [clusters, setClusters] = useState<any[]>([]);
  const [idleServers, setIdleServers] = useState<any[]>([]);
  const [selectedCluster, setSelectedCluster] = useState<any | null>(null);
  const [formData, setFormData] = useState({
    name: "",
    type: "k8s",
    manager_id: "",
    cni: "flannel",
    domain: "",
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
    if (isCreateOpen) fetchIdleServers();
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
        type: formData.type,
        manager_server_id: managerId,
        worker_server_ids: formData.worker_ids,
        cni_plugin: formData.type === "k8s" ? formData.cni : undefined,
        domain: formData.domain || undefined,
      });
      setIsCreateOpen(false);
      setFormData({ name: "", type: "k8s", manager_id: "", cni: "flannel", domain: "", worker_ids: [] });
      fetchClusters();
    } catch (err: any) {
      alert(err.message || "Failed to create cluster");
    } finally {
      setIsLoading(false);
    }
  };

  const typeLabel = (type: string) => {
    const t = clusterTypes.find((ct) => ct.id === type);
    return t?.name || type;
  };

  return (
    <div className="space-y-8 animate-fade-in max-w-6xl mx-auto">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-white">Clusters</h1>
          <p className="text-zinc-400 mt-1">Design and provision compute clusters with K8s, Swarm, or manual Docker.</p>
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
              <h3 className="text-sm font-medium text-white">Multi-Runtime</h3>
              <p className="text-xs text-zinc-500 mt-1">K8s, Docker Swarm, or manual Docker.</p>
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
              <p className="text-xs text-zinc-500 mt-1">Automated install via SSH.</p>
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
              No active clusters. Design your first cluster by choosing a runtime and assigning servers.
            </p>
            <Button variant="secondary" onClick={() => setIsCreateOpen(true)}>
              Open Designer
            </Button>
          </div>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {clusters.map((cluster) => (
            <Card
              key={cluster.id}
              className="border-zinc-800 bg-zinc-900/30 cursor-pointer hover:border-zinc-700 transition-all"
              onClick={() => setSelectedCluster(cluster)}
            >
              <CardContent className="p-6">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center gap-3">
                    <h3 className="font-semibold text-white">{cluster.name}</h3>
                    <Badge variant="outline" className="text-[10px]">
                      {typeLabel(cluster.type || "k8s")}
                    </Badge>
                  </div>
                  <Badge
                    variant="outline"
                    className={
                      cluster.status === "active"
                        ? "border-emerald-500/50 text-emerald-500"
                        : cluster.status === "provisioning"
                        ? "border-blue-500/50 text-blue-500"
                        : cluster.status === "error"
                        ? "border-red-500/50 text-red-500"
                        : "border-zinc-600 text-zinc-400"
                    }
                  >
                    {cluster.status}
                  </Badge>
                </div>
                <div className="text-sm text-zinc-400 space-y-1">
                  <p>
                    Manager:{" "}
                    {cluster.manager_server?.hostname || cluster.manager_server?.ip || `#${cluster.manager_server_id}`}
                  </p>
                  <p>Workers: {cluster.workers?.length || 0}</p>
                  {cluster.domain && <p>Domain: {cluster.domain}</p>}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Cluster Detail Dialog */}
      {selectedCluster && (
        <Dialog open={!!selectedCluster} onOpenChange={() => setSelectedCluster(null)}>
          <DialogContent className="sm:max-w-[550px] border-zinc-800 bg-zinc-950">
            <DialogHeader>
              <DialogTitle>{selectedCluster.name}</DialogTitle>
              <DialogDescription>
                {typeLabel(selectedCluster.type || "k8s")} — {selectedCluster.status}
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4 mt-4 text-sm text-zinc-400">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-zinc-500 text-xs mb-1">Manager</p>
                  <p className="text-white">
                    {selectedCluster.manager_server?.hostname || selectedCluster.manager_server?.ip}
                  </p>
                </div>
                <div>
                  <p className="text-zinc-500 text-xs mb-1">Workers</p>
                  <p className="text-white">{selectedCluster.workers?.length || 0} nodes</p>
                </div>
                {selectedCluster.type === "k8s" && (
                  <div>
                    <p className="text-zinc-500 text-xs mb-1">CNI Plugin</p>
                    <p className="text-white">{selectedCluster.cni_plugin || "flannel"}</p>
                  </div>
                )}
                {selectedCluster.domain && (
                  <div>
                    <p className="text-zinc-500 text-xs mb-1">Domain</p>
                    <p className="text-white">{selectedCluster.domain}</p>
                  </div>
                )}
              </div>
              {selectedCluster.workers?.length > 0 && (
                <div>
                  <p className="text-zinc-500 text-xs mb-2">Worker Nodes</p>
                  <div className="space-y-1">
                    {selectedCluster.workers.map((w: any) => (
                      <div key={w.id} className="flex items-center gap-2 text-zinc-300">
                        <Server size={12} />
                        <span>{w.hostname || w.ip}</span>
                        <Badge variant="outline" className="text-[10px] ml-auto">{w.status}</Badge>
                      </div>
                    ))}
                  </div>
                </div>
              )}
              {selectedCluster.error_message && (
                <div className="p-3 bg-red-500/10 border border-red-500/30 rounded text-red-400 text-xs">
                  {selectedCluster.error_message}
                </div>
              )}
            </div>
          </DialogContent>
        </Dialog>
      )}

      {/* Create Cluster Dialog */}
      {isCreateOpen && (
        <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
          <DialogContent className="sm:max-w-[600px] max-h-[90vh] overflow-y-auto border-zinc-800 bg-zinc-950">
            <DialogHeader>
              <DialogTitle>Design New Cluster</DialogTitle>
              <DialogDescription>Choose a runtime, assign servers, and provision automatically.</DialogDescription>
            </DialogHeader>

            <form onSubmit={handleCreate} className="space-y-5 mt-4">
              {/* Cluster Type */}
              <div className="space-y-2">
                <label className="text-xs font-medium text-zinc-400">Cluster Type</label>
                <div className="grid grid-cols-3 gap-3">
                  {clusterTypes.map((ct) => (
                    <div
                      key={ct.id}
                      className={`p-4 border rounded-xl cursor-pointer transition-all ${
                        formData.type === ct.id
                          ? "border-indigo-500 bg-indigo-500/10"
                          : "border-zinc-800 hover:border-zinc-700"
                      }`}
                      onClick={() => setFormData({ ...formData, type: ct.id })}
                    >
                      <ct.icon className="w-5 h-5 text-zinc-400 mb-2" />
                      <h4 className="text-sm font-medium text-white">{ct.name}</h4>
                      <p className="text-[10px] text-zinc-500 mt-1">{ct.description}</p>
                    </div>
                  ))}
                </div>
              </div>

              {/* Cluster Name */}
              <div className="space-y-2">
                <label className="text-xs font-medium text-zinc-400">Cluster Name</label>
                <Input
                  placeholder="production-v1"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  required
                />
              </div>

              {/* Domain (optional) */}
              <div className="space-y-2">
                <label className="text-xs font-medium text-zinc-400">Domain (optional)</label>
                <Input
                  placeholder="app.example.com"
                  value={formData.domain}
                  onChange={(e) => setFormData({ ...formData, domain: e.target.value })}
                />
              </div>

              {/* Manager Node */}
              <div className="space-y-2">
                <label className="text-xs font-medium text-zinc-400">Manager Node</label>
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
                    <option value="" disabled>
                      No idle servers available
                    </option>
                  )}
                </select>
                <p className="text-[10px] text-zinc-500 flex items-center gap-1">
                  <AlertCircle size={10} />
                  Must be a registered server in &quot;Ready&quot; state.
                </p>
              </div>

              {/* Worker Nodes */}
              <div className="space-y-2">
                <label className="text-xs font-medium text-zinc-400">Worker Nodes (optional)</label>
                <div className="space-y-2 max-h-32 overflow-y-auto">
                  {idleServers
                    .filter((s) => s.id !== parseInt(formData.manager_id))
                    .map((s) => (
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
                        <span className="text-sm text-zinc-300">
                          {s.hostname || s.ip} ({s.ip})
                        </span>
                      </label>
                    ))}
                  {idleServers.filter((s) => s.id !== parseInt(formData.manager_id)).length === 0 && (
                    <p className="text-xs text-zinc-500">No other idle servers</p>
                  )}
                </div>
              </div>

              {/* CNI Plugin (K8s only) */}
              {formData.type === "k8s" && (
                <div className="space-y-2">
                  <label className="text-xs font-medium text-zinc-400">CNI Plugin</label>
                  <div className="grid grid-cols-2 gap-2">
                    <div
                      className={`border rounded-md p-3 cursor-pointer transition-all ${
                        formData.cni === "flannel"
                          ? "border-indigo-500 bg-indigo-500/10"
                          : "border-zinc-800 hover:border-zinc-700"
                      }`}
                      onClick={() => setFormData({ ...formData, cni: "flannel" })}
                    >
                      <div className="text-sm font-medium text-white">Flannel</div>
                      <div className="text-xs text-zinc-500">Lightweight overlay</div>
                    </div>
                    <div
                      className={`border rounded-md p-3 cursor-pointer transition-all ${
                        formData.cni === "calico"
                          ? "border-indigo-500 bg-indigo-500/10"
                          : "border-zinc-800 hover:border-zinc-700"
                      }`}
                      onClick={() => setFormData({ ...formData, cni: "calico" })}
                    >
                      <div className="text-sm font-medium text-white">Calico</div>
                      <div className="text-xs text-zinc-500">Network policy support</div>
                    </div>
                  </div>
                </div>
              )}

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
