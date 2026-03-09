import { expect, test } from '@playwright/test';
import { changeBootstrapPassword, loginUser, uniqueSuffix } from './helpers';

test('admin can rotate bootstrap password and publish a template', async ({ page }) => {
  const suffix = uniqueSuffix();
  const newPassword = `Admin-${suffix}!`;
  const templateID = `e2e-template-${suffix}`;
  const templateName = `E2E Template ${suffix}`;

  await loginUser(page, 'admin', 'admin');
  await page.waitForURL('**/admin');
  await changeBootstrapPassword(page, newPassword);

  await expect(page.getByRole('heading', { name: 'Admin panel' })).toBeVisible();
  await page.getByTestId('template-id').fill(templateID);
  await page.getByTestId('template-name').fill(templateName);
  await page.getByTestId('template-description').fill('Template created by the end-to-end suite.');
  await page.getByTestId('template-image').fill('registry.example.com/openclaw/e2e-template');
  await page.getByTestId('template-version').fill('2.0.0');
  await page.getByTestId('template-port').fill('9090');
  await page.getByTestId('template-submit').click();

  await expect(page.getByText('Template created')).toBeVisible();
  await page.getByRole('menuitem', { name: 'Templates' }).click();
  await expect(page).toHaveURL(/\/templates$/);
  await expect(page.getByText(templateName)).toBeVisible();
  await expect(page.getByText('registry.example.com/openclaw/e2e-template')).toBeVisible();
});
