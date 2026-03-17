/**
 * SaldiviaSection — Settings panel for Saldivia-specific controls.
 * Controls: Query Mode, Max Sub-queries, Follow-up Retries, Show Decomposition.
 */

import React from "react";
import { Stack, Text, Select, Input, Switch } from "@kui/react";
import { useSettingsStore } from "../../store/useSettingsStore";

export const SaldiviaSection: React.FC = () => {
  const {
    crossdocQueryMode,
    crossdocMaxSubQueries,
    crossdocSynthesisModel,
    crossdocFollowUpRetries,
    crossdocShowDecomposition,
    set,
  } = useSettingsStore();

  return (
    <Stack gap="4">
      {/* Query Mode */}
      <Stack gap="1">
        <Text kind="label/bold/sm">Query Mode</Text>
        <Text kind="body/regular/xs" style={{ color: "var(--text-color-subtle)" }}>
          Standard uses the default RAG pipeline. Crossdoc decomposes questions into many
          sub-queries for better cross-document coverage.
        </Text>
        <Select
          value={crossdocQueryMode ?? "standard"}
          onChange={(e: React.ChangeEvent<HTMLSelectElement>) =>
            set({ crossdocQueryMode: e.target.value as "standard" | "crossdoc" })
          }
        >
          <option value="standard">Standard</option>
          <option value="crossdoc">Crossdoc (multi-document)</option>
        </Select>
      </Stack>

      {/* Max Sub-queries (only visible in crossdoc mode) */}
      {(crossdocQueryMode ?? "standard") === "crossdoc" && (
        <>
          <Stack gap="1">
            <Text kind="label/bold/sm">Max Sub-queries</Text>
            <Text kind="body/regular/xs" style={{ color: "var(--text-color-subtle)" }}>
              0 = unlimited. Higher values give better coverage but increase latency.
            </Text>
            <Input
              type="number"
              min={0}
              max={100}
              value={crossdocMaxSubQueries ?? 0}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                set({ crossdocMaxSubQueries: parseInt(e.target.value) || 0 })
              }
            />
          </Stack>

          <Stack gap="1">
            <Text kind="label/bold/sm">Synthesis Model</Text>
            <Text kind="body/regular/xs" style={{ color: "var(--text-color-subtle)" }}>
              Model for final answer synthesis. Leave empty to use the default LLM.
            </Text>
            <Input
              type="text"
              placeholder="(same as default LLM)"
              value={crossdocSynthesisModel ?? ""}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                set({ crossdocSynthesisModel: e.target.value || undefined })
              }
            />
          </Stack>

          <Stack gap="1" direction="row" align="center" justify="space-between">
            <Stack gap="0">
              <Text kind="label/bold/sm">Follow-up Retries</Text>
              <Text kind="body/regular/xs" style={{ color: "var(--text-color-subtle)" }}>
                Retry failed sub-queries with synonyms and broader terms.
              </Text>
            </Stack>
            <Switch
              checked={crossdocFollowUpRetries ?? true}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                set({ crossdocFollowUpRetries: e.target.checked })
              }
            />
          </Stack>

          <Stack gap="1" direction="row" align="center" justify="space-between">
            <Stack gap="0">
              <Text kind="label/bold/sm">Show Decomposition</Text>
              <Text kind="body/regular/xs" style={{ color: "var(--text-color-subtle)" }}>
                Display sub-queries and sources in the chat (debug mode).
              </Text>
            </Stack>
            <Switch
              checked={crossdocShowDecomposition ?? false}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                set({ crossdocShowDecomposition: e.target.checked })
              }
            />
          </Stack>
        </>
      )}
    </Stack>
  );
};
