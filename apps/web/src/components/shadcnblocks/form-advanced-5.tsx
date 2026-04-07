"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldGroup,
  FieldLabel,
  FieldSet,
  FieldLegend,
} from "@/components/ui/field";
import { Switch } from "@/components/ui/switch";

export const title = "Form with Switch";

const formSchema = z.object({
  emailNotifications: z.boolean(),
  pushNotifications: z.boolean(),
  desktopNotifications: z.boolean(),
});

const Example = () => {
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      emailNotifications: true,
      pushNotifications: false,
      desktopNotifications: true,
    },
  });

  function onSubmit(values: z.infer<typeof formSchema>) {
    console.log(values);
  }

  return (
    <div className="w-full max-w-md">
      <form className="space-y-6" onSubmit={form.handleSubmit(onSubmit)}>
        <FieldSet>
          <FieldLegend>Notification Settings</FieldLegend>
          <FieldGroup>
            <Controller
              name="emailNotifications"
              control={form.control}
              render={({ field }) => (
                <Field
                  orientation="horizontal"
                  className="flex-row items-center justify-between rounded-lg border p-4"
                >
                  <FieldContent>
                    <FieldLabel htmlFor={field.name} className="text-base">
                      Email Notifications
                    </FieldLabel>
                    <FieldDescription>
                      Receive notifications via email.
                    </FieldDescription>
                  </FieldContent>
                  <Switch
                    id={field.name}
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </Field>
              )}
            />
            <Controller
              name="pushNotifications"
              control={form.control}
              render={({ field }) => (
                <Field
                  orientation="horizontal"
                  className="flex-row items-center justify-between rounded-lg border p-4"
                >
                  <FieldContent>
                    <FieldLabel htmlFor={field.name} className="text-base">
                      Push Notifications
                    </FieldLabel>
                    <FieldDescription>
                      Receive push notifications on your device.
                    </FieldDescription>
                  </FieldContent>
                  <Switch
                    id={field.name}
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </Field>
              )}
            />
            <Controller
              name="desktopNotifications"
              control={form.control}
              render={({ field }) => (
                <Field
                  orientation="horizontal"
                  className="flex-row items-center justify-between rounded-lg border p-4"
                >
                  <FieldContent>
                    <FieldLabel htmlFor={field.name} className="text-base">
                      Desktop Notifications
                    </FieldLabel>
                    <FieldDescription>
                      Receive notifications on your desktop.
                    </FieldDescription>
                  </FieldContent>
                  <Switch
                    id={field.name}
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </Field>
              )}
            />
          </FieldGroup>
        </FieldSet>
        <Button type="submit">Save Preferences</Button>
      </form>
    </div>
  );
};

export default Example;
