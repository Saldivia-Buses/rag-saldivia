"use client";

import { Upload, X, CheckCircle, AlertCircle } from "lucide-react";
import * as React from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import {
  FileUpload,
  FileUploadDropzone,
  FileUploadTrigger,
  useFileUpload,
} from "@/components/diceui/file-upload";

export const title = "Direct Upload with Status";

interface UploadStatusProps {
  file: File;
}

const UploadStatus = ({ file }: UploadStatusProps) => {
  const fileState = useFileUpload((state) => state.files.get(file));
  const progress = fileState?.progress ?? 0;
  const status = fileState?.status ?? "idle";

  return (
    <div className="flex items-center gap-3 rounded-lg border p-3">
      <div className="min-w-0 flex-1">
        <div className="mb-1 flex items-center justify-between">
          <span className="truncate text-sm font-medium">{file.name}</span>
          {status === "success" && (
            <CheckCircle className="size-4 shrink-0 text-green-500" />
          )}
          {status === "error" && (
            <AlertCircle className="size-4 shrink-0 text-destructive" />
          )}
        </div>
        {(status === "uploading" || status === "idle") && (
          <Progress value={progress} className="h-1.5" />
        )}
        <div className="mt-1 flex items-center justify-between">
          <span className="text-xs text-muted-foreground">
            {(file.size / 1024).toFixed(1)} KB
          </span>
          <span className="text-xs text-muted-foreground">
            {status === "uploading" && `${progress}%`}
            {status === "success" && "Complete"}
            {status === "error" && fileState?.error}
          </span>
        </div>
      </div>
    </div>
  );
};

const Example = () => {
  const [files, setFiles] = React.useState<File[]>([]);

  const onUpload = React.useCallback(
    async (
      files: File[],
      {
        onProgress,
        onSuccess,
        onError,
      }: {
        onProgress: (file: File, progress: number) => void;
        onSuccess: (file: File) => void;
        onError: (file: File, error: Error) => void;
      },
    ) => {
      for (const file of files) {
        try {
          // Simulate upload with random success/failure
          for (let progress = 0; progress <= 100; progress += 10) {
            await new Promise((resolve) => setTimeout(resolve, 100));
            onProgress(file, progress);
          }

          // Simulate 80% success rate
          if (Math.random() > 0.2) {
            onSuccess(file);
            toast.success(`${file.name} uploaded successfully`);
          } else {
            throw new Error("Upload failed");
          }
        } catch (error) {
          onError(file, error as Error);
          toast.error(`Failed to upload ${file.name}`);
        }
      }
    },
    [],
  );

  const handleClear = () => {
    setFiles([]);
  };

  return (
    <FileUpload
      maxFiles={5}
      maxSize={5 * 1024 * 1024}
      className="w-full max-w-md"
      value={files}
      onValueChange={setFiles}
      onUpload={onUpload}
      multiple
    >
      <FileUploadDropzone>
        <div className="flex flex-col items-center gap-1 text-center">
          <div className="flex items-center justify-center rounded-full border p-2.5">
            <Upload className="size-6 text-muted-foreground" />
          </div>
          <p className="text-sm font-medium">Direct upload</p>
          <p className="text-xs text-muted-foreground">
            Files upload automatically
          </p>
        </div>
        <FileUploadTrigger asChild>
          <Button variant="outline" size="sm" className="mt-2">
            Browse files
          </Button>
        </FileUploadTrigger>
      </FileUploadDropzone>
      {files.length > 0 && (
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium">Uploads</span>
            <Button variant="ghost" size="sm" onClick={handleClear}>
              <X className="mr-1 size-3" />
              Clear
            </Button>
          </div>
          {files.map((file, index) => (
            <UploadStatus key={index} file={file} />
          ))}
        </div>
      )}
    </FileUpload>
  );
};

export default Example;
