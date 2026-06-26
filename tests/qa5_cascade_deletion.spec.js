const { test, expect } = require('@playwright/test');

test.describe('QA-5: Cascade Deletion & Status Rollbacks (US-4 & US-8)', () => {
  let testUser;
  let challengeId;
  let exerciseId;
  const challengeName = 'E2E Deletion Challenge';

  test.beforeEach(async ({ request, page }) => {
    testUser = 'user_del_' + Date.now();

    // 1. Create exercise for testUser
    const exRes = await request.post('http://localhost:8080/api/exercises', {
      headers: { 'X-User-Id': testUser },
      data: { name: 'E2E Deletion Pushups' }
    });
    expect(exRes.status()).toBe(201);
    const exBody = await exRes.json();
    exerciseId = exBody.id;

    // 2. Create challenge for testUser
    const res = await request.post('http://localhost:8080/api/challenges', {
      headers: { 'X-User-Id': testUser },
      data: {
        name: challengeName + '-' + Date.now(),
        exercise_id: exerciseId,
        target_value: 100,
        start_date: new Date().toISOString(),
        end_date: new Date(Date.now() + 10 * 24 * 60 * 60 * 1000).toISOString()
      }
    });
    expect(res.status()).toBe(201);
    const body = await res.json();
    challengeId = body.id;

    await page.goto('http://localhost:8080');

    // 3. Switch the frontend API client user context to testUser
    await page.evaluate(async (newUserId) => {
      const { api } = await import('/js/api.js');
      const { store } = await import('/js/store.js');
      api.setUserId(newUserId);
      
      const userBadge = document.getElementById('user-badge');
      if (userBadge) userBadge.innerText = newUserId;

      // Refetch for the new user and set in store
      try {
        const exercises = await api.getExercises();
        store.setExercises(exercises);
        const challenges = await api.getChallenges();
        store.setChallenges(challenges);
        const achievements = await api.getAchievements();
        store.setAchievements(achievements);
      } catch (e) {
        console.error(e);
      }
    }, testUser);
  });

  test('TC-3.24: Deleting a workout updates the progress bar instantly', async ({ page }) => {
    // Navigate to challenge
    const card = page.locator('.challenge-card', { hasText: challengeName }).first();
    await card.click();

    // Add a workout
    await page.locator('#add-workout-btn').click();
    await page.locator('#workout-form input[name="value"]').fill('40');
    await page.locator('#workout-submit-btn').click();

    // Close First Step popup if it appears
    const popup = page.locator('.achievement-popup-overlay');
    if (await popup.isVisible()) {
      await popup.locator('#achievement-ok-btn').click();
    }

    await expect(page.locator('.progress-info')).toContainText('40 из 100');

    // Confirm dialog automatically when triggered
    page.once('dialog', async dialog => {
      expect(dialog.message()).toContain('Вы уверены, что хотите удалить эту тренировку?');
      await dialog.accept();
    });

    // Click delete workout button
    await page.locator('.delete-workout-btn').click();

    // Verify list item is removed and progress is back to 0
    await expect(page.locator('.workout-list-item')).not.toBeVisible();
    await expect(page.locator('.progress-info')).toContainText('0 из 100');
    await expect(page.locator('.toast.success')).toContainText('Тренировка успешно удалена');
  });

  test('TC-3.25: Status rollback completed -> active after deleting workout', async ({ page }) => {
    const card = page.locator('.challenge-card', { hasText: challengeName }).first();
    await card.click();

    // Verify initial status is active
    await expect(page.locator('.challenge-status')).toHaveText('Активен');

    // Add workout to complete challenge (100 reps)
    await page.locator('#add-workout-btn').click();
    await page.locator('#workout-form input[name="value"]').fill('100');
    await page.locator('#workout-submit-btn').click();

    // Close achievement popups (first_step, equator, hero)
    const popup = page.locator('.achievement-popup-overlay');
    while (await popup.isVisible()) {
      await popup.locator('#achievement-ok-btn').click();
      await page.waitForTimeout(100);
    }

    // Verify challenge is completed
    await expect(page.locator('.challenge-status')).toHaveText('Завершен');
    await expect(page.locator('#add-workout-btn')).toBeDisabled();

    // Delete the completed workout
    page.once('dialog', async dialog => {
      await dialog.accept();
    });
    await page.locator('.delete-workout-btn').click();

    // Verify challenge rolls back to active and add workout button is enabled again
    await expect(page.locator('.challenge-status')).toHaveText('Активен');
    await expect(page.locator('#add-workout-btn')).toBeEnabled();
    await expect(page.locator('.progress-info')).toContainText('0 из 100');
  });

  test('TC-4.1 & TC-4.2 & TC-4.3: Challenge Deletion UI Flow', async ({ page }) => {
    const card = page.locator('.challenge-card', { hasText: challengeName }).first();
    await card.click();

    // Check if delete challenge button is present
    const deleteBtn = page.locator('#delete-challenge-btn');
    await expect(deleteBtn).toBeVisible();

    // Test Cancel Delete
    page.once('dialog', async dialog => {
      expect(dialog.message()).toContain('Удалить челлендж?');
      await dialog.dismiss();
    });
    await deleteBtn.click();

    // Verify we are still on the detail page
    await expect(page.locator('.challenge-detail')).toBeVisible();

    // Test Confirm Delete
    page.once('dialog', async dialog => {
      await dialog.accept();
    });
    await deleteBtn.click();

    // Verify redirected to dashboard and challenge card is gone
    await expect(page.locator('h2').first()).toHaveText('Активные челленджи');
    const deletedCard = page.locator('.challenge-card', { hasText: challengeName });
    await expect(deletedCard).not.toBeVisible();
  });
});
