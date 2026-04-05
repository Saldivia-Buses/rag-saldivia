export default function SettingsLoading() {
  return (
    <div className="max-w-2xl mx-auto space-y-8 p-6">
      {/* Section: Profile */}
      <div className="space-y-4">
        <div className="h-5 w-20 rounded bg-surface-2 animate-pulse" />
        <div className="space-y-3">
          <div className="h-10 w-full rounded-md bg-surface-2 animate-pulse" />
          <div className="h-10 w-full rounded-md bg-surface-2 animate-pulse" />
        </div>
      </div>
      {/* Section: Password */}
      <div className="space-y-4">
        <div className="h-5 w-28 rounded bg-surface-2 animate-pulse" />
        <div className="space-y-3">
          <div className="h-10 w-full rounded-md bg-surface-2 animate-pulse" />
          <div className="h-10 w-full rounded-md bg-surface-2 animate-pulse" />
        </div>
      </div>
      {/* Section: Preferences */}
      <div className="space-y-4">
        <div className="h-5 w-32 rounded bg-surface-2 animate-pulse" />
        <div className="h-10 w-full rounded-md bg-surface-2 animate-pulse" />
      </div>
    </div>
  )
}
