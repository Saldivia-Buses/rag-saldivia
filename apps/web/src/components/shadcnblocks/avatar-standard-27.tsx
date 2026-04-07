import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

export const title = "Avatar with Error State";

const Example = () => (
  <Avatar>
    <AvatarImage
      alt="@haydenbleasel"
      src="https://invalid-url-that-will-fail.png"
    />
    <AvatarFallback className="bg-destructive/10 text-destructive">
      HB
    </AvatarFallback>
  </Avatar>
);

export default Example;
