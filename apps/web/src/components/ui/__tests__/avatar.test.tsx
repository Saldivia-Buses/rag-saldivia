import { test, expect, describe, afterEach } from "bun:test"
import { render, cleanup } from "@testing-library/react"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"

afterEach(cleanup)

describe("<Avatar />", () => {
  test("renderiza AvatarFallback con iniciales", () => {
    const { getByText } = render(<Avatar><AvatarFallback>EA</AvatarFallback></Avatar>)
    expect(getByText("EA")).toBeInTheDocument()
  })

  test("AvatarFallback tiene clases de estilo accent", () => {
    const { getByText } = render(<Avatar><AvatarFallback>RS</AvatarFallback></Avatar>)
    expect(getByText("RS").className).toContain("accent")
  })

  test("Avatar tiene rounded-full", () => {
    const { container } = render(<Avatar><AvatarFallback>JD</AvatarFallback></Avatar>)
    expect(container.firstChild?.className ?? "").toContain("rounded-full")
  })

  test("renderiza con texto de 2 caracteres", () => {
    const { getByText } = render(<Avatar><AvatarFallback>AB</AvatarFallback></Avatar>)
    expect(getByText("AB")).toHaveTextContent("AB")
  })
})
