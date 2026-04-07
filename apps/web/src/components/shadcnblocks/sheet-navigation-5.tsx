import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";

export const title = "Navigation with Search Sheet";

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
        <Input placeholder="Search menu..." />
        <nav className="flex flex-col gap-2">
          <a className="text-sm font-medium hover:underline" href="#">
            Home
          </a>
          <a className="text-sm font-medium hover:underline" href="#">
            Products
          </a>
          <a className="text-sm font-medium hover:underline" href="#">
            Services
          </a>
          <a className="text-sm font-medium hover:underline" href="#">
            Blog
          </a>
          <a className="text-sm font-medium hover:underline" href="#">
            Contact
          </a>
        </nav>
      </div>
    </SheetContent>
  </Sheet>
);

export default Example;
