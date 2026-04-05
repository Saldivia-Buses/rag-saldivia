/**
 * Global hotkeys for the messaging system.
 * - Cmd+K: open command palette
 * - Cmd+N: new channel
 * - Escape: close thread/panel
 *
 * Adapted from _archive/hooks/useGlobalHotkeys.ts.
 */
"use client"

import { useHotkeys } from "react-hotkeys-hook"

type Options = {
  onOpenPalette?: () => void
  onNewChannel?: () => void
}

export function useGlobalHotkeys({ onOpenPalette, onNewChannel }: Options = {}) {
  useHotkeys(
    "mod+k",
    (e) => {
      e.preventDefault()
      onOpenPalette?.()
    },
    { enableOnFormTags: true }
  )

  useHotkeys(
    "mod+n",
    (e) => {
      e.preventDefault()
      onNewChannel?.()
    },
    { enableOnFormTags: false }
  )
}
