const { test, expect } = require('@playwright/test');

test.describe('QA-11/12: Telegram Haptic Feedback & Native Navigation (US-11 & US-12)', () => {
  let testUser;
  let challengeId;
  let exerciseId;
  const challengeName = 'E2E Haptics Challenge';

  test.beforeEach(async ({ request, page }) => {
    // In development mode, backend ignores Authorization header and defaults requests lacking X-User-Id to 'default_user_1'.
    // Therefore, we must seed resources for 'default_user_1' so that the Telegram-mocked frontend can fetch them.
    testUser = 'default_user_1';

    // Hook up logs and errors
    page.on('pageerror', err => {
      console.log('PAGE ERROR:', err.message);
    });
    page.on('console', msg => {
      console.log('BROWSER LOG:', msg.text());
    });

    // Block loading of the real Telegram SDK to avoid interference
    await page.route('https://telegram.org/js/telegram-web-app.js', async (route) => {
      await route.fulfill({
        contentType: 'application/javascript',
        body: 'console.log("Blocked real Telegram SDK. Using mock.");'
      });
    });

    // Mock window.Telegram.WebApp to spy on calls
    await page.addInitScript(() => {
      window.telegramCalls = [];
      const mockWebApp = {
        ready: () => { window.telegramCalls.push({ type: 'ready' }); },
        expand: () => { window.telegramCalls.push({ type: 'expand' }); },
        initData: 'query_id=AAHdAM0qAAAAAN0AzSp4XXXX&user=%7B%22id%22%3A12345%2C%22first_name%22%3A%22Ivan%22%7D',
        initDataUnsafe: {
          user: { id: 12345, first_name: 'Ivan', username: 'ivan_tg' }
        },
        themeParams: {},
        onEvent: () => {},
        HapticFeedback: {
          impactOccurred: (style) => {
            window.telegramCalls.push({ type: 'haptic_impact', style });
          },
          notificationOccurred: (notificationType) => {
            window.telegramCalls.push({ type: 'haptic_notification', notificationType });
          },
          selectionChanged: () => {
            window.telegramCalls.push({ type: 'haptic_selection' });
          }
        },
        BackButton: {
          show: () => { window.telegramCalls.push({ type: 'backbutton_show' }); },
          hide: () => { window.telegramCalls.push({ type: 'backbutton_hide' }); },
          onClick: (cb) => {
            window.telegramCalls.push({ type: 'backbutton_onclick' });
            window.backButtonCallback = cb;
          },
          offClick: (cb) => {
            window.telegramCalls.push({ type: 'backbutton_offclick' });
            window.backButtonCallback = null;
          }
        },
        enableClosingConfirmation: () => {
          window.telegramCalls.push({ type: 'enable_closing' });
        },
        disableClosingConfirmation: () => {
          window.telegramCalls.push({ type: 'disable_closing' });
        }
      };

      window.Telegram = { WebApp: mockWebApp };
    });

    // 1. Create exercise for testUser (making the name unique to avoid 409 Conflict across parallel tests)
    const exRes = await request.post('http://localhost:8080/api/exercises', {
      headers: { 'X-User-Id': testUser },
      data: { name: 'E2E Pushups-' + Date.now() + '-' + Math.floor(Math.random() * 1000) }
    });
    expect(exRes.status()).toBe(201);
    const exBody = await exRes.json();
    exerciseId = exBody.id;

    // 2. Create a fresh challenge for testUser
    const res = await request.post('http://localhost:8080/api/challenges', {
      headers: { 'X-User-Id': testUser },
      data: {
        name: challengeName + '-' + Date.now() + '-' + Math.floor(Math.random() * 1000),
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

  test('TC-11.1 & TC-11.2: Haptic Feedback when logging workout or experiencing validation errors', async ({ page }) => {
    // Navigate to challenge detail
    await page.click(`.challenge-card[data-id="${challengeId}"]`);
    
    // Click Add workout
    await page.click('#add-workout-btn');

    // Disable HTML5 validation to allow custom JavaScript validation check to run
    await page.evaluate(() => {
      const form = document.querySelector('#workout-form');
      if (form) form.setAttribute('novalidate', 'novalidate');
    });

    // Try to input invalid repetition value to test validation failure haptics (TC-11.2)
    await page.fill('#workout_value', '-10');
    await page.click('#workout-submit-btn');

    // Assert that haptic_notification was called with 'error'
    let calls = await page.evaluate(() => window.telegramCalls);
    let errorCalls = calls.filter(c => c.type === 'haptic_notification' && c.notificationType === 'error');
    expect(errorCalls.length).toBeGreaterThanOrEqual(1);

    // Clear calls log for cleaner assertions
    await page.evaluate(() => { window.telegramCalls = []; });

    // Enter valid data (TC-11.1)
    await page.fill('#workout_value', '30');
    await page.click('#workout-submit-btn');

    // Wait for modal to close
    await expect(page.locator('#workout-form')).toBeHidden();

    // Assert that haptic_notification was called with 'success'
    calls = await page.evaluate(() => window.telegramCalls);
    let successCalls = calls.filter(c => c.type === 'haptic_notification' && c.notificationType === 'success');
    expect(successCalls.length).toBeGreaterThanOrEqual(1);

    // Close achievement popup if visible to clean up
    const popup = page.locator('.achievement-popup-overlay');
    if (await popup.isVisible()) {
      await popup.locator('#achievement-ok-btn').click();
    }
  });

  test('TC-11.3: Haptic Feedback triggerImpact on deleting workout', async ({ page }) => {
    // Navigate to challenge detail, log a workout via UI first
    await page.click(`.challenge-card[data-id="${challengeId}"]`);
    await page.click('#add-workout-btn');
    await page.fill('#workout_value', '20');
    await page.click('#workout-submit-btn');
    await expect(page.locator('#workout-form')).toBeHidden();

    // Close achievement popup (First Step) so it doesn't block UI clicks
    const popup = page.locator('.achievement-popup-overlay');
    if (await popup.isVisible()) {
      await popup.locator('#achievement-ok-btn').click();
    }

    // Clear calls log
    await page.evaluate(() => { window.telegramCalls = []; });

    // Listen for confirm dialog and accept it
    page.once('dialog', async dialog => {
      expect(dialog.message()).toContain('Вы уверены, что хотите удалить эту тренировку?');
      await dialog.accept();
    });

    // Click delete workout button
    await page.locator('.delete-workout-btn').click();

    // Assert that haptic_impact was called with 'medium'
    const calls = await page.evaluate(() => window.telegramCalls);
    const impactCalls = calls.filter(c => c.type === 'haptic_impact' && c.style === 'medium');
    expect(impactCalls.length).toBeGreaterThanOrEqual(1);
  });

  test('TC-11.4: Haptic Feedback triggerNotification on deleting challenge', async ({ page }) => {
    await page.click(`.challenge-card[data-id="${challengeId}"]`);

    // Clear calls log
    await page.evaluate(() => { window.telegramCalls = []; });

    // Accept confirm dialog
    page.once('dialog', async dialog => {
      await dialog.accept();
    });

    // Click delete challenge
    await page.click('#delete-challenge-btn');

    // Wait for navigation back to dashboard
    await expect(page.locator('.dashboard-header-row h2')).toHaveText('Активные челленджи');

    // Assert haptic success notification was called
    const calls = await page.evaluate(() => window.telegramCalls);
    const successCalls = calls.filter(c => c.type === 'haptic_notification' && c.notificationType === 'success');
    expect(successCalls.length).toBeGreaterThanOrEqual(1);
  });

  test('TC-12.1 & TC-12.2: Native BackButton visibility and back navigation', async ({ page }) => {
    // On dashboard initially: BackButton should be hidden
    let calls = await page.evaluate(() => window.telegramCalls);
    let hideCalls = calls.filter(c => c.type === 'backbutton_hide');
    expect(hideCalls.length).toBeGreaterThanOrEqual(1);

    // Navigate to challenge detail
    await page.click(`.challenge-card[data-id="${challengeId}"]`);

    // BackButton should be shown
    calls = await page.evaluate(() => window.telegramCalls);
    let showCalls = calls.filter(c => c.type === 'backbutton_show');
    expect(showCalls.length).toBeGreaterThanOrEqual(1);

    // Trigger click on back button via evaluated mock callback
    await page.evaluate(() => {
      if (window.backButtonCallback) window.backButtonCallback();
    });

    // Should go back to dashboard
    await expect(page.locator('.dashboard-header-row h2')).toHaveText('Активные челленджи');

    // BackButton should be hidden again
    calls = await page.evaluate(() => window.telegramCalls);
    hideCalls = calls.filter(c => c.type === 'backbutton_hide');
    expect(hideCalls.length).toBeGreaterThanOrEqual(2); // Initial hide + hide on return
  });

  test('TC-12.3 & TC-12.4: ClosingConfirmation lifecycle on dirty form and exit', async ({ page }) => {
    // Navigate to challenge creation form
    await page.click('#create-challenge-btn');

    // Clean form: closing confirmation not enabled yet
    let calls = await page.evaluate(() => window.telegramCalls);
    let enableCalls = calls.filter(c => c.type === 'enable_closing');
    expect(enableCalls.length).toBe(0);

    // Type something to make it dirty (TC-12.3)
    await page.fill('#name', 'Typing a challenge name...');

    // Should trigger enableClosingConfirmation
    calls = await page.evaluate(() => window.telegramCalls);
    enableCalls = calls.filter(c => c.type === 'enable_closing');
    expect(enableCalls.length).toBeGreaterThanOrEqual(1);

    // Clear calls log
    await page.evaluate(() => { window.telegramCalls = []; });

    // Click Cancel (TC-12.4)
    await page.click('#cancel-btn');

    // Should trigger disableClosingConfirmation and return to dashboard
    await expect(page.locator('.dashboard-header-row h2')).toHaveText('Активные челленджи');
    calls = await page.evaluate(() => window.telegramCalls);
    let disableCalls = calls.filter(c => c.type === 'disable_closing');
    expect(disableCalls.length).toBeGreaterThanOrEqual(1);
  });
});

test.describe('QA-11/12 Edge: Local Dev Fallback (Running outside Telegram)', () => {
  let testUser;
  let challengeId;
  let exerciseId;

  test.beforeEach(async ({ request, page }) => {
    testUser = 'user_edge_' + Date.now();

    // Create exercise
    const exRes = await request.post('http://localhost:8080/api/exercises', {
      headers: { 'X-User-Id': testUser },
      data: { name: 'Edge Pushups-' + Date.now() }
    });
    expect(exRes.status()).toBe(201);
    const exBody = await exRes.json();
    exerciseId = exBody.id;

    // Create challenge
    const res = await request.post('http://localhost:8080/api/challenges', {
      headers: { 'X-User-Id': testUser },
      data: {
        name: 'Edge Challenge-' + Date.now(),
        exercise_id: exerciseId,
        target_value: 100,
        start_date: new Date().toISOString(),
        end_date: new Date(Date.now() + 10 * 24 * 60 * 60 * 1000).toISOString()
      }
    });
    expect(res.status()).toBe(201);
    const body = await res.json();
    challengeId = body.id;

    // Go to homepage with NO window.Telegram injected
    await page.goto('http://localhost:8080');

    // Set user context
    await page.evaluate(async (newUserId) => {
      const { api } = await import('/js/api.js');
      const { store } = await import('/js/store.js');
      api.setUserId(newUserId);
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

  test('TC-11.4: App functions properly without window.Telegram WebApp and does not crash', async ({ page }) => {
    // Check dashboard loads
    await expect(page.locator('.dashboard-header-row h2')).toHaveText('Активные челленджи');

    // Go to details
    await page.click(`.challenge-card[data-id="${challengeId}"]`);
    await expect(page.locator('#add-workout-btn')).toBeVisible();

    // Log workout
    await page.click('#add-workout-btn');
    await page.fill('#workout_value', '20');
    await page.click('#workout-submit-btn');
    await expect(page.locator('#workout-form')).toBeHidden();

    // Workout is logged and progress is updated
    await expect(page.locator('.progress-info')).toContainText('20 из 100');
  });
});
