"use client";

import { useState } from "react";

import { Label } from "@/components/ui/label";
import { Slider } from "@/components/ui/slider";

export const title = "With Tooltip";

const Example = () => {
  const [value, setValue] = useState([45]);

  return (
    <div className="flex w-full max-w-md flex-col gap-2">
      <Label htmlFor="slider">Progress</Label>
      <div className="relative pt-6">
        <Slider id="slider" onValueChange={setValue} value={value} />
        <div
          className="absolute -top-2 -translate-x-1/2 rounded bg-primary px-2 py-1 text-xs text-primary-foreground"
          style={{ left: `${value[0]}%` }}
        >
          {value[0]}%
        </div>
      </div>
    </div>
  );
};

export default Example;
