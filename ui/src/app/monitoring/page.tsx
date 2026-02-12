"use client";

import { useState, useEffect } from "react";
import { Activity, Cpu, HardDrive, Router, Server, Network, AppWindow, Maximize2 } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { SubNav } from "@/components/layout/SubNav";
import { api } from "@/lib/api";

export default function MonitoringPage() {
    const [metrics, setMetrics] = useState<any[]>([]);
    const [infra, setInfra] = useState<{ servers: any[]; clusters: any[]; applications: any[] } | null>(null);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const [overviewRes, infraRes] = await Promise.all([
                    api.monitoring.overview(),
                    api.monitoring.infra(),
                ]);
                setMetrics(overviewRes.metrics || []);
                setInfra(infraRes);
            } catch (err) {
                console.error("Failed to fetch metrics", err);
            } finally {
                setIsLoading(false);
            }
        };
        fetchData();
        const interval = setInterval(fetchData, 30000);
        return () => clearInterval(interval);
    }, []);

    const getIcon = (name: string) => {
        switch (name) {
            case 'Server': return Server;
            case 'Network': return Network;
            case 'AppWindow': return AppWindow;
            default: return Activity;
        }
    };

    return (
        <div className="space-y-6 animate-fade-in max-w-6xl mx-auto">
            <div>
                <h1 className="text-2xl font-semibold tracking-tight text-white">Monitoring</h1>
                <p className="text-zinc-400 mt-1">Real-time system metrics and health status.</p>
            </div>

            <SubNav
                items={[
                    { label: "Overview", href: "/applications" },
                    { label: "Deployments", href: "/deployments" },
                    { label: "Monitoring", href: "/monitoring" },
                ]}
            />

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                {isLoading ? (
                    Array.from({ length: 4 }).map((_, i) => (
                        <div key={i} className="h-32 bg-zinc-900 animate-pulse rounded-lg border border-zinc-800" />
                    ))
                ) : (
                    metrics.map((m) => {
                        const Icon = getIcon(m.icon);
                        return (
                            <Card key={m.name} className="bg-zinc-900/50 border-zinc-800">
                                <CardContent className="p-6">
                                    <div className="flex items-center justify-between">
                                        <div>
                                            <p className="text-sm font-medium text-zinc-400">{m.name}</p>
                                            <div className="flex items-baseline gap-1 mt-2">
                                                <span className="text-2xl font-bold text-white">{m.value}</span>
                                                <span className="text-sm text-zinc-500">{m.unit}</span>
                                            </div>
                                        </div>
                                        <div className={`p-3 rounded-full bg-zinc-900 border border-zinc-800 ${m.color}`}>
                                            <Icon size={20} />
                                        </div>
                                    </div>
                                    {/* Progress Bar */}
                                    <div className="mt-4 h-1.5 w-full bg-zinc-800 rounded-full overflow-hidden">
                                        <div className={`h-full rounded-full ${m.track}`} style={{ width: `${m.value > 100 ? 100 : m.value}%` }} />
                                    </div>
                                </CardContent>
                            </Card>
                        );
                    })
                )}
            </div>

            {/* Infra Details - Full-screen ready for DevOps */}
            <Card className="bg-zinc-900/30 border-zinc-800">
                <CardHeader className="flex flex-row items-center justify-between">
                    <CardTitle className="flex items-center gap-2">
                        <Maximize2 size={18} className="text-zinc-400" />
                        Infrastructure Details
                    </CardTitle>
                    <span className="text-xs text-zinc-500">Full-screen for team dashboards</span>
                </CardHeader>
                <CardContent className="space-y-6 border-t border-zinc-900/50 pt-6">
                    {infra && (
                        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                            <div>
                                <h4 className="text-sm font-medium text-zinc-400 mb-3 flex items-center gap-2">
                                    <Server size={14} /> Servers ({infra.servers?.length || 0})
                                </h4>
                                <div className="space-y-2">
                                    {(infra.servers || []).slice(0, 5).map((s: any) => (
                                        <div key={s.id} className="flex justify-between text-sm bg-zinc-900/50 rounded p-2">
                                            <span className="text-white font-mono">{s.hostname || s.ip}</span>
                                            <span className={`text-xs ${s.status === "ready" ? "text-emerald-500" : "text-amber-500"}`}>{s.status}</span>
                                        </div>
                                    ))}
                                    {(infra.servers?.length || 0) > 5 && <p className="text-xs text-zinc-500">+ {(infra.servers?.length || 0) - 5} more</p>}
                                </div>
                            </div>
                            <div>
                                <h4 className="text-sm font-medium text-zinc-400 mb-3 flex items-center gap-2">
                                    <Network size={14} /> Clusters ({infra.clusters?.length || 0})
                                </h4>
                                <div className="space-y-2">
                                    {(infra.clusters || []).map((c: any) => (
                                        <div key={c.id} className="flex justify-between text-sm bg-zinc-900/50 rounded p-2">
                                            <span className="text-white">{c.name}</span>
                                            <span className="text-xs text-zinc-500">{c.worker_count || 0} workers</span>
                                        </div>
                                    ))}
                                </div>
                            </div>
                            <div>
                                <h4 className="text-sm font-medium text-zinc-400 mb-3 flex items-center gap-2">
                                    <AppWindow size={14} /> Applications ({infra.applications?.length || 0})
                                </h4>
                                <div className="space-y-2">
                                    {(infra.applications || []).slice(0, 5).map((a: any) => (
                                        <div key={a.id} className="flex justify-between text-sm bg-zinc-900/50 rounded p-2">
                                            <span className="text-white">{a.name}</span>
                                            <span className={`text-xs ${a.status === "running" ? "text-emerald-500" : "text-amber-500"}`}>{a.status}</span>
                                        </div>
                                    ))}
                                    {(infra.applications?.length || 0) > 5 && <p className="text-xs text-zinc-500">+ {(infra.applications?.length || 0) - 5} more</p>}
                                </div>
                            </div>
                        </div>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}
