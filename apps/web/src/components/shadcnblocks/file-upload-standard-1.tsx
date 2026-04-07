"use client";

import { Upload } from "lucide-react";
import * as React from "react";

import { Button } from "@/components/ui/button";
import { FileUpload, FileUploadTrigger } from "@/components/diceui/file-upload";

export const title = "Simple Button Trigger";

const Example = () => {
  const [files, setFiles] = React.useState<File[]>([]);

  return (
    <div className="flex flex-col items-center gap-3">
      <FileUpload
        value={files}
        onValueChange={setFiles}
        maxFiles={1}
        maxSize={5 * 1024 * 1024}
      >
        <FileUploadTrigger asChild>
          <Button variant="outline">
            <Upload className="mr-2 size-4" />
            Upload File
          </Button>
        </FileUploadTrigger>
      </FileUpload>
      {files.length > 0 && (
        <p className="text-sm text-muted-foreground">
          Selected: {files[0].name}
        </p>
      )}
    </div>
  );
};

export default Example;
