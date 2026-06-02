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
  const errorBox = page.locator(".bg-red-50, .text-red-600").first();

  await expect(errorBox).toBeVisible();

  // pastikan tetap di halaman login
  await expect(page).toHaveURL(/login/);
});


  test("Validasi: field kosong", async ({ page }) => {
    await page.goto("/login");

    await page.getByRole("button", { name: "Masuk" }).click();

    // Pastikan error kustom muncul
    const emailError = page.getByText("Kolom Email tidak boleh kosong");
    await expect(emailError).toBeVisible();
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

  test("Logout → redirect ke /login", async ({ page }) => {
    // Login menggunakan Quick Login
    await page.goto("/login");
    await page.getByRole("button", { name: "Manager IT" }).click();
    await expect(page).toHaveURL(/dashboard/);

    // Buka menu profil/user (klik tombol avatar)
    await page.locator('button:has(.bg-orange-400)').click();

    // Lakukan logout
    await page.getByRole("button", { name: "Logout" }).click();

    // Pastikan diarahkan kembali ke halaman login
    await expect(page).toHaveURL(/login/);
  });

});
