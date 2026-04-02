// Auth layout — completely isolated from the main app.
// No sidebar, no nav, no system hints. Just the auth form.

export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <>{children}</>;
}
