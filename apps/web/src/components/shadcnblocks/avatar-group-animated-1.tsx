"use client";

import { AvatarGroup } from "@/components/shadcnblocks/avatar-group";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

export const title = "Animated";

const Example = () => (
  <AvatarGroup animate>
    <Avatar>
      <AvatarImage src="https://github.com/haydenbleasel.png" />
      <AvatarFallback>HB</AvatarFallback>
    </Avatar>
    <Avatar>
      <AvatarImage src="https://github.com/shadcn.png" />
      <AvatarFallback>CN</AvatarFallback>
    </Avatar>
    <Avatar>
      <AvatarImage src="https://github.com/JugglerX.png" />
      <AvatarFallback>LR</AvatarFallback>
    </Avatar>
    <Avatar>
      <AvatarImage src="https://github.com/zerostaticthemes.png" />
      <AvatarFallback>SC</AvatarFallback>
    </Avatar>
  </AvatarGroup>
);

export default Example;
