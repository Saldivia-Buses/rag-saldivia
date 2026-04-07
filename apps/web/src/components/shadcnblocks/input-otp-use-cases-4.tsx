"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { AlertTriangle } from "lucide-react";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Controller } from "react-hook-form";

import { Field, FieldError } from "@/components/ui/field";
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSeparator,
  InputOTPSlot,
} from "@/components/ui/input-otp";

export const title = "Transaction Confirmation";

const formSchema = z.object({
  code: z.string().min(6, {
    message: "Confirmation code must be 6 characters.",
  }),
});

const Example = () => {
  const [isConfirming, setIsConfirming] = useState(false);

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      code: "",
    },
  });

  function onSubmit(_values: z.infer<typeof formSchema>) {
    setIsConfirming(true);
    // Simulate transaction confirmation
    setTimeout(() => {
      setIsConfirming(false);
      toast.success("Transaction confirmed successfully!");
    }, 1500);
  }

  return (
    <div className="w-full max-w-sm space-y-6">
      <div className="space-y-2 text-center">
        <h2 className="text-2xl font-semibold">Confirm Transaction</h2>
        <p className="text-sm text-muted-foreground">
          You're about to send $1,250.00 to John Doe
        </p>
      </div>
      <Alert>
        <AlertTriangle className="h-4 w-4" />
        <AlertTitle>Security Check</AlertTitle>
        <AlertDescription>
          For your security, please enter the confirmation code sent to your
          registered email address.
        </AlertDescription>
      </Alert>
      <form className="space-y-6" onSubmit={form.handleSubmit(onSubmit)}>
        <Controller
          control={form.control}
          name="code"
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <div className="flex justify-center">
                <InputOTP maxLength={6} {...field}>
                  <InputOTPGroup>
                    <InputOTPSlot className="bg-background" index={0} />
                    <InputOTPSlot className="bg-background" index={1} />
                    <InputOTPSlot className="bg-background" index={2} />
                  </InputOTPGroup>
                  <InputOTPSeparator />
                  <InputOTPGroup>
                    <InputOTPSlot className="bg-background" index={3} />
                    <InputOTPSlot className="bg-background" index={4} />
                    <InputOTPSlot className="bg-background" index={5} />
                  </InputOTPGroup>
                </InputOTP>
              </div>
              {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
            </Field>
          )}
        />
        <div className="flex gap-2">
          <Button
            className="flex-1"
            onClick={() => toast.info("Transaction cancelled")}
            type="button"
            variant="outline"
          >
            Cancel
          </Button>
          <Button className="flex-1" disabled={isConfirming} type="submit">
            {isConfirming ? "Confirming..." : "Confirm Transaction"}
          </Button>
        </div>
      </form>
    </div>
  );
};

export default Example;
