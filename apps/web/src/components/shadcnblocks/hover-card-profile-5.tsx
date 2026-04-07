"use client";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "@/components/ui/hover-card";
import { Separator } from "@/components/ui/separator";

export const title = "Profile with Stats";

const Example = () => (
  <HoverCard>
    <HoverCardTrigger asChild>
      <a className="text-sm font-medium underline" href="#">
        @haydenbleasel
      </a>
    </HoverCardTrigger>
    <HoverCardContent className="w-80">
      <div className="space-y-3">
        <div className="flex gap-4">
          <Avatar className="h-12 w-12">
            <AvatarImage src="https://github.com/haydenbleasel.png" />
            <AvatarFallback>HB</AvatarFallback>
          </Avatar>
          <div className="flex-1 space-y-1">
            <h4 className="text-sm font-semibold">Hayden Bleasel</h4>
            <p className="text-sm text-muted-foreground">
              Product designer and developer
            </p>
          </div>
        </div>
        <Separator />
        <div className="grid grid-cols-3 gap-4 text-center">
          <div>
            <div className="text-lg font-semibold">243</div>
            <div className="text-xs text-muted-foreground">Posts</div>
          </div>
          <div>
            <div className="text-lg font-semibold">2.5k</div>
            <div className="text-xs text-muted-foreground">Followers</div>
          </div>
          <div>
            <div className="text-lg font-semibold">180</div>
            <div className="text-xs text-muted-foreground">Following</div>
          </div>
        </div>
      </div>
    </HoverCardContent>
  </HoverCard>
);

export default Example;
