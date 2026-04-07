import { Minus } from "lucide-react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

export const title = "Avatar with Do Not Disturb";

const Example = () => (
  <div className="relative w-fit">
    <Avatar>
      <AvatarImage
        alt="@haydenbleasel"
        src="https://github.com/haydenbleasel.png"
      />
      <AvatarFallback>HB</AvatarFallback>
    </Avatar>
    <span className="absolute -right-1 -bottom-1 flex size-3 items-center justify-center rounded-full border-2 border-background bg-red-500">
      <Minus className="size-2 stroke-white" />
    </span>
  </div>
);

export default Example;
