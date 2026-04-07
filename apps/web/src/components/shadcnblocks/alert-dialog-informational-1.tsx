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

export const title = "Simple Information Alert";

const Example = () => (
  <AlertDialog>
    <AlertDialogTrigger render={<Button variant="outline" />}>Show Info</AlertDialogTrigger>
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>System Update</AlertDialogTitle>
        <AlertDialogDescription>
          A new system update is available. The update will be installed during
          the next maintenance window.
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogAction>OK</AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
);

export default Example;
