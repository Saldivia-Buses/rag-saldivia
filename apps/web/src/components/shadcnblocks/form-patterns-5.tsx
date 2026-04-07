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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";

export const title = "Settings Form";

const formSchema = z.object({
  language: z.string(),
  theme: z.string(),
  emailNotifications: z.boolean(),
  marketingEmails: z.boolean(),
  securityAlerts: z.boolean(),
  twoFactorAuth: z.boolean(),
});

const Example = () => {
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      language: "en",
      theme: "system",
      emailNotifications: true,
      marketingEmails: false,
      securityAlerts: true,
      twoFactorAuth: false,
    },
  });

  function onSubmit(values: z.infer<typeof formSchema>) {
    console.log(values);
  }

  return (
    <div className="w-full max-w-2xl">
      <form className="space-y-8" onSubmit={form.handleSubmit(onSubmit)}>
        <div>
          <h3 className="text-lg font-medium">Preferences</h3>
          <p className="text-sm text-muted-foreground">
            Manage your application preferences
          </p>
        </div>
        <FieldGroup>
          <Controller
            name="language"
            control={form.control}
            render={({ field }) => (
              <Field>
                <FieldLabel htmlFor={field.name}>Language</FieldLabel>
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={field.name} className="bg-background">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="en">English</SelectItem>
                    <SelectItem value="es">Spanish</SelectItem>
                    <SelectItem value="fr">French</SelectItem>
                    <SelectItem value="de">German</SelectItem>
                  </SelectContent>
                </Select>
                <FieldDescription>
                  Choose your preferred language.
                </FieldDescription>
              </Field>
            )}
          />
          <Controller
            name="theme"
            control={form.control}
            render={({ field }) => (
              <Field>
                <FieldLabel htmlFor={field.name}>Theme</FieldLabel>
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={field.name} className="bg-background">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="light">Light</SelectItem>
                    <SelectItem value="dark">Dark</SelectItem>
                    <SelectItem value="system">System</SelectItem>
                  </SelectContent>
                </Select>
                <FieldDescription>
                  Select your preferred theme.
                </FieldDescription>
              </Field>
            )}
          />
        </FieldGroup>
        <FieldSet>
          <FieldLegend>Notifications</FieldLegend>
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
                      Receive notifications via email
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
              name="marketingEmails"
              control={form.control}
              render={({ field }) => (
                <Field
                  orientation="horizontal"
                  className="flex-row items-center justify-between rounded-lg border p-4"
                >
                  <FieldContent>
                    <FieldLabel htmlFor={field.name} className="text-base">
                      Marketing Emails
                    </FieldLabel>
                    <FieldDescription>
                      Receive emails about new products and features
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
              name="securityAlerts"
              control={form.control}
              render={({ field }) => (
                <Field
                  orientation="horizontal"
                  className="flex-row items-center justify-between rounded-lg border p-4"
                >
                  <FieldContent>
                    <FieldLabel htmlFor={field.name} className="text-base">
                      Security Alerts
                    </FieldLabel>
                    <FieldDescription>
                      Receive alerts about account security
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
        <FieldSet>
          <FieldLegend>Security</FieldLegend>
          <FieldGroup>
            <Controller
              name="twoFactorAuth"
              control={form.control}
              render={({ field }) => (
                <Field
                  orientation="horizontal"
                  className="flex-row items-center justify-between rounded-lg border p-4"
                >
                  <FieldContent>
                    <FieldLabel htmlFor={field.name} className="text-base">
                      Two-Factor Authentication
                    </FieldLabel>
                    <FieldDescription>
                      Add an extra layer of security to your account
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
        <Button type="submit">Save Settings</Button>
      </form>
    </div>
  );
};

export default Example;
