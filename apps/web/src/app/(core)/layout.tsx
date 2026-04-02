import { Sidebar6 } from "@/components/sidebar6";

export default function CoreLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <Sidebar6>{children}</Sidebar6>;
}
