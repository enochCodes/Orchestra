const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("orchestra_token");
}

export function setToken(token: string) {
  if (typeof window !== "undefined") {
    localStorage.setItem("orchestra_token", token);
  }
}

export function clearToken() {
  if (typeof window !== "undefined") {
    localStorage.removeItem("orchestra_token");
    localStorage.removeItem("orchestra_user");
  }
}

export function setUser(user: { id: number; email: string; display_name: string; system_role: string }) {
  if (typeof window !== "undefined") {
    localStorage.setItem("orchestra_user", JSON.stringify(user));
  }
}

export function getUser(): { id: number; email: string; display_name: string; system_role: string } | null {
  if (typeof window === "undefined") return null;
  const s = localStorage.getItem("orchestra_user");
  if (!s) return null;
  try {
    return JSON.parse(s);
  } catch {
    return null;
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  });

  if (res.status === 401) {
    clearToken();
    if (typeof window !== "undefined") {
      window.location.href = "/login";
    }
    throw new Error("Unauthorized");
  }

  const text = await res.text();
  if (!res.ok) {
    let msg = text;
    try {
      const j = JSON.parse(text);
      msg = j.message || j.error || text;
    } catch {}
    throw new Error(msg || `Request failed: ${res.status}`);
  }

  if (!text) return undefined as T;
  try {
    return JSON.parse(text) as T;
  } catch {
    return text as T;
  }
}

export const api = {
  // Auth
  login: (email: string, password: string) =>
    request<{ token: string; expires_at: string; user: any }>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),
  me: () => request<{ id: number; email: string; display_name: string; system_role: string }>("/auth/me"),
  updateProfile: (data: { display_name?: string; avatar?: string }) =>
    request<any>("/auth/me", { method: "PATCH", body: JSON.stringify(data) }),

  // Servers
  servers: {
    list: () => request<{ servers: any[]; count: number }>("/servers"),
    idle: () => request<{ servers: any[]; count: number }>("/servers/idle"),
    get: (id: number) => request<any>(`/servers/${id}`),
    register: (data: { hostname?: string; ip: string; ssh_user: string; ssh_port?: number; ssh_key: string }) =>
      request<{ server_id: number; message: string }>("/servers/register", {
        method: "POST",
        body: JSON.stringify(data),
      }),
    update: (id: number, data: any) =>
      request<any>(`/servers/${id}`, { method: "PATCH", body: JSON.stringify(data) }),
    delete: (id: number) =>
      request<any>(`/servers/${id}`, { method: "DELETE" }),
    logs: (id: number) => request<any>(`/servers/${id}/logs`),
    teams: {
      list: () => request<{ teams: any[]; count: number }>("/servers/teams"),
      create: (data: { name: string; description?: string }) =>
        request<any>("/servers/teams", { method: "POST", body: JSON.stringify(data) }),
    },
  },

  // Clusters
  clusters: {
    list: () => request<{ clusters: any[]; count: number }>("/clusters"),
    get: (id: number) => request<any>(`/clusters/${id}`),
    design: (data: {
      name: string;
      type?: string;
      manager_server_id: number;
      worker_server_ids?: number[];
      cni_plugin?: string;
      domain?: string;
    }) =>
      request<{ cluster_id: number; message: string }>("/clusters/design", {
        method: "POST",
        body: JSON.stringify(data),
      }),
  },

  // Applications
  applications: {
    list: (clusterID?: number) =>
      request<{ applications: any[]; count: number }>(
        clusterID ? `/applications?cluster_id=${clusterID}` : "/applications"
      ),
    get: (id: number) => request<any>(`/applications/${id}`),
    create: (data: any) =>
      request<any>("/applications", { method: "POST", body: JSON.stringify(data) }),
    update: (id: number, data: any) =>
      request<any>(`/applications/${id}`, { method: "PATCH", body: JSON.stringify(data) }),
    delete: (id: number) =>
      request<any>(`/applications/${id}`, { method: "DELETE" }),
    redeploy: (id: number) =>
      request<any>(`/applications/${id}/redeploy`, { method: "POST" }),
  },

  // Deployments
  deployments: {
    list: () => request<any[]>("/deployments"),
    get: (id: number) => request<any>(`/deployments/${id}`),
    logs: (id: number) => request<any>(`/deployments/${id}/logs`),
  },

  // Environments
  environments: {
    list: (clusterID?: number) =>
      request<{ environments: any[]; count: number }>(
        clusterID ? `/environments?cluster_id=${clusterID}` : "/environments"
      ),
    get: (id: number) => request<any>(`/environments/${id}`),
    create: (data: { cluster_id: number; scope: string; name: string; variables: Record<string, string> }) =>
      request<any>("/environments", { method: "POST", body: JSON.stringify(data) }),
    update: (id: number, data: any) =>
      request<any>(`/environments/${id}`, { method: "PATCH", body: JSON.stringify(data) }),
    delete: (id: number) =>
      request<any>(`/environments/${id}`, { method: "DELETE" }),
    push: (id: number) =>
      request<any>(`/environments/${id}/push`, { method: "POST" }),
  },

  // Nginx
  nginx: {
    list: (serverID?: number) =>
      request<{ configs: any[]; count: number }>(
        serverID ? `/nginx?server_id=${serverID}` : "/nginx"
      ),
    create: (data: {
      server_id: number;
      domain: string;
      upstream_port: number;
      ssl_enabled?: boolean;
      lets_encrypt?: boolean;
      application_id?: number;
    }) =>
      request<any>("/nginx", { method: "POST", body: JSON.stringify(data) }),
    delete: (id: number) =>
      request<any>(`/nginx/${id}`, { method: "DELETE" }),
  },

  // Metadata
  metadata: {
    frameworks: () => request<any[]>("/metadata/frameworks"),
    stacks: () => request<any[]>("/metadata/stacks"),
  },

  // Activities
  activities: () => request<{ activities: any[]; count: number }>("/activities"),

  // Monitoring
  monitoring: {
    overview: () => request<{ metrics: any[] }>("/monitoring/overview"),
    status: () => request<{ components: any[] }>("/monitoring/status"),
    infra: () => request<{ servers: any[]; clusters: any[]; applications: any[] }>("/monitoring/infra"),
  },
};
