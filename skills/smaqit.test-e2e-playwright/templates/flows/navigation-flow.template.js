// Generated: {{id}} — {{title}}
// smaqit.test-e2e-playwright: navigation-flow
// Re-generate with: node [SMAQIT_SKILLS_DIR]/smaqit.test-e2e-playwright/scripts/generate-e2e.js
import { test, expect } from '@playwright/test';

test('{{title}}', async ({ page }) => {
  await page.goto('/');
  await page.locator('{{username_input_selector}}').fill('{{username}}');
  await page.locator('{{password_input_selector}}').fill('{{password}}');
  await page.locator('{{submit_selector}}').click();
  await expect(page.locator('{{username_input_selector}}')).toBeHidden({ timeout: 10000 });

  // Navigate to target view
  await page.getByText('{{nav_text}}').click();

  // Optional: additional interaction before assertion (set pre_click_selector: __none__ to skip)
  await page.locator('{{pre_click_selector}}').first().click();

  await expect(page.locator('{{target_selector}}')).toBeVisible({ timeout: {{timeout_ms}} });
});
