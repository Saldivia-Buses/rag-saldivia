"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Controller } from "react-hook-form";

import {
  Field,
  FieldDescription,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";

export const title = "RadioGroup in Form";

const formSchema = z.object({
  type: z.enum(["all", "mentions", "none"], {
    message: "You need to select a notification type.",
  }),
});

const Example = () => {
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
  });

  function onSubmit(values: z.infer<typeof formSchema>) {
    toast.success(`You selected: ${values.type}`);
  }

  return (
    <div className="w-full max-w-md">
      <form className="space-y-6" onSubmit={form.handleSubmit(onSubmit)}>
        <Controller
          control={form.control}
          name="type"
          render={({ field, fieldState }) => (
            <Field className="space-y-3" data-invalid={fieldState.invalid}>
              <FieldLabel>Notify me about...</FieldLabel>
              <RadioGroup
                className="flex flex-col space-y-1"
                defaultValue={field.value}
                onValueChange={field.onChange}
              >
                <div className="flex items-center space-x-2">
                  <RadioGroupItem id="all" value="all" />
                  <Label htmlFor="all">All new messages</Label>
                </div>
                <div className="flex items-center space-x-2">
                  <RadioGroupItem id="mentions" value="mentions" />
                  <Label htmlFor="mentions">Direct messages and mentions</Label>
                </div>
                <div className="flex items-center space-x-2">
                  <RadioGroupItem id="none" value="none" />
                  <Label htmlFor="none">Nothing</Label>
                </div>
              </RadioGroup>
              <FieldDescription>
                Choose how you want to be notified about new activity.
              </FieldDescription>
              {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
            </Field>
          )}
        />
        <Button type="submit">Submit</Button>
      </form>
    </div>
  );
};

export default Example;
