// Generated: {{id}} — {{title}}
// smaqit.test-e2e-playwright: login-flow
// Re-generate with: node [SMAQIT_SKILLS_DIR]/smaqit.test-e2e-playwright/scripts/generate-e2e.js
import { test, expect } from '@playwright/test';

const USERS = '{{usernames}}'.split(',').map(u => u.trim()).filter(Boolean);
const PASSWORD = '{{password}}';

for (const username of USERS) {
  test(`login: ${username}`, async ({ page }) => {
    await page.goto('/');
    await page.locator('{{username_input_selector}}').fill(username);
    await page.locator('{{password_input_selector}}').fill(PASSWORD);
    await page.locator('{{submit_selector}}').click();
    // Login succeeded when the login form unmounts
    await expect(page.locator('{{username_input_selector}}')).toBeHidden({ timeout: {{timeout_ms}} });
  });
}
