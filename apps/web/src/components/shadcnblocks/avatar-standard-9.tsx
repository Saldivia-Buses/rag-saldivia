import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

export const title = "Avatar Large";

const Example = () => (
  <Avatar className="size-12">
    <AvatarImage
      alt="@haydenbleasel"
      src="https://github.com/haydenbleasel.png"
    />
    <AvatarFallback className="text-base">HB</AvatarFallback>
  </Avatar>
);

export default Example;
