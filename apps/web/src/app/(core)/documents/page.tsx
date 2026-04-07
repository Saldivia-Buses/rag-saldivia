"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Upload,
  FileText,
  Clock,
  CheckCircle2,
  XCircle,
  Loader2,
  Trash2,
} from "lucide-react";
import { api } from "@/lib/api/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";

interface IngestJob {
  id: string;
  file_name: string;
  file_size: number;
  collection: string;
  status: "pending" | "processing" | "completed" | "failed";
  error?: string;
  created_at: string;
}

const STATUS_CONFIG = {
  pending: { icon: Clock, label: "Pendiente", variant: "secondary" as const },
  processing: { icon: Loader2, label: "Procesando", variant: "default" as const },
  completed: { icon: CheckCircle2, label: "Completado", variant: "default" as const },
  failed: { icon: XCircle, label: "Error", variant: "destructive" as const },
};

function formatBytes(bytes: number) {
  if (bytes === 0) return "0 B";
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / 1024 ** i).toFixed(i ? 1 : 0)} ${sizes[i]}`;
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString("es-AR", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export default function DocumentsPage() {
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const [collection, setCollection] = useState("");
  const [file, setFile] = useState<File | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ["ingest-jobs"],
    queryFn: () => api.get<{ jobs: IngestJob[] }>("/v1/ingest/jobs?limit=100"),
    refetchInterval: 5000,
  });

  const upload = useMutation({
    mutationFn: async () => {
      if (!file || !collection) throw new Error("Archivo y colección requeridos");
      const formData = new FormData();
      formData.append("file", file);
      formData.append("collection", collection);
      return api.post<IngestJob>("/v1/ingest/upload", formData, { raw: true });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["ingest-jobs"] });
      setOpen(false);
      setFile(null);
      setCollection("");
    },
  });

  const deleteJob = useMutation({
    mutationFn: (jobId: string) => api.delete(`/v1/ingest/jobs/${jobId}`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["ingest-jobs"] }),
  });

  const jobs = data?.jobs ?? [];

  return (
    <div className="flex-1 p-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-semibold">Documentos</h1>
          <p className="text-muted-foreground text-sm mt-1">
            Subí documentos para que el sistema los indexe y puedas consultarlos por chat.
          </p>
        </div>

        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger asChild>
            <Button>
              <Upload className="h-4 w-4 mr-2" />
              Subir documento
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Subir documento</DialogTitle>
            </DialogHeader>
            <form
              className="flex flex-col gap-4 mt-2"
              onSubmit={(e) => {
                e.preventDefault();
                upload.mutate();
              }}
            >
              <div className="flex flex-col gap-2">
                <Label htmlFor="collection">Colección</Label>
                <Input
                  id="collection"
                  placeholder="ej: manuales, normativas, planos"
                  value={collection}
                  onChange={(e) => setCollection(e.target.value)}
                  required
                />
              </div>
              <div className="flex flex-col gap-2">
                <Label htmlFor="file">Archivo</Label>
                <Input
                  id="file"
                  type="file"
                  accept=".pdf,.docx,.doc,.txt,.md,.csv,.xlsx,.pptx,.html,.json,.xml"
                  onChange={(e) => setFile(e.target.files?.[0] ?? null)}
                  required
                />
                {file && (
                  <p className="text-xs text-muted-foreground">
                    {file.name} ({formatBytes(file.size)})
                  </p>
                )}
              </div>
              <Button type="submit" disabled={upload.isPending || !file || !collection}>
                {upload.isPending ? (
                  <>
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    Subiendo...
                  </>
                ) : (
                  "Subir"
                )}
              </Button>
              {upload.isError && (
                <p className="text-sm text-destructive">
                  {(upload.error as Error).message}
                </p>
              )}
            </form>
          </DialogContent>
        </Dialog>
      </div>

      {isLoading ? (
        <div className="flex flex-col gap-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-16 bg-muted animate-pulse rounded-lg" />
          ))}
        </div>
      ) : jobs.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
          <FileText className="h-12 w-12 mb-4" />
          <p className="text-lg font-medium">Sin documentos</p>
          <p className="text-sm">Subí tu primer documento para empezar.</p>
        </div>
      ) : (
        <div className="flex flex-col gap-2">
          {jobs.map((job) => {
            const cfg = STATUS_CONFIG[job.status];
            const Icon = cfg.icon;
            return (
              <div
                key={job.id}
                className="flex items-center gap-4 p-4 rounded-lg border bg-card"
              >
                <FileText className="h-8 w-8 text-muted-foreground shrink-0" />
                <div className="flex-1 min-w-0">
                  <p className="font-medium truncate">{job.file_name}</p>
                  <p className="text-xs text-muted-foreground">
                    {formatBytes(job.file_size)} &middot; {job.collection} &middot;{" "}
                    {formatDate(job.created_at)}
                  </p>
                  {job.error && (
                    <p className="text-xs text-destructive mt-1">{job.error}</p>
                  )}
                </div>
                <Badge variant={cfg.variant} className="shrink-0 gap-1">
                  <Icon
                    className={`h-3 w-3 ${job.status === "processing" ? "animate-spin" : ""}`}
                  />
                  {cfg.label}
                </Badge>
                {(job.status === "completed" || job.status === "failed") && (
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => deleteJob.mutate(job.id)}
                    disabled={deleteJob.isPending}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
