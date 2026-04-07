import Link from "next/link";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

interface CardStandard4Props {
  title?: string;
  description?: string;
  href?: string;
}

export default function CardStandard4({
  title = "Card",
  description = "Description",
  href = "#",
}: CardStandard4Props) {
  return (
    <Link href={href} prefetch>
    <Card size="sm" className="w-56 overflow-hidden cursor-pointer transition-all hover:ring-primary/50 hover:scale-[1.02]">
      <CardHeader>
        <CardTitle className="!text-lg font-semibold">{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
    </Card>
    </Link>
  );
}
