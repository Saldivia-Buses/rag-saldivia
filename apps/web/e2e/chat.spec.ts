import { test, expect } from "@playwright/test";
import { login } from "./helpers/auth";

test.describe("Chat", () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test("can navigate to chat and see welcome message", async ({ page }) => {
    await page.goto("/chat");
    await page.waitForURL(/\/chat/, { timeout: 10_000 });

    // The empty state shows "En que puedo ayudarte?"
    await expect(
      page.locator("text=En que puedo ayudarte"),
    ).toBeVisible({ timeout: 10_000 });
  });

  test("chat page shows sidebar with session list", async ({ page }) => {
    await page.goto("/chat");

    // Sidebar header "Chats" and "Nuevo chat" button
    await expect(page.locator("text=Chats").first()).toBeVisible({
      timeout: 10_000,
    });
    await expect(page.locator("text=Nuevo chat")).toBeVisible();
  });

  test("can type a message in the chat input", async ({ page }) => {
    await page.goto("/chat");

    // Wait for the input area to be ready
    const textarea = page.locator("textarea");
    await expect(textarea).toBeVisible({ timeout: 10_000 });

    // Type a message
    await textarea.fill("Hola, esto es un test");
    await expect(textarea).toHaveValue("Hola, esto es un test");
  });

  test("can create a new chat session via button", async ({ page }) => {
    await page.goto("/chat");

    // Click "Nuevo chat" button in sidebar
    const newChatButton = page.locator("text=Nuevo chat");
    await expect(newChatButton).toBeVisible({ timeout: 10_000 });
    await newChatButton.click();

    // After creating a session, the URL should have ?s= parameter
    await page.waitForURL(/\/chat\?s=/, { timeout: 10_000 });
  });

  test("can send a message and see it appear", async ({ page }) => {
    await page.goto("/chat");

    const textarea = page.locator("textarea");
    await expect(textarea).toBeVisible({ timeout: 10_000 });

    // Type and submit a message
    await textarea.fill("Hola, esto es un test de Playwright");
    await textarea.press("Enter");

    // The user message should appear in the chat content area (not sidebar).
    // The message also shows as session title in sidebar, so scope to the
    // chat area which contains the user bubble with class "bg-muted".
    await expect(
      page.locator(".bg-muted >> text=Hola, esto es un test de Playwright").first(),
    ).toBeVisible({ timeout: 15_000 });
  });
});
