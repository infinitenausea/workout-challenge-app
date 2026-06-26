const { test, expect } = require('@playwright/test');

test.describe('QA-2: Exercise and Challenge Creation (US-1 & US-2)', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to homepage before each test
    await page.goto('http://localhost:8080');
  });

  test('TC-1.4: UI toggles for custom exercises', async ({ page }) => {
    // Click create challenge button
    await page.locator('#create-challenge-btn').click();
    await expect(page.locator('#challenge-form')).toBeVisible();

    const select = page.locator('#exercise_id');
    const customGroup = page.locator('#custom-exercise-group');

    // Initially custom exercise input group should be hidden
    await expect(customGroup).not.toBeVisible();

    // Select custom option
    await select.selectOption('custom');
    await expect(customGroup).toBeVisible();

    // Select predefined option
    await select.selectOption({ index: 1 });
    await expect(customGroup).not.toBeVisible();
  });

  test('TC-1.6: UI validation for empty exercise name when selecting custom', async ({ page }) => {
    await page.locator('#create-challenge-btn').click();
    await page.locator('#name').fill('Empty Exercise Challenge');
    await page.locator('#exercise_id').selectOption('custom');
    
    // Verify HTML5 required attribute is set
    const customInput = page.locator('#custom_exercise_name');
    await expect(customInput).toHaveAttribute('required', 'required');

    // Clear custom exercise input
    await customInput.fill('');
    await page.locator('#target_value').fill('100');

    // Remove required attribute to test fallback JS validation
    await page.evaluate(() => {
      document.getElementById('custom_exercise_name').removeAttribute('required');
    });

    // Click submit
    await page.locator('button[type="submit"]').click();

    // Should display toast error
    const toast = page.locator('.toast.error');
    await expect(toast).toBeVisible();
    await expect(toast).toContainText('Введите название нового упражнения');
  });

  test('TC-2.5: UI form validation/submit blocking for invalid dates', async ({ page }) => {
    await page.locator('#create-challenge-btn').click();
    await page.locator('#name').fill('Invalid Date Challenge');
    await page.locator('#exercise_id').selectOption({ index: 1 });
    await page.locator('#target_value').fill('100');

    // Start date in future, end date today (earlier)
    const today = new Date();
    const futureDate = new Date();
    futureDate.setDate(today.getDate() + 5);

    const formatDate = (d) => d.toISOString().split('T')[0];

    await page.locator('#start_date').fill(formatDate(futureDate));
    await page.locator('#end_date').fill(formatDate(today));

    await page.locator('button[type="submit"]').click();

    const toast = page.locator('.toast.error');
    await expect(toast).toBeVisible();
    await expect(toast).toContainText('Дата окончания не может быть раньше даты старта');
  });

  test('TC-1.5 & TC-2.4: Success flows for custom exercise and challenge creation', async ({ page }) => {
    const uniqId = Date.now();
    const challengeName = `E2E Challenge ${uniqId}`;
    const customExName = `E2E Exercise ${uniqId}`;

    await page.locator('#create-challenge-btn').click();
    await page.locator('#name').fill(challengeName);
    
    // Create custom exercise
    await page.locator('#exercise_id').selectOption('custom');
    await page.locator('#custom_exercise_name').fill(customExName);
    await page.locator('#target_value').fill('150');

    const today = new Date();
    const end = new Date();
    end.setDate(today.getDate() + 10);
    const formatDate = (d) => d.toISOString().split('T')[0];

    await page.locator('#start_date').fill(formatDate(today));
    await page.locator('#end_date').fill(formatDate(end));

    // Submit
    await page.locator('button[type="submit"]').click();

    // Verify success toast for exercise and challenge
    const toast = page.locator('.toast.success');
    await expect(toast.first()).toBeVisible();
    
    // Redirect to dashboard
    await expect(page.locator('.dashboard-header-row h2')).toHaveText('Активные челленджи');
    
    // Verify the challenge card on the dashboard
    const card = page.locator('.challenge-card', { hasText: challengeName });
    await expect(card).toBeVisible();
    await expect(card.locator('.progress-info')).toContainText('0 из 150');
  });
});
