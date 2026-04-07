import { Loader2 } from "lucide-react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

export const title = "Avatar with Loading State";

const Example = () => (
  <div className="relative">
    <Avatar>
      <AvatarImage
        alt="@haydenbleasel"
        src="https://github.com/haydenbleasel.png"
      />
      <AvatarFallback>HB</AvatarFallback>
    </Avatar>
    <div className="absolute inset-0 flex items-center justify-center rounded-full bg-background/80">
      <Loader2 className="size-4 animate-spin text-muted-foreground" />
    </div>
  </div>
);

export default Example;
