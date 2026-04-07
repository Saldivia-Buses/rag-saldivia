"use client";

import { ImagePlus, X } from "lucide-react";
import * as React from "react";

import { Button } from "@/components/ui/button";
import { FileUpload, FileUploadTrigger } from "@/components/diceui/file-upload";

export const title = "Gallery Image Add";

const Example = () => {
  const [files, setFiles] = React.useState<File[]>([]);

  const handleRemove = (index: number) => {
    setFiles(files.filter((_, i) => i !== index));
  };

  return (
    <div className="w-full max-w-md">
      <FileUpload
        value={files}
        onValueChange={setFiles}
        accept="image/*"
        maxFiles={6}
        maxSize={5 * 1024 * 1024}
        multiple
      >
        <div className="grid grid-cols-3 gap-2">
          {files.map((file, index) => (
            <div key={index} className="group relative aspect-square">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={URL.createObjectURL(file)}
                alt={file.name}
                className="h-full w-full rounded-lg object-cover"
              />
              <Button
                variant="secondary"
                size="icon"
                className="absolute top-1 right-1 size-6 opacity-0 transition-opacity group-hover:opacity-100"
                onClick={() => handleRemove(index)}
              >
                <X className="size-3" />
              </Button>
            </div>
          ))}
          {files.length < 6 && (
            <FileUploadTrigger asChild>
              <button className="flex aspect-square cursor-pointer flex-col items-center justify-center gap-1 rounded-lg border-2 border-dashed transition-colors hover:border-primary hover:bg-primary/5">
                <ImagePlus className="size-6 text-muted-foreground" />
                <span className="text-xs text-muted-foreground">Add</span>
              </button>
            </FileUploadTrigger>
          )}
        </div>
      </FileUpload>
      <p className="mt-2 text-center text-xs text-muted-foreground">
        {files.length}/6 images
      </p>
    </div>
  );
};

export default Example;
