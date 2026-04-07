import { Lightbulb } from "lucide-react";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/shadcnblocks/alert-dialog";
import { Button } from "@/components/ui/button";

export const title = "Information with Centered Icon";

const Example = () => (
  <AlertDialog>
    <AlertDialogTrigger render={<Button variant="outline" />}>Show Tip</AlertDialogTrigger>
    <AlertDialogContent>
      <AlertDialogHeader className="items-center">
        <div className="flex size-12 items-center justify-center rounded-full bg-yellow-100 dark:bg-yellow-900">
          <Lightbulb className="size-6 text-yellow-600 dark:text-yellow-400" />
        </div>
        <AlertDialogTitle>Pro Tip</AlertDialogTitle>
        <AlertDialogDescription className="text-center">
          Use keyboard shortcuts to navigate faster. Press{" "}
          <kbd className="rounded bg-muted px-1.5 py-0.5 text-xs font-semibold">
            Cmd+K
          </kbd>{" "}
          to open the command palette.
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogAction>Got it</AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
);

export default Example;
