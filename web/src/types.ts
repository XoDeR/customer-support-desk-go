export type Role = "customer" | "agent" | "admin";
export type Status = "open" | "pending" | "resolved" | "closed";
export type Priority = "low" | "medium" | "high" | "urgent";
export type Category = "billing" | "technical" | "account" | "other";

export interface User {
  id: string;
  email: string;
  role: Role;
  status: string;
  created_at: string;
}

export interface Ticket {
  id: string;
  title: string;
  description: string;
  customer_id: string;
  assignee_id?: string | null;
  team_id?: string | null;
  status: Status;
  priority: Priority;
  category: Category | string;
  sla_due_at?: string | null;
  sla_paused_at?: string | null;
  sla_remaining_seconds?: number | null;
  breached_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface Comment {
  id: string;
  ticket_id: string;
  author_id: string;
  body: string;
  visibility: "public" | "internal" | string;
  created_at: string;
  optimistic?: boolean;
}

export interface Agent {
  id: string;
  email: string;
  role: Role;
}

export interface TimelineEvent {
  id: string;
  event_type: string;
  payload?: unknown;
  created_at: string;
}

export interface Attachment {
  id: string;
  ticket_id: string;
  filename: string;
  mime_type: string;
  size_bytes: number;
  created_at: string;
}

export interface CannedReply {
  id: string;
  title: string;
  body: string;
  team_id?: string | null;
  created_by: string;
}

export interface SavedFilter {
  id: string;
  name: string;
  query: {
    q?: string;
    status?: string;
    priority?: string;
    [key: string]: unknown;
  };
}

export interface Tag {
  id: string;
  name: string;
}
