"use client";

import { AvatarGroup } from "@/components/shadcnblocks/avatar-group";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

export const title = "With Tooltips";

const Example = () => (
  <TooltipProvider>
    <AvatarGroup>
      <Tooltip>
        <TooltipTrigger asChild>
          <Avatar>
            <AvatarImage src="https://github.com/haydenbleasel.png" />
            <AvatarFallback>HB</AvatarFallback>
          </Avatar>
        </TooltipTrigger>
        <TooltipContent>
          <p>Hayden Bleasel</p>
        </TooltipContent>
      </Tooltip>
      <Tooltip>
        <TooltipTrigger asChild>
          <Avatar>
            <AvatarImage src="https://github.com/shadcn.png" />
            <AvatarFallback>CN</AvatarFallback>
          </Avatar>
        </TooltipTrigger>
        <TooltipContent>
          <p>shadcn</p>
        </TooltipContent>
      </Tooltip>
      <Tooltip>
        <TooltipTrigger asChild>
          <Avatar>
            <AvatarImage src="https://github.com/leerob.png" />
            <AvatarFallback>LR</AvatarFallback>
          </Avatar>
        </TooltipTrigger>
        <TooltipContent>
          <p>Lee Robinson</p>
        </TooltipContent>
      </Tooltip>
      <Tooltip>
        <TooltipTrigger asChild>
          <Avatar>
            <AvatarImage src="https://github.com/serafimcloud.png" />
            <AvatarFallback>SC</AvatarFallback>
          </Avatar>
        </TooltipTrigger>
        <TooltipContent>
          <p>Serafim Cloud</p>
        </TooltipContent>
      </Tooltip>
    </AvatarGroup>
  </TooltipProvider>
);

export default Example;
