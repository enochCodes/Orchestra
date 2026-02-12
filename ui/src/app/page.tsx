"use client";

import { useState, useEffect } from "react";
import {
  Server,
  Network,
  Box,
  Rocket,
  ArrowUpRight,
  Clock,
  CheckCircle2,
  AlertTriangle,
  Cpu,
  Activity,
  GitCommit,
  Loader2,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import Link from "next/link";
import { api } from "@/lib/api";

const activityIcons: Record<string, typeof CheckCircle2> = {
  cluster_provisioned: CheckCircle2,
  server_registered: Server,
  app_deployed: GitCommit,
  deployment_failed: AlertTriangle,
  user_login: Activity,
  cluster_created: Network,
};

export default function DashboardPage() {
  const [metrics, setMetrics] = useState<any[]>([]);
  const [activities, setActivities] = useState<any[]>([]);
  const [status, setStatus] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetch = async () => {
      try {
        const [overviewRes, activitiesRes, statusRes] = await Promise.all([
          api.monitoring.overview(),
          api.activities(),
          api.monitoring.status(),
        ]);
        setMetrics(overviewRes.metrics || []);
        setActivities(activitiesRes.activities || []);
        setStatus(statusRes.components || []);
      } catch (err) {
        console.error("Failed to fetch dashboard data", err);
      } finally {
        setLoading(false);
      }
    };
    fetch();
  }, []);

  const formatTime = (dateStr: string) => {
    const d = new Date(dateStr);
    const now = new Date();
    const diff = Math.floor((now.getTime() - d.getTime()) / 60000);
    if (diff < 1) return "Just now";
    if (diff < 60) return `${diff} mins ago`;
    const h = Math.floor(diff / 60);
    if (h < 24) return `${h} hour${h > 1 ? "s" : ""} ago`;
    return d.toLocaleDateString();
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Loader2 className="w-8 h-8 animate-spin text-zinc-500" />
      </div>
    );
  }

  const stats = [
    { label: "Total Servers", value: String(metrics.find((m) => m.name === "Total Servers")?.value ?? 0), icon: Server, color: "text-indigo-400", bg: "bg-indigo-500/10" },
    { label: "Active Clusters", value: String(metrics.find((m) => m.name === "Active Clusters")?.value ?? 0), icon: Network, color: "text-emerald-400", bg: "bg-emerald-500/10" },
    { label: "Applications", value: String(metrics.find((m) => m.name === "Running Apps")?.value ?? 0), icon: Box, color: "text-cyan-400", bg: "bg-cyan-500/10" },
    { label: "Deployments", value: String(metrics.find((m) => m.name === "Deployments")?.value ?? 0), icon: Rocket, color: "text-amber-400", bg: "bg-amber-500/10" },
  ];

  return (
    <div className="space-y-8 animate-fade-in">
      <div className="flex items-end justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-white">Dashboard</h1>
          <p className="text-zinc-400 mt-2">Overview of your infrastructure health and activity.</p>
        </div>
        <div className="flex gap-3">
          <Button variant="outline" size="sm">
            <Clock className="w-4 h-4 mr-2" />
            Last 24 Hours
          </Button>
          <Button size="sm">
            <Activity className="w-4 h-4 mr-2" />
            Live View
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat) => {
          const Icon = stat.icon;
          return (
            <Card key={stat.label} className="border-zinc-800 bg-zinc-900/50 hover:bg-zinc-900/80 transition-all duration-300">
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-zinc-400">{stat.label}</p>
                    <div className="text-3xl font-bold text-white mt-2">{stat.value}</div>
                  </div>
                  <div className={`p-3 rounded-xl ${stat.bg}`}>
                    <Icon className={`w-6 h-6 ${stat.color}`} />
                  </div>
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <Card className="lg:col-span-2 border-zinc-800">
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
            <CardDescription>Latest actions across your infrastructure.</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-6">
              {activities.length === 0 ? (
                <p className="text-sm text-zinc-500">No recent activity.</p>
              ) : (
                activities.slice(0, 10).map((item, i) => {
                  const Icon = activityIcons[item.type] || Activity;
                  const typeClass = item.type?.includes("failed") ? "warning" : item.type?.includes("registered") || item.type?.includes("login") ? "info" : "success";
                  return (
                    <div key={item.id} className="flex gap-4">
                      <div className="mt-1 relative">
                        <div className="absolute top-8 left-1/2 -ml-px w-0.5 h-full bg-zinc-800 last:hidden" />
                        <div
                          className={`w-8 h-8 rounded-full border-2 border-zinc-900 flex items-center justify-center 
                            ${typeClass === "success" ? "bg-emerald-500/20 text-emerald-500" : ""}
                            ${typeClass === "warning" ? "bg-amber-500/20 text-amber-500" : ""}
                            ${typeClass === "info" ? "bg-blue-500/20 text-blue-500" : ""}
                          `}
                        >
                          <Icon size={14} strokeWidth={2.5} />
                        </div>
                      </div>
                      <div className="flex-1 pb-6 border-b border-zinc-800/50 last:border-0 last:pb-0">
                        <p className="text-sm font-medium text-zinc-200">{item.message}</p>
                        <p className="text-xs text-zinc-500 mt-1">{formatTime(item.created_at)}</p>
                      </div>
                    </div>
                  );
                })
              )}
            </div>
          </CardContent>
        </Card>

        <div className="space-y-6">
          <Card className="border-zinc-800 bg-gradient-to-br from-indigo-900/10 via-zinc-900/50 to-zinc-900/50">
            <CardHeader>
              <CardTitle>Quick Actions</CardTitle>
            </CardHeader>
            <CardContent className="grid gap-3">
              <Link href="/servers">
                <Button className="w-full justify-start" variant="secondary">
                  <Server className="w-4 h-4 mr-2" />
                  Register Server
                </Button>
              </Link>
              <Link href="/clusters">
                <Button className="w-full justify-start" variant="secondary">
                  <Network className="w-4 h-4 mr-2" />
                  Design Cluster
                </Button>
              </Link>
              <Link href="/applications">
                <Button className="w-full justify-start" variant="secondary">
                  <Box className="w-4 h-4 mr-2" />
                  Deploy Application
                </Button>
              </Link>
            </CardContent>
          </Card>

          <Card className="border-zinc-800">
            <CardHeader>
              <CardTitle>System Status</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {status.length === 0 ? (
                <p className="text-sm text-zinc-500">Unable to fetch status.</p>
              ) : (
                status.map((svc: any) => (
                  <div key={svc.name} className="flex items-center justify-between text-sm">
                    <span className="text-zinc-400">{svc.name}</span>
                    <Badge variant="outline" className="gap-2 border-zinc-700 bg-zinc-800/50">
                      <span className={`w-1.5 h-1.5 rounded-full ${svc.healthy ? "bg-emerald-500" : "bg-red-500"} animate-pulse`} />
                      {svc.status}
                    </Badge>
                  </div>
                ))
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
