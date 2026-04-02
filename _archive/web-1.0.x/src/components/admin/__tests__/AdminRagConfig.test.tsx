import { test, expect, describe, afterEach, mock } from "bun:test"
import { render, cleanup } from "@testing-library/react"
import { AdminRagConfig } from "@/components/admin/AdminRagConfig"

afterEach(cleanup)

mock.module("@/app/actions/config", () => ({
  actionUpdateRagParams: mock(() => Promise.resolve()),
  actionResetRagParams: mock(() => Promise.resolve()),
}))

const defaultParams = {
  temperature: 0.2,
  top_p: 0.7,
  max_tokens: 1024,
  vdb_top_k: 10,
  reranker_top_k: 5,
  use_guardrails: false,
  use_reranker: true,
  chunk_size: 512,
  chunk_overlap: 50,
  embedding_model: "nvidia/nv-embedqa-e5-v5",
}

describe("<AdminRagConfig />", () => {
  test("renders all slider labels", () => {
    const { getByText } = render(
      <AdminRagConfig params={defaultParams} defaults={defaultParams} />
    )
    expect(getByText("Temperature")).toBeInTheDocument()
    expect(getByText("Top P")).toBeInTheDocument()
    expect(getByText("Max Tokens")).toBeInTheDocument()
    expect(getByText("VDB Top-K")).toBeInTheDocument()
    expect(getByText("Reranker Top-K")).toBeInTheDocument()
    expect(getByText("Chunk Size")).toBeInTheDocument()
    expect(getByText("Chunk Overlap")).toBeInTheDocument()
  })

  test("shows current values for params", () => {
    const { getByText } = render(
      <AdminRagConfig params={defaultParams} defaults={defaultParams} />
    )
    expect(getByText("0.2")).toBeInTheDocument()
    expect(getByText("0.7")).toBeInTheDocument()
    expect(getByText("1024")).toBeInTheDocument()
    expect(getByText("10")).toBeInTheDocument()
  })

  test("save button exists", () => {
    const { getByRole } = render(
      <AdminRagConfig params={defaultParams} defaults={defaultParams} />
    )
    expect(getByRole("button", { name: /Guardar cambios/ })).toBeInTheDocument()
  })
})
