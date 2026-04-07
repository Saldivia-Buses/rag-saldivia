"use client";

import { ImageIcon, FileText, X } from "lucide-react";
import * as React from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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

export const title = "Multiple File Upload Fields";

const Example = () => {
  const [images, setImages] = React.useState<File[]>([]);
  const [documents, setDocuments] = React.useState<File[]>([]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    console.log("Images:", images);
    console.log("Documents:", documents);
    alert(
      `Submitted ${images.length} image(s) and ${documents.length} document(s)`,
    );
  };

  return (
    <form onSubmit={handleSubmit} className="w-full max-w-md space-y-6">
      <div className="space-y-2">
        <Label htmlFor="title">Project Title</Label>
        <Input
          id="title"
          name="title"
          placeholder="Enter project title"
          required
        />
      </div>

      <div className="space-y-2">
        <Label>Project Images</Label>
        <FileUpload
          value={images}
          onValueChange={setImages}
          accept="image/*"
          maxFiles={4}
          maxSize={5 * 1024 * 1024}
          multiple
        >
          <FileUploadDropzone className="py-4">
            <div className="flex items-center gap-2">
              <ImageIcon className="size-4 text-muted-foreground" />
              <span className="text-sm text-muted-foreground">
                Drop images or
              </span>
              <FileUploadTrigger asChild>
                <Button variant="link" size="sm" className="h-auto p-0">
                  browse
                </Button>
              </FileUploadTrigger>
            </div>
          </FileUploadDropzone>
          <FileUploadList>
            {images.map((file, index) => (
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
        <p className="text-xs text-muted-foreground">
          Up to 4 images, 5MB each
        </p>
      </div>

      <div className="space-y-2">
        <Label>Supporting Documents</Label>
        <FileUpload
          value={documents}
          onValueChange={setDocuments}
          accept=".pdf,.doc,.docx"
          maxFiles={3}
          maxSize={10 * 1024 * 1024}
          multiple
        >
          <FileUploadDropzone className="py-4">
            <div className="flex items-center gap-2">
              <FileText className="size-4 text-muted-foreground" />
              <span className="text-sm text-muted-foreground">
                Drop documents or
              </span>
              <FileUploadTrigger asChild>
                <Button variant="link" size="sm" className="h-auto p-0">
                  browse
                </Button>
              </FileUploadTrigger>
            </div>
          </FileUploadDropzone>
          <FileUploadList>
            {documents.map((file, index) => (
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
        <p className="text-xs text-muted-foreground">
          Up to 3 documents, 10MB each
        </p>
      </div>

      <Button type="submit" className="w-full">
        Create Project
      </Button>
    </form>
  );
};

export default Example;
