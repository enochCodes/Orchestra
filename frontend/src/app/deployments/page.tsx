"use client";

import { useState, useEffect } from "react";
import { Activity, Clock, CheckCircle2, XCircle } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { SubNav } from "@/components/layout/SubNav";
import { api } from "@/lib/api";

export default function DeploymentsPage() {
    const [deployments, setDeployments] = useState<any[]>([]);

    useEffect(() => {
        const fetchDeployments = async () => {
            try {
                const data = await api.deployments.list();
                setDeployments(Array.isArray(data) ? data : []);
            } catch (err) {
                console.error("Failed to fetch deployments", err);
            }
        };
        fetchDeployments();
    }, []);

    const getStatusStyles = (status: string) => {
        switch (status) {
            case 'live':
            case 'success':
                return 'bg-emerald-500/10 border-emerald-500/50 text-emerald-500';
            case 'failed':
                return 'bg-red-500/10 border-red-500/50 text-red-500';
            default:
                return 'bg-blue-500/10 border-blue-500/50 text-blue-500';
        }
    };

    return (
        <div className="space-y-6 animate-fade-in max-w-4xl mx-auto">
            <div>
                <h1 className="text-2xl font-semibold tracking-tight text-white">Deployments</h1>
                <p className="text-zinc-400 mt-1">Audit log of application updates and rollbacks.</p>
            </div>

            <SubNav
                items={[
                    { label: "Overview", href: "/applications" },
                    { label: "Deployments", href: "/deployments" },
                    { label: "Monitoring", href: "/monitoring" },
                ]}
            />

            <div className="space-y-4">
                {deployments.length === 0 ? (
                    <div className="text-center py-12 text-zinc-500 bg-zinc-900/30 rounded-lg border border-zinc-800">
                        No deployments recorded yet.
                    </div>
                ) : (
                    deployments.map((dep, i) => (
                        <div key={dep.id} className="relative pl-8 pb-8 last:pb-0">
                            {/* Timeline Line */}
                            {i !== deployments.length - 1 && (
                                <div className="absolute left-[11px] top-8 bottom-0 w-px bg-zinc-800" />
                            )}

                            {/* Icon */}
                            <div className={`absolute left-0 top-1 w-6 h-6 rounded-full border flex items-center justify-center ${getStatusStyles(dep.status)}`}>
                                {dep.status === 'live' || dep.status === 'success' ? <CheckCircle2 size={14} /> :
                                    dep.status === 'failed' ? <XCircle size={14} /> : <Clock size={14} />}
                            </div>

                            <div className="bg-zinc-900/50 border border-zinc-800 rounded-lg p-4 hover:border-zinc-700 transition-colors">
                                <div className="flex items-center justify-between mb-2">
                                    <div className="flex items-center gap-2">
                                        <h3 className="text-sm font-semibold text-white">{dep.application?.name || 'Unknown Application'}</h3>
                                        <Badge variant="outline" className="text-[10px] h-5">{dep.version}</Badge>
                                    </div>
                                    <span className="text-xs text-zinc-500 flex items-center gap-1">
                                        <Clock size={12} /> {new Date(dep.created_at).toLocaleString()}
                                    </span>
                                </div>
                                <div className="text-sm text-zinc-400">
                                    Status: <span className="text-zinc-300 font-medium capitalize">{dep.status}</span>
                                </div>
                                {dep.status === 'failed' && (
                                    <div className="mt-3 bg-red-500/5 border border-red-500/20 rounded p-2 text-xs text-red-400 font-mono">
                                        {dep.logs || "Error: Container failed to start within timeout."}
                                    </div>
                                )}
                            </div>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
}
