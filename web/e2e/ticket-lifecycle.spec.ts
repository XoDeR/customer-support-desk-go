import { expect, test } from "@playwright/test";

const apiURL = process.env.E2E_API_URL ?? "http://localhost:8080/api/v1";
const agentEmail = process.env.E2E_AGENT_EMAIL ?? "agent@example.com";
const agentPassword = process.env.E2E_AGENT_PASSWORD ?? "agent-password-change-me";

test.describe("ticket lifecycle smoke", () => {
  test("customer creates ticket; agent assigns and replies publicly", async ({ page, request }) => {
    const health = await request.get(`${apiURL.replace(/\/api\/v1$/, "")}/health`);
    test.skip(!health.ok(), "API is not running — start make run-api first");

    const stamp = Date.now();
    const customerEmail = `customer.e2e.${stamp}@example.com`;
    const password = "customer-e2e-password";
    const subject = `E2E password reset ${stamp}`;

    await page.goto("/register");
    await page.getByLabel("Email").fill(customerEmail);
    await page.getByLabel("Password").fill(password);
    await page.getByRole("button", { name: /create account/i }).click();
    await expect(page.getByText(/account created/i)).toBeVisible();

    await page.goto("/login");
    await page.getByLabel("Email").fill(customerEmail);
    await page.getByLabel("Password").fill(password);
    await page.getByRole("button", { name: /sign in/i }).click();
    await expect(page.getByRole("heading", { name: /my tickets/i })).toBeVisible();

    await page.getByRole("link", { name: /new ticket/i }).first().click();
    await page.getByLabel("Subject").fill(subject);
    await page.getByLabel("How can we help?").fill("Playwright lifecycle smoke test body.");
    await page.getByRole("button", { name: /submit request/i }).click();
    await expect(page.getByRole("heading", { name: subject })).toBeVisible();

    await page.getByRole("button", { name: /sign out/i }).click();
    await page.getByLabel("Email").fill(agentEmail);
    await page.getByLabel("Password").fill(agentPassword);
    await page.getByRole("button", { name: /sign in/i }).click();
    await expect(page.getByRole("heading", { name: /ticket queue/i })).toBeVisible();

    await page.getByRole("link", { name: subject }).click();
    await expect(page.getByRole("heading", { name: subject })).toBeVisible();

    await page.getByLabel("Status").selectOption("pending");
    await page.getByPlaceholder("Write a reply…").fill("Thanks — we are investigating.");
    await page.getByRole("button", { name: /send reply/i }).click();
    await expect(page.getByText("Thanks — we are investigating.")).toBeVisible();
  });
});
