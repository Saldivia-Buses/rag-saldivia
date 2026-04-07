"use client";

import { AvatarGroup } from "@/components/shadcnblocks/avatar-group";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

const Example = () => (
  <AvatarGroup>
    <Avatar>
      <AvatarImage src="https://github.com/haydenbleasel.png" />
      <AvatarFallback>HB</AvatarFallback>
    </Avatar>
    <Avatar>
      <AvatarImage src="https://github.com/shadcn.png" />
      <AvatarFallback>CN</AvatarFallback>
    </Avatar>
    <Avatar>
      <AvatarImage src="https://github.com/leerob.png" />
      <AvatarFallback>LR</AvatarFallback>
    </Avatar>
    <Avatar>
      <AvatarImage src="https://github.com/JugglerX.png" />
      <AvatarFallback>SC</AvatarFallback>
    </Avatar>
  </AvatarGroup>
);

export default Example;
