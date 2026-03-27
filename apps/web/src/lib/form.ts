import { useForm, type DefaultValues, type FieldValues } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"

/**
 * Helper que combina useForm + zodResolver para reducir boilerplate.
 * Inferencia automática del tipo desde el schema Zod.
 */
export function createForm<T extends FieldValues>(
  schema: z.ZodType<T>,
  defaultValues: DefaultValues<T>
) {
  return useForm<T>({
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    resolver: zodResolver(schema as any),
    defaultValues,
  })
}
