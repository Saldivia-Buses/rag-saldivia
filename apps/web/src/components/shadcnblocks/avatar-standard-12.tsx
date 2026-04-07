import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

export const title = "Avatar with Shadow";

const Example = () => (
  <Avatar className="shadow-lg">
    <AvatarImage
      alt="@haydenbleasel"
      src="https://github.com/haydenbleasel.png"
    />
    <AvatarFallback>HB</AvatarFallback>
  </Avatar>
);

export default Example;
