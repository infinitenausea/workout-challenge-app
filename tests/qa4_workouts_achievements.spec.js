const { test, expect } = require('@playwright/test');

test.describe('QA-4: Workouts & Achievements (US-3 & US-6)', () => {
  let testUser;
  let challengeId;
  let exerciseId;
  const challengeName = 'E2E Workout Challenge';

  test.beforeEach(async ({ request, page }) => {
    testUser = 'user_' + Date.now();

    // 1. Create exercise for testUser
    const exRes = await request.post('http://localhost:8080/api/exercises', {
      headers: { 'X-User-Id': testUser },
      data: { name: 'E2E Pushups' }
    });
    expect(exRes.status()).toBe(201);
    const exBody = await exRes.json();
    exerciseId = exBody.id;

    // 2. Create a fresh challenge for testUser
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
      } catch (e) {
        console.error(e);
      }
    }, testUser);
  });

  test('TC-3.20 & TC-3.21: Logging workout with instant progress update and Equator achievement', async ({ page }) => {
    // Click on our newly created challenge card
    const card = page.locator('.challenge-card', { hasText: challengeName }).first();
    await card.click();
    await page.waitForSelector('.challenge-detail');

    // Verify progress starts at 0%
    await expect(page.locator('.progress-info')).toContainText('0 из 100');

    // Click "Добавить тренировку"
    await page.locator('#add-workout-btn').click();
    await expect(page.locator('.modal-overlay')).toBeVisible();

    // Fill workout form
    await page.locator('#workout-form input[name="value"]').fill('30');
    await page.locator('#workout-submit-btn').click();

    // Verify modal is closed and toast message appeared
    await expect(page.locator('.modal-overlay')).not.toBeVisible();
    await expect(page.locator('.toast.success')).toContainText('Тренировка добавлена! +30 повторений');

    // Verify First Step achievement popup appears and close all active popups
    const popup = page.locator('.achievement-popup-overlay');
    await expect(popup).toBeVisible();
    while (await popup.isVisible()) {
      await popup.locator('#achievement-ok-btn').click();
      await page.waitForTimeout(100);
    }
    await expect(popup).not.toBeVisible();

    // Verify progress bar instantly updated to 30% without reload
    await expect(page.locator('.progress-info')).toContainText('30 из 100');
    await expect(page.locator('.progress-percent')).toHaveText('30%');

    // Verify workout list contains the item
    await expect(page.locator('.workout-list-item')).toContainText('+30 повторений');

    // Log another workout to reach 50% (equator milestone)
    await page.locator('#add-workout-btn').click();
    await page.locator('#workout-form input[name="value"]').fill('20');
    await page.locator('#workout-submit-btn').click();

    // Verify Equator achievement popup appears and close all active popups
    await expect(popup).toBeVisible();
    while (await popup.isVisible()) {
      await popup.locator('#achievement-ok-btn').click();
      await page.waitForTimeout(100);
    }
    await expect(popup).not.toBeVisible();

    // Verify progress is 50%
    await expect(page.locator('.progress-info')).toContainText('50 из 100');
  });

  test('TC-3.22 & TC-3.23: Workout form validation & TC-3.28 close modal behavior', async ({ page }) => {
    const card = page.locator('.challenge-card', { hasText: challengeName }).first();
    await card.click();
    
    // Open modal
    await page.locator('#add-workout-btn').click();
    await expect(page.locator('.modal-overlay')).toBeVisible();

    // Test Esc close
    await page.keyboard.press('Escape');
    await expect(page.locator('.modal-overlay')).not.toBeVisible();

    // Open modal again
    await page.locator('#add-workout-btn').click();
    await expect(page.locator('.modal-overlay')).toBeVisible();

    // Test backdrop overlay click close
    await page.locator('.modal-overlay').click({ position: { x: 5, y: 5 } });
    await expect(page.locator('.modal-overlay')).not.toBeVisible();

    // Open modal again to test validation
    await page.locator('#add-workout-btn').click();
    
    // HTML5 validation or client-side empty validation check
    await page.locator('#workout-form input[name="value"]').fill('');
    await page.locator('#workout-submit-btn').click();
    
    // The modal stays open due to validation
    await expect(page.locator('.modal-overlay')).toBeVisible();
  });

  test('TC-3.30: Multiple achievements triggered in a single workout', async ({ request, page }) => {
    // Create a new small challenge (target 10) to trigger first_step, equator, and hero all at once
    const res = await request.post('http://localhost:8080/api/challenges', {
      headers: { 'X-User-Id': testUser },
      data: {
        name: 'Quick Challenge ' + Date.now(),
        exercise_id: exerciseId,
        target_value: 10,
        start_date: new Date().toISOString(),
        end_date: new Date(Date.now() + 10 * 24 * 60 * 60 * 1000).toISOString()
      }
    });
    const body = await res.json();
    const quickChallengeId = body.id;

    // Refresh UI store to load this new challenge
    await page.evaluate(async () => {
      const { api } = await import('/js/api.js');
      const { store } = await import('/js/store.js');
      const challenges = await api.getChallenges();
      store.setChallenges(challenges);
    });

    const card = page.locator('.challenge-card', { hasText: 'Quick Challenge' }).first();
    await card.click();

    // Open modal and add 10 reps (100% of target)
    await page.locator('#add-workout-btn').click();
    await page.locator('#workout-form input[name="value"]').fill('10');
    await page.locator('#workout-submit-btn').click();

    // We should see a sequence of popups: close all of them
    const popup1 = page.locator('.achievement-popup-overlay');
    await expect(popup1).toBeVisible();
    while (await popup1.isVisible()) {
      await popup1.locator('#achievement-ok-btn').click();
      await page.waitForTimeout(100);
    }
    // All popups closed
    await expect(popup1).not.toBeVisible();
  });

  test('TC-14.4 & TC-14.5 & TC-14.6: Verify achievement grid on details page, information popup on click, and absence on dashboard', async ({ page }) => {
    // TC-14.6: Verify achievements card is NOT on the dashboard
    await expect(page.locator('text=Мои Достижения')).not.toBeVisible();

    // Click on challenge card
    const card = page.locator('.challenge-card', { hasText: challengeName }).first();
    await card.click();
    await page.waitForSelector('.challenge-detail');

    // TC-14.4: Verify achievements grid is visible in details page
    const grid = page.locator('.achievements-grid');
    await expect(grid).toBeVisible();
    
    // Check that we have 8 achievement items in the grid
    const items = grid.locator('.achievement-grid-item');
    await expect(items).toHaveCount(8);

    // TC-14.5: Click on the first achievement and verify information popup appears
    const firstAch = items.first();
    await firstAch.click();
    
    const popup = page.locator('.achievement-popup-overlay');
    await expect(popup).toBeVisible();
    await expect(popup.locator('.achievement-popup-card h2')).toHaveText('Первый шаг');
    
    // Verify it closes on clicking Close button
    await popup.locator('#achievement-close-btn').click();
    await expect(popup).not.toBeVisible();
  });
});
