"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldSet,
  FieldLegend,
} from "@/components/ui/field";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";

export const title = "Form with Radio Buttons";

const formSchema = z.object({
  plan: z
    .string()
    .min(1, {
      message: "You need to select a plan.",
    })
    .refine((val) => ["free", "pro", "enterprise"].includes(val), {
      message: "You need to select a plan.",
    }),
});

const Example = () => {
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
  });

  function onSubmit(values: z.infer<typeof formSchema>) {
    console.log(values);
  }

  return (
    <div className="w-full max-w-md">
      <form className="space-y-4" onSubmit={form.handleSubmit(onSubmit)}>
        <Controller
          name="plan"
          control={form.control}
          render={({ field, fieldState }) => (
            <FieldSet>
              <FieldLegend>Select your plan</FieldLegend>
              <FieldDescription>
                Choose the plan that best fits your needs.
              </FieldDescription>
              <RadioGroup
                name={field.name}
                value={field.value}
                onValueChange={field.onChange}
              >
                <FieldGroup>
                  <Field
                    orientation="horizontal"
                    data-invalid={fieldState.invalid}
                  >
                    <RadioGroupItem
                      value="free"
                      id="plan-free"
                      aria-invalid={fieldState.invalid}
                    />
                    <FieldLabel htmlFor="plan-free" className="font-normal">
                      Free - $0/month
                    </FieldLabel>
                  </Field>
                  <Field
                    orientation="horizontal"
                    data-invalid={fieldState.invalid}
                  >
                    <RadioGroupItem
                      value="pro"
                      id="plan-pro"
                      aria-invalid={fieldState.invalid}
                    />
                    <FieldLabel htmlFor="plan-pro" className="font-normal">
                      Pro - $9/month
                    </FieldLabel>
                  </Field>
                  <Field
                    orientation="horizontal"
                    data-invalid={fieldState.invalid}
                  >
                    <RadioGroupItem
                      value="enterprise"
                      id="plan-enterprise"
                      aria-invalid={fieldState.invalid}
                    />
                    <FieldLabel
                      htmlFor="plan-enterprise"
                      className="font-normal"
                    >
                      Enterprise - Contact us
                    </FieldLabel>
                  </Field>
                </FieldGroup>
              </RadioGroup>
              {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
            </FieldSet>
          )}
        />
        <Button type="submit">Continue</Button>
      </form>
    </div>
  );
};

export default Example;
