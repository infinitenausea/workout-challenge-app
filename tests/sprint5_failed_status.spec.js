const { test, expect } = require('@playwright/test');

test.describe('Sprint 5: Failed Status & UI Lock', () => {
  test('Challenge should be marked as failed and UI locked after cron job', async ({ page, request }) => {
    test.setTimeout(90000); // Allow 90 seconds for this test due to cron wait
    
    const yesterday = new Date();
    yesterday.setDate(yesterday.getDate() - 1);
    const dayBefore = new Date();
    dayBefore.setDate(dayBefore.getDate() - 2);

    const format = (d) => {
        return d.toISOString();
    };

    // We must use default_user_1 because the frontend hardcodes it in MVP
    console.log('Creating expired challenge for default_user_1...');
    
    // First, make sure exercise 1 exists or just create a custom one
    const exResp = await request.post('http://localhost:8080/api/exercises', {
      headers: { 'X-User-Id': 'default_user_1' },
      data: { name: 'QA Test Exercise' }
    });
    
    const challengesResp = await request.post('http://localhost:8080/api/challenges', {
      headers: { 'X-User-Id': 'default_user_1' },
      data: {
        name: 'QA Expired Challenge ' + Date.now(),
        exercise_id: 1, // assuming 1 exists, but if not we can just pass any valid ID
        target_value: 100,
        start_date: format(dayBefore),
        end_date: format(yesterday)
      }
    });

    expect(challengesResp.status()).toBe(201);
    
    console.log('Challenge created. Waiting 65s for the 1-minute cron job to run...');
    await page.waitForTimeout(65000);

    console.log('Navigating to dashboard...');
    await page.goto('http://localhost:8080');

    // Wait for dashboard challenges to load
    await page.waitForSelector('.challenge-card');

    // Find the challenge we just created (should be the first one since it's ordered by start_date DESC or we can just look for "Провален")
    const card = page.locator('.challenge-card', { hasText: 'QA Expired Challenge' }).first();
    await expect(card).toBeVisible();

    // The status text should say "Провален"
    const statusBadge = card.locator('.challenge-status');
    await expect(statusBadge).toHaveText('Провален');
    await expect(statusBadge).toHaveClass(/badge-failed/);

    // Go to detail page
    await card.click();
    await page.waitForSelector('.challenge-detail');

    const detailStatusBadge = page.locator('.challenge-detail-header .challenge-status');
    await expect(detailStatusBadge).toHaveText('Провален');
    await expect(detailStatusBadge).toHaveClass(/badge-failed/);

    // Check if the add workout button is disabled
    const addBtn = page.locator('#add-workout-btn');
    await expect(addBtn).toBeDisabled();
    
    console.log('All checks passed successfully!');
  });
});
