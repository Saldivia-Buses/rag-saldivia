"use client";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "@/components/ui/hover-card";

export const title = "Simple Profile Hover Card";

const Example = () => (
  <HoverCard>
    <HoverCardTrigger asChild>
      <a className="text-sm font-medium underline" href="#">
        @haydenbleasel
      </a>
    </HoverCardTrigger>
    <HoverCardContent>
      <div className="flex gap-4">
        <Avatar>
          <AvatarImage src="https://github.com/haydenbleasel.png" />
          <AvatarFallback>HB</AvatarFallback>
        </Avatar>
        <div className="space-y-1">
          <h4 className="text-sm font-semibold">Hayden Bleasel</h4>
          <p className="text-sm text-muted-foreground">
            Product designer and developer
          </p>
        </div>
      </div>
    </HoverCardContent>
  </HoverCard>
);

export default Example;
