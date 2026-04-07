"use client";

import { Upload, X } from "lucide-react";
import * as React from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
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

export const title = "Contact Form with Attachments";

const Example = () => {
  const [files, setFiles] = React.useState<File[]>([]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const formData = new FormData(e.target as HTMLFormElement);
    console.log("Form submitted with", files.length, "attachments");
    alert(`Form submitted with ${files.length} attachment(s)`);
  };

  return (
    <form onSubmit={handleSubmit} className="w-full max-w-md space-y-4">
      <div className="space-y-2">
        <Label htmlFor="name">Name</Label>
        <Input id="name" name="name" placeholder="Your name" required />
      </div>

      <div className="space-y-2">
        <Label htmlFor="email">Email</Label>
        <Input
          id="email"
          name="email"
          type="email"
          placeholder="you@example.com"
          required
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="message">Message</Label>
        <Textarea
          id="message"
          name="message"
          placeholder="Your message..."
          required
        />
      </div>

      <div className="space-y-2">
        <Label>Attachments (optional)</Label>
        <FileUpload
          value={files}
          onValueChange={setFiles}
          maxFiles={3}
          maxSize={5 * 1024 * 1024}
          multiple
        >
          <FileUploadDropzone className="py-4">
            <div className="flex items-center gap-2">
              <Upload className="size-4 text-muted-foreground" />
              <span className="text-sm text-muted-foreground">
                Drop files or
              </span>
              <FileUploadTrigger asChild>
                <Button variant="link" size="sm" className="h-auto p-0">
                  browse
                </Button>
              </FileUploadTrigger>
            </div>
          </FileUploadDropzone>
          <FileUploadList>
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
      </div>

      <Button type="submit" className="w-full">
        Send Message
      </Button>
    </form>
  );
};

export default Example;
