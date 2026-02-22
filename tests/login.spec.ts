import { test, expect } from '@playwright/test';



test('clic sur le bouton Google redirige vers /login', async ({ page }) => {
  await page.goto('http://localhost:8080/');

  await page.click('text=Se connecter avec Google');

  await expect(page).toHaveURL(/accounts\.google\.com/);
});


test('la route /login redirige vers Google', async ({ page }) => {
  await page.goto('http://localhost:8080/login');

  await expect(page).toHaveURL(/accounts\.google\.com/);
});


test('callback sans vrai code Google renvoie une erreur', async ({ page }) => {
  await page.goto('http://localhost:8080/callback?code=fake');

  const content = await page.content();
  expect(content).toContain('Impossible d\'échanger le code');
});

