"use client"

import dynamic from "next/dynamic"

const ReactScanDynamic = dynamic(
  () => import("./ReactScan"),
  { ssr: false }
)

export function ReactScanProvider() {
  if (process.env.NODE_ENV !== "development") return null
  return <ReactScanDynamic />
}
