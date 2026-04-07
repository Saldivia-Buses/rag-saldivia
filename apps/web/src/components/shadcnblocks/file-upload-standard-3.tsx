"use client";

import { Upload, X } from "lucide-react";
import * as React from "react";

import { Button } from "@/components/ui/button";
import {
  FileUpload,
  FileUploadItem,
  FileUploadItemDelete,
  FileUploadItemMetadata,
  FileUploadItemPreview,
  FileUploadList,
  FileUploadTrigger,
} from "@/components/diceui/file-upload";

export const title = "Single File Upload";

const Example = () => {
  const [files, setFiles] = React.useState<File[]>([]);

  return (
    <FileUpload
      value={files}
      onValueChange={setFiles}
      maxFiles={1}
      maxSize={10 * 1024 * 1024}
      className="w-full max-w-md"
    >
      <FileUploadTrigger asChild>
        <Button variant="secondary" className="w-full">
          <Upload className="mr-2 size-4" />
          Select File
        </Button>
      </FileUploadTrigger>
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
