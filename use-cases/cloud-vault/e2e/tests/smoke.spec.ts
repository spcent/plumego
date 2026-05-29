import { test, expect } from '@playwright/test';

test.describe('Cloud Vault Smoke Tests', () => {
  test('should load homepage', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveTitle(/Cloud Vault/);
  });

  test('should navigate to Documents page', async ({ page }) => {
    await page.goto('/');
    await page.click('text=Documents');
    await expect(page.locator('h1')).toContainText('Documents');
  });

  test('should navigate to Search page', async ({ page }) => {
    await page.goto('/');
    await page.click('text=Search');
    await expect(page.locator('h1')).toContainText('Search');
  });

  test('should navigate to Collections page', async ({ page }) => {
    await page.goto('/');
    await page.click('text=Collections');
    await expect(page.locator('h1')).toContainText('Collections');
  });

  test('should navigate to Tags page', async ({ page }) => {
    await page.goto('/');
    await page.click('text=Tags');
    await expect(page.locator('h1')).toContainText('Tags');
  });

  test('should navigate to Import page', async ({ page }) => {
    await page.goto('/');
    await page.click('text=Import');
    await expect(page.locator('h1')).toContainText('Import');
  });

  test('should navigate to System Health page', async ({ page }) => {
    await page.goto('/');
    await page.click('text=System');
    await expect(page.locator('h1')).toContainText('System Health');
  });

  test('should navigate to Organize page', async ({ page }) => {
    await page.goto('/');
    await page.click('text=Organize');
    await expect(page.locator('h1')).toContainText('Organize');
  });

  test('should navigate to AI Tasks page', async ({ page }) => {
    await page.goto('/');
    await page.click('text=AI Tasks');
    await expect(page.locator('h1')).toContainText('AI Tasks');
  });

  test('should navigate to Prompts page', async ({ page }) => {
    await page.goto('/');
    await page.click('text=Prompts');
    await expect(page.locator('h1')).toContainText('Prompts');
  });

  test('health API should return ok', async ({ request }) => {
    const response = await request.get('/api/v1/system/health');
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data.status).toBe('ok');
    expect(data.database).toBe('ok');
    expect(data.storage).toBe('ok');
  });

  test('stats API should return counts', async ({ request }) => {
    const response = await request.get('/api/v1/system/stats');
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(typeof data.document_count).toBe('number');
    expect(typeof data.collection_count).toBe('number');
    expect(typeof data.tag_count).toBe('number');
  });

  test('doctor API should run checks', async ({ request }) => {
    const response = await request.post('/api/v1/system/doctor', {
      data: {
        checks: ['storage_objects', 'document_versions'],
      },
    });
    expect(response.ok()).toBeTruthy();
    const data = await response.json();
    expect(data.overall_status).toBeDefined();
    expect(Array.isArray(data.checks)).toBeTruthy();
  });
});
