import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

export const title = "Avatar Extra Large";

const Example = () => (
  <Avatar className="size-20">
    <AvatarImage
      alt="@haydenbleasel"
      src="https://github.com/haydenbleasel.png"
    />
    <AvatarFallback className="text-xl">HB</AvatarFallback>
  </Avatar>
);

export default Example;
