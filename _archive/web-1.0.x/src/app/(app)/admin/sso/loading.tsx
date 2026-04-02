import { SkeletonTable } from "@/components/ui/skeleton"

export default function AdminSsoLoading() {
  return (
    <div style={{ padding: "0" }}>
      <SkeletonTable rows={3} cols={4} />
    </div>
  )
}
