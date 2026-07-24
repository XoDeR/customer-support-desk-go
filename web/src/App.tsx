import { useCallback, useEffect, useMemo, useState } from "react";
import {
  Link,
  Navigate,
  Outlet,
  Route,
  Routes,
  useLocation,
  useNavigate,
  useParams,
  useSearchParams,
} from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { authApi, opsApi, ticketsApi } from "./api";
import { useAuthStore } from "./store";
import type { Comment, Priority, Status, Ticket } from "./types";
import { Badge, Button, Card, Empty, ErrorState, Loading, TicketBadges } from "./components/ui";
import { useRealtime } from "./hooks/useRealtime";

const isStaff = (role?: string) => role === "agent" || role === "admin";
const homeFor = (role?: string) => (isStaff(role) ? "/agent/tickets" : "/tickets");

const formatDate = (value?: string | null) =>
  value
    ? new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" }).format(new Date(value))
    : "—";

function AuthPage({ register = false }: { register?: boolean }) {
  const navigate = useNavigate();
  const setTokens = useAuthStore((s) => s.setTokens);
  const setUser = useAuthStore((s) => s.setUser);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const registered = Boolean((useLocation().state as { registered?: boolean } | null)?.registered);

  const mutation = useMutation({
    mutationFn: async () => {
      if (register) {
        await authApi.register(email, password);
        return null;
      }
      const tokens = await authApi.login(email, password);
      setTokens(tokens);
      return authApi.me();
    },
    onSuccess: (user) => {
      if (register) navigate("/login", { state: { registered: true } });
      else if (user) {
        setUser(user);
        navigate(homeFor(user.role));
      }
    },
  });

  return (
    <main className="grid min-h-screen place-items-center bg-slate-50 p-5">
      <Card className="w-full max-w-md p-8">
        <div className="mb-7">
          <p className="text-sm font-semibold text-indigo-600">CUSTOMER SUPPORT & SLA DESK</p>
          <h1 className="mt-2 text-2xl font-bold text-slate-900">
            {register ? "Create your account" : "Welcome back"}
          </h1>
          <p className="mt-2 text-sm text-slate-500">
            {register
              ? "Register as a customer and start tracking support requests."
              : "Sign in to manage tickets, replies, and SLA status."}
          </p>
        </div>
        {registered && (
          <p className="mb-4 rounded-lg bg-emerald-50 p-3 text-sm text-emerald-700">
            Account created. You can sign in now.
          </p>
        )}
        {mutation.error && <ErrorState error={mutation.error as Error} />}
        <form
          className="mt-4 space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            mutation.mutate();
          }}
        >
          <label className="block text-sm font-medium text-slate-700">
            Email
            <input
              required
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="field mt-1"
            />
          </label>
          <label className="block text-sm font-medium text-slate-700">
            Password
            <input
              required
              minLength={8}
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="field mt-1"
            />
          </label>
          <Button className="w-full" disabled={mutation.isPending}>
            {mutation.isPending ? "Please wait…" : register ? "Create account" : "Sign in"}
          </Button>
        </form>
        <p className="mt-6 text-center text-sm text-slate-500">
          {register ? "Already have an account?" : "New customer?"}{" "}
          <Link className="font-semibold text-indigo-600" to={register ? "/login" : "/register"}>
            {register ? "Sign in" : "Create an account"}
          </Link>
        </p>
      </Card>
    </main>
  );
}

function Shell() {
  const user = useAuthStore((s) => s.user);
  const clear = useAuthStore((s) => s.clear);
  const navigate = useNavigate();
  const client = useQueryClient();

  const onEvent = useCallback(() => {
    client.invalidateQueries({ queryKey: ["tickets"] });
    client.invalidateQueries({ queryKey: ["ticket"] });
    client.invalidateQueries({ queryKey: ["comments"] });
    client.invalidateQueries({ queryKey: ["timeline"] });
  }, [client]);

  useRealtime(onEvent);

  return (
    <div className="min-h-screen bg-slate-50">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-5">
          <Link to={homeFor(user?.role)} className="font-bold tracking-tight text-slate-900">
            Support<span className="text-indigo-600">Desk</span>
          </Link>
          <nav className="flex items-center gap-4 text-sm">
            <Link className="text-slate-600 hover:text-slate-900" to={homeFor(user?.role)}>
              {isStaff(user?.role) ? "Ticket queue" : "My tickets"}
            </Link>
            {isStaff(user?.role) && (
              <Link className="text-slate-600 hover:text-slate-900" to="/agent/tools">
                Agent tools
              </Link>
            )}
            {!isStaff(user?.role) && (
              <Link className="text-slate-600 hover:text-slate-900" to="/tickets/new">
                New ticket
              </Link>
            )}
            <span className="hidden text-slate-400 sm:block">{user?.email}</span>
            {user?.role && <Badge>{user.role}</Badge>}
            <button
              className="font-medium text-slate-600 hover:text-slate-900"
              onClick={async () => {
                try {
                  await authApi.logout();
                } finally {
                  clear();
                  navigate("/login");
                }
              }}
            >
              Sign out
            </button>
          </nav>
        </div>
      </header>
      <main className="mx-auto max-w-7xl p-5 sm:p-8">
        <Outlet />
      </main>
    </div>
  );
}

function Protected({ staffOnly = false }: { staffOnly?: boolean }) {
  const tokens = useAuthStore((s) => s.tokens);
  const user = useAuthStore((s) => s.user);
  const setUser = useAuthStore((s) => s.setUser);
  const clear = useAuthStore((s) => s.clear);

  const me = useQuery({
    queryKey: ["me"],
    queryFn: authApi.me,
    enabled: !!tokens && !user,
    retry: false,
  });

  useEffect(() => {
    if (me.data) setUser(me.data);
  }, [me.data, setUser]);

  useEffect(() => {
    if (me.isError) clear();
  }, [me.isError, clear]);

  if (!tokens) return <Navigate to="/login" replace />;
  if (!user && me.isLoading) return <Loading label="Checking session…" />;
  if (!user && me.isError) return <Navigate to="/login" replace />;
  if (staffOnly && !isStaff(user?.role)) return <Navigate to="/tickets" replace />;
  return <Shell />;
}

function TicketsList({ staff = false }: { staff?: boolean }) {
  const client = useQueryClient();
  const [params, setParams] = useSearchParams();
  const [filterName, setFilterName] = useState("");
  const q = params.get("q") ?? "";
  const status = params.get("status") ?? "";
  const priority = params.get("priority") ?? "";

  const { data, isLoading, error } = useQuery({
    queryKey: ["tickets", q],
    queryFn: () => ticketsApi.list(new URLSearchParams(q ? { q } : {})),
  });
  const savedFilters = useQuery({
    queryKey: ["saved-filters"],
    queryFn: opsApi.savedFilters,
    enabled: staff,
  });

  const tickets = useMemo(
    () => (data ?? []).filter((t) => (!status || t.status === status) && (!priority || t.priority === priority)),
    [data, status, priority],
  );

  const change = (key: string, value: string) => {
    const next = new URLSearchParams(params);
    if (value) next.set(key, value);
    else next.delete(key);
    setParams(next);
  };

  const saveFilter = useMutation({
    mutationFn: () =>
      opsApi.createSavedFilter(filterName.trim(), {
        ...(q ? { q } : {}),
        ...(status ? { status } : {}),
        ...(priority ? { priority } : {}),
      }),
    onSuccess: () => {
      setFilterName("");
      client.invalidateQueries({ queryKey: ["saved-filters"] });
    },
  });

  const deleteFilter = useMutation({
    mutationFn: (id: string) => opsApi.deleteSavedFilter(id),
    onSuccess: () => client.invalidateQueries({ queryKey: ["saved-filters"] }),
  });

  if (isLoading) return <Loading label="Loading tickets…" />;
  if (error) return <ErrorState error={error as Error} />;

  return (
    <>
      <div className="mb-7 flex flex-wrap items-end justify-between gap-4">
        <div>
          <p className="text-sm font-semibold text-indigo-600">
            {staff ? "AGENT WORKSPACE" : "CUSTOMER PORTAL"}
          </p>
          <h1 className="mt-1 text-3xl font-bold text-slate-900">
            {staff ? "Ticket queue" : "My tickets"}
          </h1>
        </div>
        {!staff && (
          <Link to="/tickets/new">
            <Button>New ticket</Button>
          </Link>
        )}
      </div>

      <Card className="mb-5 p-4">
        <div className="grid gap-3 sm:grid-cols-3">
          <input
            aria-label="Search tickets"
            value={q}
            onChange={(e) => change("q", e.target.value)}
            placeholder="Search tickets…"
            className="field"
          />
          <select
            aria-label="Filter by status"
            className="field"
            value={status}
            onChange={(e) => change("status", e.target.value)}
          >
            <option value="">All statuses</option>
            {(["open", "pending", "resolved", "closed"] as Status[]).map((x) => (
              <option key={x} value={x}>
                {x}
              </option>
            ))}
          </select>
          <select
            aria-label="Filter by priority"
            className="field"
            value={priority}
            onChange={(e) => change("priority", e.target.value)}
          >
            <option value="">All priorities</option>
            {(["low", "medium", "high", "urgent"] as Priority[]).map((x) => (
              <option key={x} value={x}>
                {x}
              </option>
            ))}
          </select>
        </div>
        {staff && (
          <div className="mt-4 space-y-3 border-t border-slate-100 pt-4">
            <div className="flex flex-wrap gap-2">
              {(savedFilters.data ?? []).map((filter) => (
                <div key={filter.id} className="flex items-center gap-1">
                  <button
                    type="button"
                    className="rounded-full border border-slate-200 bg-white px-3 py-1 text-xs font-medium text-slate-700 hover:border-indigo-300 hover:text-indigo-700"
                    onClick={() => {
                      const next = new URLSearchParams();
                      if (typeof filter.query.q === "string") next.set("q", filter.query.q);
                      if (typeof filter.query.status === "string") next.set("status", filter.query.status);
                      if (typeof filter.query.priority === "string") {
                        next.set("priority", filter.query.priority);
                      }
                      setParams(next);
                    }}
                  >
                    {filter.name}
                  </button>
                  <button
                    type="button"
                    aria-label={`Delete filter ${filter.name}`}
                    className="text-xs text-slate-400 hover:text-red-600"
                    onClick={() => deleteFilter.mutate(filter.id)}
                  >
                    ×
                  </button>
                </div>
              ))}
              {!savedFilters.data?.length && !savedFilters.isLoading && (
                <p className="text-xs text-slate-500">No saved filters yet.</p>
              )}
            </div>
            <form
              className="flex flex-wrap gap-2"
              onSubmit={(e) => {
                e.preventDefault();
                if (filterName.trim()) saveFilter.mutate();
              }}
            >
              <input
                className="field max-w-xs"
                placeholder="Name for current filters"
                value={filterName}
                onChange={(e) => setFilterName(e.target.value)}
              />
              <Button type="submit" disabled={saveFilter.isPending || !filterName.trim()}>
                {saveFilter.isPending ? "Saving…" : "Save filter"}
              </Button>
            </form>
            {saveFilter.error && <ErrorState error={saveFilter.error as Error} />}
          </div>
        )}
      </Card>

      {!tickets.length ? (
        <Empty
          title="No tickets found"
          detail="Try adjusting the filters or create a new support request."
        />
      ) : (
        <Card className="divide-y divide-slate-100">
          {tickets.map((ticket) => (
            <Link
              key={ticket.id}
              to={staff ? `/agent/tickets/${ticket.id}` : `/tickets/${ticket.id}`}
              className="block p-5 transition hover:bg-slate-50"
            >
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <h2 className="font-semibold text-slate-900">{ticket.title}</h2>
                  <p className="mt-1 line-clamp-1 text-sm text-slate-500">{ticket.description}</p>
                  <p className="mt-2 text-xs text-slate-400">Updated {formatDate(ticket.updated_at)}</p>
                </div>
                <TicketBadges ticket={ticket} />
              </div>
            </Link>
          ))}
        </Card>
      )}
    </>
  );
}

function NewTicket() {
  const navigate = useNavigate();
  const [form, setForm] = useState({
    title: "",
    description: "",
    category: "technical",
    priority: "medium" as Priority,
  });

  const mutation = useMutation({
    mutationFn: () => ticketsApi.create(form),
    onSuccess: (ticket) => navigate(`/tickets/${ticket.id}`),
  });

  return (
    <div className="mx-auto max-w-2xl">
      <Link className="text-sm font-medium text-indigo-600" to="/tickets">
        ← My tickets
      </Link>
      <h1 className="mt-4 text-3xl font-bold">Create a support request</h1>
      <Card className="mt-6 p-6">
        {mutation.error && <ErrorState error={mutation.error as Error} />}
        <form
          className="mt-4 space-y-5"
          onSubmit={(e) => {
            e.preventDefault();
            mutation.mutate();
          }}
        >
          <label className="block text-sm font-medium">
            Subject
            <input
              required
              className="field mt-1"
              value={form.title}
              onChange={(e) => setForm({ ...form, title: e.target.value })}
            />
          </label>
          <label className="block text-sm font-medium">
            How can we help?
            <textarea
              required
              rows={7}
              className="field mt-1"
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
            />
          </label>
          <div className="grid gap-4 sm:grid-cols-2">
            <label className="block text-sm font-medium">
              Category
              <select
                className="field mt-1"
                value={form.category}
                onChange={(e) => setForm({ ...form, category: e.target.value })}
              >
                {["technical", "billing", "account", "other"].map((x) => (
                  <option key={x} value={x}>
                    {x}
                  </option>
                ))}
              </select>
            </label>
            <label className="block text-sm font-medium">
              Priority
              <select
                className="field mt-1"
                value={form.priority}
                onChange={(e) => setForm({ ...form, priority: e.target.value as Priority })}
              >
                {(["low", "medium", "high", "urgent"] as Priority[]).map((x) => (
                  <option key={x} value={x}>
                    {x}
                  </option>
                ))}
              </select>
            </label>
          </div>
          <Button disabled={mutation.isPending}>
            {mutation.isPending ? "Creating…" : "Submit request"}
          </Button>
        </form>
      </Card>
    </div>
  );
}

function TicketDetail({ staff = false }: { staff?: boolean }) {
  const { id = "" } = useParams();
  const user = useAuthStore((s) => s.user);
  const client = useQueryClient();
  const navigate = useNavigate();
  const [body, setBody] = useState("");
  const [internal, setInternal] = useState(false);
  const [file, setFile] = useState<File | null>(null);

  const ticket = useQuery({ queryKey: ["ticket", id], queryFn: () => ticketsApi.get(id) });
  const comments = useQuery({ queryKey: ["comments", id], queryFn: () => ticketsApi.comments(id) });
  const attachments = useQuery({
    queryKey: ["attachments", id],
    queryFn: () => ticketsApi.attachments(id),
  });
  const timeline = useQuery({
    queryKey: ["timeline", id],
    queryFn: () => ticketsApi.timeline(id),
    enabled: staff,
  });
  const canned = useQuery({
    queryKey: ["canned-replies"],
    queryFn: opsApi.cannedReplies,
    enabled: staff,
  });
  const tags = useQuery({
    queryKey: ["tags"],
    queryFn: opsApi.tags,
    enabled: staff,
  });

  const commentMutation = useMutation({
    mutationFn: () => ticketsApi.addComment(id, body, internal ? "internal" : "public"),
    onMutate: async () => {
      await client.cancelQueries({ queryKey: ["comments", id] });
      const previous = client.getQueryData<Comment[]>(["comments", id]);
      const optimistic: Comment = {
        id: `temporary-${Date.now()}`,
        ticket_id: id,
        author_id: user?.id ?? "",
        body,
        visibility: internal ? "internal" : "public",
        created_at: new Date().toISOString(),
        optimistic: true,
      };
      client.setQueryData<Comment[]>(["comments", id], (old = []) => [...old, optimistic]);
      return { previous };
    },
    onError: (_error, _variables, context) => {
      client.setQueryData(["comments", id], context?.previous);
    },
    onSuccess: () => setBody(""),
    onSettled: () => {
      client.invalidateQueries({ queryKey: ["comments", id] });
      client.invalidateQueries({ queryKey: ["timeline", id] });
    },
  });

  const update = useMutation({
    mutationFn: (data: object) => ticketsApi.patch(id, data),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: ["ticket", id] });
      client.invalidateQueries({ queryKey: ["timeline", id] });
      client.invalidateQueries({ queryKey: ["tickets"] });
    },
  });

  const escalate = useMutation({
    mutationFn: () => ticketsApi.escalate(id),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: ["ticket", id] });
      client.invalidateQueries({ queryKey: ["timeline", id] });
    },
  });

  const upload = useMutation({
    mutationFn: async () => {
      if (!file) throw new Error("Choose a file before uploading.");
      return ticketsApi.upload(id, file);
    },
    onSuccess: () => {
      setFile(null);
      client.invalidateQueries({ queryKey: ["attachments", id] });
    },
  });

  const attachTag = useMutation({
    mutationFn: (tagId: string) => opsApi.attachTag(id, tagId),
    onSuccess: () => client.invalidateQueries({ queryKey: ["timeline", id] }),
  });

  if (ticket.isLoading || comments.isLoading) return <Loading label="Loading ticket…" />;
  if (ticket.error || comments.error) {
    return <ErrorState error={(ticket.error ?? comments.error) as Error} />;
  }

  const t = ticket.data!;

  return (
    <div className="grid gap-6 lg:grid-cols-[1fr_19rem]">
      <div>
        <button
          className="text-sm font-medium text-indigo-600"
          onClick={() => navigate(staff ? "/agent/tickets" : "/tickets")}
        >
          ← Back to tickets
        </button>
        <div className="mt-4 flex flex-wrap justify-between gap-4">
          <div>
            <p className="text-sm text-slate-500">#{t.id.slice(0, 8)}</p>
            <h1 className="mt-1 text-3xl font-bold text-slate-900">{t.title}</h1>
          </div>
          <TicketBadges ticket={t} />
        </div>

        <Card className="mt-6 p-6">
          <p className="whitespace-pre-wrap text-slate-700">{t.description}</p>
        </Card>

        <h2 className="mt-8 text-lg font-bold">Conversation</h2>
        <div className="mt-3 space-y-3">
          {comments.data!.length === 0 ? (
            <Empty title="No replies yet" detail="Replies and updates will appear here." />
          ) : (
            comments.data!.map((comment) => (
              <Card
                key={comment.id}
                className={`p-4 ${comment.visibility === "internal" ? "border-amber-200 bg-amber-50" : ""}`}
              >
                <div className="mb-2 flex justify-between text-xs text-slate-500">
                  <span>
                    {comment.visibility === "internal"
                      ? "Internal note"
                      : comment.author_id === user?.id
                        ? "You"
                        : "Support team"}
                    {comment.optimistic ? " · Sending…" : ""}
                  </span>
                  <span>{formatDate(comment.created_at)}</span>
                </div>
                <p className="whitespace-pre-wrap text-sm text-slate-700">{comment.body}</p>
              </Card>
            ))
          )}
        </div>

        <Card className="mt-5 p-5">
          {commentMutation.error && <ErrorState error={commentMutation.error as Error} />}
          <form
            onSubmit={(e) => {
              e.preventDefault();
              if (body.trim()) commentMutation.mutate();
            }}
          >
            {staff && (
              <label className="mb-3 block text-sm font-medium text-slate-700">
                Insert canned reply
                <select
                  className="field mt-1"
                  defaultValue=""
                  onChange={(e) => {
                    const reply = canned.data?.find((item) => item.id === e.target.value);
                    if (reply) setBody(reply.body);
                    e.target.value = "";
                  }}
                >
                  <option value="">Choose a template…</option>
                  {(canned.data ?? []).map((reply) => (
                    <option key={reply.id} value={reply.id}>
                      {reply.title}
                    </option>
                  ))}
                </select>
              </label>
            )}
            <textarea
              className="field"
              rows={4}
              value={body}
              onChange={(e) => setBody(e.target.value)}
              placeholder={internal ? "Add an internal note…" : "Write a reply…"}
            />
            {staff && (
              <label className="mt-3 flex items-center gap-2 text-sm text-slate-600">
                <input
                  type="checkbox"
                  checked={internal}
                  onChange={(e) => setInternal(e.target.checked)}
                />
                Internal note (customers cannot see this)
              </label>
            )}
            <div className="mt-4 flex flex-wrap items-center justify-between gap-3">
              <label className="text-sm text-slate-600">
                Attach file
                <input
                  className="ml-2 text-xs"
                  type="file"
                  onChange={(e) => setFile(e.target.files?.[0] ?? null)}
                />
              </label>
              <Button disabled={commentMutation.isPending || !body.trim()}>
                {commentMutation.isPending ? "Sending…" : "Send reply"}
              </Button>
            </div>
          </form>
          {file && (
            <div className="mt-3 flex items-center gap-3">
              <button
                className="text-sm font-medium text-indigo-600"
                onClick={() => upload.mutate()}
                disabled={upload.isPending}
              >
                {upload.isPending ? "Uploading…" : `Upload ${file.name}`}
              </button>
              {upload.error && <span className="text-sm text-red-600">{(upload.error as Error).message}</span>}
            </div>
          )}
        </Card>

        {staff && (
          <Card className="mt-5 p-5">
            <h2 className="font-bold">Tags</h2>
            <div className="mt-3 flex flex-wrap gap-2">
              {(tags.data ?? []).map((tag) => (
                <button
                  key={tag.id}
                  type="button"
                  className="rounded-full border border-slate-200 px-3 py-1 text-xs font-medium text-slate-700 hover:border-indigo-300 hover:text-indigo-700"
                  disabled={attachTag.isPending}
                  onClick={() => attachTag.mutate(tag.id)}
                >
                  {tag.name}
                </button>
              ))}
              {!tags.data?.length && <p className="text-sm text-slate-500">No tags yet — create some in Agent tools.</p>}
            </div>
            {attachTag.error && (
              <div className="mt-3">
                <ErrorState error={attachTag.error as Error} />
              </div>
            )}
          </Card>
        )}

        <h2 className="mt-8 text-lg font-bold">Attachments</h2>
        <div className="mt-3">
          {attachments.isLoading ? (
            <Loading label="Loading attachments…" />
          ) : attachments.error ? (
            <ErrorState error={attachments.error as Error} />
          ) : !attachments.data?.length ? (
            <Empty title="No attachments" detail="Upload a PDF, image, or text file (max 10 MiB)." />
          ) : (
            <Card className="divide-y divide-slate-100">
              {attachments.data.map((item) => (
                <div key={item.id} className="flex items-center justify-between gap-3 p-4 text-sm">
                  <div>
                    <p className="font-medium text-slate-800">{item.filename}</p>
                    <p className="text-xs text-slate-500">
                      {item.mime_type} · {(item.size_bytes / 1024).toFixed(1)} KB · {formatDate(item.created_at)}
                    </p>
                  </div>
                </div>
              ))}
            </Card>
          )}
        </div>

        {staff && (
          <>
            <h2 className="mt-8 text-lg font-bold">Timeline</h2>
            <div className="mt-3">
              {timeline.isLoading ? (
                <Loading label="Loading timeline…" />
              ) : timeline.error ? (
                <ErrorState error={timeline.error as Error} />
              ) : !timeline.data?.length ? (
                <Empty title="No timeline events" detail="Status changes and audit events will show up here." />
              ) : (
                <Card className="divide-y divide-slate-100">
                  {timeline.data.map((event) => (
                    <div key={event.id} className="p-4 text-sm">
                      <div className="flex justify-between gap-3">
                        <p className="font-medium text-slate-800">{event.event_type}</p>
                        <span className="text-xs text-slate-500">{formatDate(event.created_at)}</span>
                      </div>
                      {event.payload != null && (
                        <pre className="mt-2 overflow-x-auto rounded-lg bg-slate-50 p-2 text-xs text-slate-600">
                          {typeof event.payload === "string"
                            ? event.payload
                            : JSON.stringify(event.payload, null, 2)}
                        </pre>
                      )}
                    </div>
                  ))}
                </Card>
              )}
            </div>
          </>
        )}
      </div>

      <aside className="space-y-5">
        <Card className="p-5">
          <h2 className="font-bold">Ticket details</h2>
          <dl className="mt-4 space-y-3 text-sm">
            <div className="flex justify-between gap-3">
              <dt className="text-slate-500">Created</dt>
              <dd>{formatDate(t.created_at)}</dd>
            </div>
            <div className="flex justify-between gap-3">
              <dt className="text-slate-500">Category</dt>
              <dd className="capitalize">{t.category}</dd>
            </div>
            <div>
              <dt className="text-slate-500">SLA due</dt>
              <dd className="mt-1 font-medium">
                {t.sla_paused_at ? "Paused" : formatDate(t.sla_due_at)}
              </dd>
            </div>
            {t.assignee_id && (
              <div className="flex justify-between gap-3">
                <dt className="text-slate-500">Assignee</dt>
                <dd className="truncate text-xs">{t.assignee_id}</dd>
              </div>
            )}
          </dl>
        </Card>
        {staff && (
          <AgentControls
            ticket={t}
            updating={update.isPending}
            update={update.mutate}
            escalateError={escalate.error as Error | null}
            updateError={update.error as Error | null}
            escalating={escalate.isPending}
            onEscalate={() => escalate.mutate()}
          />
        )}
      </aside>
    </div>
  );
}

function AgentControls({
  ticket,
  update,
  updating,
  escalating,
  onEscalate,
  escalateError,
  updateError,
}: {
  ticket: Ticket;
  update: (data: object) => void;
  updating: boolean;
  escalating: boolean;
  onEscalate: () => void;
  escalateError: Error | null;
  updateError: Error | null;
}) {
  const agents = useQuery({ queryKey: ["agents"], queryFn: ticketsApi.agents });

  return (
    <Card className="p-5">
      <h2 className="font-bold">Agent actions</h2>
      {(updateError || escalateError) && (
        <div className="mt-3">
          <ErrorState error={(updateError ?? escalateError)!} />
        </div>
      )}
      <label className="mt-4 block text-sm font-medium">
        Status
        <select
          className="field mt-1"
          value={ticket.status}
          disabled={updating}
          onChange={(e) => update({ status: e.target.value as Status })}
        >
          {(["open", "pending", "resolved", "closed"] as Status[]).map((x) => (
            <option key={x} value={x}>
              {x}
            </option>
          ))}
        </select>
      </label>
      <label className="mt-4 block text-sm font-medium">
        Priority
        <select
          className="field mt-1"
          value={ticket.priority}
          disabled={updating}
          onChange={(e) => update({ priority: e.target.value as Priority })}
        >
          {(["low", "medium", "high", "urgent"] as Priority[]).map((x) => (
            <option key={x} value={x}>
              {x}
            </option>
          ))}
        </select>
      </label>
      <label className="mt-4 block text-sm font-medium">
        Assignee
        <select
          className="field mt-1"
          value={ticket.assignee_id ?? ""}
          disabled={updating || agents.isLoading}
          onChange={(e) => update({ assignee_id: e.target.value || null })}
        >
          <option value="">Unassigned</option>
          {agents.data?.map((agent) => (
            <option value={agent.id} key={agent.id}>
              {agent.email}
            </option>
          ))}
        </select>
      </label>
      <Button
        className="mt-5 w-full bg-red-600 hover:bg-red-700"
        onClick={onEscalate}
        disabled={escalating}
      >
        {escalating ? "Escalating…" : "Escalate ticket"}
      </Button>
    </Card>
  );
}

function AgentTools() {
  const client = useQueryClient();
  const [reply, setReply] = useState({ title: "", body: "" });
  const [tagName, setTagName] = useState("");

  const canned = useQuery({ queryKey: ["canned-replies"], queryFn: opsApi.cannedReplies });
  const tags = useQuery({ queryKey: ["tags"], queryFn: opsApi.tags });

  const createReply = useMutation({
    mutationFn: () => opsApi.createCannedReply(reply.title.trim(), reply.body.trim()),
    onSuccess: () => {
      setReply({ title: "", body: "" });
      client.invalidateQueries({ queryKey: ["canned-replies"] });
    },
  });
  const deleteReply = useMutation({
    mutationFn: (id: string) => opsApi.deleteCannedReply(id),
    onSuccess: () => client.invalidateQueries({ queryKey: ["canned-replies"] }),
  });
  const createTag = useMutation({
    mutationFn: () => opsApi.createTag(tagName.trim()),
    onSuccess: () => {
      setTagName("");
      client.invalidateQueries({ queryKey: ["tags"] });
    },
  });

  return (
    <div className="mx-auto max-w-4xl space-y-8">
      <div>
        <p className="text-sm font-semibold text-indigo-600">AGENT WORKSPACE</p>
        <h1 className="mt-1 text-3xl font-bold text-slate-900">Agent tools</h1>
        <p className="mt-2 text-sm text-slate-500">
          Manage canned replies and tags used across the ticket queue.
        </p>
      </div>

      <Card className="p-6">
        <h2 className="text-lg font-bold">Canned replies</h2>
        <form
          className="mt-4 space-y-3"
          onSubmit={(e) => {
            e.preventDefault();
            if (reply.title.trim() && reply.body.trim()) createReply.mutate();
          }}
        >
          <input
            className="field"
            placeholder="Title"
            value={reply.title}
            onChange={(e) => setReply({ ...reply, title: e.target.value })}
            required
          />
          <textarea
            className="field"
            rows={4}
            placeholder="Reply body"
            value={reply.body}
            onChange={(e) => setReply({ ...reply, body: e.target.value })}
            required
          />
          <Button disabled={createReply.isPending}>
            {createReply.isPending ? "Saving…" : "Add canned reply"}
          </Button>
        </form>
        {createReply.error && (
          <div className="mt-3">
            <ErrorState error={createReply.error as Error} />
          </div>
        )}
        <div className="mt-5 divide-y divide-slate-100">
          {(canned.data ?? []).map((item) => (
            <div key={item.id} className="flex items-start justify-between gap-3 py-3">
              <div>
                <p className="font-medium text-slate-900">{item.title}</p>
                <p className="mt-1 whitespace-pre-wrap text-sm text-slate-600">{item.body}</p>
              </div>
              <button
                type="button"
                className="text-sm text-red-600 hover:underline"
                onClick={() => deleteReply.mutate(item.id)}
              >
                Delete
              </button>
            </div>
          ))}
          {!canned.data?.length && !canned.isLoading && (
            <Empty title="No canned replies" detail="Create templates for common agent responses." />
          )}
        </div>
      </Card>

      <Card className="p-6">
        <h2 className="text-lg font-bold">Tags</h2>
        <form
          className="mt-4 flex flex-wrap gap-2"
          onSubmit={(e) => {
            e.preventDefault();
            if (tagName.trim()) createTag.mutate();
          }}
        >
          <input
            className="field max-w-xs"
            placeholder="Tag name"
            value={tagName}
            onChange={(e) => setTagName(e.target.value)}
            required
          />
          <Button disabled={createTag.isPending}>
            {createTag.isPending ? "Saving…" : "Add tag"}
          </Button>
        </form>
        {createTag.error && (
          <div className="mt-3">
            <ErrorState error={createTag.error as Error} />
          </div>
        )}
        <div className="mt-4 flex flex-wrap gap-2">
          {(tags.data ?? []).map((tag) => (
            <Badge key={tag.id}>{tag.name}</Badge>
          ))}
          {!tags.data?.length && !tags.isLoading && (
            <p className="text-sm text-slate-500">No tags yet.</p>
          )}
        </div>
      </Card>
    </div>
  );
}

function Landing() {
  const user = useAuthStore((s) => s.user);
  const tokens = useAuthStore((s) => s.tokens);
  return <Navigate to={tokens ? homeFor(user?.role) : "/login"} replace />;
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<AuthPage />} />
      <Route path="/register" element={<AuthPage register />} />
      <Route element={<Protected />}>
        <Route path="/tickets" element={<TicketsList />} />
        <Route path="/tickets/new" element={<NewTicket />} />
        <Route path="/tickets/:id" element={<TicketDetail />} />
      </Route>
      <Route element={<Protected staffOnly />}>
        <Route path="/agent/tickets" element={<TicketsList staff />} />
        <Route path="/agent/tickets/:id" element={<TicketDetail staff />} />
        <Route path="/agent/tools" element={<AgentTools />} />
      </Route>
      <Route path="*" element={<Landing />} />
    </Routes>
  );
}
