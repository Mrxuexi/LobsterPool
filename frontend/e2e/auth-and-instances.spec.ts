import { expect, test } from '@playwright/test';
import { registerUser, selectAntdOption, uniqueSuffix } from './helpers';

test('member can register, create an instance, and delete it', async ({ page }) => {
  const suffix = uniqueSuffix();
  const username = `e2e-member-${suffix}`;
  const password = `Pass-${suffix}!`;
  const instanceName = `E2E Instance ${suffix}`;

  await registerUser(page, username, password);

  await expect(page.getByRole('heading', { name: 'Instances', exact: true })).toBeVisible();
  await page.getByTestId('sidebar-create-instance').click();

  await expect(page).toHaveURL(/\/instances\/create$/);
  await page.getByTestId('instance-name').fill(instanceName);
  await selectAntdOption(page, 'instance-template', 'Mattermost Bot');
  await selectAntdOption(page, 'instance-cluster', 'Mock Cluster');
  await page.getByTestId('instance-api-key').fill('sk-e2e-test-key');
  await page.getByTestId('instance-bot-token').fill('mm-e2e-test-token');
  await page.getByTestId('instance-submit').click();

  await page.waitForURL(/\/instances\/[^/]+$/);
  await expect(page.getByRole('heading', { name: instanceName })).toBeVisible();
  await expect(page.getByText('mock', { exact: true })).toBeVisible();
  await expect(page.getByRole('link', { name: /svc\.cluster\.local/i }).first()).toBeVisible();

  await page.getByTestId('instance-detail-delete').click();
  await page.getByTestId('instance-detail-delete-confirm').click();
  await page.waitForURL('**/instances');
  await expect(page.getByText(instanceName)).toHaveCount(0);
});
