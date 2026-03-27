"use client"

import React from "react"
import { AlertTriangle } from "lucide-react"
import { Button } from "@/components/ui/button"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"

interface ErrorBoundaryProps {
  children: React.ReactNode
  fallback?: React.ReactNode
  onReset?: () => void
}

interface ErrorBoundaryState {
  hasError: boolean
  error: Error | null
}

export class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, info: React.ErrorInfo) {
    console.error("[ErrorBoundary] error capturado:", error, info.componentStack)
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null })
    this.props.onReset?.()
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback
      }

      const message =
        process.env.NODE_ENV === "production"
          ? "Ha ocurrido un error inesperado."
          : (this.state.error?.message ?? "Error desconocido.")

      return (
        <div className="flex flex-col items-center justify-center h-full gap-4 p-8">
          <EmptyPlaceholder>
            <EmptyPlaceholder.Icon icon={AlertTriangle} />
            <EmptyPlaceholder.Title>Algo salió mal</EmptyPlaceholder.Title>
            <EmptyPlaceholder.Description>{message}</EmptyPlaceholder.Description>
          </EmptyPlaceholder>
          <Button onClick={this.handleReset} variant="outline">
            Reintentar
          </Button>
        </div>
      )
    }

    return this.props.children
  }
}
