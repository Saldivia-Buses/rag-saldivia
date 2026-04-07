import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";

export const title = "Mobile Menu Sheet";

const Example = () => (
  <Sheet>
    <SheetTrigger asChild>
      <Button variant="outline">Menu</Button>
    </SheetTrigger>
    <SheetContent side="left">
      <SheetHeader>
        <SheetTitle>Menu</SheetTitle>
      </SheetHeader>
      <nav className="flex flex-col gap-4 p-4">
        <a className="text-sm font-medium hover:underline" href="#">
          Home
        </a>
        <a className="text-sm font-medium hover:underline" href="#">
          Products
        </a>
        <a className="text-sm font-medium hover:underline" href="#">
          About
        </a>
        <a className="text-sm font-medium hover:underline" href="#">
          Contact
        </a>
      </nav>
    </SheetContent>
  </Sheet>
);

export default Example;
