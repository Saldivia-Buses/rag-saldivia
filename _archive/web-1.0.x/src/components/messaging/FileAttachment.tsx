/**
 * FileAttachment — renders file preview in messages.
 * Images show inline, PDFs/others show as download link.
 */
"use client"

import { FileText, Download } from "lucide-react"

type FileMetadata = {
  fileName: string
  fileSize: number
  mimeType: string
  url: string
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1048576) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1048576).toFixed(1)} MB`
}

function isImage(mimeType: string) {
  return mimeType.startsWith("image/")
}

export function FileAttachment({ file }: { file: FileMetadata }) {
  if (isImage(file.mimeType)) {
    return (
      <div className="mt-2 max-w-sm">
        <a href={file.url} target="_blank" rel="noopener noreferrer">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={file.url}
            alt={file.fileName}
            className="rounded-lg border border-border max-h-64 object-contain"
          />
        </a>
        <p className="text-xs text-fg-subtle mt-1">{file.fileName} · {formatSize(file.fileSize)}</p>
      </div>
    )
  }

  const isPdf = file.mimeType === "application/pdf"
  const Icon = isPdf ? FileText : Download

  return (
    <a
      href={file.url}
      target="_blank"
      rel="noopener noreferrer"
      className="mt-2 inline-flex items-center gap-2 px-3 py-2 rounded-lg border border-border bg-surface hover:bg-surface-2 transition-colors max-w-sm"
    >
      <Icon className="h-5 w-5 text-accent shrink-0" />
      <div className="min-w-0 flex-1">
        <p className="text-sm text-fg font-medium truncate">{file.fileName}</p>
        <p className="text-xs text-fg-subtle">{formatSize(file.fileSize)}</p>
      </div>
    </a>
  )
}
