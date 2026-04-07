import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

export const title = "Avatar with Tooltip";

const Example = () => (
  <TooltipProvider>
    <Tooltip>
      <TooltipTrigger asChild>
        <Avatar className="cursor-pointer">
          <AvatarImage
            alt="@haydenbleasel"
            src="https://github.com/haydenbleasel.png"
          />
          <AvatarFallback>HB</AvatarFallback>
        </Avatar>
      </TooltipTrigger>
      <TooltipContent>
        <p>Hayden Bleasel</p>
      </TooltipContent>
    </Tooltip>
  </TooltipProvider>
);

export default Example;
