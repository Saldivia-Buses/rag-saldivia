"use client";

import { CloudUpload, X } from "lucide-react";
import * as React from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";

import { Button } from "@/components/ui/button";
import { Controller } from "react-hook-form";

import {
  Field,
  FieldDescription,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";
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

export const title = "React Hook Form Integration";

const formSchema = z.object({
  files: z
    .array(z.instanceof(File))
    .min(1, "Please upload at least one file")
    .max(2, "Maximum 2 files allowed")
    .refine((files) => files.every((file) => file.size <= 5 * 1024 * 1024), {
      message: "File size must be less than 5MB",
    }),
});

type FormValues = z.infer<typeof formSchema>;

const Example = () => {
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      files: [],
    },
  });

  const onSubmit = React.useCallback((data: FormValues) => {
    console.log("Submitted files:", data.files);
    alert(`Submitted ${data.files.length} file(s)`);
  }, []);

  return (
    <form
      onSubmit={form.handleSubmit(onSubmit)}
      className="w-full max-w-md space-y-4"
    >
      <Controller
        control={form.control}
        name="files"
        render={({ field, fieldState }) => (
          <Field data-invalid={fieldState.invalid}>
            <FieldLabel>Attachments</FieldLabel>
            <FileUpload
              value={field.value}
              onValueChange={field.onChange}
              accept="image/*"
              maxFiles={2}
              maxSize={5 * 1024 * 1024}
              onFileReject={(_, message) => {
                form.setError("files", { message });
              }}
              multiple
            >
              <FileUploadDropzone className="flex-row flex-wrap border-dotted text-center">
                <CloudUpload className="size-4" />
                <span className="text-sm">Drag and drop or</span>
                <FileUploadTrigger asChild>
                  <Button variant="link" size="sm" className="h-auto p-0">
                    choose files
                  </Button>
                </FileUploadTrigger>
                <span className="text-sm">to upload</span>
              </FileUploadDropzone>
              <FileUploadList>
                {field.value?.map((file, index) => (
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
            <FieldDescription>
              Upload up to 2 images (max 5MB each)
            </FieldDescription>
            {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
          </Field>
        )}
      />
      <Button type="submit">Submit</Button>
    </form>
  );
};

export default Example;
