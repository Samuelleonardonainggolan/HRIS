import { test, expect } from "@playwright/test";

test.describe("Modul Mengelola Data Departemen", () => {
  // Hook untuk login sebelum setiap tes
  test.beforeEach(async ({ page }) => {
    await page.goto("/login");
    await page.getByRole("button", { name: "Manager HR" }).click();
    await expect(page).toHaveURL(/dashboard/);

    // Navigasi ke halaman Manajemen Departemen
    await page.goto("/dashboard/manager-hr/departemen");
    await expect(page).toHaveURL(/\/dashboard\/manager-hr\/departemen/);
  });

  test("Menambah departemen baru", async ({ page }) => {
    // Verifikasi halaman list terbuka
    await expect(page.getByRole("heading", { name: "Manajemen Departemen" })).toBeVisible();

    // Klik tombol Tambah Departemen
    await page.getByRole("button", { name: "Tambah Departemen" }).click();
    await expect(page).toHaveURL(/tambah-departemen/);

    // Verifikasi halaman tambah terbuka
    await expect(page.getByRole("heading", { name: "Tambah Departemen Baru" })).toBeVisible();

    // Buat nama departemen unik agar tidak duplikat
    const randomSuffix = Math.floor(Math.random() * 10000);
    const testDeptName = `Departemen Test ${randomSuffix}`;

    // Isi Nama Departemen (kode akan ter-generate otomatis)
    await page.locator("#name").fill(testDeptName);

    // Isi Deskripsi
    await page.locator("#description").fill("Deskripsi departemen untuk keperluan pengujian otomatis.");

    // Simpan departemen
    await page.getByRole("button", { name: "Simpan Departemen" }).click();

    // Verifikasi berhasil kembali ke halaman list
    await expect(page).toHaveURL(/\/dashboard\/manager-hr\/departemen/);
    await expect(page.getByText("Departemen berhasil ditambahkan").first()).toBeVisible();
  });

  test("Mengubah departemen", async ({ page }) => {
    // Tunggu tabel dimuat
    await page.waitForTimeout(1000);

    // Klik tombol Edit (ikon pensil) pada baris departemen pertama
    const editButton = page.locator('button[title="Edit"]').first();
    await editButton.click();

    // Pastikan halaman edit terbuka (URL mengandung ?edit=)
    await expect(page).toHaveURL(/tambah-departemen\?edit=/);
    await expect(page.getByRole("heading", { name: "Edit Departemen" })).toBeVisible();

    // Ubah deskripsi departemen
    await page.locator("#description").fill("Deskripsi diperbarui melalui pengujian otomatis.");

    // Simpan perubahan
    await page.getByRole("button", { name: "Simpan Perubahan" }).click();

    // Verifikasi berhasil kembali ke halaman list
    await expect(page).toHaveURL(/\/dashboard\/manager-hr\/departemen/);
    await expect(page.getByText("Departemen berhasil diperbarui").first()).toBeVisible();
  });

  test("Menonaktifkan departemen", async ({ page }) => {
    // Tunggu tabel dimuat
    await page.waitForTimeout(1000);

    // Cari baris departemen yang berstatus "Aktif" untuk dinonaktifkan
    const deactivateButton = page.locator('button[title="Nonaktifkan"]').first();
    await deactivateButton.click();

    // ConfirmationDialog akan muncul, klik tombol konfirmasi "Ya, Nonaktifkan"
    await expect(page.getByRole("heading", { name: "Nonaktifkan Departemen" })).toBeVisible();
    await page.getByRole("button", { name: "Ya, Nonaktifkan" }).click();

    // Verifikasi pesan sukses muncul
    await expect(page.getByText("Departemen berhasil dinonaktifkan").first()).toBeVisible();
  });
});
