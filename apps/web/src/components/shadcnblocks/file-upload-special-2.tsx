"use client";

import { Paperclip, Send, X } from "lucide-react";
import * as React from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  FileUpload,
  FileUploadItem,
  FileUploadItemDelete,
  FileUploadItemMetadata,
  FileUploadItemPreview,
  FileUploadList,
  FileUploadTrigger,
} from "@/components/diceui/file-upload";

export const title = "Chat Input with Attachments";

const Example = () => {
  const [files, setFiles] = React.useState<File[]>([]);
  const [message, setMessage] = React.useState("");

  const handleSend = () => {
    if (message.trim() || files.length > 0) {
      console.log("Sending:", { message, files });
      setMessage("");
      setFiles([]);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div className="w-full max-w-md">
      <FileUpload
        value={files}
        onValueChange={setFiles}
        maxFiles={5}
        maxSize={5 * 1024 * 1024}
        multiple
      >
        <FileUploadList className="mb-2">
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
        <div className="flex items-center gap-2 rounded-lg border bg-background p-2">
          <FileUploadTrigger asChild>
            <Button variant="ghost" size="icon" className="size-8 shrink-0">
              <Paperclip className="size-4" />
            </Button>
          </FileUploadTrigger>
          <Input
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Type a message..."
            className="flex-1 border-0 bg-transparent focus-visible:ring-0"
          />
          <Button
            size="icon"
            className="size-8 shrink-0"
            onClick={handleSend}
            disabled={!message.trim() && files.length === 0}
          >
            <Send className="size-4" />
          </Button>
        </div>
      </FileUpload>
    </div>
  );
};

export default Example;
