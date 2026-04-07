import { PartyPopper } from "lucide-react";

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

export const title = "Success with Celebration";

const Example = () => (
  <AlertDialog>
    <AlertDialogTrigger render={<Button />}>Complete Action</AlertDialogTrigger>
    <AlertDialogContent>
      <AlertDialogHeader className="items-center">
        <div className="flex size-12 items-center justify-center rounded-full bg-gradient-to-br from-purple-100 to-pink-100 dark:from-purple-900 dark:to-pink-900">
          <PartyPopper className="size-6 text-purple-600 dark:text-purple-400" />
        </div>
        <AlertDialogTitle>Milestone Achieved!</AlertDialogTitle>
        <AlertDialogDescription className="text-center">
          Congratulations! You've completed 100 tasks. Keep up the great work!
        </AlertDialogDescription>
      </AlertDialogHeader>
      <div className="rounded-lg bg-gradient-to-r from-purple-50 to-pink-50 p-4 text-center dark:from-purple-950 dark:to-pink-950">
        <div className="text-xs font-medium text-muted-foreground">
          ACHIEVEMENT UNLOCKED
        </div>
        <div className="mt-1 text-lg font-semibold">Century Club</div>
      </div>
      <AlertDialogFooter>
        <AlertDialogAction>Awesome!</AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
);

export default Example;
