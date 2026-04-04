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
  title = "3-Bedroom House",
  description = "A luxurious 3-bedroom house with a modern design.",
  href = "#",
}: CardStandard4Props) {
  return (
    <a href={href}>
    <Card size="sm" className="w-56 overflow-hidden cursor-pointer transition-all hover:ring-primary/50 hover:scale-[1.02]">
      <CardHeader>
        <CardTitle className="!text-lg font-semibold">{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent className="p-0">
        <img
          alt=""
          height={1380}
          src="https://images.unsplash.com/photo-1570129477492-45c003edd2be?q=80&w=2070&auto=format&fit=crop&ixlib=rb-4.1.0&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D"
          width={2070}
        />
      </CardContent>
    </Card>
    </a>
  );
}
