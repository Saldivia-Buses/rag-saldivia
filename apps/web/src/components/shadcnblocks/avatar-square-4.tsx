import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

export const title = "Avatar with Online Indicator";

const Example = () => (
  <div className="relative w-fit">
    <Avatar className="rounded-lg">
      <AvatarImage
        alt="@haydenbleasel"
        src="https://github.com/haydenbleasel.png"
      />
      <AvatarFallback className="rounded-lg">HB</AvatarFallback>
    </Avatar>
    <span className="absolute -right-1 -bottom-1 size-3 rounded-full border-2 border-background bg-green-500" />
  </div>
);

export default Example;
