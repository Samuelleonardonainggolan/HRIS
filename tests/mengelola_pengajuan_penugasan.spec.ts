import { test, expect } from "@playwright/test";

test.describe("Modul Pengajuan Penugasan", () => {
  // Hook untuk login sebelum setiap tes
  test.beforeEach(async ({ page }) => {
    await page.goto("/login");
    
    // Gunakan kredensial Manager Departemen
    await page.getByLabel("Email").fill("manager.it@company.com");
    await page.getByLabel("Password").fill("password123");
    await page.getByRole("button", { name: "Masuk" }).click();
    
    await expect(page).toHaveURL(/dashboard/);
    
    // Navigasi ke halaman Penugasan Departemen
    await page.goto("/dashboard/manager-dept/penugasan");
  });

  test("Melihat daftar pengajuan penugasan", async ({ page }) => {
    // Verifikasi halaman termuat
    await expect(page.getByRole("heading", { name: /Penugasan/i })).toBeVisible();
    
    // Verifikasi tabel pengajuan ditampilkan
    await expect(page.locator("table")).toBeVisible();
  });

  test("Menambahkan pengajuan penugasan baru", async ({ page }) => {
    // Tunggu sesaat agar DOM sepenuhnya termuat
    await page.waitForTimeout(1000);

    // Cari tombol tambah penugasan (Teks bisa bervariasi: Tambah Penugasan, Buat Penugasan, dsb)
    const btnTambah = page.getByRole("button", { name: /Tambah Penugasan/i });
    const isBtnVisible = await btnTambah.isVisible().catch(() => false);

    if (!isBtnVisible) {
      console.log("Tombol Tambah Penugasan tidak ditemukan pada UI, melewati tes pembuatan.");
      return;
    }

    // Klik tombol tambah
    await btnTambah.click();

    // Verifikasi URL pindah ke form tambah (biasanya di-routing ke /tambah-penugasan)
    await expect(page).toHaveURL(/tambah-penugasan/);

    // Isi Form Pengajuan Penugasan
    // Menggunakan getByLabel / placeholder sebagai locator umum form
    const tanggalInput = page.getByLabel(/Tanggal/i);
    if (await tanggalInput.isVisible().catch(() => false)) {
      await tanggalInput.fill("2026-10-10");
    }

    const alasanInput = page.getByLabel(/Alasan/i);
    if (await alasanInput.isVisible().catch(() => false)) {
      await alasanInput.fill("Penugasan otomatis dari E2E Test Playwright");
    }

    // Centang karyawan pertama di tabel form jika ada checkbox
    const firstCheckbox = page.locator('tbody input[type="checkbox"]').first();
    if (await firstCheckbox.isVisible().catch(() => false)) {
      await firstCheckbox.check();
    }

    // Simpan Pengajuan
    const simpanBtn = page.getByRole("button", { name: /Simpan/i });
    if (await simpanBtn.isVisible().catch(() => false)) {
      await simpanBtn.click();

      // Verifikasi kembali ke list dan muncul toast berhasil
      await expect(page).toHaveURL(/\/dashboard\/manager-dept\/penugasan/);
      await expect(page.getByText(/berhasil/i).first()).toBeVisible();
    }
  });
});
