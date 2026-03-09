import { expect, type Page } from '@playwright/test';

export function uniqueSuffix(): string {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

export async function registerUser(page: Page, username: string, password: string): Promise<void> {
  await page.goto('/login');
  await page.getByRole('tab', { name: 'Register' }).click();
  await page.getByTestId('auth-username').fill(username);
  await page.getByTestId('auth-password').fill(password);
  await page.getByTestId('auth-submit').click();
  await page.waitForURL('**/instances');
}

export async function loginUser(page: Page, username: string, password: string): Promise<void> {
  await page.goto('/login');
  await page.getByRole('tab', { name: 'Login' }).click();
  await page.getByTestId('auth-username').fill(username);
  await page.getByTestId('auth-password').fill(password);
  await page.getByTestId('auth-submit').click();
}

export async function changeBootstrapPassword(page: Page, newPassword: string): Promise<void> {
  await expect(page.getByRole('dialog', { name: 'Change the default password' })).toBeVisible();
  await page.getByTestId('change-password-new').fill(newPassword);
  await page.getByTestId('change-password-confirm').fill(newPassword);
  await page.getByTestId('change-password-submit').click();
  await expect(page.getByRole('dialog', { name: 'Change the default password' })).toBeHidden();
}

export async function selectAntdOption(page: Page, testId: string, optionLabel: string): Promise<void> {
  await page.getByTestId(testId).click();
  await page.locator('.ant-select-item-option').filter({ hasText: optionLabel }).first().click();
}
