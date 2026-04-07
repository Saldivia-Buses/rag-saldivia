"use client";

import { Camera, ImagePlus, X } from "lucide-react";
import * as React from "react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { FileUpload, FileUploadTrigger } from "@/components/diceui/file-upload";

export const title = "Profile with Cover and Avatar";

const Example = () => {
  const [coverFiles, setCoverFiles] = React.useState<File[]>([]);
  const [avatarFiles, setAvatarFiles] = React.useState<File[]>([]);

  const coverPreview =
    coverFiles.length > 0 ? URL.createObjectURL(coverFiles[0]) : null;

  const avatarPreview =
    avatarFiles.length > 0
      ? URL.createObjectURL(avatarFiles[0])
      : "https://github.com/shadcn.png";

  return (
    <div className="w-full max-w-md overflow-hidden rounded-lg border">
      {/* Cover Image */}
      <FileUpload
        value={coverFiles}
        onValueChange={setCoverFiles}
        accept="image/*"
        maxFiles={1}
        maxSize={5 * 1024 * 1024}
      >
        <FileUploadTrigger asChild>
          <div className="group relative h-32 cursor-pointer bg-gradient-to-r from-primary/20 to-primary/10">
            {coverPreview && (
              // eslint-disable-next-line @next/next/no-img-element
              <img
                src={coverPreview}
                alt="Cover"
                className="h-full w-full object-cover"
              />
            )}
            <div className="absolute inset-0 flex items-center justify-center bg-black/40 opacity-0 transition-opacity group-hover:opacity-100">
              <div className="flex items-center gap-2 text-white">
                <ImagePlus className="size-5" />
                <span className="text-sm font-medium">
                  {coverPreview ? "Change cover" : "Add cover"}
                </span>
              </div>
            </div>
            {coverPreview && (
              <Button
                variant="secondary"
                size="icon"
                className="absolute right-2 bottom-2 size-7 opacity-0 group-hover:opacity-100"
                onClick={(e) => {
                  e.stopPropagation();
                  setCoverFiles([]);
                }}
              >
                <X className="size-4" />
              </Button>
            )}
          </div>
        </FileUploadTrigger>
      </FileUpload>

      {/* Avatar */}
      <div className="relative px-4 pb-4">
        <FileUpload
          value={avatarFiles}
          onValueChange={setAvatarFiles}
          accept="image/*"
          maxFiles={1}
          maxSize={2 * 1024 * 1024}
        >
          <FileUploadTrigger asChild>
            <button className="group relative -mt-12 cursor-pointer rounded-full ring-4 ring-background">
              <Avatar className="size-24">
                <AvatarImage src={avatarPreview} alt="Avatar" />
                <AvatarFallback>CN</AvatarFallback>
              </Avatar>
              <div className="absolute inset-0 flex items-center justify-center rounded-full bg-black/50 opacity-0 transition-opacity group-hover:opacity-100">
                <Camera className="size-6 text-white" />
              </div>
            </button>
          </FileUploadTrigger>
        </FileUpload>

        <div className="mt-3">
          <h3 className="text-lg font-semibold">John Doe</h3>
          <p className="text-sm text-muted-foreground">@johndoe</p>
          <p className="mt-2 text-sm">
            Software developer passionate about building great products.
          </p>
        </div>
      </div>
    </div>
  );
};

export default Example;
