"use client";

import { Upload, X } from "lucide-react";
import * as React from "react";

import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import {
  FileUpload,
  FileUploadItem,
  FileUploadItemDelete,
  FileUploadItemMetadata,
  FileUploadItemPreview,
  FileUploadList,
  FileUploadTrigger,
} from "@/components/diceui/file-upload";

export const title = "File Upload with Label";

const Example = () => {
  const [files, setFiles] = React.useState<File[]>([]);

  return (
    <div className="w-full max-w-md space-y-2">
      <Label>Attachment</Label>
      <FileUpload
        value={files}
        onValueChange={setFiles}
        maxFiles={3}
        maxSize={5 * 1024 * 1024}
        multiple
      >
        <FileUploadTrigger asChild>
          <Button variant="outline" className="w-full">
            <Upload className="mr-2 size-4" />
            Choose Files
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
    </div>
  );
};

export default Example;
