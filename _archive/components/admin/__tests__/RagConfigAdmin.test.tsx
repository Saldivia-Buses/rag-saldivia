import { test, expect, describe, afterEach, mock } from "bun:test"
import { render, cleanup } from "@testing-library/react"
import { RagConfigAdmin } from "@/components/admin/RagConfigAdmin"

afterEach(cleanup)

mock.module("@/app/actions/config", () => ({
  actionUpdateRagParams: mock(() => Promise.resolve()),
  actionResetRagParams: mock(() => Promise.resolve()),
}))

const defaultParams = {
  temperature: 0.7,
  top_p: 0.9,
  max_tokens: 1024,
  vdb_top_k: 10,
  reranker_top_k: 5,
  use_reranker: true,
  use_guardrails: false,
}

describe("<RagConfigAdmin />", () => {
  test("renderiza todos los sliders", () => {
    const { getByText } = render(<RagConfigAdmin params={defaultParams} />)
    expect(getByText("Temperature")).toBeInTheDocument()
    expect(getByText("Top P")).toBeInTheDocument()
    expect(getByText("Max tokens")).toBeInTheDocument()
    expect(getByText("VDB Top-K")).toBeInTheDocument()
  })

  test("muestra los valores actuales de los parámetros", () => {
    const { getByText } = render(<RagConfigAdmin params={defaultParams} />)
    expect(getByText("0.7")).toBeInTheDocument()
  })

  test("botón Guardar cambios presente", () => {
    const { getByRole } = render(<RagConfigAdmin params={defaultParams} />)
    expect(getByRole("button", { name: /Guardar/ })).toBeInTheDocument()
  })

  test("botón Resetear presente", () => {
    const { getByRole } = render(<RagConfigAdmin params={defaultParams} />)
    expect(getByRole("button", { name: /Resetear/ })).toBeInTheDocument()
  })

  test("renderiza toggle de Reranker", () => {
    const { getByText } = render(<RagConfigAdmin params={defaultParams} />)
    expect(getByText("Reranker")).toBeInTheDocument()
  })

  test("renderiza toggle de Guardrails", () => {
    const { getByText } = render(<RagConfigAdmin params={defaultParams} />)
    expect(getByText("Guardrails")).toBeInTheDocument()
  })
})
