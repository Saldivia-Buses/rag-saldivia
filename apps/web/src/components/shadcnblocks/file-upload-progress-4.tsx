"use client";

import { Upload, X, CheckCircle2 } from "lucide-react";
import * as React from "react";

import { Button } from "@/components/ui/button";
import {
  FileUpload,
  FileUploadDropzone,
  FileUploadItem,
  FileUploadItemDelete,
  FileUploadItemPreview,
  FileUploadList,
  FileUploadTrigger,
  useFileUpload,
} from "@/components/diceui/file-upload";

export const title = "Progress with Percentage";

interface FileItemWithProgressProps {
  file: File;
}

const FileItemWithProgress = ({ file }: FileItemWithProgressProps) => {
  const fileState = useFileUpload((state) => state.files.get(file));
  const progress = fileState?.progress ?? 0;
  const status = fileState?.status ?? "idle";

  return (
    <FileUploadItem value={file}>
      <FileUploadItemPreview />
      <div className="flex flex-1 flex-col gap-1">
        <div className="flex items-center justify-between">
          <span className="truncate text-sm font-medium">{file.name}</span>
          <span className="text-xs text-muted-foreground">
            {status === "success" ? (
              <CheckCircle2 className="size-4 text-green-500" />
            ) : (
              `${progress}%`
            )}
          </span>
        </div>
        <div className="h-1.5 w-full overflow-hidden rounded-full bg-primary/20">
          <div
            className="h-full bg-primary transition-all duration-300"
            style={{ width: `${progress}%` }}
          />
        </div>
        <span className="text-xs text-muted-foreground">
          {(file.size / 1024).toFixed(1)} KB
        </span>
      </div>
      <FileUploadItemDelete asChild>
        <Button variant="ghost" size="icon" className="size-7">
          <X className="size-4" />
        </Button>
      </FileUploadItemDelete>
    </FileUploadItem>
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
      }: {
        onProgress: (file: File, progress: number) => void;
        onSuccess: (file: File) => void;
        onError: (file: File, error: Error) => void;
      },
    ) => {
      for (const file of files) {
        for (let progress = 0; progress <= 100; progress += 5) {
          await new Promise((resolve) => setTimeout(resolve, 60));
          onProgress(file, progress);
        }
        onSuccess(file);
      }
    },
    [],
  );

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
          <p className="text-sm font-medium">Drag & drop files here</p>
          <p className="text-xs text-muted-foreground">
            Shows percentage during upload
          </p>
        </div>
        <FileUploadTrigger asChild>
          <Button variant="outline" size="sm" className="mt-2">
            Browse files
          </Button>
        </FileUploadTrigger>
      </FileUploadDropzone>
      <FileUploadList>
        {files.map((file, index) => (
          <FileItemWithProgress key={index} file={file} />
        ))}
      </FileUploadList>
    </FileUpload>
  );
};

export default Example;
