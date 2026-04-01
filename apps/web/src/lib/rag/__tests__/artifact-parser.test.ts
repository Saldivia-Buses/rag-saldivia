/**
 * Tests for artifact-parser.ts — pure function tests, no DOM/mocks needed.
 * Corre con: bun test apps/web/src/lib/rag/__tests__/artifact-parser.test.ts
 */

import { describe, test, expect } from "bun:test"
import {
  extractArtifacts,
  extractStreamingArtifact,
  stripArtifactTags,
  extractCodeBlocks,
} from "../artifact-parser"

// ── extractArtifacts ──

describe("extractArtifacts", () => {
  test("parses a single code artifact with language and title", () => {
    const text = `Here is some code:
<artifact id="a1" type="code" title="Hello World" language="typescript" version="1">console.log("hello")</artifact>
Done.`

    const { cleanText, artifacts } = extractArtifacts(text)

    expect(artifacts).toHaveLength(1)
    expect(artifacts[0]!.id).toBe("a1")
    expect(artifacts[0]!.type).toBe("code")
    expect(artifacts[0]!.title).toBe("Hello World")
    expect(artifacts[0]!.content).toBe('console.log("hello")')
    expect(artifacts[0]!.language).toBe("typescript")
    expect(artifacts[0]!.version).toBe(1)
    expect(cleanText).toContain("[ARTIFACT:a1]")
    expect(cleanText).not.toContain("<artifact")
  })

  test("parses HTML artifact", () => {
    const html = `<div><h1>Title</h1></div>`
    const text = `<artifact id="h1" type="html" title="Page">${html}</artifact>`

    const { artifacts } = extractArtifacts(text)

    expect(artifacts).toHaveLength(1)
    expect(artifacts[0]!.type).toBe("html")
    expect(artifacts[0]!.content).toBe(html)
  })

  test("parses SVG artifact", () => {
    const svg = `<svg width="100" height="100"><circle cx="50" cy="50" r="40"/></svg>`
    const text = `<artifact id="s1" type="svg" title="Circle">${svg}</artifact>`

    const { artifacts } = extractArtifacts(text)

    expect(artifacts).toHaveLength(1)
    expect(artifacts[0]!.type).toBe("svg")
    expect(artifacts[0]!.content).toBe(svg)
  })

  test("parses mermaid artifact with default language and title", () => {
    const diagram = `graph TD\n  A --> B`
    const text = `<artifact id="m1" type="mermaid">${diagram}</artifact>`

    const { artifacts } = extractArtifacts(text)

    expect(artifacts).toHaveLength(1)
    expect(artifacts[0]!.type).toBe("mermaid")
    expect(artifacts[0]!.title).toBe("Diagrama")
    expect(artifacts[0]!.language).toBe("mermaid")
  })

  test("parses table artifact", () => {
    const table = `| Col1 | Col2 |\n|------|------|\n| A    | B    |`
    const text = `<artifact id="t1" type="table" title="Data">${table}</artifact>`

    const { artifacts } = extractArtifacts(text)

    expect(artifacts).toHaveLength(1)
    expect(artifacts[0]!.type).toBe("table")
    expect(artifacts[0]!.title).toBe("Data")
  })

  test("parses text artifact", () => {
    const text = `<artifact id="x1" type="text" title="Notes">Some plain text content</artifact>`

    const { artifacts } = extractArtifacts(text)

    expect(artifacts).toHaveLength(1)
    expect(artifacts[0]!.type).toBe("text")
    expect(artifacts[0]!.content).toBe("Some plain text content")
  })

  test("parses multiple artifacts in one string", () => {
    const text = `Intro text.
<artifact id="c1" type="code" language="python" title="Script">print("hi")</artifact>
Middle text.
<artifact id="c2" type="html" title="Page"><p>hello</p></artifact>
End text.`

    const { cleanText, artifacts } = extractArtifacts(text)

    expect(artifacts).toHaveLength(2)
    expect(artifacts[0]!.id).toBe("c1")
    expect(artifacts[1]!.id).toBe("c2")
    expect(cleanText).toContain("Intro text.")
    expect(cleanText).toContain("[ARTIFACT:c1]")
    expect(cleanText).toContain("Middle text.")
    expect(cleanText).toContain("[ARTIFACT:c2]")
    expect(cleanText).toContain("End text.")
  })

  test("preserves content between artifacts", () => {
    const text = `Before first.
<artifact id="a1" type="code" language="js" title="A">code1</artifact>
Between them here.
<artifact id="a2" type="code" language="js" title="B">code2</artifact>
After last.`

    const { cleanText } = extractArtifacts(text)

    expect(cleanText).toContain("Before first.")
    expect(cleanText).toContain("Between them here.")
    expect(cleanText).toContain("After last.")
  })

  test("handles nested code blocks inside artifact content", () => {
    // Content has backtick code blocks but the artifact tags delimit the real boundary
    const inner = "```js\nconsole.log('nested')\n```"
    const text = `<artifact id="n1" type="code" title="Nested" language="markdown">${inner}</artifact>`

    const { artifacts } = extractArtifacts(text)

    expect(artifacts).toHaveLength(1)
    expect(artifacts[0]!.content).toBe(inner)
  })

  test("handles empty content artifact", () => {
    const text = `<artifact id="e1" type="code" title="Empty" language="js"></artifact>`

    const { artifacts } = extractArtifacts(text)

    expect(artifacts).toHaveLength(1)
    expect(artifacts[0]!.content).toBe("")
  })

  test("malformed unclosed artifact tag yields no artifacts", () => {
    const text = `<artifact id="broken" type="code" title="Oops">this never closes`

    const { cleanText, artifacts } = extractArtifacts(text)

    expect(artifacts).toHaveLength(0)
    // Original text is returned untouched since regex doesn't match
    expect(cleanText).toBe(text)
  })

  test("no artifacts returns empty array and original text", () => {
    const text = "Just plain text with no artifacts at all."

    const { cleanText, artifacts } = extractArtifacts(text)

    expect(artifacts).toHaveLength(0)
    expect(cleanText).toBe(text)
  })

  test("missing type attribute defaults to code", () => {
    const text = `<artifact id="d1" title="No Type">some code</artifact>`

    const { artifacts } = extractArtifacts(text)

    expect(artifacts).toHaveLength(1)
    expect(artifacts[0]!.type).toBe("code")
  })

  test("missing id attribute gets auto-generated id", () => {
    const text = `<artifact type="code" title="Auto ID">stuff</artifact>`

    const { cleanText, artifacts } = extractArtifacts(text)

    expect(artifacts).toHaveLength(1)
    // Auto-generated IDs follow the pattern art_N
    expect(artifacts[0]!.id).toMatch(/^art_\d+$/)
    expect(cleanText).toContain(`[ARTIFACT:${artifacts[0]!.id}]`)
  })

  test("missing title defaults to 'Sin titulo' for non-mermaid", () => {
    const text = `<artifact id="nt1" type="code" language="js">x = 1</artifact>`

    const { artifacts } = extractArtifacts(text)

    expect(artifacts[0]!.title).toBe("Sin título")
  })

  test("missing title defaults to 'Diagrama' for mermaid type", () => {
    const text = `<artifact id="mt1" type="mermaid">graph TD\n  A-->B</artifact>`

    const { artifacts } = extractArtifacts(text)

    expect(artifacts[0]!.title).toBe("Diagrama")
  })

  test("extra/unknown attributes are silently ignored", () => {
    const text = `<artifact id="ex1" type="code" title="Extra" language="go" author="llm" priority="high" version="3">func main() {}</artifact>`

    const { artifacts } = extractArtifacts(text)

    expect(artifacts).toHaveLength(1)
    expect(artifacts[0]!.id).toBe("ex1")
    expect(artifacts[0]!.language).toBe("go")
    expect(artifacts[0]!.version).toBe(3)
    // Extra attrs don't break parsing and are not on the typed result
    expect(artifacts[0]!).not.toHaveProperty("author")
    expect(artifacts[0]!).not.toHaveProperty("priority")
  })

  test("special characters in content are preserved", () => {
    const content = `const x = a < b && c > d;\nconst s = "he said \\"hello\\"";`
    const text = `<artifact id="sp1" type="code" title="Special" language="js">${content}</artifact>`

    const { artifacts } = extractArtifacts(text)

    expect(artifacts[0]!.content).toBe(content)
  })

  test("angle brackets in artifact content are preserved", () => {
    const content = `<div class="test"><span>inner</span></div>`
    const text = `<artifact id="ang1" type="html" title="Tags">${content}</artifact>`

    const { artifacts } = extractArtifacts(text)

    expect(artifacts[0]!.content).toBe(content)
  })

  test("content with whitespace is trimmed", () => {
    const text = `<artifact id="ws1" type="code" title="Whitespace" language="js">
  const x = 1;
  const y = 2;
</artifact>`

    const { artifacts } = extractArtifacts(text)

    expect(artifacts[0]!.content).toBe("const x = 1;\n  const y = 2;")
  })

  test("version defaults to 1 when missing", () => {
    const text = `<artifact id="v0" type="code" title="No Ver">x</artifact>`

    const { artifacts } = extractArtifacts(text)

    expect(artifacts[0]!.version).toBe(1)
  })

  test("version defaults to 1 when non-numeric", () => {
    const text = `<artifact id="vn" type="code" title="Bad Ver" version="abc">x</artifact>`

    const { artifacts } = extractArtifacts(text)

    expect(artifacts[0]!.version).toBe(1)
  })

  test("version parses numeric string correctly", () => {
    const text = `<artifact id="v5" type="code" title="V5" version="5">x</artifact>`

    const { artifacts } = extractArtifacts(text)

    expect(artifacts[0]!.version).toBe(5)
  })

  test("language is undefined for non-mermaid types without language attr", () => {
    const text = `<artifact id="nl1" type="html" title="No Lang"><p>hi</p></artifact>`

    const { artifacts } = extractArtifacts(text)

    expect(artifacts[0]!.language).toBeUndefined()
  })

  test("very long content is handled", () => {
    const longContent = "x".repeat(100_000)
    const text = `<artifact id="long1" type="text" title="Long">${longContent}</artifact>`

    const { artifacts } = extractArtifacts(text)

    expect(artifacts).toHaveLength(1)
    expect(artifacts[0]!.content).toHaveLength(100_000)
  })

  test("multiline content with various newline patterns", () => {
    const content = "line1\nline2\nline3\n\nline5"
    const text = `<artifact id="ml1" type="code" title="Multiline" language="text">${content}</artifact>`

    const { artifacts } = extractArtifacts(text)

    expect(artifacts[0]!.content).toBe(content)
  })

  test("placeholder format is correct with newline padding", () => {
    const text = `before<artifact id="fmt1" type="code" title="X">y</artifact>after`

    const { cleanText } = extractArtifacts(text)

    expect(cleanText).toBe("before\n\n[ARTIFACT:fmt1]\n\nafter")
  })

  test("called consecutively, regex lastIndex resets (no stale state)", () => {
    const text1 = `<artifact id="r1" type="code" title="First">a</artifact>`
    const text2 = `<artifact id="r2" type="code" title="Second">b</artifact>`

    const result1 = extractArtifacts(text1)
    const result2 = extractArtifacts(text2)

    expect(result1.artifacts).toHaveLength(1)
    expect(result1.artifacts[0]!.id).toBe("r1")
    expect(result2.artifacts).toHaveLength(1)
    expect(result2.artifacts[0]!.id).toBe("r2")
  })
})

// ── extractStreamingArtifact ──

describe("extractStreamingArtifact", () => {
  test("detects partial artifact at end of text", () => {
    const text = `Here is some text.
<artifact id="s1" type="code" title="In Progress" language="python">def hello():
  print("world")`

    const result = extractStreamingArtifact(text)

    expect(result).not.toBeNull()
    expect(result!.id).toBe("s1")
    expect(result!.type).toBe("code")
    expect(result!.title).toBe("In Progress")
    expect(result!.language).toBe("python")
    expect(result!.isStreaming).toBe(true)
    expect(result!.version).toBe(1)
    expect(result!.content).toContain('def hello():')
  })

  test("returns null when no partial artifact exists", () => {
    const text = "Just normal text, nothing special here."

    const result = extractStreamingArtifact(text)

    expect(result).toBeNull()
  })

  test("returns null when all artifacts are complete", () => {
    const text = `<artifact id="c1" type="code" title="Done">complete</artifact> Some trailing text.`

    const result = extractStreamingArtifact(text)

    expect(result).toBeNull()
  })

  test("ignores complete artifacts, detects trailing partial", () => {
    const text = `<artifact id="done1" type="code" title="Done">finished</artifact>
Middle text.
<artifact id="stream1" type="html" title="Loading"><div>still going`

    const result = extractStreamingArtifact(text)

    expect(result).not.toBeNull()
    expect(result!.id).toBe("stream1")
    expect(result!.type).toBe("html")
    expect(result!.content).toContain("<div>still going")
  })

  test("defaults to type code when type attr missing", () => {
    const text = `<artifact id="noType" title="Test">partial content`

    const result = extractStreamingArtifact(text)

    expect(result).not.toBeNull()
    expect(result!.type).toBe("code")
  })

  test("defaults id to 'streaming' when id attr missing", () => {
    const text = `<artifact type="code" title="Test">partial`

    const result = extractStreamingArtifact(text)

    expect(result).not.toBeNull()
    expect(result!.id).toBe("streaming")
  })

  test("defaults title to 'Generando...' when title attr missing", () => {
    const text = `<artifact id="gt1" type="code">some code`

    const result = extractStreamingArtifact(text)

    expect(result).not.toBeNull()
    expect(result!.title).toBe("Generando...")
  })

  test("mermaid type gets language 'mermaid' by default", () => {
    const text = `<artifact id="sm1" type="mermaid">graph TD`

    const result = extractStreamingArtifact(text)

    expect(result).not.toBeNull()
    expect(result!.language).toBe("mermaid")
  })

  test("empty content after opening tag returns empty string content", () => {
    const text = `<artifact id="empty1" type="code" title="Empty">`

    const result = extractStreamingArtifact(text)

    expect(result).not.toBeNull()
    expect(result!.content).toBe("")
  })

  test("called consecutively, regex lastIndex resets", () => {
    const text1 = `<artifact id="x1" type="code">partial1`
    const text2 = `<artifact id="x2" type="html">partial2`

    const r1 = extractStreamingArtifact(text1)
    const r2 = extractStreamingArtifact(text2)

    expect(r1).not.toBeNull()
    expect(r1!.id).toBe("x1")
    expect(r2).not.toBeNull()
    expect(r2!.id).toBe("x2")
  })
})

// ── stripArtifactTags ──

describe("stripArtifactTags", () => {
  test("strips complete artifact tags", () => {
    const text = `Before.
<artifact id="s1" type="code" title="X">code here</artifact>
After.`

    const result = stripArtifactTags(text)

    expect(result).not.toContain("<artifact")
    expect(result).not.toContain("</artifact>")
    expect(result).not.toContain("code here")
    expect(result).toContain("Before.")
    expect(result).toContain("After.")
  })

  test("strips partial artifact tag at end (streaming)", () => {
    const text = `Some text.
<artifact id="p1" type="code" title="Partial">still streaming`

    const result = stripArtifactTags(text)

    expect(result).not.toContain("<artifact")
    expect(result).not.toContain("still streaming")
    expect(result).toContain("Some text.")
  })

  test("strips both complete and partial artifacts", () => {
    const text = `Intro.
<artifact id="c1" type="code" title="Done">complete</artifact>
Middle.
<artifact id="p1" type="code">partial`

    const result = stripArtifactTags(text)

    expect(result).not.toContain("<artifact")
    expect(result).toContain("Intro.")
    expect(result).toContain("Middle.")
    expect(result).not.toContain("complete")
    expect(result).not.toContain("partial")
  })

  test("returns original text (trimmed) when no artifacts", () => {
    const text = "  Just plain text.  "

    const result = stripArtifactTags(text)

    expect(result).toBe("Just plain text.")
  })

  test("returns empty string when text is only an artifact", () => {
    const text = `<artifact id="only" type="code" title="All">everything</artifact>`

    const result = stripArtifactTags(text)

    expect(result).toBe("")
  })

  test("handles multiple complete artifacts", () => {
    const text = `A<artifact id="a1" type="code" title="1">x</artifact>B<artifact id="a2" type="code" title="2">y</artifact>C`

    const result = stripArtifactTags(text)

    expect(result).toContain("A")
    expect(result).toContain("B")
    expect(result).toContain("C")
    expect(result).not.toContain("<artifact")
  })

  test("result is trimmed", () => {
    const text = `   \n<artifact id="t1" type="code" title="X">y</artifact>\n   `

    const result = stripArtifactTags(text)

    expect(result).toBe(result.trim())
  })
})

// ── extractCodeBlocks ──

describe("extractCodeBlocks", () => {
  test("extracts code block with 3+ lines as artifact", () => {
    const text = "Some text\n```python\nline1\nline2\nline3\n```\nMore text"

    const artifacts = extractCodeBlocks(text)

    expect(artifacts).toHaveLength(1)
    expect(artifacts[0]!.type).toBe("code")
    expect(artifacts[0]!.language).toBe("python")
    expect(artifacts[0]!.title).toBe("Código python")
    expect(artifacts[0]!.content).toContain("line1")
    expect(artifacts[0]!.content).toContain("line3")
    expect(artifacts[0]!.version).toBe(1)
  })

  test("skips code blocks with fewer than 3 lines", () => {
    const text = "```js\nshort\n```"

    const artifacts = extractCodeBlocks(text)

    expect(artifacts).toHaveLength(0)
  })

  test("skips code blocks with exactly 2 lines", () => {
    const text = "```js\nline1\nline2\n```"

    const artifacts = extractCodeBlocks(text)

    expect(artifacts).toHaveLength(0)
  })

  test("extracts code block with exactly 3 lines", () => {
    const text = "```js\nline1\nline2\nline3\n```"

    const artifacts = extractCodeBlocks(text)

    expect(artifacts).toHaveLength(1)
  })

  test("mermaid code block gets type mermaid and title Diagrama", () => {
    const text = "```mermaid\ngraph TD\n  A --> B\n  B --> C\n```"

    const artifacts = extractCodeBlocks(text)

    expect(artifacts).toHaveLength(1)
    expect(artifacts[0]!.type).toBe("mermaid")
    expect(artifacts[0]!.title).toBe("Diagrama")
    expect(artifacts[0]!.language).toBe("mermaid")
  })

  test("code block without language tag defaults to text", () => {
    const text = "```\nline1\nline2\nline3\n```"

    const artifacts = extractCodeBlocks(text)

    expect(artifacts).toHaveLength(1)
    expect(artifacts[0]!.language).toBe("text")
    expect(artifacts[0]!.title).toBe("Código text")
  })

  test("extracts multiple code blocks", () => {
    const text = `Some text
\`\`\`python
a = 1
b = 2
c = 3
\`\`\`
Middle
\`\`\`rust
fn main() {
    println!("hello");
    return;
}
\`\`\`
End`

    const artifacts = extractCodeBlocks(text)

    expect(artifacts).toHaveLength(2)
    expect(artifacts[0]!.language).toBe("python")
    expect(artifacts[1]!.language).toBe("rust")
  })

  test("returns empty array when no code blocks", () => {
    const text = "Just plain text, no code blocks."

    const artifacts = extractCodeBlocks(text)

    expect(artifacts).toHaveLength(0)
  })

  test("auto-generates unique ids", () => {
    const text = `\`\`\`js
line1
line2
line3
\`\`\`
\`\`\`py
line1
line2
line3
\`\`\``

    const artifacts = extractCodeBlocks(text)

    expect(artifacts).toHaveLength(2)
    expect(artifacts[0]!.id).toMatch(/^art_\d+$/)
    expect(artifacts[1]!.id).toMatch(/^art_\d+$/)
    expect(artifacts[0]!.id).not.toBe(artifacts[1]!.id)
  })

  test("content is trimEnd'd but not trimStart'd", () => {
    const text = "```js\n  indented1\n  indented2\n  indented3\n```"

    const artifacts = extractCodeBlocks(text)

    expect(artifacts).toHaveLength(1)
    // Leading whitespace on lines is preserved, trailing whitespace is trimmed
    expect(artifacts[0]!.content.startsWith("  indented1")).toBe(true)
  })

  test("mixed qualifying and non-qualifying blocks", () => {
    const text = `\`\`\`js
short
\`\`\`
\`\`\`python
long1
long2
long3
\`\`\``

    const artifacts = extractCodeBlocks(text)

    expect(artifacts).toHaveLength(1)
    expect(artifacts[0]!.language).toBe("python")
  })
})

// ── Cross-function interaction ──

describe("cross-function interactions", () => {
  test("extractArtifacts + extractStreamingArtifact: complete artifact not detected as streaming", () => {
    const text = `<artifact id="comp1" type="code" title="Done">complete content</artifact>`

    const { artifacts } = extractArtifacts(text)
    const streaming = extractStreamingArtifact(text)

    expect(artifacts).toHaveLength(1)
    expect(streaming).toBeNull()
  })

  test("stripArtifactTags handles same text as extractArtifacts without interference", () => {
    const text = `Prose here.
<artifact id="x1" type="code" title="Test">content</artifact>
More prose.`

    const { cleanText, artifacts } = extractArtifacts(text)
    const stripped = stripArtifactTags(text)

    expect(artifacts).toHaveLength(1)
    expect(cleanText).toContain("[ARTIFACT:x1]")
    expect(stripped).not.toContain("[ARTIFACT")
    expect(stripped).toContain("Prose here.")
    expect(stripped).toContain("More prose.")
  })

  test("text with both artifact tags and code blocks", () => {
    const text = `<artifact id="a1" type="code" title="Tagged">tagged code</artifact>
\`\`\`python
line1
line2
line3
\`\`\``

    const { artifacts: tagArtifacts } = extractArtifacts(text)
    const codeBlocks = extractCodeBlocks(text)

    expect(tagArtifacts).toHaveLength(1)
    expect(tagArtifacts[0]!.id).toBe("a1")
    // extractCodeBlocks still finds the fenced block (independent parse)
    expect(codeBlocks).toHaveLength(1)
    expect(codeBlocks[0]!.language).toBe("python")
  })
})
