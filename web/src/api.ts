import { useAuthStore, type Tokens } from "./store";
import type { Agent, Attachment, Comment, Ticket, TimelineEvent, User } from "./types";

export const API_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080/api/v1";

interface Envelope<T> {
  success: boolean;
  data: T;
  error?: string;
}

async function refreshTokens(): Promise<Tokens | null> {
  const refresh_token = useAuthStore.getState().tokens?.refresh_token;
  if (!refresh_token) return null;

  const response = await fetch(`${API_URL}/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token }),
  });
  if (!response.ok) return null;

  const body = (await response.json()) as Envelope<Tokens>;
  if (!body.success) return null;

  useAuthStore.getState().setTokens(body.data);
  return body.data;
}

export async function api<T>(path: string, init: RequestInit = {}, retry = true): Promise<T> {
  const token = useAuthStore.getState().tokens?.access_token;
  const headers = new Headers(init.headers);

  if (token) headers.set("Authorization", `Bearer ${token}`);
  if (init.body && !(init.body instanceof FormData) && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(`${API_URL}${path}`, { ...init, headers });

  if (response.status === 401 && retry && (await refreshTokens())) {
    return api<T>(path, init, false);
  }
  if (response.status === 401) {
    useAuthStore.getState().clear();
  }

  const body = (await response.json().catch(() => null)) as Envelope<T> | null;
  if (!response.ok || !body?.success) {
    throw new Error(body?.error ?? `Request failed (${response.status})`);
  }
  return body.data;
}

export const authApi = {
  login: (email: string, password: string) =>
    api<Tokens>("/auth/login", { method: "POST", body: JSON.stringify({ email, password }) }),
  register: (email: string, password: string) =>
    api<User>("/auth/register", { method: "POST", body: JSON.stringify({ email, password }) }),
  me: () => api<User>("/me"),
  logout: () =>
    api<unknown>("/auth/logout", {
      method: "POST",
      body: JSON.stringify({ refresh_token: useAuthStore.getState().tokens?.refresh_token }),
    }),
};

export const ticketsApi = {
  list: (params = new URLSearchParams()) => api<Ticket[]>(`/tickets?${params}`),
  search: (q: string) => api<Ticket[]>(`/tickets/search?${new URLSearchParams({ q })}`),
  get: (id: string) => api<Ticket>(`/tickets/${id}`),
  create: (data: Pick<Ticket, "title" | "description" | "category" | "priority">) =>
    api<Ticket>("/tickets", { method: "POST", body: JSON.stringify(data) }),
  patch: (
    id: string,
    data: Partial<Pick<Ticket, "status" | "priority" | "assignee_id" | "team_id">>,
  ) => api<Ticket>(`/tickets/${id}`, { method: "PATCH", body: JSON.stringify(data) }),
  escalate: (id: string) => api<Ticket>(`/tickets/${id}/escalate`, { method: "POST" }),
  comments: (id: string) => api<Comment[]>(`/tickets/${id}/comments`),
  addComment: (id: string, body: string, visibility = "public") =>
    api<Comment>(`/tickets/${id}/comments`, {
      method: "POST",
      body: JSON.stringify({ body, visibility }),
    }),
  agents: () => api<Agent[]>("/agents"),
  timeline: (id: string) => api<TimelineEvent[]>(`/tickets/${id}/timeline`),
  attachments: (id: string) => api<Attachment[]>(`/tickets/${id}/attachments`),
  upload: (id: string, file: File) => {
    const form = new FormData();
    form.append("file", file);
    return api<Attachment>(`/tickets/${id}/attachments`, { method: "POST", body: form });
  },
};
