import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

export const title = "Avatar with Hover Effect";

const Example = () => (
  <Avatar className="transition-transform hover:scale-110 hover:shadow-lg">
    <AvatarImage
      alt="@haydenbleasel"
      src="https://github.com/haydenbleasel.png"
    />
    <AvatarFallback>HB</AvatarFallback>
  </Avatar>
);

export default Example;
