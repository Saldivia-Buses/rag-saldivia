import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";

export const title = "Sidebar Navigation Sheet";

const Example = () => (
  <Sheet>
    <SheetTrigger asChild>
      <Button variant="outline">Navigation</Button>
    </SheetTrigger>
    <SheetContent side="left">
      <SheetHeader>
        <SheetTitle>Navigation</SheetTitle>
      </SheetHeader>
      <nav className="flex flex-col gap-4 p-4">
        <div className="flex flex-col gap-2">
          <h3 className="text-sm font-semibold">Dashboard</h3>
          <a
            className="pl-4 text-sm text-muted-foreground hover:text-foreground"
            href="#"
          >
            Overview
          </a>
          <a
            className="pl-4 text-sm text-muted-foreground hover:text-foreground"
            href="#"
          >
            Analytics
          </a>
        </div>
        <div className="flex flex-col gap-2">
          <h3 className="text-sm font-semibold">Settings</h3>
          <a
            className="pl-4 text-sm text-muted-foreground hover:text-foreground"
            href="#"
          >
            Profile
          </a>
          <a
            className="pl-4 text-sm text-muted-foreground hover:text-foreground"
            href="#"
          >
            Account
          </a>
        </div>
      </nav>
    </SheetContent>
  </Sheet>
);

export default Example;
