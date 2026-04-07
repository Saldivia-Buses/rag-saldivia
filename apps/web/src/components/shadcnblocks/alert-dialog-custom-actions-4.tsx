import { Download } from "lucide-react";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/shadcnblocks/alert-dialog";
import { Button } from "@/components/ui/button";

export const title = "Actions with Loading State";

const Example = () => (
  <AlertDialog>
    <AlertDialogTrigger render={<Button variant="outline" />}><Download className="size-4" />Download
              </AlertDialogTrigger>
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>Download Report</AlertDialogTitle>
        <AlertDialogDescription>
          Your report is ready to download. This file contains sensitive
          information and should be handled securely.
        </AlertDialogDescription>
      </AlertDialogHeader>
      <div className="rounded-md bg-muted p-3 text-sm">
        <div className="flex items-center justify-between">
          <span>File size:</span>
          <span className="font-medium">2.4 MB</span>
        </div>
        <div className="flex items-center justify-between">
          <span>Format:</span>
          <span className="font-medium">PDF</span>
        </div>
      </div>
      <AlertDialogFooter>
        <AlertDialogCancel>Cancel</AlertDialogCancel>
        <AlertDialogAction>
          <Download className="size-4" />
          Download File
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
);

export default Example;
