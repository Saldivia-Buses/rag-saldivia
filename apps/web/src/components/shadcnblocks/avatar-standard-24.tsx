import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

export const title = "Avatar with Ring Animation";

const Example = () => (
  <div className="relative w-fit">
    <Avatar className="animate-pulse ring-2 ring-green-500 ring-offset-2 ring-offset-background">
      <AvatarImage
        alt="@haydenbleasel"
        src="https://github.com/haydenbleasel.png"
      />
      <AvatarFallback>HB</AvatarFallback>
    </Avatar>
    <span className="absolute -right-1 -bottom-1 size-3 rounded-full border-2 border-background bg-green-500" />
  </div>
);

export default Example;
