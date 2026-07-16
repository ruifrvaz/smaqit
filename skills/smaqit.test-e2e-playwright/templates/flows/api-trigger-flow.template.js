// Generated: {{id}} — {{title}}
// smaqit.test-e2e-playwright: api-trigger-flow
// Re-generate with: node [SMAQIT_SKILLS_DIR]/smaqit.test-e2e-playwright/scripts/generate-e2e.js
import { test, expect } from '@playwright/test';

test('{{title}}', { timeout: {{timeout_ms}} }, async ({ page }) => {
  await page.goto('/');
  await page.locator('{{username_input_selector}}').fill('{{username}}');
  await page.locator('{{password_input_selector}}').fill('{{password}}');
  await page.locator('{{submit_selector}}').click();
  await expect(page.locator('{{username_input_selector}}')).toBeHidden({ timeout: 10000 });

  // Navigate to target view
  await page.getByText('{{nav_text}}').click();

  // Optional: additional interaction before trigger (set pre_click_selector: __none__ to skip)
  await page.locator('{{pre_click_selector}}').first().click();

  // Trigger async API call
  await page.getByText('{{trigger_text}}').click();

  // Confirm loading state appeared (proves the async call started)
  await expect(page.getByText('{{loading_text}}')).toBeVisible({ timeout: 5000 });

  // Wait for loading to complete
  await expect(page.getByText('{{loading_text}}')).toBeHidden({ timeout: {{timeout_ms}} });
});
