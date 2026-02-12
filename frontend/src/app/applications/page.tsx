"use client";

import { useState, useEffect } from "react";
import {
  AppWindow, Plus, GitBranch, Terminal, Layers, Box, Code2, Server,
  Trash2, ChevronRight, Lock, Layout, Globe, Cpu, CheckCircle2, Image, FolderOpen
} from "lucide-react";
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { SubNav } from "@/components/layout/SubNav";
import { api } from "@/lib/api";
import Link from "next/link";

interface Framework {
  id: string;
  name: string;
  description: string;
  default_build: string;
  default_start: string;
}

interface AppType {
  id: string;
  name: string;
  frameworks: Framework[];
}

export default function ApplicationsPage() {
  const [isDeployOpen, setIsDeployOpen] = useState(false);
  const [manageApp, setManageApp] = useState<any | null>(null);
  const [step, setStep] = useState(1);
  const [isLoading, setIsLoading] = useState(false);
  const [appTypes, setAppTypes] = useState<AppType[]>([]);
  const [clusters, setClusters] = useState<any[]>([]);
  const [formData, setFormData] = useState({
    name: "",
    cluster_id: 0,
    repo: "",
    branch: "main",
    source_type: "git" as "git" | "manual" | "docker_image",
    docker_image: "",
    manual_path: "",
    type: "web_service",
    frameworkId: "",
    buildCmd: "",
    startCmd: "",
    env: {
      production: [{ key: "", value: "" }],
      preview: [{ key: "", value: "" }]
    }
  });
  const [activeEnvTab, setActiveEnvTab] = useState("production");
  const [apps, setApps] = useState<any[]>([]);

  const fetchApps = async () => {
    try {
      const data = await api.applications.list();
      setApps(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error("Failed to fetch applications", err);
    }
  };

  const fetchClusters = async () => {
    try {
      const res = await api.clusters.list();
      setClusters(res.clusters || []);
    } catch (err) {
      console.error("Failed to fetch clusters", err);
    }
  };

  useEffect(() => {
    fetchApps();
    fetchClusters();
  }, []);

  useEffect(() => {
    if (isDeployOpen) {
      api.metadata.frameworks().then(setAppTypes).catch(console.error);
      fetchClusters();
    }
  }, [isDeployOpen]);

  const handleTypeSelect = (typeId: string) => {
    setFormData(prev => ({ ...prev, type: typeId, frameworkId: "" }));
  };

  const handleFrameworkSelect = (fw: Framework) => {
    setFormData(prev => ({
      ...prev,
      frameworkId: fw.id,
      buildCmd: fw.default_build,
      startCmd: fw.default_start
    }));
  };

  const handleEnvChange = (index: number, field: 'key' | 'value', value: string) => {
    const currentEnvs = [...formData.env[activeEnvTab as 'production' | 'preview']];
    currentEnvs[index][field] = value;
    setFormData(prev => ({
      ...prev,
      env: { ...prev.env, [activeEnvTab]: currentEnvs }
    }));
  };

  const addEnvField = () => {
    const currentEnvs = [...formData.env[activeEnvTab as 'production' | 'preview'], { key: "", value: "" }];
    setFormData(prev => ({ ...prev, env: { ...prev.env, [activeEnvTab]: currentEnvs } }));
  };

  const removeEnvField = (index: number) => {
    const currentEnvs = [...formData.env[activeEnvTab as 'production' | 'preview']];
    currentEnvs.splice(index, 1);
    setFormData(prev => ({ ...prev, env: { ...prev.env, [activeEnvTab]: currentEnvs } }));
  };

  const handleDeploy = async () => {
    setIsLoading(true);
    const formatEnvs = (arr: { key: string; value: string }[]) => {
      return arr.reduce((acc, curr) => {
        if (curr.key) acc[curr.key] = curr.value;
        return acc;
      }, {} as Record<string, string>);
    };

    const clusterId = formData.cluster_id || (clusters[0]?.id);
    if (!clusterId) {
      alert("Please select a cluster");
      setIsLoading(false);
      return;
    }

    const payload: any = {
      name: formData.name,
      cluster_id: clusterId,
      namespace: "default",
      build_type: formData.frameworkId,
      build_cmd: formData.buildCmd,
      start_cmd: formData.startCmd,
      env_vars: {
        production: formatEnvs(formData.env.production),
        preview: formatEnvs(formData.env.preview)
      },
      source_type: formData.source_type,
    };

    if (formData.source_type === "git") {
      payload.repo_url = formData.repo;
      payload.branch = formData.branch;
    } else if (formData.source_type === "docker_image") {
      payload.docker_image = formData.docker_image;
    } else if (formData.source_type === "manual") {
      payload.manual_path = formData.manual_path;
    }

    try {
      await api.applications.create(payload);
      fetchApps();
      setIsDeployOpen(false);
      setStep(1);
      setFormData({
        name: "",
        cluster_id: 0,
        repo: "",
        branch: "main",
        source_type: "git",
        docker_image: "",
        manual_path: "",
        type: "web_service",
        frameworkId: "",
        buildCmd: "",
        startCmd: "",
        env: { production: [{ key: "", value: "" }], preview: [{ key: "", value: "" }] }
      });
    } catch (e: any) {
      alert(e.message || "Error deploying app");
    } finally {
      setIsLoading(false);
    }
  };

  const activeAppType = appTypes.find(t => t.id === formData.type);

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-white">Applications</h1>
          <p className="text-zinc-400 mt-1">Manage your polyglot workloads and deployments.</p>
        </div>
        <Button onClick={() => setIsDeployOpen(true)}>
          <Plus className="w-4 h-4 mr-2" />
          New Deployment
        </Button>
      </div>

      <SubNav
        items={[
          { label: "Overview", href: "/applications" },
          { label: "Deployments", href: "/deployments" },
          { label: "Monitoring", href: "/monitoring" },
          { label: "Environment", href: "/settings" },
        ]}
      />

      <div className="rounded-lg border border-zinc-800 overflow-hidden bg-zinc-900/30">
        <Table>
          <TableHeader className="bg-zinc-900/50">
            <TableRow className="hover:bg-transparent border-zinc-800">
              <TableHead>Name</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Cluster</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Replicas</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {apps.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center py-12 text-zinc-500">
                  No applications found. Deploy one to get started.
                </TableCell>
              </TableRow>
            ) : (
              apps.map((app) => (
                <TableRow key={app.id} className="border-zinc-800 hover:bg-zinc-900/40">
                  <TableCell className="font-medium text-white">
                    <div className="flex items-center gap-3">
                      <div className="p-2 bg-zinc-900 rounded border border-zinc-800">
                        {app.build_type?.includes("go") ? <Terminal size={16} /> :
                          app.build_type?.includes("python") ? <Code2 size={16} /> :
                            app.build_type?.includes("next") ? <Globe size={16} /> :
                              <AppWindow size={16} />}
                      </div>
                      {app.name}
                    </div>
                  </TableCell>
                  <TableCell className="text-zinc-400 capitalize">{app.build_type || app.source_type || "docker"}</TableCell>
                  <TableCell className="text-zinc-400">{app.cluster?.name || "—"}</TableCell>
                  <TableCell>
                    <Badge variant={app.status === "running" ? "success" : "warning"}>
                      {app.status}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-zinc-400 font-mono text-xs">{app.replicas || 1}/1</TableCell>
                  <TableCell className="text-right">
                    <Button variant="ghost" size="sm" onClick={() => setManageApp(app)}>
                      Manage
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      <Dialog open={isDeployOpen} onOpenChange={setIsDeployOpen}>
        <DialogContent className="sm:max-w-[700px] max-h-[90vh] overflow-y-auto border-zinc-800 bg-zinc-950 text-white">
          <DialogHeader>
            <DialogTitle>Deploy New Application</DialogTitle>
            <DialogDescription>
              Configure your application deployment pipeline.
            </DialogDescription>
          </DialogHeader>

          <div className="flex items-center justify-between px-8 py-4 border-b border-zinc-900">
            {[1, 2, 3, 4, 5].map((s) => (
              <div key={s} className="flex flex-col items-center gap-2">
                <div className={`w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold transition-all ${step >= s ? "bg-indigo-600 text-white" : "bg-zinc-800 text-zinc-500"}`}>
                  {step > s ? <CheckCircle2 size={14} /> : s}
                </div>
                <span className={`text-[10px] uppercase tracking-wider font-medium ${step >= s ? "text-indigo-400" : "text-zinc-600"}`}>
                  {s === 1 ? "Source" : s === 2 ? "Stack" : s === 3 ? "Config" : s === 4 ? "Env" : "Review"}
                </span>
              </div>
            ))}
          </div>

          <div className="py-6 min-h-[300px]">
            {step === 1 && (
              <div className="space-y-6">
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
                  <label className="text-xs font-medium text-zinc-400">Deployment Source</label>
                  <div className="grid grid-cols-3 gap-3">
                    <div
                      className={`p-4 border rounded-xl cursor-pointer transition-all ${formData.source_type === "git" ? "border-indigo-500 bg-indigo-500/10" : "border-zinc-800 hover:border-zinc-700"}`}
                      onClick={() => setFormData({ ...formData, source_type: "git" })}
                    >
                      <GitBranch className="w-6 h-6 text-zinc-400 mb-2" />
                      <h3 className="font-semibold text-sm">Git Repo</h3>
                      <p className="text-xs text-zinc-500">Clone from repository</p>
                    </div>
                    <div
                      className={`p-4 border rounded-xl cursor-pointer transition-all ${formData.source_type === "manual" ? "border-indigo-500 bg-indigo-500/10" : "border-zinc-800 hover:border-zinc-700"}`}
                      onClick={() => setFormData({ ...formData, source_type: "manual" })}
                    >
                      <FolderOpen className="w-6 h-6 text-zinc-400 mb-2" />
                      <h3 className="font-semibold text-sm">Manual Path</h3>
                      <p className="text-xs text-zinc-500">Folder upload</p>
                    </div>
                    <div
                      className={`p-4 border rounded-xl cursor-pointer transition-all ${formData.source_type === "docker_image" ? "border-indigo-500 bg-indigo-500/10" : "border-zinc-800 hover:border-zinc-700"}`}
                      onClick={() => setFormData({ ...formData, source_type: "docker_image" })}
                    >
                      <Image className="w-6 h-6 text-zinc-400 mb-2" />
                      <h3 className="font-semibold text-sm">Docker Image</h3>
                      <p className="text-xs text-zinc-500">Pre-built image</p>
                    </div>
                  </div>
                </div>

                {formData.source_type === "git" && (
                  <>
                    <div className="space-y-2">
                      <label className="text-xs font-medium text-zinc-400">Git Repository</label>
                      <Input placeholder="https://github.com/..." value={formData.repo} onChange={e => setFormData({ ...formData, repo: e.target.value })} />
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <label className="text-xs font-medium text-zinc-400">Project Name</label>
                        <Input placeholder="my-app" value={formData.name} onChange={e => setFormData({ ...formData, name: e.target.value })} />
                      </div>
                      <div className="space-y-2">
                        <label className="text-xs font-medium text-zinc-400">Branch</label>
                        <Input placeholder="main" value={formData.branch} onChange={e => setFormData({ ...formData, branch: e.target.value })} />
                      </div>
                    </div>
                  </>
                )}

                {formData.source_type === "docker_image" && (
                  <div className="space-y-2">
                    <label className="text-xs font-medium text-zinc-400">Docker Image</label>
                    <Input placeholder="nginx:latest" value={formData.docker_image} onChange={e => setFormData({ ...formData, docker_image: e.target.value })} />
                    <div className="space-y-2 mt-2">
                      <label className="text-xs font-medium text-zinc-400">App Name</label>
                      <Input placeholder="my-app" value={formData.name} onChange={e => setFormData({ ...formData, name: e.target.value })} />
                    </div>
                  </div>
                )}

                {formData.source_type === "manual" && (
                  <div className="space-y-2">
                    <label className="text-xs font-medium text-zinc-400">Path / Folder</label>
                    <Input placeholder="/path/to/app" value={formData.manual_path} onChange={e => setFormData({ ...formData, manual_path: e.target.value })} />
                    <div className="space-y-2 mt-2">
                      <label className="text-xs font-medium text-zinc-400">App Name</label>
                      <Input placeholder="my-app" value={formData.name} onChange={e => setFormData({ ...formData, name: e.target.value })} />
                    </div>
                  </div>
                )}
              </div>
            )}

            {step === 2 && (
              <div className="space-y-4">
                <h3 className="text-sm font-medium text-white mb-4">Select Technology Stack</h3>
                <div className="grid grid-cols-2 gap-3">
                  {activeAppType?.frameworks?.map(fw => (
                    <div
                      key={fw.id}
                      className={`flex items-start gap-4 p-4 border rounded-xl cursor-pointer transition-all ${formData.frameworkId === fw.id ? "border-indigo-500 bg-indigo-500/10" : "border-zinc-800 hover:border-zinc-700"}`}
                      onClick={() => handleFrameworkSelect(fw)}
                    >
                      <div className="mt-1">
                        {fw.id.includes("go") ? <Terminal className="text-cyan-400" /> :
                          fw.id.includes("node") ? <Server className="text-green-400" /> :
                            fw.id.includes("rust") ? <Cpu className="text-orange-400" /> :
                              fw.id.includes("python") ? <Code2 className="text-blue-400" /> :
                                <Layout className="text-pink-400" />}
                      </div>
                      <div>
                        <h4 className="font-medium text-sm text-white">{fw.name}</h4>
                        <p className="text-xs text-zinc-500 mt-1">{fw.description}</p>
                      </div>
                      {formData.frameworkId === fw.id && <CheckCircle2 className="ml-auto text-indigo-500" size={16} />}
                    </div>
                  ))}
                  {formData.source_type === "docker_image" && (
                    <p className="text-sm text-zinc-500 col-span-2">Docker image deployments use minimal config.</p>
                  )}
                </div>
              </div>
            )}

            {step === 3 && (
              <div className="space-y-6">
                <div className="p-4 bg-zinc-900/50 border border-zinc-800 rounded-lg">
                  <h4 className="text-sm font-medium text-white mb-4 flex items-center gap-2">
                    <Terminal size={14} className="text-zinc-400" /> Build Configuration
                  </h4>
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <label className="text-xs font-medium text-zinc-400">Build Command</label>
                      <Input className="font-mono text-sm" value={formData.buildCmd} onChange={e => setFormData({ ...formData, buildCmd: e.target.value })} />
                    </div>
                    <div className="space-y-2">
                      <label className="text-xs font-medium text-zinc-400">Start Command</label>
                      <Input className="font-mono text-sm" value={formData.startCmd} onChange={e => setFormData({ ...formData, startCmd: e.target.value })} />
                    </div>
                  </div>
                </div>
              </div>
            )}

            {step === 4 && (
              <div className="space-y-4">
                <Tabs value={activeEnvTab} onValueChange={setActiveEnvTab} className="w-full">
                  <TabsList className="grid w-full grid-cols-2 bg-zinc-900">
                    <TabsTrigger value="production">Production</TabsTrigger>
                    <TabsTrigger value="preview">Preview</TabsTrigger>
                  </TabsList>
                  <TabsContent value="production" className="mt-4 space-y-4">
                    <div className="space-y-2 max-h-[250px] overflow-y-auto pr-2">
                      {formData.env.production.map((field, i) => (
                        <div key={i} className="flex gap-2">
                          <Input placeholder="KEY" className="font-mono text-xs flex-1" value={field.key} onChange={e => handleEnvChange(i, "key", e.target.value)} />
                          <Input placeholder="VALUE" className="font-mono text-xs flex-1" type="password" value={field.value} onChange={e => handleEnvChange(i, "value", e.target.value)} />
                          <Button variant="ghost" size="icon" onClick={() => removeEnvField(i)} className="text-zinc-500 hover:text-red-400">
                            <Trash2 size={14} />
                          </Button>
                        </div>
                      ))}
                      <Button variant="outline" size="sm" onClick={addEnvField} className="w-full border-dashed border-zinc-700 text-zinc-500 hover:text-white">
                        <Plus size={12} className="mr-2" /> Add Environment Variable
                      </Button>
                    </div>
                  </TabsContent>
                  <TabsContent value="preview" className="mt-4 space-y-4">
                    <div className="p-4 bg-zinc-900/30 rounded border border-zinc-800 text-center text-zinc-500 text-xs">
                      Same logic for Preview environments...
                    </div>
                  </TabsContent>
                </Tabs>
              </div>
            )}
          </div>

          <DialogFooter className="flex justify-between sm:justify-between border-t border-zinc-900 pt-6">
            {step > 1 ? (
              <Button variant="ghost" onClick={() => setStep(step - 1)}>Back</Button>
            ) : <div />}

            {step < 4 ? (
              <Button onClick={() => setStep(step + 1)} disabled={!formData.name || (formData.source_type === "git" && !formData.repo) || (formData.source_type === "docker_image" && !formData.docker_image) || (step === 2 && formData.source_type !== "docker_image" && !formData.frameworkId)}>
                Continue <ChevronRight size={14} className="ml-1" />
              </Button>
            ) : (
              <Button onClick={handleDeploy} disabled={isLoading} className="bg-indigo-600 hover:bg-indigo-700">
                {isLoading ? "Deploying..." : "Deploy Project"}
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {manageApp && (
        <Dialog open={!!manageApp} onOpenChange={() => setManageApp(null)}>
          <DialogContent className="sm:max-w-[500px] border-zinc-800 bg-zinc-950">
            <DialogHeader>
              <DialogTitle>Manage Application</DialogTitle>
              <DialogDescription>
                {manageApp.name} — {manageApp.status}
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4 mt-4">
              <div className="text-sm text-zinc-400">
                <p>Cluster: {manageApp.cluster?.name || "—"}</p>
                <p>Replicas: {manageApp.replicas || 1}</p>
                <p>Source: {manageApp.source_type || "git"}</p>
              </div>
              <div className="flex gap-2">
                <Link href={`/deployments?app=${manageApp.id}`}>
                  <Button variant="outline" size="sm">View Deployments</Button>
                </Link>
                <Button variant="outline" size="sm" onClick={() => setManageApp(null)}>Close</Button>
              </div>
            </div>
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
}
