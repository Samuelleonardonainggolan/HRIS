import { test, expect } from "@playwright/test";

test.describe("Modul Registrasi Wajah & Absensi", () => {
  // Hook untuk login sebelum setiap tes
  test.beforeEach(async ({ page }) => {
    await page.goto("/login");
    // Login menggunakan kredensial asli
    await page.getByLabel("Email").fill("manager.hr@company.com");
    await page.getByLabel("Password").fill("password123");
    await page.getByRole("button", { name: "Masuk" }).click();
    await expect(page).toHaveURL(/dashboard/);
  });

  test("Melihat halaman Persetujuan Registrasi Wajah", async ({ page }) => {
    // Navigasi ke halaman Persetujuan Registrasi Wajah
    await page.goto("/dashboard/manager-hr/persetujuan-registrasi-wajah");
    
    // Verifikasi berada di halaman yang tepat
    await expect(page).toHaveURL(/\/dashboard\/manager-hr\/persetujuan-registrasi-wajah/);
    
    // Verifikasi judul halaman muncul
    await expect(page.getByRole("heading", { name: "Persetujuan Registrasi Wajah" })).toBeVisible();
    
    // Pastikan deskripsi halaman juga muncul
    await expect(page.getByText("Kelola permintaan registrasi wajah karyawan")).toBeVisible();
  });

  test("Melihat halaman Data Kehadiran (Presensi)", async ({ page }) => {
    // Navigasi ke halaman Presensi
    await page.goto("/dashboard/manager-hr/presensi");
    
    // Verifikasi berada di halaman yang tepat
    await expect(page).toHaveURL(/\/dashboard\/manager-hr\/presensi/);
    
    // Verifikasi judul halaman muncul
    await expect(page.getByRole("heading", { name: "Presensi Karyawan" })).toBeVisible();
    
    // Verifikasi elemen ringkasan (Total Kehadiran) muncul di layar
    await expect(page.getByText("Total Kehadiran", { exact: false }).first()).toBeVisible();
  });
});
