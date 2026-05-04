import { test, expect } from "@playwright/test";

test.describe("Login Page - Blackbox Testing", () => {

  test("Login sukses → redirect ke /dashboard", async ({ page }) => {
    await page.goto("/login");

    await page.getByLabel("Email").fill("manager.hr@company.com");
    await page.getByLabel("Password").fill("password123");

    await page.getByRole("button", { name: "Masuk" }).click();

    await expect(page).toHaveURL(/dashboard/);
  });

test("Login gagal → tampil error", async ({ page }) => {
  await page.goto("/login");

  await page.getByLabel("Email").fill("manager@company.com");
  await page.getByLabel("Password").fill("salah");

  await page.getByRole("button", { name: "Masuk" }).click();

  // Tunggu error muncul (tanpa peduli teks)
  const errorBox = page.locator(".bg-red-50");

  await expect(errorBox).toBeVisible();

  // pastikan tetap di halaman login
  await expect(page).toHaveURL(/login/);
});


  test("Validasi: field kosong", async ({ page }) => {
    await page.goto("/login");

    await page.getByRole("button", { name: "Masuk" }).click();

    const email = page.locator("#email");

    const isInvalid = await email.evaluate((el: HTMLInputElement) => !el.validity.valid);
    expect(isInvalid).toBeTruthy();
  });

  test("Quick Login (Manager IT) → sukses", async ({ page }) => {
    await page.goto("/login");

    await page.getByRole("button", { name: "Manager IT" }).click();

    await expect(page).toHaveURL(/dashboard/);
  });

  test("Akses /dashboard tanpa login → redirect ke /login", async ({ page }) => {
    await page.goto("/dashboard");

    await expect(page).toHaveURL(/login/);
  });

});
