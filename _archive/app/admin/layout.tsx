export default function AdminLayout({ children }: { children: React.ReactNode }) {
  return (
    <div data-density="compact" className="h-full">
      {children}
    </div>
  )
}
