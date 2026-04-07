"use client";

import { Camera, Trash2 } from "lucide-react";
import * as React from "react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { FileUpload, FileUploadTrigger } from "@/components/diceui/file-upload";

export const title = "Avatar with Menu";

const Example = () => {
  const [files, setFiles] = React.useState<File[]>([]);
  const defaultAvatar = "https://github.com/shadcn.png";

  const avatarPreview =
    files.length > 0 ? URL.createObjectURL(files[0]) : defaultAvatar;

  const handleRemove = () => {
    setFiles([]);
  };

  return (
    <div className="flex flex-col items-center gap-4">
      <FileUpload
        value={files}
        onValueChange={setFiles}
        accept="image/*"
        maxFiles={1}
        maxSize={2 * 1024 * 1024}
      >
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button className="group relative cursor-pointer rounded-full focus:ring-2 focus:ring-ring focus:ring-offset-2 focus:outline-none">
              <Avatar className="size-24">
                <AvatarImage src={avatarPreview} alt="Avatar" />
                <AvatarFallback>CN</AvatarFallback>
              </Avatar>
              <div className="absolute inset-0 flex items-center justify-center rounded-full bg-black/50 opacity-0 transition-opacity group-hover:opacity-100">
                <Camera className="size-6 text-white" />
              </div>
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="center">
            <FileUploadTrigger asChild>
              <DropdownMenuItem>
                <Camera className="mr-2 size-4" />
                Upload photo
              </DropdownMenuItem>
            </FileUploadTrigger>
            {files.length > 0 && (
              <DropdownMenuItem
                className="text-destructive focus:text-destructive"
                onClick={handleRemove}
              >
                <Trash2 className="mr-2 size-4" />
                Remove photo
              </DropdownMenuItem>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </FileUpload>
      <div className="text-center">
        <p className="font-medium">John Doe</p>
        <p className="text-sm text-muted-foreground">Click avatar to change</p>
      </div>
    </div>
  );
};

export default Example;
