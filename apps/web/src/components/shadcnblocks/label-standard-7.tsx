import { Label } from "@/components/ui/label";

export const title = "Error Label";

const Example = () => (
  <Label className="text-destructive">
    Password
    <span className="ml-1 text-xs font-normal">• Password is too weak</span>
  </Label>
);

export default Example;
