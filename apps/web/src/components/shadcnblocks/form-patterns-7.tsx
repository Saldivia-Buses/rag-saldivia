"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import type { CheckedState } from "@radix-ui/react-checkbox";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldSet,
  FieldLegend,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

export const title = "Search Filter Form";

const categories = [
  { id: "electronics", label: "Electronics" },
  { id: "clothing", label: "Clothing" },
  { id: "books", label: "Books" },
  { id: "home", label: "Home & Garden" },
] as const;

const formSchema = z.object({
  query: z.string(),
  sortBy: z.string(),
  priceMin: z.coerce.number().min(0).optional(),
  priceMax: z.coerce.number().min(0).optional(),
  categories: z.array(z.string()),
  inStock: z.boolean(),
});

type FormInput = z.input<typeof formSchema>;
type FormOutput = z.output<typeof formSchema>;

const Example = () => {
  const form = useForm<FormInput, unknown, FormOutput>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      query: "",
      sortBy: "relevance",
      priceMin: undefined,
      priceMax: undefined,
      categories: [],
      inStock: false,
    },
  });

  function onSubmit(values: FormOutput) {
    console.log(values);
  }

  const watchedCategories = form.watch("categories");

  return (
    <div className="w-full max-w-md">
      <form className="space-y-6" onSubmit={form.handleSubmit(onSubmit)}>
        <div>
          <h2 className="text-2xl font-bold">Search Products</h2>
          <p className="text-sm text-muted-foreground">
            Filter and find what you're looking for
          </p>
        </div>
        <FieldGroup>
          <Controller
            name="query"
            control={form.control}
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel htmlFor={field.name}>Search</FieldLabel>
                <Input
                  {...field}
                  id={field.name}
                  aria-invalid={fieldState.invalid}
                  className="bg-background"
                  placeholder="Search products..."
                  type="search"
                />
                {fieldState.invalid && (
                  <FieldError errors={[fieldState.error]} />
                )}
              </Field>
            )}
          />
          <Controller
            name="sortBy"
            control={form.control}
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel htmlFor={field.name}>Sort By</FieldLabel>
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger
                    id={field.name}
                    aria-invalid={fieldState.invalid}
                    className="bg-background"
                  >
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="relevance">Relevance</SelectItem>
                    <SelectItem value="price-low">
                      Price: Low to High
                    </SelectItem>
                    <SelectItem value="price-high">
                      Price: High to Low
                    </SelectItem>
                    <SelectItem value="newest">Newest First</SelectItem>
                  </SelectContent>
                </Select>
                {fieldState.invalid && (
                  <FieldError errors={[fieldState.error]} />
                )}
              </Field>
            )}
          />
        </FieldGroup>
        <FieldSet>
          <FieldLegend variant="label">Price Range</FieldLegend>
          <FieldGroup>
            <div className="grid grid-cols-2 gap-4">
              <Controller
                name="priceMin"
                control={form.control}
                render={({ field, fieldState }) => (
                  <Field data-invalid={fieldState.invalid}>
                    <FieldLabel htmlFor={field.name}>Min</FieldLabel>
                    <Input
                      {...field}
                      id={field.name}
                      aria-invalid={fieldState.invalid}
                      className="bg-background"
                      placeholder="$0"
                      type="number"
                      value={String(field.value ?? "")}
                      onChange={(e) => field.onChange(e.target.value)}
                    />
                    {fieldState.invalid && (
                      <FieldError errors={[fieldState.error]} />
                    )}
                  </Field>
                )}
              />
              <Controller
                name="priceMax"
                control={form.control}
                render={({ field, fieldState }) => (
                  <Field data-invalid={fieldState.invalid}>
                    <FieldLabel htmlFor={field.name}>Max</FieldLabel>
                    <Input
                      {...field}
                      id={field.name}
                      aria-invalid={fieldState.invalid}
                      className="bg-background"
                      placeholder="$1000"
                      type="number"
                      value={String(field.value ?? "")}
                      onChange={(e) => field.onChange(e.target.value)}
                    />
                    {fieldState.invalid && (
                      <FieldError errors={[fieldState.error]} />
                    )}
                  </Field>
                )}
              />
            </div>
          </FieldGroup>
        </FieldSet>
        <FieldSet>
          <FieldLegend variant="label">Categories</FieldLegend>
          <FieldDescription>
            Select the categories you want to search in.
          </FieldDescription>
          <FieldGroup>
            {categories.map((category) => (
              <Field orientation="horizontal" key={category.id}>
                <Checkbox
                  id={`category-${category.id}`}
                  checked={watchedCategories.includes(category.id)}
                  onCheckedChange={(checked: CheckedState) => {
                    const current = form.getValues("categories");
                    if (checked === true) {
                      form.setValue("categories", [...current, category.id]);
                    } else {
                      form.setValue(
                        "categories",
                        current.filter((value) => value !== category.id),
                      );
                    }
                  }}
                />
                <FieldLabel
                  htmlFor={`category-${category.id}`}
                  className="font-normal"
                >
                  {category.label}
                </FieldLabel>
              </Field>
            ))}
          </FieldGroup>
        </FieldSet>
        <Controller
          name="inStock"
          control={form.control}
          render={({ field }) => (
            <Field orientation="horizontal">
              <Checkbox
                id={field.name}
                checked={field.value}
                onCheckedChange={(checked: CheckedState) =>
                  field.onChange(checked === true)
                }
              />
              <FieldContent>
                <FieldLabel htmlFor={field.name}>In Stock Only</FieldLabel>
                <FieldDescription>
                  Show only products that are currently available.
                </FieldDescription>
              </FieldContent>
            </Field>
          )}
        />
        <div className="flex gap-2">
          <Button className="flex-1" type="submit">
            Apply Filters
          </Button>
          <Button onClick={() => form.reset()} type="button" variant="outline">
            Reset
          </Button>
        </div>
      </form>
    </div>
  );
};

export default Example;
