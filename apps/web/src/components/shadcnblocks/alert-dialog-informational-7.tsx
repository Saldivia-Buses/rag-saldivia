import { Sparkles } from "lucide-react";

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

export const title = "Information with Highlighted Content";

const Example = () => (
  <AlertDialog>
    <AlertDialogTrigger render={<Button variant="outline" />}><Sparkles className="size-4" />What's New
              </AlertDialogTrigger>
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>New in Version 2.0</AlertDialogTitle>
        <AlertDialogDescription>
          Check out the latest features and improvements:
        </AlertDialogDescription>
      </AlertDialogHeader>
      <div className="space-y-2">
        <div className="rounded-lg border bg-muted/50 p-3">
          <div className="text-sm font-medium">Dark Mode</div>
          <p className="text-xs text-muted-foreground">
            Toggle between light and dark themes
          </p>
        </div>
        <div className="rounded-lg border bg-muted/50 p-3">
          <div className="text-sm font-medium">Improved Performance</div>
          <p className="text-xs text-muted-foreground">
            Up to 50% faster load times
          </p>
        </div>
        <div className="rounded-lg border bg-muted/50 p-3">
          <div className="text-sm font-medium">New Integrations</div>
          <p className="text-xs text-muted-foreground">
            Connect with your favorite tools
          </p>
        </div>
      </div>
      <AlertDialogFooter>
        <AlertDialogAction>Awesome!</AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
);

export default Example;
