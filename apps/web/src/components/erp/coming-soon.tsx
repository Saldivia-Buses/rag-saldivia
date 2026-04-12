import { Construction } from "lucide-react";

export function ComingSoon({ title, description }: { title: string; description: string }) {
  return (
    <div className="flex-1 flex flex-col items-center justify-center py-24 text-center">
      <Construction className="h-16 w-16 text-muted-foreground/30 mb-6" />
      <h2 className="text-lg font-semibold mb-2">{title}</h2>
      <p className="text-sm text-muted-foreground max-w-md">{description}</p>
    </div>
  );
}
