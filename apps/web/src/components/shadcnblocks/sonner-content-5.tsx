"use client";

import { toast } from "sonner";

import { Button } from "@/components/ui/button";

export const title = "Toast with Rich HTML Content";

const Example = () => (
  <Button
    onClick={() =>
      toast(
        <div className="flex items-start gap-3">
          <div className="flex size-10 shrink-0 items-center justify-center rounded-full bg-primary text-primary-foreground">
            <span className="text-sm font-semibold">JD</span>
          </div>
          <div className="flex flex-col gap-1">
            <div className="flex items-center gap-2">
              <strong className="text-sm">John Doe</strong>
              <span className="rounded-full bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
                New
              </span>
            </div>
            <p className="text-sm">Hey! Just finished the design mockups 🎨</p>
            <span className="text-xs text-muted-foreground">2 minutes ago</span>
          </div>
        </div>,
      )
    }
    variant="outline"
  >
    Show Toast
  </Button>
);

export default Example;
