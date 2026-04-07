import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

export const title = "Avatar with Colored Border";

const Example = () => (
  <Avatar className="border-2 border-blue-500">
    <AvatarImage
      alt="@haydenbleasel"
      src="https://github.com/haydenbleasel.png"
    />
    <AvatarFallback>HB</AvatarFallback>
  </Avatar>
);

export default Example;
