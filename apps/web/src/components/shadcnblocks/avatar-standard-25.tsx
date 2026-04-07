import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

export const title = "Avatar with Custom Colors";

const Example = () => (
  <Avatar>
    <AvatarImage
      alt="@haydenbleasel"
      src="https://github.com/haydenbleasel.png"
    />
    <AvatarFallback className="bg-blue-500 text-white">HB</AvatarFallback>
  </Avatar>
);

export default Example;
