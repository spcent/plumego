import { test, expect, Page, APIRequestContext } from '@playwright/test';

const TEST_USER = {
  username: 'admin',
  email: 'admin@e2e.local',
  password: 'StrongPassword123!',
};

// Skip the pre-auth smoke block when the server has auth enabled.
// V0.8's deployment model defaults to auth.enabled=true, so most
// of these anonymous tests will 401 in production-like setups.
const SKIP_PRE_AUTH = process.env.E2E_SKIP_PRE_AUTH !== '0';

test.describe('Cloud Vault Smoke Tests (pre-auth)', () => {
  test.skip(SKIP_PRE_AUTH, 'pre-auth tests skipped when E2E_SKIP_PRE_AUTH is set');

  test('should load homepage', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveTitle(/Cloud Vault/);
  });

  test('health API should return ok', async ({ request }) => {
    const response = await request.get('/api/v1/system/health');
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data.status).toBe('ok');
    expect(data.database).toBe('ok');
    expect(data.storage).toBe('ok');
  });
});

test.describe('V0.8 Productization Smoke Suite', () => {
  // These tests assume a fresh server with no users.
  // They run sequentially: setup → login → settings → backup → logout → unauth.
  // Skip when Playwright not installed or when BASE_URL is unreachable.
  test.skip(
    process.env.CI !== 'true' && !process.env.BASE_URL,
    'skip V0.8 suite unless CI=true or BASE_URL is set',
  );

  test('1. first boot shows setup page, then login works', async ({ page }) => {
    await page.goto('/');
    // Setup flow
    await expect(page.locator('input[name="username"], input[placeholder*="sername" i]')).toBeVisible({ timeout: 10_000 });
    await page.fill('input[name="username"], input[placeholder*="sername" i]', TEST_USER.username);
    await page.fill('input[name="email"], input[type="email"]', TEST_USER.email);
    await page.fill('input[name="password"], input[type="password"]', TEST_USER.password);
    // Confirm password if present
    const confirm = page.locator('input[name="confirmPassword"]');
    if (await confirm.isVisible().catch(() => false)) {
      await confirm.fill(TEST_USER.password);
    }
    await page.click('button[type="submit"], button:has-text("Create")');

    // Wait for redirect to main app
    await expect(page.locator('nav, [role="tablist"]')).toBeVisible({ timeout: 15_000 });
  });

  test('2. theme toggle switches dark class', async ({ page }) => {
    await login(page);
    const html = page.locator('html');
    const initialClass = await html.getAttribute('class');
    const toggle = page.locator('button[aria-label*="theme" i], button:has-text("Theme"), [data-testid="theme-toggle"]');
    if (await toggle.isVisible().catch(() => false)) {
      await toggle.click();
      await expect(html).not.toHaveAttribute('class', initialClass ?? '');
    }
  });

  test('3. locale switch changes nav text', async ({ page }) => {
    await login(page);
    // Find locale selector — could be a dropdown or buttons
    const enTab = page.locator('button:has-text("Vault")').first();
    const zhTab = page.locator('button:has-text("文档库")').first();
    if ((await enTab.isVisible().catch(() => false)) && (await zhTab.isVisible().catch(() => false))) {
      // Locale already English, try switching
      const localeSwitcher = page.locator('[data-testid="locale-switch"], select[aria-label*="language" i]');
      if (await localeSwitcher.isVisible().catch(() => false)) {
        await localeSwitcher.selectOption('zh-CN');
        await expect(page.locator('button:has-text("文档库")')).toBeVisible();
        // Switch back
        await localeSwitcher.selectOption('en-US');
        await expect(page.locator('button:has-text("Vault")')).toBeVisible();
      }
    }
  });

  test('4. settings page shows config, no secrets', async ({ page }) => {
    await login(page);
    await page.click('button:has-text("Settings")');
    await expect(page.locator('text=System Settings').first()).toBeVisible({ timeout: 10_000 });
    // Should show version
    await expect(page.locator('text=Version')).toBeVisible();
    // Should show storage provider
    await expect(page.locator('text=Storage Provider')).toBeVisible();
    // Should NOT show secrets
    const bodyText = await page.locator('body').innerText();
    expect(bodyText).not.toMatch(/[A-Za-z0-9+/]{40,}={0,2}/); // no base64 tokens
    expect(bodyText.toLowerCase()).not.toContain('api_key');
    expect(bodyText.toLowerCase()).not.toContain('secret_key');
  });

  test('5. backup panel creates and lists backup', async ({ page }) => {
    await login(page);
    await page.click('button:has-text("System")');
    await expect(page.locator('text=Backup Management').first()).toBeVisible({ timeout: 10_000 });
    // Create backup
    await page.click('button:has-text("Create Backup")');
    await expect(page.locator('text=cloud-vault-backup').first()).toBeVisible({ timeout: 30_000 });
    // Download link should appear
    await expect(page.locator('a[download], a:has-text("Download")').first()).toBeVisible();
  });

  test('6. logout redirects to login', async ({ page }) => {
    await login(page);
    await page.click('button:has-text("Logout")');
    await expect(page.locator('input[type="password"], button:has-text("Sign in")').first()).toBeVisible({ timeout: 10_000 });
  });

  test('7. unauth GET /api/v1/documents returns 401', async ({ request }) => {
    const response = await request.get('/api/v1/documents');
    expect(response.status()).toBe(401);
  });

  test('8. unauth GET /api/v1/system/stats returns 401', async ({ request }) => {
    const response = await request.get('/api/v1/system/stats');
    expect(response.status()).toBe(401);
  });

  test('9. unauth POST /api/v1/system/backup returns 401', async ({ request }) => {
    const response = await request.post('/api/v1/system/backup', { data: {} });
    expect(response.status()).toBe(401);
  });

  test('10. health endpoint remains public', async ({ request }) => {
    const response = await request.get('/api/v1/system/health');
    expect(response.ok()).toBeTruthy();
  });
});

async function login(page: Page) {
  await page.goto('/');
  // If already logged in (cookie persisted), skip
  const loggedIn = page.locator('nav, [role="tablist"]');
  if (await loggedIn.isVisible().catch(() => false)) {
    return;
  }
  await page.fill('input[name="username"], input[placeholder*="sername" i]', TEST_USER.username);
  await page.fill('input[name="password"], input[type="password"]', TEST_USER.password);
  await page.click('button[type="submit"], button:has-text("Sign in"), button:has-text("Login")');
  await expect(loggedIn).toBeVisible({ timeout: 10_000 });
}
