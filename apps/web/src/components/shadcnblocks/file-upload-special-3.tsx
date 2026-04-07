"use client";

import { ImagePlus, X } from "lucide-react";
import * as React from "react";

import { Button } from "@/components/ui/button";
import { FileUpload, FileUploadTrigger } from "@/components/diceui/file-upload";

export const title = "Cover Image Upload";

const Example = () => {
  const [files, setFiles] = React.useState<File[]>([]);

  const coverPreview = files.length > 0 ? URL.createObjectURL(files[0]) : null;

  const handleRemove = () => {
    setFiles([]);
  };

  return (
    <div className="w-full max-w-md">
      <FileUpload
        value={files}
        onValueChange={setFiles}
        accept="image/*"
        maxFiles={1}
        maxSize={5 * 1024 * 1024}
      >
        <FileUploadTrigger asChild>
          <div className="group relative cursor-pointer overflow-hidden rounded-lg border-2 border-dashed">
            {coverPreview ? (
              <div className="relative aspect-[3/1]">
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src={coverPreview}
                  alt="Cover"
                  className="h-full w-full object-cover"
                />
                <div className="absolute inset-0 flex items-center justify-center bg-black/40 opacity-0 transition-opacity group-hover:opacity-100">
                  <span className="text-sm font-medium text-white">
                    Click to change
                  </span>
                </div>
              </div>
            ) : (
              <div className="flex aspect-[3/1] flex-col items-center justify-center gap-2 bg-muted/50 transition-colors group-hover:bg-muted">
                <ImagePlus className="size-8 text-muted-foreground" />
                <span className="text-sm text-muted-foreground">
                  Add cover image
                </span>
              </div>
            )}
          </div>
        </FileUploadTrigger>
      </FileUpload>
      {coverPreview && (
        <Button
          variant="ghost"
          size="sm"
          className="mt-2 text-destructive"
          onClick={handleRemove}
        >
          <X className="mr-1 size-4" />
          Remove cover
        </Button>
      )}
    </div>
  );
};

export default Example;
