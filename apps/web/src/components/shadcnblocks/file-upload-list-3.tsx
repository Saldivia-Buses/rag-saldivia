"use client";

import { Upload, X } from "lucide-react";
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
} from "@/components/diceui/file-upload";

export const title = "Compact File List";

const Example = () => {
  const [files, setFiles] = React.useState<File[]>([]);

  return (
    <FileUpload
      maxFiles={5}
      maxSize={5 * 1024 * 1024}
      className="w-full max-w-md"
      value={files}
      onValueChange={setFiles}
      multiple
    >
      <FileUploadDropzone className="py-3">
        <div className="flex items-center gap-2">
          <Upload className="size-4 text-muted-foreground" />
          <span className="text-sm">Drop files or</span>
          <FileUploadTrigger asChild>
            <Button variant="link" size="sm" className="h-auto p-0">
              browse
            </Button>
          </FileUploadTrigger>
        </div>
      </FileUploadDropzone>
      <FileUploadList className="gap-1">
        {files.map((file, index) => (
          <FileUploadItem key={index} value={file} className="p-2">
            <FileUploadItemPreview className="size-8" />
            <FileUploadItemMetadata size="sm" />
            <FileUploadItemDelete asChild>
              <Button variant="ghost" size="icon" className="size-6">
                <X className="size-3" />
              </Button>
            </FileUploadItemDelete>
          </FileUploadItem>
        ))}
      </FileUploadList>
    </FileUpload>
  );
};

export default Example;
