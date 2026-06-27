const { test, expect } = require('@playwright/test');

test.describe('QA-3: Dashboard UI & Achievements (US-5 & US-7)', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost:8080');
  });

  test('TC-3.18 & TC-3.19: Navigation from dashboard to details and back', async ({ page }) => {
    // Wait for challenge cards to load
    // If no challenges exist, we'll create one first
    const cardCount = await page.locator('.challenge-card').count();
    if (cardCount === 0) {
      await page.locator('#create-challenge-btn').click();
      await page.locator('#name').fill('Nav Test Challenge');
      await page.locator('#exercise_id').selectOption({ index: 1 });
      await page.locator('#target_value').fill('100');
      await page.locator('button[type="submit"]').click();
      await page.waitForSelector('.challenge-card');
    }

    const firstCard = page.locator('.challenge-card').first();
    const challengeTitle = await firstCard.locator('.challenge-title').innerText();

    // Click the card to navigate to details page
    await firstCard.click();
    await expect(page.locator('.challenge-detail')).toBeVisible();

    // Verify detail page contains challenge title
    await expect(page.locator('.challenge-detail h2')).toHaveText(challengeTitle);

    // Click back button
    await page.locator('#back-btn').click();

    // Verify we are back on the dashboard
    await expect(page.locator('.dashboard-header-row h2')).toHaveText('Активные челленджи');
    await expect(page.locator('.challenge-card').first()).toBeVisible();
  });
});
