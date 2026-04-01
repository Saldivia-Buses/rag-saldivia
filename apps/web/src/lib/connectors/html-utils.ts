/**
 * HTML to text conversion — shared by Confluence and Web Crawler connectors.
 *
 * Strips HTML tags while preserving basic structure (paragraphs, headings,
 * line breaks). No external dependencies.
 */

/** Strip HTML tags and convert to readable plain text. */
export function stripHtml(html: string): string {
  return html
    // Block elements → newlines
    .replace(/<\/?(p|div|br|h[1-6]|li|tr|blockquote|pre|hr)[^>]*>/gi, "\n")
    // Remove all remaining tags
    .replace(/<[^>]+>/g, "")
    // Decode common HTML entities
    .replace(/&amp;/g, "&")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'")
    .replace(/&nbsp;/g, " ")
    // Collapse multiple newlines
    .replace(/\n{3,}/g, "\n\n")
    // Trim
    .trim()
}
