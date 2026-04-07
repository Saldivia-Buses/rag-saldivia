"use client";

import { Upload, X, Loader2, CheckCircle2 } from "lucide-react";
import * as React from "react";

import { Button } from "@/components/ui/button";
import {
  FileUpload,
  FileUploadDropzone,
  FileUploadItem,
  FileUploadItemDelete,
  FileUploadItemMetadata,
  FileUploadItemPreview,
  FileUploadList,
  FileUploadTrigger,
  useFileUpload,
} from "@/components/diceui/file-upload";

export const title = "Upload Status Indicator";

interface FileItemWithStatusProps {
  file: File;
}

const FileItemWithStatus = ({ file }: FileItemWithStatusProps) => {
  const fileState = useFileUpload((state) => state.files.get(file));
  const status = fileState?.status ?? "idle";

  return (
    <FileUploadItem value={file}>
      <FileUploadItemPreview />
      <FileUploadItemMetadata />
      <div className="flex items-center gap-2">
        {status === "uploading" && (
          <Loader2 className="size-4 animate-spin text-primary" />
        )}
        {status === "success" && (
          <CheckCircle2 className="size-4 text-green-500" />
        )}
        <FileUploadItemDelete asChild>
          <Button variant="ghost" size="icon" className="size-7">
            <X className="size-4" />
          </Button>
        </FileUploadItemDelete>
      </div>
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
        for (let progress = 0; progress <= 100; progress += 10) {
          await new Promise((resolve) => setTimeout(resolve, 100));
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
            Shows spinner while uploading
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
          <FileItemWithStatus key={index} file={file} />
        ))}
      </FileUploadList>
    </FileUpload>
  );
};

export default Example;
