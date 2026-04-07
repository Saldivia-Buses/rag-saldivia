"use client";

import { Upload, X, AlertCircle } from "lucide-react";
import * as React from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Alert, AlertDescription } from "@/components/ui/alert";
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

export const title = "Custom Validation";

const Example = () => {
  const [files, setFiles] = React.useState<File[]>([]);
  const [error, setError] = React.useState<string | null>(null);

  const onFileValidate = React.useCallback((file: File) => {
    // Custom validation: filename must not contain spaces
    if (file.name.includes(" ")) {
      return "Filename cannot contain spaces";
    }
    // Custom validation: file must be less than 2MB
    if (file.size > 2 * 1024 * 1024) {
      return "File must be less than 2MB";
    }
    return null;
  }, []);

  const onFileReject = React.useCallback((file: File, message: string) => {
    setError(`${file.name}: ${message}`);
    toast.error(message, {
      description: `"${file.name}" failed validation`,
    });
  }, []);

  const onValueChange = React.useCallback((newFiles: File[]) => {
    setFiles(newFiles);
    setError(null);
  }, []);

  return (
    <FileUpload
      maxFiles={5}
      maxSize={5 * 1024 * 1024}
      className="w-full max-w-md"
      value={files}
      onValueChange={onValueChange}
      onFileValidate={onFileValidate}
      onFileReject={onFileReject}
      multiple
    >
      <FileUploadDropzone>
        <div className="flex flex-col items-center gap-1 text-center">
          <div className="flex items-center justify-center rounded-full border p-2.5">
            <Upload className="size-6 text-muted-foreground" />
          </div>
          <p className="text-sm font-medium">Custom validation</p>
          <p className="text-xs text-muted-foreground">
            No spaces in filename, max 2MB
          </p>
        </div>
        <FileUploadTrigger asChild>
          <Button variant="outline" size="sm" className="mt-2">
            Browse files
          </Button>
        </FileUploadTrigger>
      </FileUploadDropzone>
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
      <FileUploadList>
        {files.map((file, index) => (
          <FileUploadItem key={index} value={file}>
            <FileUploadItemPreview />
            <FileUploadItemMetadata />
            <FileUploadItemDelete asChild>
              <Button variant="ghost" size="icon" className="size-7">
                <X className="size-4" />
              </Button>
            </FileUploadItemDelete>
          </FileUploadItem>
        ))}
      </FileUploadList>
    </FileUpload>
  );
};

export default Example;
