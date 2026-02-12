"use client";

import { useState, useEffect } from "react";
import { Plus, Trash2, Upload, Lock, Key, CheckCircle2, XCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { api } from "@/lib/api";

export default function EnvironmentsPage() {
  const [envs, setEnvs] = useState<any[]>([]);
  const [clusters, setClusters] = useState<any[]>([]);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editEnv, setEditEnv] = useState<any | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [formData, setFormData] = useState({
    cluster_id: 0,
    scope: "production",
    name: "",
    variables: [{ key: "", value: "" }] as { key: string; value: string }[],
  });

  const fetchEnvs = async () => {
    try {
      const res = await api.environments.list();
      setEnvs(res.environments || []);
    } catch (err) {
      console.error("Failed to fetch environments", err);
    }
  };

  const fetchClusters = async () => {
    try {
      const res = await api.clusters.list();
      setClusters(res.clusters || []);
    } catch {}
  };

  useEffect(() => {
    fetchEnvs();
    fetchClusters();
  }, []);

  const handleCreate = async () => {
    if (!formData.name || !formData.cluster_id) {
      alert("Name and cluster are required");
      return;
    }
    setIsLoading(true);
    const vars: Record<string, string> = {};
    formData.variables.forEach((v) => {
      if (v.key) vars[v.key] = v.value;
    });
    try {
      await api.environments.create({
        cluster_id: formData.cluster_id,
        scope: formData.scope,
        name: formData.name,
        variables: vars,
      });
      setIsCreateOpen(false);
      setFormData({ cluster_id: 0, scope: "production", name: "", variables: [{ key: "", value: "" }] });
      fetchEnvs();
    } catch (err: any) {
      alert(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  const handlePush = async (id: number) => {
    try {
      await api.environments.push(id);
      alert("Environment push queued!");
      fetchEnvs();
    } catch (err: any) {
      alert(err.message);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm("Delete this environment?")) return;
    try {
      await api.environments.delete(id);
      fetchEnvs();
    } catch (err: any) {
      alert(err.message);
    }
  };

  const handleSaveEdit = async () => {
    if (!editEnv) return;
    setIsLoading(true);
    const vars: Record<string, string> = {};
    editEnv.variablesList.forEach((v: any) => {
      if (v.key) vars[v.key] = v.value;
    });
    try {
      await api.environments.update(editEnv.id, { variables: vars });
      setEditEnv(null);
      fetchEnvs();
    } catch (err: any) {
      alert(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  const openEdit = (env: any) => {
    const list = Object.entries(env.variables || {}).map(([key, value]) => ({
      key,
      value: value as string,
    }));
    if (list.length === 0) list.push({ key: "", value: "" });
    setEditEnv({ ...env, variablesList: list });
  };

  return (
    <div className="space-y-8 animate-fade-in max-w-6xl mx-auto">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-white">Environments</h1>
          <p className="text-zinc-400 mt-1">
            Manage environment variables per cluster. Push configs to all servers in one click.
          </p>
        </div>
        <Button onClick={() => { setIsCreateOpen(true); fetchClusters(); }}>
          <Plus className="w-4 h-4 mr-2" />
          New Environment
        </Button>
      </div>

      {envs.length === 0 ? (
        <Card className="border-dashed border border-zinc-800 bg-transparent shadow-none">
          <div className="h-64 flex flex-col items-center justify-center text-center p-8">
            <Key className="w-8 h-8 text-zinc-500 mb-4" />
            <h3 className="text-lg font-medium text-white">No Environments</h3>
            <p className="text-zinc-500 max-w-sm mt-2 mb-6 text-sm">
              Create an environment to manage variables for a cluster scope (production, staging, preview).
            </p>
            <Button variant="secondary" onClick={() => setIsCreateOpen(true)}>Create Environment</Button>
          </div>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {envs.map((env) => (
            <Card key={env.id} className="border-zinc-800 bg-zinc-900/30">
              <CardContent className="p-6">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2">
                    <h3 className="font-semibold text-white">{env.name}</h3>
                    <Badge variant="outline" className="text-[10px]">{env.scope}</Badge>
                  </div>
                  <div className="flex items-center gap-2">
                    {env.synced ? (
                      <Badge variant="outline" className="border-emerald-500/50 text-emerald-500 text-[10px]">
                        <CheckCircle2 size={10} className="mr-1" /> Synced
                      </Badge>
                    ) : (
                      <Badge variant="outline" className="border-yellow-500/50 text-yellow-500 text-[10px]">
                        <XCircle size={10} className="mr-1" /> Unsynced
                      </Badge>
                    )}
                  </div>
                </div>
                <p className="text-xs text-zinc-500 mb-3">
                  Cluster: {env.cluster?.name || `#${env.cluster_id}`} — {Object.keys(env.variables || {}).length} variables
                </p>
                <div className="flex gap-2">
                  <Button variant="outline" size="sm" onClick={() => openEdit(env)}>Edit</Button>
                  <Button variant="outline" size="sm" onClick={() => handlePush(env.id)}>
                    <Upload size={12} className="mr-1" /> Push
                  </Button>
                  <Button variant="ghost" size="sm" className="text-red-400 hover:text-red-300" onClick={() => handleDelete(env.id)}>
                    <Trash2 size={12} />
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Create Dialog */}
      <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
        <DialogContent className="sm:max-w-[600px] max-h-[85vh] overflow-y-auto border-zinc-800 bg-zinc-950">
          <DialogHeader>
            <DialogTitle>New Environment</DialogTitle>
            <DialogDescription>Define environment variables for a cluster scope.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4 mt-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-xs font-medium text-zinc-400">Cluster</label>
                <select
                  className="w-full rounded-md border border-zinc-800 bg-zinc-900/50 px-3 py-2 text-sm text-zinc-200"
                  value={formData.cluster_id}
                  onChange={(e) => setFormData({ ...formData, cluster_id: parseInt(e.target.value) })}
                >
                  <option value={0}>Select cluster</option>
                  {clusters.map((c) => (
                    <option key={c.id} value={c.id}>{c.name}</option>
                  ))}
                </select>
              </div>
              <div className="space-y-2">
                <label className="text-xs font-medium text-zinc-400">Scope</label>
                <select
                  className="w-full rounded-md border border-zinc-800 bg-zinc-900/50 px-3 py-2 text-sm text-zinc-200"
                  value={formData.scope}
                  onChange={(e) => setFormData({ ...formData, scope: e.target.value })}
                >
                  <option value="production">Production</option>
                  <option value="staging">Staging</option>
                  <option value="preview">Preview</option>
                </select>
              </div>
            </div>
            <div className="space-y-2">
              <label className="text-xs font-medium text-zinc-400">Name</label>
              <Input
                placeholder="production-v1"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <label className="text-xs font-medium text-zinc-400">Variables</label>
              <div className="space-y-2 max-h-[200px] overflow-y-auto pr-1">
                {formData.variables.map((v, i) => (
                  <div key={i} className="flex gap-2">
                    <Input
                      placeholder="KEY"
                      className="font-mono text-xs flex-1"
                      value={v.key}
                      onChange={(e) => {
                        const vars = [...formData.variables];
                        vars[i].key = e.target.value;
                        setFormData({ ...formData, variables: vars });
                      }}
                    />
                    <Input
                      placeholder="VALUE"
                      className="font-mono text-xs flex-1"
                      type="password"
                      value={v.value}
                      onChange={(e) => {
                        const vars = [...formData.variables];
                        vars[i].value = e.target.value;
                        setFormData({ ...formData, variables: vars });
                      }}
                    />
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => {
                        const vars = [...formData.variables];
                        vars.splice(i, 1);
                        setFormData({ ...formData, variables: vars });
                      }}
                      className="text-zinc-500 hover:text-red-400"
                    >
                      <Trash2 size={14} />
                    </Button>
                  </div>
                ))}
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    setFormData({ ...formData, variables: [...formData.variables, { key: "", value: "" }] })
                  }
                  className="w-full border-dashed border-zinc-700 text-zinc-500"
                >
                  <Plus size={12} className="mr-2" /> Add Variable
                </Button>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setIsCreateOpen(false)}>Cancel</Button>
            <Button onClick={handleCreate} isLoading={isLoading}>Create</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit Dialog */}
      {editEnv && (
        <Dialog open={!!editEnv} onOpenChange={() => setEditEnv(null)}>
          <DialogContent className="sm:max-w-[600px] max-h-[85vh] overflow-y-auto border-zinc-800 bg-zinc-950">
            <DialogHeader>
              <DialogTitle>Edit: {editEnv.name}</DialogTitle>
              <DialogDescription>Update environment variables. Push to sync to servers.</DialogDescription>
            </DialogHeader>
            <div className="space-y-2 mt-4 max-h-[300px] overflow-y-auto pr-1">
              {editEnv.variablesList.map((v: any, i: number) => (
                <div key={i} className="flex gap-2">
                  <Input
                    placeholder="KEY"
                    className="font-mono text-xs flex-1"
                    value={v.key}
                    onChange={(e) => {
                      const list = [...editEnv.variablesList];
                      list[i].key = e.target.value;
                      setEditEnv({ ...editEnv, variablesList: list });
                    }}
                  />
                  <Input
                    placeholder="VALUE"
                    className="font-mono text-xs flex-1"
                    type="password"
                    value={v.value}
                    onChange={(e) => {
                      const list = [...editEnv.variablesList];
                      list[i].value = e.target.value;
                      setEditEnv({ ...editEnv, variablesList: list });
                    }}
                  />
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => {
                      const list = [...editEnv.variablesList];
                      list.splice(i, 1);
                      setEditEnv({ ...editEnv, variablesList: list });
                    }}
                    className="text-zinc-500 hover:text-red-400"
                  >
                    <Trash2 size={14} />
                  </Button>
                </div>
              ))}
              <Button
                variant="outline"
                size="sm"
                onClick={() =>
                  setEditEnv({
                    ...editEnv,
                    variablesList: [...editEnv.variablesList, { key: "", value: "" }],
                  })
                }
                className="w-full border-dashed border-zinc-700 text-zinc-500"
              >
                <Plus size={12} className="mr-2" /> Add Variable
              </Button>
            </div>
            <DialogFooter>
              <Button variant="ghost" onClick={() => setEditEnv(null)}>Cancel</Button>
              <Button onClick={handleSaveEdit} isLoading={isLoading}>Save Changes</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
}
