import type { ButtonHTMLAttributes, PropsWithChildren } from "react";
import type { Priority, Status, Ticket } from "../types";

export function Spinner() {
  return <div className="h-6 w-6 animate-spin rounded-full border-2 border-slate-200 border-t-indigo-600" />;
}

export function Loading({ label = "Loading…" }: { label?: string }) {
  return (
    <div className="flex min-h-52 flex-col items-center justify-center gap-3 text-sm text-slate-500">
      <Spinner />
      <span>{label}</span>
    </div>
  );
}

export function ErrorState({ error }: { error: Error }) {
  return (
    <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
      {error.message}
    </div>
  );
}

export function Empty({ title, detail }: { title: string; detail: string }) {
  return (
    <div className="rounded-xl border border-dashed border-slate-300 bg-white p-10 text-center">
      <h3 className="font-semibold text-slate-800">{title}</h3>
      <p className="mt-1 text-sm text-slate-500">{detail}</p>
    </div>
  );
}

export function Card({ children, className = "" }: PropsWithChildren<{ className?: string }>) {
  return (
    <section className={`rounded-xl border border-slate-200 bg-white shadow-sm ${className}`}>
      {children}
    </section>
  );
}

const statusClasses: Record<Status, string> = {
  open: "bg-blue-50 text-blue-700",
  pending: "bg-amber-50 text-amber-700",
  resolved: "bg-emerald-50 text-emerald-700",
  closed: "bg-slate-100 text-slate-600",
};

const priorityClasses: Record<Priority, string> = {
  low: "bg-slate-100 text-slate-600",
  medium: "bg-violet-50 text-violet-700",
  high: "bg-orange-50 text-orange-700",
  urgent: "bg-red-50 text-red-700",
};

type BadgeTone = Status | Priority | "slate" | "breach";

export function Badge({ children, tone = "slate" }: PropsWithChildren<{ tone?: BadgeTone }>) {
  const classes =
    tone === "slate"
      ? "bg-slate-100 text-slate-600"
      : tone === "breach"
        ? "bg-red-50 text-red-700"
        : statusClasses[tone as Status] ?? priorityClasses[tone as Priority] ?? "bg-slate-100 text-slate-600";

  return (
    <span className={`inline-flex rounded-full px-2.5 py-1 text-xs font-semibold capitalize ${classes}`}>
      {children}
    </span>
  );
}

function slaLabel(ticket: Ticket): string | null {
  if (ticket.breached_at) return "SLA breached";
  if (ticket.sla_paused_at) return "SLA paused";
  if (!ticket.sla_due_at) return null;
  const due = new Date(ticket.sla_due_at).getTime();
  const hoursLeft = (due - Date.now()) / 3_600_000;
  if (hoursLeft < 0) return "SLA overdue";
  if (hoursLeft < 4) return "SLA soon";
  return null;
}

export function TicketBadges({ ticket }: { ticket: Ticket }) {
  const sla = slaLabel(ticket);
  return (
    <div className="flex flex-wrap gap-2">
      <Badge tone={ticket.status}>{ticket.status}</Badge>
      <Badge tone={ticket.priority}>{ticket.priority}</Badge>
      {sla && <Badge tone="breach">{sla}</Badge>}
    </div>
  );
}

export function Button({
  children,
  className = "",
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      className={`rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-indigo-700 disabled:cursor-not-allowed disabled:opacity-50 ${className}`}
      {...props}
    >
      {children}
    </button>
  );
}
