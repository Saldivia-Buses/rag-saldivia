import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

export const title = "Avatar with Badge Bottom Left";

const Example = () => (
  <div className="relative w-fit">
    <Avatar>
      <AvatarImage
        alt="@haydenbleasel"
        src="https://github.com/haydenbleasel.png"
      />
      <AvatarFallback>HB</AvatarFallback>
    </Avatar>
    <span className="absolute -bottom-2 -left-2 flex size-5 items-center justify-center rounded-full border-2 border-background bg-red-500 text-xs font-medium text-white">
      5
    </span>
  </div>
);

export default Example;
