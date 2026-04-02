"use client";

import { useState, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  PromptInput,
  PromptInputTextarea,
  PromptInputActions,
  PromptInputAction,
} from "@/components/ui/prompt-input";
import {
  ChatContainerRoot,
  ChatContainerContent,
  ChatContainerScrollAnchor,
} from "@/components/ui/chat-container";
import { cn } from "@/lib/utils";
import {
  ArrowUpIcon,
  PlusIcon,
  SquareIcon,
  Trash2Icon,
} from "lucide-react";

function formatRelativeDate(date: Date): string {
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  if (diffMin < 1) return "Justo ahora";
  if (diffMin < 60) return `Hace ${diffMin} min`;
  const diffH = Math.floor(diffMin / 60);
  if (diffH < 24) return `Hace ${diffH}h`;
  const diffD = Math.floor(diffH / 24);
  if (diffD < 7) return `Hace ${diffD}d`;
  return date.toLocaleDateString("es-AR", { day: "numeric", month: "short" });
}

// --- Types ---

interface Message {
  id: string;
  role: "user" | "assistant";
  content: string;
}

interface Session {
  id: string;
  title: string;
  createdAt: Date;
  messages: Message[];
}

// --- Chat Sessions Sidebar ---

function ChatSidebar({
  sessions,
  activeSessionId,
  onSelectSession,
  onNewSession,
  onDeleteSession,
}: {
  sessions: Session[];
  activeSessionId: string | null;
  onSelectSession: (id: string) => void;
  onNewSession: () => void;
  onDeleteSession: (id: string) => void;
}) {
  return (
    <div className="w-72 shrink-0 border-r flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b">
        <h2 className="font-semibold text-lg">Chats</h2>
        <Button
          onClick={onNewSession}
          variant="outline"
          size="sm"
          className="gap-1.5"
        >
          <PlusIcon className="size-3.5" />
          Nuevo chat
        </Button>
      </div>

      {/* List */}
      <ScrollArea className="flex-1">
        <div className="flex flex-col">
          {sessions.length === 0 && (
            <p className="px-4 py-12 text-sm text-center text-muted-foreground">
              Tus conversaciones aparecerán acá
            </p>
          )}
          {sessions.map((session) => (
            <div
              key={session.id}
              className={cn(
                "group flex items-start justify-between gap-2 border-b px-4 py-3 cursor-pointer transition-colors",
                activeSessionId === session.id
                  ? "bg-accent/50"
                  : "hover:bg-muted/50"
              )}
              onClick={() => onSelectSession(session.id)}
            >
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium truncate">
                  {session.title}
                </p>
                <p className="text-xs text-muted-foreground mt-0.5">
                  {formatRelativeDate(session.createdAt)}
                </p>
              </div>
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  onDeleteSession(session.id);
                }}
                className="size-6 shrink-0 mt-0.5 flex items-center justify-center rounded opacity-0 group-hover:opacity-100 hover:bg-destructive/10 hover:text-destructive transition-all"
              >
                <Trash2Icon className="size-3.5" />
              </button>
            </div>
          ))}
        </div>
      </ScrollArea>
    </div>
  );
}

// --- Page ---

export default function ChatPage() {
  const [input, setInput] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [sessions, setSessions] = useState<Session[]>([]);
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null);

  const activeSession = sessions.find((s) => s.id === activeSessionId);
  const messages = activeSession?.messages ?? [];

  const createSession = useCallback(
    (firstMessage?: string) => {
      const session: Session = {
        id: crypto.randomUUID(),
        title: firstMessage?.slice(0, 50) || "Nueva conversación",
        createdAt: new Date(),
        messages: [],
      };
      setSessions((prev) => [session, ...prev]);
      setActiveSessionId(session.id);
      return session.id;
    },
    []
  );

  const handleSubmit = useCallback(() => {
    if (!input.trim() || isLoading) return;

    let sessionId = activeSessionId;
    if (!sessionId) {
      sessionId = createSession(input);
    }

    const userMessage: Message = {
      id: crypto.randomUUID(),
      role: "user",
      content: input,
    };

    // Update session title if it's the first message
    setSessions((prev) =>
      prev.map((s) =>
        s.id === sessionId
          ? {
              ...s,
              title: s.messages.length === 0 ? input.slice(0, 50) : s.title,
              messages: [...s.messages, userMessage],
            }
          : s
      )
    );
    setInput("");

    // TODO: connect to RAG backend
    setIsLoading(true);
    setTimeout(() => {
      const assistantMessage: Message = {
        id: crypto.randomUUID(),
        role: "assistant",
        content:
          "Esta es una respuesta de ejemplo. Cuando el backend esté conectado, las respuestas vendrán del servicio RAG con streaming via SSE.",
      };
      setSessions((prev) =>
        prev.map((s) =>
          s.id === sessionId
            ? { ...s, messages: [...s.messages, assistantMessage] }
            : s
        )
      );
      setIsLoading(false);
    }, 1000);
  }, [input, isLoading, activeSessionId, createSession]);

  const handleDeleteSession = useCallback(
    (id: string) => {
      setSessions((prev) => prev.filter((s) => s.id !== id));
      if (activeSessionId === id) {
        setActiveSessionId(null);
      }
    },
    [activeSessionId]
  );

  return (
    <div className="flex flex-1 min-h-0">
      {/* Sessions sidebar */}
      <ChatSidebar
        sessions={sessions}
        activeSessionId={activeSessionId}
        onSelectSession={setActiveSessionId}
        onNewSession={() => createSession()}
        onDeleteSession={handleDeleteSession}
      />

      {/* Chat area */}
      <div className="flex flex-1 flex-col min-h-0 min-w-0">
        {/* Messages */}
        <ChatContainerRoot className="flex-1">
          <ChatContainerContent className="max-w-3xl mx-auto w-full px-4 py-6 gap-6">
            {messages.length === 0 && !isLoading && (
              <div className="flex flex-1 items-center justify-center">
                <div className="text-center">
                  <h2 className="text-lg font-semibold">
                    ¿En qué puedo ayudarte?
                  </h2>
                  <p className="text-sm text-muted-foreground mt-1">
                    Preguntá sobre tus documentos o cualquier tema de tu
                    empresa.
                  </p>
                </div>
              </div>
            )}

            {messages.map((message) => (
              <div
                key={message.id}
                className={cn(
                  "flex",
                  message.role === "user" ? "justify-end" : "justify-start"
                )}
              >
                <div
                  className={cn(
                    "max-w-[80%] text-sm whitespace-pre-wrap",
                    message.role === "user"
                      ? "rounded-2xl bg-muted px-4 py-2.5"
                      : "text-foreground leading-relaxed"
                  )}
                >
                  {message.content}
                </div>
              </div>
            ))}

            {isLoading && (
              <div className="flex justify-start">
                <div className="text-sm text-muted-foreground">
                  <span className="inline-flex gap-1">
                    <span className="size-1.5 rounded-full bg-current animate-bounce [animation-delay:-0.3s]" />
                    <span className="size-1.5 rounded-full bg-current animate-bounce [animation-delay:-0.15s]" />
                    <span className="size-1.5 rounded-full bg-current animate-bounce" />
                  </span>
                </div>
              </div>
            )}

            <ChatContainerScrollAnchor />
          </ChatContainerContent>
        </ChatContainerRoot>

        {/* Input */}
        <div className="border-t bg-background p-4">
          <div className="mx-auto max-w-3xl">
            <PromptInput
              value={input}
              onValueChange={setInput}
              isLoading={isLoading}
              onSubmit={handleSubmit}
            >
              <PromptInputTextarea placeholder="Escribí tu mensaje..." />
              <PromptInputActions>
                <PromptInputAction tooltip={isLoading ? "Detener" : "Enviar"}>
                  <Button
                    variant="default"
                    size="icon"
                    className="size-8 rounded-full"
                    disabled={!isLoading && !input.trim()}
                    onClick={
                      isLoading ? () => setIsLoading(false) : handleSubmit
                    }
                  >
                    {isLoading ? (
                      <SquareIcon className="size-3 fill-current" />
                    ) : (
                      <ArrowUpIcon className="size-4" />
                    )}
                  </Button>
                </PromptInputAction>
              </PromptInputActions>
            </PromptInput>
          </div>
        </div>
      </div>
    </div>
  );
}
