import { test, expect } from '@playwright/test';

test('la page d’accueil affiche le bouton Google', async ({ page }) => {
  await page.goto('http://localhost:8080/');

  const button = page.locator('text=Se connecter avec Google');
  await expect(button).toBeVisible();
});