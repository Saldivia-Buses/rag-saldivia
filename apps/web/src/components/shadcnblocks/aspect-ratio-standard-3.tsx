import { AspectRatio } from "@/components/shadcnblocks/aspect-ratio";

export const title = "1:1 Aspect Ratio (Square)";

const Example = () => (
  <div className="w-full max-w-md">
    <AspectRatio ratio={1}>
      <div className="flex size-full items-center justify-center rounded-md bg-muted">
        <span className="text-sm text-muted-foreground">1:1</span>
      </div>
    </AspectRatio>
  </div>
);

export default Example;
