import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

export const title = "Avatar Small";

const Example = () => (
  <Avatar className="size-6">
    <AvatarImage
      alt="@haydenbleasel"
      src="https://github.com/haydenbleasel.png"
    />
    <AvatarFallback className="text-xs">HB</AvatarFallback>
  </Avatar>
);

export default Example;
