import { Avatar } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";

export const title = "Navigation with User Profile Sheet";

const Example = () => (
  <Sheet>
    <SheetTrigger asChild>
      <Button variant="outline">Menu</Button>
    </SheetTrigger>
    <SheetContent side="left">
      <SheetHeader>
        <SheetTitle>Menu</SheetTitle>
      </SheetHeader>
      <div className="flex flex-col gap-4 p-4">
        <div className="flex items-center gap-3 rounded-md bg-muted p-3">
          <Avatar>
            <div className="flex h-full w-full items-center justify-center bg-primary text-sm text-primary-foreground">
              JD
            </div>
          </Avatar>
          <div className="flex flex-col">
            <span className="text-sm font-medium">John Doe</span>
            <span className="text-xs text-muted-foreground">
              john@example.com
            </span>
          </div>
        </div>
        <nav className="flex flex-col gap-2">
          <a className="text-sm font-medium hover:underline" href="#">
            Dashboard
          </a>
          <a className="text-sm font-medium hover:underline" href="#">
            Projects
          </a>
          <a className="text-sm font-medium hover:underline" href="#">
            Settings
          </a>
          <a className="text-sm font-medium hover:underline" href="#">
            Sign Out
          </a>
        </nav>
      </div>
    </SheetContent>
  </Sheet>
);

export default Example;
