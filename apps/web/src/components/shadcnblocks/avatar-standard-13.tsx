import { Avatar, AvatarFallback } from "@/components/ui/avatar";

export const title = "Avatar with Gradient Fallback";

const Example = () => (
  <Avatar>
    <AvatarFallback className="bg-gradient-to-br from-purple-500 to-pink-500 text-white">
      HB
    </AvatarFallback>
  </Avatar>
);

export default Example;
