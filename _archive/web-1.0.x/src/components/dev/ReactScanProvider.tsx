"use client"

import dynamic from "next/dynamic"

const ReactScanDynamic = dynamic(
  () => import("./ReactScan"),
  { ssr: false }
)

export function ReactScanProvider() {
  if (process.env.NODE_ENV !== "development") return null
  if (process.env.NEXT_PUBLIC_DISABLE_REACT_SCAN === "true") return null
  return <ReactScanDynamic />
}
