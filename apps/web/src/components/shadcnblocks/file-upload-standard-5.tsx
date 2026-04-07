"use client";

import { Upload } from "lucide-react";

import { Button } from "@/components/ui/button";
import { FileUpload, FileUploadTrigger } from "@/components/diceui/file-upload";

export const title = "Disabled State";

const Example = () => {
  return (
    <FileUpload disabled className="w-full max-w-md">
      <FileUploadTrigger asChild>
        <Button variant="outline" className="w-full" disabled>
          <Upload className="mr-2 size-4" />
          Upload Disabled
        </Button>
      </FileUploadTrigger>
    </FileUpload>
  );
};

export default Example;
