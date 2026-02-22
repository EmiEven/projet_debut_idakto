import { test, expect } from '@playwright/test';

test('accès à /secret sans session renvoie vers /login', async ({ page }) => {
  await page.goto('http://localhost:8080/secret');
  await expect(page).toHaveURL(/accounts\.google\.com/);
});
