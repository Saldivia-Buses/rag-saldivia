"use client";

import { useState, useCallback, useRef, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  PromptInput,
  PromptInputTextarea,
  PromptInputActions,
} from "@/components/ui/prompt-input";
import {
  ChatContainerRoot,
  ChatContainerContent,
  ChatContainerScrollAnchor,
} from "@/components/ui/chat-container";
import { cn } from "@/lib/utils";
import { api, ApiError } from "@/lib/api/client";
import {
  Reasoning,
  ReasoningTrigger,
  ReasoningContent,
} from "@/components/ai-elements/reasoning";
import { Streamdown } from "streamdown";
import { code } from "@streamdown/code";
import { createMathPlugin } from "@streamdown/math";
import "katex/dist/katex.min.css";

const mathPlugin = createMathPlugin({ singleDollarTextMath: true });
import { MODELS, DEFAULT_MODEL, type LLMModel } from "@/lib/models";
import {
  ModelSelector,
  ModelSelectorTrigger,
  ModelSelectorContent,
  ModelSelectorInput,
  ModelSelectorList,
  ModelSelectorEmpty,
  ModelSelectorGroup,
  ModelSelectorItem,
  ModelSelectorLogo,
  ModelSelectorName,
} from "@/components/ai-elements/model-selector";
import {
  ArrowUpIcon,
  ChevronDownIcon,
  PlusIcon,
  SquareIcon,
  Trash2Icon,
} from "lucide-react";

function formatRelativeDate(dateStr: string): string {
  const date = new Date(dateStr);
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

interface ApiSession {
  id: string;
  title: string;
  collection: string;
  created_at: string;
  updated_at: string;
}

interface ApiMessage {
  id: string;
  role: "user" | "assistant" | "system";
  content: string;
  thinking?: string | null;
  sources?: Array<{ document_name: string; content: string; score: number }>;
  created_at: string;
}

// --- Chat Sessions Sidebar ---

function ChatSidebar({
  sessions,
  isLoading: sessionsLoading,
  activeSessionId,
  onSelectSession,
  onNewSession,
  onDeleteSession,
}: {
  sessions: ApiSession[];
  isLoading: boolean;
  activeSessionId: string | null;
  onSelectSession: (id: string) => void;
  onNewSession: () => void;
  onDeleteSession: (id: string) => void;
}) {
  return (
    <div className="hidden md:flex w-72 flex-col min-h-0 absolute left-0 top-0 bottom-0 z-10 bg-background">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3">
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
          {sessionsLoading && (
            <div className="px-4 py-12 text-sm text-center text-muted-foreground">
              Cargando...
            </div>
          )}
          {!sessionsLoading && sessions.length === 0 && (
            <p className="px-4 py-12 text-sm text-center text-muted-foreground">
              Tus conversaciones apareceran aca
            </p>
          )}
          {sessions.map((session) => (
            <div
              key={session.id}
              className={cn(
                "group flex items-start justify-between gap-2 px-3 py-2.5 cursor-pointer transition-colors rounded-md mx-2",
                activeSessionId === session.id
                  ? "bg-accent/60"
                  : "hover:bg-muted",
              )}
              onClick={() => onSelectSession(session.id)}
            >
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium truncate">{session.title}</p>
                <p className="text-[13px] text-muted-foreground mt-0.5">
                  {formatRelativeDate(session.created_at)}
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
  const [isStreaming, setIsStreaming] = useState(false);
  const [streamingContent, setStreamingContent] = useState("");
  const [thinkingContent, setThinkingContent] = useState("");
  const router = useRouter();
  const searchParams = useSearchParams();
  const activeSessionId = searchParams.get("s");
  const [selectedModel, setSelectedModel] = useState<LLMModel>(DEFAULT_MODEL);
  const [modelSelectorOpen, setModelSelectorOpen] = useState(false);
  const setActiveSessionId = useCallback(
    (id: string | null) => {
      if (id) {
        router.push(`/chat?s=${id}`);
      } else {
        router.push("/chat");
      }
    },
    [router],
  );

  const abortRef = useRef<AbortController | null>(null);
  const streamRef = useRef(""); // tracks streaming content synchronously
  const thinkingRef = useRef(""); // tracks thinking content synchronously
  const queryClient = useQueryClient();

  // Fetch sessions
  const { data: sessions = [], isLoading: sessionsLoading } = useQuery({
    queryKey: ["chat", "sessions"],
    queryFn: () => api.get<ApiSession[]>("/v1/chat/sessions"),
  });

  // Fetch messages for active session
  const { data: messages = [] } = useQuery({
    queryKey: ["chat", "messages", activeSessionId],
    queryFn: () =>
      api.get<ApiMessage[]>(
        `/v1/chat/sessions/${activeSessionId}/messages`,
      ),
    enabled: !!activeSessionId,
  });

  // Create session mutation
  const createSessionMutation = useMutation({
    mutationFn: (title: string) =>
      api.post<ApiSession>("/v1/chat/sessions", { title }),
    onSuccess: (session) => {
      queryClient.invalidateQueries({ queryKey: ["chat", "sessions"] });
      setActiveSessionId(session.id);
    },
  });

  // Delete session mutation
  const deleteSessionMutation = useMutation({
    mutationFn: (id: string) => api.delete(`/v1/chat/sessions/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["chat", "sessions"] });
    },
  });

  // Send message mutation
  const sendMessageMutation = useMutation({
    mutationFn: ({
      sessionId,
      content,
    }: {
      sessionId: string;
      content: string;
    }) =>
      api.post<ApiMessage>(`/v1/chat/sessions/${sessionId}/messages`, {
        role: "user",
        content,
      }),
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({
        queryKey: ["chat", "messages", vars.sessionId],
      });
    },
  });

  const handleNewSession = useCallback(() => {
    createSessionMutation.mutate("Nueva conversacion");
  }, [createSessionMutation]);

  const handleDeleteSession = useCallback(
    (id: string) => {
      deleteSessionMutation.mutate(id);
      if (activeSessionId === id) {
        setActiveSessionId(null);
      }
    },
    [activeSessionId, deleteSessionMutation, setActiveSessionId],
  );

  const handleSubmit = useCallback(async () => {
    if (!input.trim() || isStreaming) return;

    const content = input;
    setInput("");

    // Create session if none is active
    let sessionId = activeSessionId;
    if (!sessionId) {
      try {
        const session = await api.post<ApiSession>("/v1/chat/sessions", {
          title: content.slice(0, 50),
        });
        sessionId = session.id;
        setActiveSessionId(session.id);
        queryClient.invalidateQueries({ queryKey: ["chat", "sessions"] });
      } catch {
        return;
      }
    }

    // Send user message
    try {
      await sendMessageMutation.mutateAsync({ sessionId, content });
    } catch {
      return;
    }

    // Stream RAG response
    setIsStreaming(true);
    setStreamingContent("");
    setThinkingContent("");
    streamRef.current = "";
    thinkingRef.current = "";

    try {
      // Build full conversation history for context
      const history = messages.map((m) => ({
        role: m.role as "user" | "assistant",
        content: m.content,
      }));
      history.push({ role: "user", content });

      for await (const chunk of api.streamWithThinking("/v1/rag/generate", {
        messages: history,
        model: selectedModel.id,
        stream: true,
        use_knowledge_base: true,
        ...(selectedModel.supportsReasoning
          ? { reasoning: { effort: "medium" } }
          : {}),
      })) {
        if (chunk.type === "thinking") {
          thinkingRef.current += chunk.text;
          setThinkingContent(thinkingRef.current);
        } else {
          streamRef.current += chunk.text;
          setStreamingContent(streamRef.current);
        }
      }

      // Store assistant message in backend using the ref (always current)
      const finalContent = streamRef.current;
      const finalThinking = thinkingRef.current || undefined;
      if (finalContent) {
        const saved = await api.post<ApiMessage>(`/v1/chat/sessions/${sessionId}/messages`, {
          role: "assistant",
          content: finalContent,
          thinking: finalThinking,
        });

        // Refetch messages — streaming content stays visible until data arrives
        await queryClient.invalidateQueries({
          queryKey: ["chat", "messages", sessionId],
        });
      }
    } catch (err) {
      // Save error as assistant message so it persists in the conversation
      const errorMsg = err instanceof ApiError
        ? `Error: ${err.message}`
        : "Error: no se pudo obtener respuesta";
      try {
        await api.post(`/v1/chat/sessions/${sessionId}/messages`, {
          role: "assistant",
          content: errorMsg,
        });
        await queryClient.invalidateQueries({
          queryKey: ["chat", "messages", sessionId],
        });
      } catch {
        // If saving fails, at least show it in streaming
        streamRef.current = errorMsg;
        setStreamingContent(streamRef.current);
      }
    } finally {
      setIsStreaming(false);
      setStreamingContent("");
      setThinkingContent("");
      streamRef.current = "";
      thinkingRef.current = "";
    }
  }, [input, isStreaming, activeSessionId, queryClient, sendMessageMutation]);

  const handleStop = useCallback(() => {
    abortRef.current?.abort();
    setIsStreaming(false);
  }, []);

  // Combine persisted messages with streaming content
  const displayMessages = [
    ...messages,
    ...((streamingContent || thinkingContent)
      ? [
          {
            id: "streaming",
            role: "assistant" as const,
            content: streamingContent,
            thinking: thinkingContent,
            sources: [] as Array<{ document_name: string; content: string; score: number }>,
            created_at: new Date().toISOString(),
          },
        ]
      : []),
  ];

  return (
    <div className="flex flex-1 min-h-0 relative">
      {/* Sessions sidebar */}
      <ChatSidebar
        sessions={sessions}
        isLoading={sessionsLoading}
        activeSessionId={activeSessionId}
        onSelectSession={setActiveSessionId}
        onNewSession={handleNewSession}
        onDeleteSession={handleDeleteSession}
      />

      {/* Chat area */}
      <div className="flex flex-1 flex-col min-h-0 min-w-0 relative">
        {/* Messages */}
        <ChatContainerRoot className="flex-1 [mask-image:linear-gradient(to_bottom,transparent,black_1.5rem,black_calc(100%-2rem),transparent)]">
          <ChatContainerContent className="max-w-4xl mx-auto w-full px-6 py-6 gap-6">
            {displayMessages.length === 0 && !isStreaming && (
              <div className="flex flex-1 items-center justify-center">
                <div className="text-center">
                  <h2 className="text-lg font-semibold">
                    En que puedo ayudarte?
                  </h2>
                  <p className="text-sm text-muted-foreground mt-1">
                    Pregunta sobre tus documentos o cualquier tema de tu
                    empresa.
                  </p>
                </div>
              </div>
            )}

            {displayMessages.map((message) => (
              <div
                key={message.id}
                className={cn(
                  "flex",
                  message.role === "user" ? "justify-end" : "justify-start",
                )}
              >
                <div
                  className={cn(
                    "max-w-[80%] text-sm",
                    message.role === "user"
                      ? "rounded-2xl bg-muted px-4 py-2.5 whitespace-pre-wrap"
                      : "text-foreground leading-relaxed prose prose-sm dark:prose-invert prose-p:my-1 prose-headings:my-2 prose-pre:my-2 prose-ul:my-1 prose-ol:my-1 max-w-none",
                  )}
                >
                  {/* Reasoning block (above response content) */}
                  {(message.thinking || null) && (
                    <Reasoning
                      isStreaming={message.id === "streaming" && isStreaming && !message.content}
                    >
                      <ReasoningTrigger
                        getThinkingMessage={(streaming, duration) => {
                          if (streaming || duration === 0)
                            return <span className="text-muted-foreground">Pensando...</span>;
                          if (duration === undefined)
                            return <span>Penso unos segundos</span>;
                          return <span>Penso {duration} segundos</span>;
                        }}
                      />
                      <ReasoningContent>
                        {message.thinking || null || ""}
                      </ReasoningContent>
                    </Reasoning>
                  )}
                  {message.role === "user" ? (
                    message.content
                  ) : (
                    message.content && (
                      <Streamdown plugins={{ code, math: mathPlugin }}>
                        {message.content}
                      </Streamdown>
                    )
                  )}
                  {message.sources && message.sources.length > 0 && (
                    <div className="mt-3 flex flex-wrap gap-1.5">
                      {message.sources.map((src, i) => (
                        <span
                          key={i}
                          className="inline-flex items-center rounded-md bg-muted px-2 py-0.5 text-xs text-muted-foreground"
                        >
                          {src.document_name}
                        </span>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            ))}

            {isStreaming && !streamingContent && !thinkingContent && (
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

        {/* Input — no border, seamless with chat */}
        <div className="bg-background p-4 pb-6 relative z-10">
          <div className="max-w-4xl mx-auto w-full flex flex-col gap-2">
            <PromptInput
              value={input}
              onValueChange={setInput}
              isLoading={isStreaming}
              onSubmit={handleSubmit}
            >
              <PromptInputTextarea placeholder="Escribi tu mensaje..." />
              <PromptInputActions className="justify-between">
                {/* Model selector */}
                <ModelSelector
                  open={modelSelectorOpen}
                  onOpenChange={setModelSelectorOpen}
                >
                  <ModelSelectorTrigger
                    render={
                      <Button
                        variant="ghost"
                        size="sm"
                        className="gap-2 text-sm text-muted-foreground hover:text-foreground h-9 px-3"
                      />
                    }
                  >
                    <ModelSelectorLogo
                      provider={selectedModel.providerLogo}
                      className="size-5"
                    />
                    <span className="hidden sm:inline">
                      {selectedModel.name}
                    </span>
                    <ChevronDownIcon className="size-3.5 opacity-50" />
                  </ModelSelectorTrigger>
                  <ModelSelectorContent
                    title="Elegir modelo"
                    className="sm:max-w-md"
                  >
                    <ModelSelectorInput
                      placeholder="Buscar modelo..."
                      autoFocus
                      onKeyDown={(e) => e.stopPropagation()}
                    />
                    <ModelSelectorList className="max-h-[400px]">
                      <ModelSelectorEmpty>
                        No se encontraron modelos
                      </ModelSelectorEmpty>
                      <ModelSelectorGroup heading="Modelos disponibles">
                        {MODELS.map((model) => (
                          <ModelSelectorItem
                            key={model.id}
                            value={model.id}
                            onSelect={() => {
                              setSelectedModel(model);
                              setModelSelectorOpen(false);
                            }}
                            className="flex items-center gap-3.5 py-3 px-3"
                          >
                            <ModelSelectorLogo
                              provider={model.providerLogo}
                              className="size-6 shrink-0"
                            />
                            <div className="flex-1 min-w-0">
                              <ModelSelectorName className="text-sm font-medium">
                                {model.name}
                              </ModelSelectorName>
                              <p className="text-xs text-muted-foreground mt-0.5">
                                {model.description}
                              </p>
                            </div>
                            <span className="text-xs text-muted-foreground shrink-0">
                              {model.contextWindow}
                            </span>
                          </ModelSelectorItem>
                        ))}
                      </ModelSelectorGroup>
                    </ModelSelectorList>
                  </ModelSelectorContent>
                </ModelSelector>

                {/* Send / Stop */}
                <Button
                  variant="default"
                  size="icon"
                  className="size-9 rounded-full"
                  disabled={!isStreaming && !input.trim()}
                  onClick={isStreaming ? handleStop : handleSubmit}
                >
                  {isStreaming ? (
                    <SquareIcon className="size-3 fill-current" />
                  ) : (
                    <ArrowUpIcon className="size-4" />
                  )}
                </Button>
              </PromptInputActions>
            </PromptInput>
          </div>
        </div>
      </div>
    </div>
  );
}
