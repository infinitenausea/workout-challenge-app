const { test, expect } = require('@playwright/test');

test.describe('QA-8: Telegram SDK Integration (US-10)', () => {
  test.beforeEach(async ({ page }) => {
    // Block the real Telegram SDK from loading so it doesn't overwrite our mock window.Telegram
    await page.route('https://telegram.org/js/telegram-web-app.js', async (route) => {
      await route.fulfill({
        contentType: 'application/javascript',
        body: 'console.log("Blocked real Telegram SDK. Using mock.");'
      });
    });

    // Output console logs and page errors for debugging
    page.on('pageerror', err => {
      console.log('PAGE ERROR:', err.message);
    });
    page.on('console', msg => {
      console.log('BROWSER LOG:', msg.text());
    });
  });

  test('TC-10.1: Fallback to X-User-Id header when running outside Telegram', async ({ page }) => {
    let hasXUserIdHeader = false;
    let authHeaderValue = null;
    
    await page.route('**/api/**', async (route) => {
      const request = route.request();
      const headers = request.headers();
      if (headers['x-user-id'] === 'default_user_1') {
        hasXUserIdHeader = true;
      }
      authHeaderValue = headers['authorization'];
      await route.continue();
    });

    // Navigate to homepage
    await page.goto('http://localhost:8080');

    // Wait for exercises or challenges load (dashboard render)
    await expect(page.locator('.dashboard-header-row h2')).toHaveText('Активные челленджи');

    // Assertions
    expect(hasXUserIdHeader).toBe(true);
    expect(authHeaderValue).toBeUndefined();
  });

  test('TC-10.1: Pass Authorization header when Telegram initData is mock-injected', async ({ page }) => {
    // Mock window.Telegram.WebApp on page init
    await page.addInitScript(() => {
      const mockWebApp = {
        ready: () => {},
        expand: () => {},
        initData: 'query_id=AAHdAM0qAAAAAN0AzSp4XXXX&user=%7B%22id%22%3A12345%2C%22first_name%22%3A%22Ivan%22%7D',
        initDataUnsafe: {
          user: {
            id: 12345,
            first_name: 'Ivan',
            username: 'ivan_tg'
          }
        },
        themeParams: {},
        onEvent: () => {}
      };

      window.Telegram = {
        WebApp: mockWebApp
      };
    });

    let hasXUserIdHeader = false;
    let authHeaderValue = null;

    await page.route('**/api/**', async (route) => {
      const request = route.request();
      const headers = request.headers();
      if (headers['x-user-id']) {
        hasXUserIdHeader = true;
      }
      if (headers['authorization']) {
        authHeaderValue = headers['authorization'];
      }
      await route.continue();
    });

    // Start waiting for the response before navigating
    const responsePromise = page.waitForResponse('**/api/exercises');

    await page.goto('http://localhost:8080');

    // Wait for the response to resolve
    await responsePromise;

    // Wait for the badge to be updated
    const badge = page.locator('#user-badge');
    await expect(badge).toHaveText('ivan_tg');

    // Since initData is provided, the API client should send Authorization, and NOT X-User-Id.
    expect(authHeaderValue).toBe('Bearer query_id=AAHdAM0qAAAAAN0AzSp4XXXX&user=%7B%22id%22%3A12345%2C%22first_name%22%3A%22Ivan%22%7D');
    expect(hasXUserIdHeader).toBe(false);
  });

  test('TC-10.2: Automatically switch theme when themeChanged event is fired in Telegram', async ({ page }) => {
    // Inject Telegram WebApp mock with event listener support
    await page.addInitScript(() => {
      const callbacks = {};
      const mockWebApp = {
        ready: () => {},
        expand: () => {},
        initData: '',
        initDataUnsafe: {},
        colorScheme: 'dark',
        themeParams: {},
        onEvent: (eventType, callback) => {
          if (!callbacks[eventType]) callbacks[eventType] = [];
          callbacks[eventType].push(callback);
        },
        triggerEvent: (eventType) => {
          if (callbacks[eventType]) {
            callbacks[eventType].forEach(cb => cb());
          }
        }
      };

      window.Telegram = {
        WebApp: mockWebApp
      };
    });

    await page.goto('http://localhost:8080');

    // Wait for page load
    await expect(page.locator('.dashboard-header-row h2')).toHaveText('Активные челленджи');

    // Check that body does NOT have light-theme class initially (dark mode by default)
    const body = page.locator('body');
    await expect(body).not.toHaveClass(/light-theme/);

    // Simulate switching theme to light inside mock
    await page.evaluate(() => {
      window.Telegram.WebApp.colorScheme = 'light';
      window.Telegram.WebApp.triggerEvent('themeChanged');
    });

    // Check that body now HAS light-theme class
    await expect(body).toHaveClass(/light-theme/);

    // Simulate switching theme back to dark inside mock
    await page.evaluate(() => {
      window.Telegram.WebApp.colorScheme = 'dark';
      window.Telegram.WebApp.triggerEvent('themeChanged');
    });

    // Check that body does NOT have light-theme class anymore
    await expect(body).not.toHaveClass(/light-theme/);
  });
});
